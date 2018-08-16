package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/imdario/mergo"
	"github.com/moooofly/hunter-agent/opts"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// Config defines the configuration of the Hunter agent.
// It includes json tags to deserialize configuration from a file
// using the same names that the flags in the command line uses.
type Config struct {
	Pidfile string `json:"pidfile,omitempty"`
	Root    string `json:"data-root,omitempty"`
	RawLogs bool   `json:"raw-logs,omitempty"`

	// ShutdownTimeout is the timeout value (in seconds) the agent will wait for something
	// to stop when Hunter agent is being shutdown
	ShutdownTimeout int `json:"shutdown-timeout,omitempty"`

	Debug    bool     `json:"debug,omitempty"`
	LogLevel string   `json:"log-level,omitempty"`
	Hosts    []string `json:"hosts,omitempty"`

	Brokers   []string `json:"brokers,omitempty"`
	Topic     string   `json:topic,omitempty`
	Partition string   `json:partition,omitempty`

	MetricsAddress string `json:"metrics-addr,omitempty"`

	sync.Mutex
	ValuesSet map[string]interface{} `json:"-"`
}

// IsValueSet returns true if a configuration value
// was explicitly set in the configuration file.
func (conf *Config) IsValueSet(name string) bool {
	if conf.ValuesSet == nil {
		return false
	}
	_, ok := conf.ValuesSet[name]
	return ok
}

// New returns a new fully initialized Config struct
func New() *Config {
	config := Config{}
	// FIXME: do some initialization here
	return &config
}

// Reload reads the configuration in the host and reloads the daemon and server.
func Reload(configFile string, flags *pflag.FlagSet, reload func(*Config)) error {
	logrus.Infof("Got signal to reload configuration, reloading from: %s", configFile)

	newConfig := New()

	reload(newConfig)
	return nil
}

// boolValue is an interface that boolean value flags implement
// to tell the command line how to make -name equivalent to -name=true.
type boolValue interface {
	IsBoolFlag() bool
}

// MergeDaemonConfigurations reads a configuration file,
// loads the file configuration in an isolated structure,
// and merges the configuration provided from flags on top
// if there are no conflicts.
func MergeDaemonConfigurations(flagsConfig *Config, flags *pflag.FlagSet, configFile string) (*Config, error) {
	fileConfig, err := getConflictFreeConfiguration(configFile, flags)
	if err != nil {
		return nil, err
	}

	if err := Validate(fileConfig); err != nil {
		return nil, fmt.Errorf("configuration validation from file failed (%v)", err)
	}

	// merge flags configuration on top of the file configuration
	if err := mergo.Merge(fileConfig, flagsConfig); err != nil {
		return nil, err
	}

	// We need to validate again once both fileConfig and flagsConfig
	// have been merged
	if err := Validate(fileConfig); err != nil {
		return nil, fmt.Errorf("merged configuration validation from file and command line flags failed (%v)", err)
	}

	return fileConfig, nil
}

// getConflictFreeConfiguration loads the configuration from a JSON file.
// It compares that configuration with the one provided by the flags,
// and returns an error if there are conflicts.
func getConflictFreeConfiguration(configFile string, flags *pflag.FlagSet) (*Config, error) {
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	var reader io.Reader
	if flags != nil {
		var jsonConfig map[string]interface{}
		reader = bytes.NewReader(b)
		if err := json.NewDecoder(reader).Decode(&jsonConfig); err != nil {
			return nil, err
		}

		configSet := configValuesSet(jsonConfig)

		if err := findConfigurationConflicts(configSet, flags); err != nil {
			return nil, err
		}

		// Override flag values to make sure the values set in the config file with nullable values, like `false`,
		// are not overridden by default truthy values from the flags that were not explicitly set.
		// See https://github.com/docker/docker/issues/20289 for an example.
		//
		// TODO: Rewrite configuration logic to avoid same issue with other nullable values, like numbers.
		namedOptions := make(map[string]interface{})
		for key, value := range configSet {
			f := flags.Lookup(key)
			if f == nil { // ignore named flags that don't match
				namedOptions[key] = value
				continue
			}

			if _, ok := f.Value.(boolValue); ok {
				f.Value.Set(fmt.Sprintf("%v", value))
			}
		}
		if len(namedOptions) > 0 {
			// set also default for mergeVal flags that are boolValue at the same time.
			flags.VisitAll(func(f *pflag.Flag) {
				if opt, named := f.Value.(opts.NamedOption); named {
					v, set := namedOptions[opt.Name()]
					_, boolean := f.Value.(boolValue)
					if set && boolean {
						f.Value.Set(fmt.Sprintf("%v", v))
					}
				}
			})
		}

		config.ValuesSet = configSet
	}

	reader = bytes.NewReader(b)
	if err := json.NewDecoder(reader).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate validates some specific configs.
func Validate(config *Config) error {
	// validate some configurations here
	return nil
}

// configValuesSet returns the configuration values explicitly set in the file.
func configValuesSet(config map[string]interface{}) map[string]interface{} {
	flatten := make(map[string]interface{})
	for k, v := range config {
		if m, isMap := v.(map[string]interface{}); isMap {
			for km, vm := range m {
				flatten[km] = vm
			}
			continue
		}

		flatten[k] = v
	}
	return flatten
}

// findConfigurationConflicts iterates over the provided flags searching for
// duplicated configurations and unknown keys. It returns an error with all the conflicts if
// it finds any.
func findConfigurationConflicts(config map[string]interface{}, flags *pflag.FlagSet) error {
	// 1. Search keys from the file that we don't recognize as flags.
	unknownKeys := make(map[string]interface{})
	for key, value := range config {
		if flag := flags.Lookup(key); flag == nil {
			unknownKeys[key] = value
		}
	}

	// 2. Discard values that implement NamedOption.
	// Their configuration name differs from their flag name, like `labels` and `label`.
	if len(unknownKeys) > 0 {
		unknownNamedConflicts := func(f *pflag.Flag) {
			if namedOption, ok := f.Value.(opts.NamedOption); ok {
				if _, valid := unknownKeys[namedOption.Name()]; valid {
					delete(unknownKeys, namedOption.Name())
				}
			}
		}
		flags.VisitAll(unknownNamedConflicts)
	}

	if len(unknownKeys) > 0 {
		var unknown []string
		for key := range unknownKeys {
			unknown = append(unknown, key)
		}
		return fmt.Errorf("the following directives don't match any configuration option: %s", strings.Join(unknown, ", "))
	}

	var conflicts []string
	printConflict := func(name string, flagValue, fileValue interface{}) string {
		return fmt.Sprintf("%s: (from flag: %v, from file: %v)", name, flagValue, fileValue)
	}

	// 3. Search keys that are present as a flag and as a file option.
	duplicatedConflicts := func(f *pflag.Flag) {
		// search option name in the json configuration payload if the value is a named option
		if namedOption, ok := f.Value.(opts.NamedOption); ok {
			if optsValue, ok := config[namedOption.Name()]; ok {
				conflicts = append(conflicts, printConflict(namedOption.Name(), f.Value.String(), optsValue))
			}
		} else {
			// search flag name in the json configuration payload
			for _, name := range []string{f.Name, f.Shorthand} {
				if value, ok := config[name]; ok {
					conflicts = append(conflicts, printConflict(name, f.Value.String(), value))
					break
				}
			}
		}
	}

	flags.Visit(duplicatedConflicts)

	if len(conflicts) > 0 {
		return fmt.Errorf("the following directives are specified both as a flag and in the configuration file: %s", strings.Join(conflicts, ", "))
	}
	return nil
}
