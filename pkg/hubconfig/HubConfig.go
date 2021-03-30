// Package hubconfig with the hub configuration struct and methods
package hubconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// HubConfigName the configuration file name of the hub
const HubConfigName = "hub.yaml"

// HubLogFile the file name of the hub logging
const HubLogFile = "hub.log"

// DefaultCertsFolder with the location of certificates
const DefaultCertsFolder = "./certs"

// DefaultPort with MQTT TLS port
const DefaultPort = 8883

type test interface {
	hello()
}

// HubConfig with hub configuration parameters
// Intended for hub plugins to provide hub services
type HubConfig struct {
	Logging struct {
		Loglevel   string `yaml:"logLevel"`   // debug, info, warning, error. Default is warning
		LogFile    string `yaml:"logFile"`    // hub logging to file
		TimeFormat string `yaml:"timeFormat"` // go default ISO8601 ("2006-01-02T15:04:05.000-0700")
	} `yaml:"logging"`

	// Configuration of hub client messaging
	Messenger struct {
		Address     string `yaml:"address"`           // address with hostname or ip of the message bus
		Port        int    `yaml:"port,omitempty"`    // optional port, default is 8883 for MQTT TLS
		CertsFolder string `yaml:"certsFolder"`       // Folder containing certificates, default is {home}/certs
		Signing     bool   `yaml:"signing,omitempty"` // Message signing to be used by all publishers, default is false
		Timeout     int    `yaml:"timeout,omitempty"` // Client connection timeout in seconds. 0 for indefinite
	} `yaml:"messenger"`

	Home         string   `yaml:"home"`         // application home directory. Default is parent of executable.
	Zone         string   `yaml:"zone"`         // zone this hub belongs to. Used as prefix in ThingID, default is local
	ConfigFolder string   `yaml:"configFolder"` // location of configuration files. Default is ./config
	PluginFolder string   `yaml:"pluginFolder"` // location of plugin binaries. Default is ./bin
	Plugins      []string `yaml:"plugins"`      // names of plugins to start
	// internal
}

// ConfigArgs configuration commandline arguments
type ConfigArgs struct {
	name         string
	defaultValue string
	description  string
}

// CreateDefaultHubConfig with default values
// homeFolder is the home of the application, log and configuration folders.
// Use "" for default: parent of application binary
// When relative path is given, it is relative to the application binary
func CreateDefaultHubConfig(homeFolder string) *HubConfig {
	appBin, _ := os.Executable()
	binFolder := path.Dir(appBin)
	if homeFolder == "" {
		homeFolder = path.Dir(binFolder)
	} else if !path.IsAbs(homeFolder) {
		// turn relative home folder in absolute path
		homeFolder = path.Join(binFolder, homeFolder)
	}
	logrus.Infof("AppBin is: %s; Home is: %s", appBin, homeFolder)
	config := &HubConfig{
		// ConfigFolder: path.Join(homeFolder, "config"),
		Home:         homeFolder,
		ConfigFolder: path.Join(homeFolder, "config"),
		Plugins:      make([]string, 0),
		PluginFolder: path.Join(homeFolder, "./bin"),
		Zone:         "local",
	}
	// config.Messenger.CertsFolder = path.Join(homeFolder, "certs")
	config.Messenger.CertsFolder = path.Join(homeFolder, DefaultCertsFolder)
	// config.Messenger.CaCertFile = certsetup.CaCertFile
	// config.Messenger.ServerCertFile = certsetup.ServerCertFile
	// config.Messenger.ClientCertFile = certsetup.ClientCertFile
	// config.Messenger.ClientKeyFile = certsetup.ClientKeyFile
	config.Messenger.Address = "localhost"
	config.Messenger.Port = DefaultPort
	config.Logging.Loglevel = "warning"
	// config.Logging.LogFile = path.Join(homeFolder, "logs/"+HubLogFile)
	config.Logging.LogFile = path.Join(homeFolder, "./logs/"+HubLogFile)
	return config
}

// LoadConfig loads the configuration from file into the given config
// Returns nil if successful
func LoadConfig(configFile string, config interface{}) error {
	var err error
	var rawConfig []byte
	rawConfig, err = ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Warningf("Unable to load config file: %s", err)
		return err
	}
	logrus.Infof("Loaded config file '%s'", configFile)

	err = yaml.Unmarshal(rawConfig, config)
	if err != nil {
		logrus.Errorf("Error parsing config file '%s': %s", configFile, err)
		return err
	}
	return nil
}

// ValidateConfig checks if values in the hub configuration are correct
// Returns an error if the config is invalid
func ValidateConfig(config *HubConfig) error {
	if _, err := os.Stat(config.Home); os.IsNotExist(err) {
		logrus.Errorf("Home folder '%s' not found\n", config.Home)
		return err
	}
	if _, err := os.Stat(config.ConfigFolder); os.IsNotExist(err) {
		logrus.Errorf("Configuration folder '%s' not found\n", config.ConfigFolder)
		return err
	}

	loggingFolder := path.Dir(config.Logging.LogFile)
	if _, err := os.Stat(loggingFolder); os.IsNotExist(err) {
		logrus.Errorf("Logging folder '%s' not found\n", loggingFolder)
		return err
	}

	if _, err := os.Stat(config.Messenger.CertsFolder); os.IsNotExist(err) {
		logrus.Errorf("TLS certificate folder '%s' not found\n", config.Messenger.CertsFolder)
		return err
	}
	// // Pluginfolder is either empty or must exist
	// if config.PluginFolder != "" {
	// 	if _, err := os.Stat(config.PluginFolder); os.IsNotExist(err) {
	// 		logrus.Warningf("Plugins folder '%s' not found.\n", config.PluginFolder)
	// 		return err
	// 	}
	// }

	// Address must exist
	if config.Messenger.Address == "" {
		err := fmt.Errorf("Message bus address not provided\n")
		logrus.Error(err)
		return err
	}

	return nil
}
