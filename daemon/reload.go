package daemon // import "github.com/moooofly/hunter-agent/daemon"

import (
	"encoding/json"

	"github.com/moooofly/hunter-agent/daemon/config"
	"github.com/sirupsen/logrus"
)

// Reload reads configuration changes and modifies the
// daemon according to those changes.
// These are the settings that Reload changes:
// - Daemon debug log level
// - Daemon shutdown timeout (in seconds)
func (daemon *Daemon) Reload(conf *config.Config) (err error) {
	daemon.configStore.Lock()

	defer func() {
		jsonString, _ := json.Marshal(daemon.configStore)

		daemon.configStore.Unlock()
		logrus.Infof("Reloaded configuration: %s", jsonString)
	}()

	daemon.reloadDebug(conf)
	daemon.reloadShutdownTimeout(conf)
	daemon.reloadQueueSize(conf)

	return nil
}

// reloadDebug updates configuration with queue size option
func (daemon *Daemon) reloadQueueSize(conf *config.Config) {
	// update corresponding configuration
	if conf.IsValueSet("queue-size") {
		daemon.configStore.QueueSize = conf.QueueSize
		logrus.Debugf("Reset Queue Size: %d", daemon.configStore.QueueSize)
	}
}

// reloadDebug updates configuration with Debug option
func (daemon *Daemon) reloadDebug(conf *config.Config) {
	// update corresponding configuration
	if conf.IsValueSet("debug") {
		daemon.configStore.Debug = conf.Debug
	}
}

// reloadShutdownTimeout updates configuration with daemon shutdown timeout option
func (daemon *Daemon) reloadShutdownTimeout(conf *config.Config) {
	// update corresponding configuration
	if conf.IsValueSet("shutdown-timeout") {
		daemon.configStore.ShutdownTimeout = conf.ShutdownTimeout
		logrus.Debugf("Reset Shutdown Timeout: %d", daemon.configStore.ShutdownTimeout)
	}
}
