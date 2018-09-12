package daemon // import "github.com/moooofly/hunter-agent/daemon"

import (
	"encoding/json"
	"fmt"

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
	attributes := map[string]string{}

	defer func() {
		jsonString, _ := json.Marshal(daemon.configStore)

		daemon.configStore.Unlock()
		logrus.Infof("Reloaded configuration: %s", jsonString)
	}()

	daemon.reloadDebug(conf, attributes)
	daemon.reloadShutdownTimeout(conf, attributes)
	daemon.reloadQueueSize(conf, attributes)

	return nil
}

// reloadDebug updates configuration with queue size option
// and updates the passed attributes
func (daemon *Daemon) reloadQueueSize(conf *config.Config, attributes map[string]string) {
	// update corresponding configuration
	if conf.IsValueSet("queue-size") {
		daemon.configStore.QueueSize = conf.QueueSize
		logrus.Debugf("Reset Queue Size: %d", daemon.configStore.QueueSize)
	}
	// prepare reload event attributes with updatable configurations
	attributes["queue-size"] = fmt.Sprintf("%t", daemon.configStore.QueueSize)
}

// reloadDebug updates configuration with Debug option
// and updates the passed attributes
func (daemon *Daemon) reloadDebug(conf *config.Config, attributes map[string]string) {
	// update corresponding configuration
	if conf.IsValueSet("debug") {
		daemon.configStore.Debug = conf.Debug
	}
	// prepare reload event attributes with updatable configurations
	attributes["debug"] = fmt.Sprintf("%t", daemon.configStore.Debug)
}

// reloadShutdownTimeout updates configuration with daemon shutdown timeout option
// and updates the passed attributes
func (daemon *Daemon) reloadShutdownTimeout(conf *config.Config, attributes map[string]string) {
	// update corresponding configuration
	if conf.IsValueSet("shutdown-timeout") {
		daemon.configStore.ShutdownTimeout = conf.ShutdownTimeout
		logrus.Debugf("Reset Shutdown Timeout: %d", daemon.configStore.ShutdownTimeout)
	}

	// prepare reload event attributes with updatable configurations
	attributes["shutdown-timeout"] = fmt.Sprintf("%d", daemon.configStore.ShutdownTimeout)
}
