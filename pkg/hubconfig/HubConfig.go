// Package hubconfig with the hub configuration struct and methods
package hubconfig

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubserve-go/pkg/hubnet"
	"gopkg.in/yaml.v2"
)

// HubConfigName the configuration file name of the hub
const HubConfigName = "hub.yaml"

// DefaultCertsFolder with the location of certificates
const DefaultCertsFolder = "./certs"

// location of config files wrt home folder
const DefaultConfigFolder = "./config"

// location of log files wrt home folder
const DefaultLogsFolder = "./logs"

// auth
const (
	DefaultAclFile  = "hub.acl"
	DefaultUnpwFile = "hub.passwd"
)

// DefaultPort for MQTT or Websocket over TLS port
const DefaultMqttPortUnpw = 8883
const DefaultMqttPortCert = 8884
const DefaultMqttPortWS = 9001

// HubConfig with hub configuration parameters
// Intended for hub plugins to provide hub services
type HubConfig struct {
	// logging
	Loglevel   string `yaml:"logLevel"`   // debug, info, warning, error. Default is warning
	LogsFolder string `yaml:"logsFolder"` // location of Wost log files
	// LogFile   string `yaml:"logFile"`   // log filename is pluginID.log

	// MQTT message bus configuration
	MqttAddress  string `yaml:"mqttAddress,omitempty"`  // address with hostname or ip of the message bus
	MqttPortCert int    `yaml:"mqttPortCert,omitempty"` // MQTT TLS port for certificate based authentication
	MqttPortUnpw int    `yaml:"mqttPortUnpw,omitempty"` // MQTT TLS port for login/password authentication
	MqttPortWS   int    `yaml:"mqttPortWS,omitempty"`   // Websocket TLS port for login/password authentication
	MqttTimeout  int    `yaml:"mqttTimeout,omitempty"`  // plugin mqtt connection timeout in seconds. 0 for indefinite

	// auth
	AclStorePath  string `yaml:"aclStore"`  // path to the ACL store
	UnpwStorePath string `yaml:"unpwStore"` // path to the uername/password store

	// zoning
	Zone string `yaml:"zone"` // zone this hub belongs to. Used as prefix in ThingID, default is local

	// Folders
	Home         string   `yaml:"home"`         // application home directory. Default is parent of executable.
	CertsFolder  string   `yaml:"certsFolder"`  // Folder containing certificates, default is {home}/certs
	ConfigFolder string   `yaml:"configFolder"` // location of configuration files. Default is ./config
	Plugins      []string `yaml:"plugins"`      // names of plugins to start
}

// ConfigArgs configuration commandline arguments
// type ConfigArgs struct {
// 	name         string
// 	defaultValue string
// 	description  string
// }

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
		Home:         homeFolder,
		ConfigFolder: path.Join(homeFolder, DefaultConfigFolder),
		Plugins:      make([]string, 0),
		Zone:         "local",
	}
	// config.Messenger.CertsFolder = path.Join(homeFolder, "certs")
	config.CertsFolder = path.Join(homeFolder, DefaultCertsFolder)
	// config.CaCertFile = certsetup.CaCertFile
	// config.Messenger.ServerCertFile = certsetup.ServerCertFile
	// config.Messenger.ClientCertFile = certsetup.ClientCertFile
	// config.Messenger.ClientKeyFile = certsetup.ClientKeyFile
	config.MqttAddress = hubnet.GetOutboundIP("").String()
	config.MqttPortCert = DefaultMqttPortCert
	config.MqttPortUnpw = DefaultMqttPortUnpw
	config.MqttPortWS = DefaultMqttPortWS
	config.Loglevel = "warning"
	config.LogsFolder = path.Join(homeFolder, "logs")
	config.AclStorePath = path.Join(config.ConfigFolder, DefaultAclFile)
	config.UnpwStorePath = path.Join(config.ConfigFolder, DefaultUnpwFile)
	return config
}

// LoadConfig loads the configuration from file into the given config
//  configFile path to yaml configuration file
//  config interface to typed structure matching the config. Must have yaml tags
//  substituteMap map to substitude {{.key}} with value from map, nil to ignore
// Returns nil if successful
func LoadConfig(configFile string, config interface{}, substituteMap map[string]string) error {
	var err error
	var rawConfig []byte
	rawConfig, err = ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Infof("Unable to load config file: %s", err)
		return err
	}
	logrus.Infof("Loaded config file '%s'", configFile)
	rawText := string(rawConfig)
	if substituteMap != nil {
		rawText = SubstituteText(rawText, substituteMap)
	}

	err = yaml.Unmarshal([]byte(rawText), config)
	if err != nil {
		logrus.Errorf("Error parsing config file '%s': %s", configFile, err)
		return err
	}
	return nil
}

// LoadHubConfig loads the hub configuration
// See also LoadCommandlineConfig() for including commandline options.
// The following defaults are used:
//  - The home folder is the parent of the application binary
//  - the config folder is the 'config' subdirectory of home
//  - the certs folder is the 'certs' subdirectory of home
//  - the logs folder is the 'logs' subdirectory of home
//  - the configfile to load is {config}/hub.yaml
//
//  configFile to load. Use "" for default {home}/config/hub.yaml
//  homeFolder to use. Use "" for default {appbin}/..
//  pluginID to substitute optional {{.pluginID}} in the hub.yaml file
//
// Returns the hub configuration and error code in case of error
func LoadHubConfig(configFile string, homeFolder string, pluginID string) (*HubConfig, error) {
	substituteMap := make(map[string]string)
	substituteMap["pluginID"] = pluginID

	if homeFolder == "" {
		appPath, _ := os.Executable()
		binFolder := path.Dir(appPath)
		homeFolder = path.Dir(binFolder)
	}

	// set configuration defaults
	hubConfig := CreateDefaultHubConfig(homeFolder)

	substituteMap["home"] = hubConfig.Home
	substituteMap["config"] = hubConfig.ConfigFolder
	substituteMap["logs"] = hubConfig.LogsFolder
	substituteMap["certs"] = hubConfig.CertsFolder

	if configFile == "" {
		configFile = path.Join(hubConfig.ConfigFolder, HubConfigName)
	}
	logrus.Infof("Using %s as hub config file", configFile)
	err1 := LoadConfig(configFile, hubConfig, substituteMap)
	if err1 != nil {
		return hubConfig, err1
	}

	// make sure folders have an absolute path
	if !path.IsAbs(hubConfig.CertsFolder) {
		hubConfig.CertsFolder = path.Join(homeFolder, hubConfig.CertsFolder)
	}
	// if !path.IsAbs(hubConfig.PluginFolder) {
	// 	hubConfig.PluginFolder = path.Join(homeFolder, hubConfig.PluginFolder)
	// }

	// disabled as invalid hub config should not prevent providing help
	//err2 := ValidateHubConfig(hubConfig)
	// if err2 != nil {
	// 	return hubConfig, err2
	// }
	return hubConfig, nil
}

// LoadPluginConfig loads the plugin configuration from file
//
// This uses the -home and -c commandline arguments commands without the flag package.
// Intended for plugins or Hub apps that have their own commandlines.
//
// The plugin configuration is the {pluginName}.yaml. If no pluginName is given it is ignored.
//
//  configFolder with the location of the plugin configuration
//  pluginName is the plugin instance name used to determine the config filename
//  pluginConfig is the configuration to load. nil to only load the hub config.
// Returns nil on success or error
func LoadPluginConfig(configFolder string, pluginName string, pluginConfig interface{}, substituteMap map[string]string) error {
	var err error
	// plugin config is optional
	if pluginName != "" && pluginConfig != nil {
		pluginConfigFile := path.Join(configFolder, pluginName+".yaml")
		err = LoadConfig(pluginConfigFile, pluginConfig, substituteMap)
		if err != nil {
			logrus.Infof("Plugin configuration file %s not found. Ignored", pluginConfigFile)
		}
	}

	return err
}

// Substitute template strings in the text
//  text to substitude template strings, eg "hello {{.destination}}"
//  substituteMap with replacement keywords, eg {"destination":"world"}
// Returns text with template strings replaced
func SubstituteText(text string, substituteMap map[string]string) string {
	var msg bytes.Buffer

	tpl, err := template.New("").Parse(text)
	_ = err
	tpl.Execute(&msg, substituteMap)
	return msg.String()
}

// ValidateHubConfig checks if values in the hub configuration are correct
// Returns an error if the config is invalid
func ValidateHubConfig(config *HubConfig) error {
	if _, err := os.Stat(config.Home); os.IsNotExist(err) {
		logrus.Errorf("Home folder '%s' not found\n", config.Home)
		return err
	}
	if _, err := os.Stat(config.ConfigFolder); os.IsNotExist(err) {
		logrus.Errorf("Configuration folder '%s' not found\n", config.ConfigFolder)
		return err
	}

	if _, err := os.Stat(config.LogsFolder); os.IsNotExist(err) {
		logrus.Errorf("Logging folder '%s' not found\n", config.LogsFolder)
		return err
	}

	if _, err := os.Stat(config.CertsFolder); os.IsNotExist(err) {
		logrus.Errorf("TLS certificate folder '%s' not found\n", config.CertsFolder)
		return err
	}
	// // Pluginfolder is either empty or must exist
	// if config.PluginFolder != "" {
	// 	if _, err := os.Stat(config.PluginFolder); os.IsNotExist(err) {
	// 		logrus.Warningf("Plugins folder '%s' not found.\n", config.PluginFolder)
	// 		return err
	// 	}
	// }

	return nil
}
