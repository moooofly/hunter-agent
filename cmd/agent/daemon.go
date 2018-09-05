package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/moooofly/hunter-agent/cli/debug"
	"github.com/moooofly/hunter-agent/daemon"
	"github.com/moooofly/hunter-agent/daemon/config"
	"github.com/moooofly/hunter-agent/daemon/listeners"
	"github.com/moooofly/hunter-agent/gen-go/dumpproto"
	dopts "github.com/moooofly/hunter-agent/opts"
	"github.com/moooofly/hunter-agent/pkg/pidfile"
	customsig "github.com/moooofly/hunter-agent/pkg/signal"
	"google.golang.org/grpc"

	"github.com/census-instrumentation/opencensus-proto/gen-go/exporterproto"
	"github.com/census-instrumentation/opencensus-proto/gen-go/traceproto"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/sys/unix"
)

// RFC3339NanoFixed is time.RFC3339Nano with nanoseconds padded using zeros to
// ensure the formatted time isalways the same number of characters.
const RFC3339NanoFixed = "2006-01-02T15:04:05.000000000Z07:00"

type DaemonCli struct {
	*config.Config
	configFile *string
	flags      *pflag.FlagSet

	// TODO: add more service here

	d *daemon.Daemon
}

// NewDaemonCli returns a daemon CLI
func NewDaemonCli() *DaemonCli {
	return &DaemonCli{}
}

func (cli *DaemonCli) start(opts *daemonOptions) (err error) {
	stopc := make(chan bool)
	defer close(stopc)

	if cli.Config, err = loadDaemonCliConfig(opts); err != nil {
		return err
	}
	cli.configFile = &opts.configFile
	cli.flags = opts.flags

	if cli.Config.Debug {
		debug.Enable()
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: RFC3339NanoFixed,
		DisableColors:   cli.Config.RawLogs,
		FullTimestamp:   true,
	})

	if err := setDefaultUmask(); err != nil {
		return fmt.Errorf("Failed to set umask: %v", err)
	}

	// Create the daemon root before we create ANY other files
	if err := daemon.CreateDaemonRoot(cli.Config); err != nil {
		return err
	}

	if cli.Pidfile != "" {
		pf, err := pidfile.New(cli.Pidfile)
		if err != nil {
			return fmt.Errorf("Error starting daemon: %v", err)
		}
		defer func() {
			if err := pf.Remove(); err != nil {
				logrus.Error(err)
			}
		}()
	}

	if len(cli.Config.Hosts) == 0 {
		cli.Config.Hosts = make([]string, 1)
	}

	// setup grpc server here
	// TODO: setup more serivice here
	hosts, err := loadListeners(cli)
	if err != nil {
		return fmt.Errorf("Failed to load listeners: %v", err)
	}

	customsig.Trap(func() {
		cli.stop()
		<-stopc // wait for daemonCli.start() to return
	}, logrus.StandardLogger())

	d, err := daemon.NewDaemon(cli.Config)
	if err != nil {
		return fmt.Errorf("Error starting daemon: %v", err)
	}

	d.StoreHosts(hosts)

	logrus.Info("Daemon has completed initialization")

	cli.d = d

	// reload the configuration by USR2 signal.
	cli.setupConfigReloadTrap()

	serveAPIWait := make(chan error)
	// TODO: add flow control metrics here
	if cli.Config.MetricsAddress == "" {
		return fmt.Errorf("cli.Config.MetricsAddress must not be \"\" currently")
	}
	startMetricsServer(cli.Config.MetricsAddress, serveAPIWait)
	errAPI := <-serveAPIWait

	shutdownDaemon(d)

	if errAPI != nil {
		return fmt.Errorf("Shutting down due to ServeAPI error: %v", errAPI)
	}

	return nil
}

func (cli *DaemonCli) reloadConfig() {
	reload := func(c *config.Config) {

		if err := cli.d.Reload(c); err != nil {
			logrus.Errorf("Error reconfiguring the daemon: %v", err)
			return
		}

		if c.IsValueSet("debug") {
			debugEnabled := debug.IsEnabled()
			switch {
			case debugEnabled && !c.Debug: // disable debug
				debug.Disable()
			case c.Debug && !debugEnabled: // enable debug
				debug.Enable()
			}
		}
	}

	if err := config.Reload(*cli.configFile, cli.flags, reload); err != nil {
		logrus.Error(err)
	}
}

func (cli *DaemonCli) stop() {
	// do some cleanups
	logrus.Debug("---> do some cleanups here.")
}

// shutdownDaemon just wraps daemon.Shutdown() to handle a timeout in case
// d.Shutdown() is waiting too long to kill container or worst it's
// blocked there
func shutdownDaemon(d *daemon.Daemon) {
	shutdownTimeout := d.ShutdownTimeout()
	ch := make(chan struct{})
	go func() {
		d.Shutdown()
		close(ch)
	}()
	if shutdownTimeout < 0 {
		<-ch
		logrus.Debug("Clean shutdown succeeded")
		return
	}
	select {
	case <-ch:
		logrus.Debug("Clean shutdown succeeded")
	case <-time.After(time.Duration(shutdownTimeout) * time.Second):
		logrus.Error("Force shutdown daemon")
	}
}

func loadDaemonCliConfig(opts *daemonOptions) (*config.Config, error) {
	flags := opts.flags
	conf := opts.daemonConfig
	conf.Debug = opts.Debug
	conf.LogLevel = opts.LogLevel
	conf.Hosts = opts.Hosts
	conf.Brokers = opts.Brokers
	conf.Topic = opts.Topic
	conf.Partition = opts.Partition

	if opts.configFile != "" {
		c, err := config.MergeDaemonConfigurations(conf, flags, opts.configFile)
		if err != nil {
			if flags.Changed("config-file") || !os.IsNotExist(err) {
				return nil, fmt.Errorf("unable to configure the hunter-agent with file %s: %v", opts.configFile, err)
			}
		}
		// the merged configuration can be nil if the config file didn't exist.
		// leave the current configuration as it is if when that happens.
		if c != nil {
			conf = c
		}
	}

	if err := config.Validate(conf); err != nil {
		return nil, err
	}

	setLogLevel(conf.LogLevel)

	return conf, nil
}

func loadListeners(cli *DaemonCli) ([]string, error) {
	var hosts []string
	for i := 0; i < len(cli.Config.Hosts); i++ {
		var err error
		if cli.Config.Hosts[i], err = dopts.ParseHost(cli.Config.Hosts[i]); err != nil {
			return nil, fmt.Errorf("error parsing -H %s : %v", cli.Config.Hosts[i], err)
		}

		protoAddr := cli.Config.Hosts[i]
		protoAddrParts := strings.SplitN(protoAddr, "://", 2)
		if len(protoAddrParts) != 2 {
			return nil, fmt.Errorf("bad format %s, expected PROTO://ADDR", protoAddr)
		}

		proto := protoAddrParts[0]
		addr := protoAddrParts[1]

		ls, err := listeners.Init(proto, addr)
		if err != nil {
			return nil, err
		}

		go serveGRPC(ls[0], cli)

		logrus.Debugf("--> Serve %s", ls[0].Addr().String())
		logrus.Infof("Listener created on %s (%s)", proto, addr)
		hosts = append(hosts, protoAddrParts[1])

		// TODO: add more service here
	}

	return hosts, nil
}

// ---------

func newAsyncProducer(brokerList []string) sarama.AsyncProducer {

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		logrus.Fatalln("Failed to start Sarama producer:", err)
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range producer.Errors() {
			logrus.Println("Failed to write access log entry:", err)
		}
	}()

	return producer
}

func serveGRPC(l net.Listener, cli *DaemonCli) {
	s := grpc.NewServer()
	p := newAsyncProducer(cli.Config.Brokers)
	defer func() {
		if err := p.Close(); err != nil {
			logrus.Println("Failed to close producer", err)
		}
	}()

	exporterproto.RegisterExportServer(s, &server{
		topic:     cli.Topic,
		partition: cli.Partition,
		producer:  p,
	})
	if err := s.Serve(l); err != nil {
		logrus.Errorf("Failed to serve: %v", err)
	}
}

func dumpSpans(spans []*traceproto.Span) ([]byte, error) {
	ds := &dumpproto.DumpSpans{Spans: spans}
	serialized, err := proto.Marshal(ds)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

type server struct {
	topic     string
	partition string
	producer  sarama.AsyncProducer
}

func (s *server) ExportSpan(stream exporterproto.Export_ExportSpanServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// debug
		for _, sp := range in.Spans {
			fmt.Printf("---> span: %v\n", sp)
		}

		dump, err := dumpSpans(in.Spans)
		if err != nil {
			logrus.Errorf("dumpSpans err: %v", err)
			return err
		}

		s.producer.Input() <- &sarama.ProducerMessage{
			Topic: s.topic,
			Key:   sarama.StringEncoder(s.partition),
			Value: sarama.ByteEncoder(dump),
		}
	}
}

func (s *server) ExportMetrics(stream exporterproto.Export_ExportMetricsServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Println(in)
	}
}

// ---------

const defaultDaemonConfigFile = "/etc/hunter/agent.json"

// setDefaultUmask sets the umask to 0022 to avoid problems
// caused by custom umask
func setDefaultUmask() error {
	desiredUmask := 0022
	unix.Umask(desiredUmask)
	if umask := unix.Umask(desiredUmask); umask != desiredUmask {
		return fmt.Errorf("failed to set umask: expected %#o, got %#o", desiredUmask, umask)
	}

	return nil
}

// setupConfigReloadTrap configures the USR2 signal to reload the configuration.
func (cli *DaemonCli) setupConfigReloadTrap() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, unix.SIGHUP)
	go func() {
		for range c {
			cli.reloadConfig()
		}
	}()
}
