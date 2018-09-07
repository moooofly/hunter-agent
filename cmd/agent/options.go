package main

import (
	"fmt"
	"os"

	"github.com/moooofly/hunter-agent/daemon/config"
	"github.com/moooofly/hunter-agent/opts"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type daemonOptions struct {
	// from config file
	daemonConfig *config.Config
	// from command line
	flags      *pflag.FlagSet
	configFile string

	// common options
	Debug     bool
	LogLevel  string
	Hosts     []string
	Brokers   []string
	Topic     string
	Partition string
}

// newDaemonOptions returns a new daemonOptions
func newDaemonOptions(config *config.Config) *daemonOptions {
	return &daemonOptions{
		daemonConfig: config,
	}
}

// InstallFlags adds flags for the common options on the FlagSet
func (o *daemonOptions) InstallCommonOptionsFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.configFile, "config-file", defaultDaemonConfigFile, "Hunter agent configuration file")

	flags.BoolVarP(&o.Debug, "debug", "D", false, "Enable debug mode")
	flags.StringVarP(&o.LogLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)

	hostOpt := opts.NewNamedListOptsRef("hosts", &o.Hosts, opts.ValidateHost)
	flags.VarP(hostOpt, "host", "H", "Hunter agent's listening address")

	brokerhostOpt := opts.NewNamedListOptsRef("brokers", &o.Brokers, opts.ValidateBrokerHost)
	flags.VarP(brokerhostOpt, "broker", "b", "The kafka broker host")

	flags.StringVar(&o.Topic, "topic", "jaeger-spans-test-001", `The Kafka topic`)
	flags.StringVar(&o.Partition, "partition", "", `The Kafka partition (If set, only one partition can be used, otherwise use traceid of span for multiple partitions instead)`)
}

// setLogLevel sets the logrus logging level
func setLogLevel(logLevel string) {
	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse logging level: %s\n", logLevel)
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
