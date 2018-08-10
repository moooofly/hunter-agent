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

	machineMemory uint64

	diskUsageRunning int32
	pruneRunning     int32
	hosts            map[string]bool // hosts stores the addresses the daemon is listening on
	startupDone      chan struct{}
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
	d.setupDumpStackTrap(stackDumpDir)

	if err := configureMaxThreads(config); err != nil {
		logrus.Warnf("Failed to configure golang's threads limit: %v", err)
	}

	//go d.execCommandGC()

	logrus.WithFields(logrus.Fields{
		"version": version.Version,
		"commit":  version.GitCommit,
	}).Info("Hunter agent")

	return d, nil
}

func (daemon *Daemon) waitForStartupDone() {
	<-daemon.startupDone
}

// FIXME: shutdown my service
/*
func (daemon *Daemon) shutdownContainer(c *container.Container) error {
	stopTimeout := c.StopTimeout()

	// If container failed to exit in stopTimeout seconds of SIGTERM, then using the force
	if err := daemon.containerStop(c, stopTimeout); err != nil {
		return fmt.Errorf("Failed to stop container %s with error: %v", c.ID, err)
	}

	// Wait without timeout for the container to exit.
	// Ignore the result.
	<-c.Wait(context.Background(), container.WaitConditionNotRunning)
	return nil
}
*/

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

	/*
		if daemon.containers != nil {
			logrus.Debugf("daemon configured with a %d seconds minimum shutdown timeout", daemon.configStore.ShutdownTimeout)
			logrus.Debugf("start clean shutdown of all containers with a %d seconds timeout...", daemon.ShutdownTimeout())
				daemon.containers.ApplyAll(func(c *container.Container) {
					if !c.IsRunning() {
						return
					}
					logrus.Debugf("stopping %s", c.ID)
					if err := daemon.shutdownContainer(c); err != nil {
						logrus.Errorf("Stop container error: %v", err)
						return
					}
					logrus.Debugf("container stopped %s", c.ID)
				})
		}
	*/

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
	// setup the daemons oom_score_adj
	/*
		if err := setupOOMScoreAdj(config.OOMScoreAdjust); err != nil {
			return err
		}
	*/
	return nil
}

/*
func setupOOMScoreAdj(score int) error {
	f, err := os.OpenFile("/proc/self/oom_score_adj", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	stringScore := strconv.Itoa(score)
	_, err = f.WriteString(stringScore)
	if os.IsPermission(err) {
		// Setting oom_score_adj does not work in an
		// unprivileged container. Ignore the error, but log
		// it if we appear not to be in that situation.
		if !rsystem.RunningInUserNS() {
			logrus.Debugf("Permission denied writing %q to /proc/self/oom_score_adj", stringScore)
		}
		return nil
	}

	return err
}
*/

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
