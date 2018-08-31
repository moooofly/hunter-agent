// Package daemon exposes the functions that occur on the host server
// that the Docker daemon is running.
//
// In implementing the various functions of the daemon, there is often
// a method-specific struct for configuring the runtime behavior.
package daemon // import "github.com/moooofly/hunter-agent/daemon"

import (
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/moooofly/hunter-agent/daemon/config"
	stackdump "github.com/moooofly/hunter-agent/pkg/signal"
	"github.com/moooofly/hunter-agent/version"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// Daemon holds information about the Docker daemon.
type Daemon struct {
	configStore *config.Config
	shutdown    bool

	hosts       map[string]bool // hosts stores the addresses the daemon is listening on
	startupDone chan struct{}
}

// StoreHosts stores the addresses the daemon is listening on
func (daemon *Daemon) StoreHosts(hosts []string) {
	if daemon.hosts == nil {
		daemon.hosts = make(map[string]bool)
	}
	for _, h := range hosts {
		daemon.hosts[h] = true
	}
}

// NewDaemon sets up everything for the daemon to be able to service
// requests from the webserver.
func NewDaemon(config *config.Config) (daemon *Daemon, err error) {

	if err := setupDaemonProcess(config); err != nil {
		return nil, err
	}

	d := &Daemon{
		configStore: config,
		startupDone: make(chan struct{}),
	}
	// Ensure the daemon is properly shutdown if there is a failure during
	// initialization
	defer func() {
		if err != nil {
			if err := d.Shutdown(); err != nil {
				logrus.Error(err)
			}
		}
	}()

	// set up SIGUSR1 handler on Unix-like systems to dump Go routine stacks
	stackDumpDir := config.Root
	logrus.Debugf("--> config.Root: %v", config.Root)
	d.setupDumpStackTrap(stackDumpDir)

	if err := configureMaxThreads(config); err != nil {
		logrus.Warnf("Failed to configure golang's threads limit: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"version":    version.Version,
		"commit":     version.GitCommit,
		"build time": version.BuildTime,
	}).Info("Hunter agent")

	return d, nil
}

func (daemon *Daemon) waitForStartupDone() {
	<-daemon.startupDone
}

// ShutdownTimeout returns the timeout (in seconds) before containers are forcibly
// killed during shutdown. The default timeout can be configured both on the daemon
// and per container, and the longest timeout will be used. A grace-period of
// 5 seconds is added to the configured timeout.
//
// A negative (-1) timeout means "indefinitely", which means that containers
// are not forcibly killed, and the daemon shuts down after all containers exit.
func (daemon *Daemon) ShutdownTimeout() int {
	shutdownTimeout := daemon.configStore.ShutdownTimeout
	if shutdownTimeout < 0 {
		return -1
	}

	//graceTimeout := 5
	// FIXME: gracefull shutdown
	/*
		for _, c := range daemon.containers.List() {
			stopTimeout := c.StopTimeout()
			if stopTimeout < 0 {
				return -1
			}
			if stopTimeout+graceTimeout > shutdownTimeout {
				shutdownTimeout = stopTimeout + graceTimeout
			}
		}
	*/
	return shutdownTimeout
}

// Shutdown stops the daemon.
func (daemon *Daemon) Shutdown() error {
	daemon.shutdown = true

	// FIXME: do something here
	return nil
}

// IsShuttingDown tells whether the daemon is shutting down or not
func (daemon *Daemon) IsShuttingDown() bool {
	return daemon.shutdown
}

// ------

// configureMaxThreads sets the Go runtime max threads threshold
// which is 90% of the kernel setting from /proc/sys/kernel/threads-max
func configureMaxThreads(config *config.Config) error {
	mt, err := ioutil.ReadFile("/proc/sys/kernel/threads-max")
	if err != nil {
		return err
	}
	mtint, err := strconv.Atoi(strings.TrimSpace(string(mt)))
	if err != nil {
		return err
	}
	maxThreads := (mtint / 100) * 90
	debug.SetMaxThreads(maxThreads)
	logrus.Debugf("Golang's threads limit set to %d", maxThreads)
	return nil
}

// setupDaemonProcess sets various settings for the daemon's process
func setupDaemonProcess(config *config.Config) error {
	return nil
}

func (d *Daemon) setupDumpStackTrap(root string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, unix.SIGUSR1)
	go func() {
		for range c {
			path, err := stackdump.DumpStacks(root)
			if err != nil {
				logrus.WithError(err).Error("failed to write goroutines dump")
			} else {
				logrus.Infof("goroutine stacks written to %s", path)
			}
		}
	}()
}
