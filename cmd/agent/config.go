package main

import (
	"github.com/moooofly/hunter-agent/daemon/config"
	"github.com/spf13/pflag"
)

var defaultPidFile = "/var/run/hunter-agent.pid"
var defaultDataRoot = "/var/lib/hunter"

// defaultShutdownTimeout is the default shutdown timeout for the daemon
const defaultShutdownTimeout = 15
const defaultQueueSize = 1000

// installConfigFlags adds flags to the pflag.FlagSet to configure the daemon
func installDaemonConfigFlags(conf *config.Config, flags *pflag.FlagSet) {
	flags.StringVarP(&conf.Pidfile, "pidfile", "p", defaultPidFile, "Path to use for daemon PID file")
	flags.StringVar(&conf.Root, "data-root", defaultDataRoot, "Root directory for keeping some files")
	flags.BoolVar(&conf.RawLogs, "raw-logs", false, "Full timestamps without ANSI coloring")
	flags.IntVar(&conf.ShutdownTimeout, "shutdown-timeout", defaultShutdownTimeout, "Set the default shutdown timeout")

	flags.IntVar(&conf.QueueSize, "queue-size", defaultQueueSize, "Set maximum size of internal queues for flow control")

	// FIXME: put this one here is ok?
	flags.StringVar(&conf.MetricsAddress, "metrics-addr", "", "Set default address and port to serve the metrics api on")
}
