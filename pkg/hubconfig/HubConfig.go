// Package hubconfig with the hub configuration struct and methods
package hubconfig

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"text/template"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// HubConfigName the configuration file name of the hub
const HubConfigName = "hub.yaml"

// HubLogFile the file name of the hub logging
const HubLogFile = "hub.log"

// DefaultCertsFolder with the location of certificates
const DefaultCertsFolder = "./certs"

// DefaultPort for MQTT or Websocket over TLS port
const DefaultPortWS = 8883
const DefaultPortMqtt = 8884

type test interface {
	hello()
}

// HubConfig with hub configuration parameters
// Intended for hub plugins to provide hub services
type HubConfig struct {
	// logging
	Loglevel string `yaml:"logLevel"` // debug, info, warning, error. Default is warning
	LogFile  string `yaml:"logFile"`  // hub logging to file

	// MQTT message bus configuration
	MqttAddress    string `yaml:"mqttAddress,omitempty"`    // address with hostname or ip of the message bus
	MqttCertPort   int    `yaml:"mqttCertPort,omitempty"`   // MQTT TLS port for certificate based authentication
	MqttUnpwPortWS int    `yaml:"mqttUnpwPortWS,omitempty"` // Websocket TLS port for login/password authentication
	MqttTimeout    int    `yaml:"mqttTimeout,omitempty"`    // Client connection timeout in seconds. 0 for indefinite

	// zoning
	Zone string `yaml:"zone"` // zone this hub belongs to. Used as prefix in ThingID, default is local

	// Folders
	Home         string `yaml:"home"`         // application home directory. Default is parent of executable.
	CertsFolder  string `yaml:"certsFolder"`  // Folder containing certificates, default is {home}/certs
	ConfigFolder string `yaml:"configFolder"` // location of configuration files. Default is ./config
	// PluginFolder string   `yaml:"pluginFolder"` // location of plugin binaries. Default is ./bin
	Plugins []string `yaml:"plugins"` // names of plugins to start
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
		// PluginFolder: path.Join(homeFolder, "./bin"),
		Zone: "local",
	}
	// config.Messenger.CertsFolder = path.Join(homeFolder, "certs")
	config.CertsFolder = path.Join(homeFolder, DefaultCertsFolder)
	// config.Messenger.CaCertFile = certsetup.CaCertFile
	// config.Messenger.ServerCertFile = certsetup.ServerCertFile
	// config.Messenger.ClientCertFile = certsetup.ClientCertFile
	// config.Messenger.ClientKeyFile = certsetup.ClientKeyFile
	config.MqttAddress = GetOutboundIP("").String()
	config.MqttCertPort = DefaultPortMqtt
	config.MqttUnpwPortWS = DefaultPortWS
	config.Loglevel = "warning"
	// config.Logging.LogFile = path.Join(homeFolder, "logs/"+HubLogFile)
	config.LogFile = path.Join(homeFolder, "./logs/"+HubLogFile)
	return config
}

// Get the default outbound IP address to reach the given hostname.
// Use a local hostname if a subnet other than the default one should be used.
// Use "" for the default route address
//  destination to reach or "" to use 1.1.1.1 (no connection will be established)
func GetOutboundIP(destination string) net.IP {
	if destination == "" {
		destination = "1.1.1.1"
	}
	// This dial command doesn't actually create a connection
	conn, err := net.Dial("udp", destination+":80")
	if err != nil {
		logrus.Errorf("GetIPAddr: %s", err)
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
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
// This uses the -home and -c commandline arguments commands without the flag package.
// Intended for plugins or Hub apps that have their own commandlines but still need to use the
// Hub's base configuration.
// See also LoadCommandlineConfig() for including plugin and commandline options.
//
// This checks the following commandline arguments:
//  - Commandline "-c"  specifies an alternative hub configuration file
//  - Commandline "--home" sets the home folder as the base of ./config, ./logs and ./bin directories
//
//  homeFolder overrides the default home folder. Leave empty to use parent of application binary.
//  pluginID to substitute optional {{.pluginID}} in the hub.yaml file
// The current working directory is changed to this folder
//  Returns the hub configuration and error code in case of error
func LoadHubConfig(homeFolder string, pluginID string) (*HubConfig, error) {
	substituteMap := make(map[string]string)
	substituteMap["pluginID"] = pluginID

	args := os.Args[1:]
	if homeFolder == "" {
		// Option --home overrides the default home folder. Intended for testing.
		for index, arg := range args {
			if arg == "--home" || arg == "-home" {
				homeFolder = args[index+1]
				// make relative paths absolute
				if !path.IsAbs(homeFolder) {
					cwd, _ := os.Getwd()
					homeFolder = path.Join(cwd, homeFolder)
				}
				break
			}
		}
	}

	// set configuration defaults
	hubConfig := CreateDefaultHubConfig(homeFolder)
	hubConfigFile := path.Join(hubConfig.ConfigFolder, HubConfigName)

	// Option -c overrides the default hub config file. Intended for testing.
	args = os.Args[1:]
	for index, arg := range args {
		if arg == "-c" {
			hubConfigFile = args[index+1]
			// make relative paths absolute
			if !path.IsAbs(hubConfigFile) {
				hubConfigFile = path.Join(homeFolder, hubConfigFile)
			}
			logrus.Infof("Commandline option '-c %s' overrides defaulthub configfile", hubConfigFile)
			break
		}
	}
	logrus.Infof("Using %s as hub config file", hubConfigFile)
	err1 := LoadConfig(hubConfigFile, hubConfig, substituteMap)
	if err1 != nil {
		// panic("Unable to continue without hub.yaml")
		return hubConfig, err1
	}

	// make sure folders have an absolute path
	if !path.IsAbs(hubConfig.CertsFolder) {
		hubConfig.CertsFolder = path.Join(homeFolder, hubConfig.CertsFolder)
	}
	// if !path.IsAbs(hubConfig.PluginFolder) {
	// 	hubConfig.PluginFolder = path.Join(homeFolder, hubConfig.PluginFolder)
	// }

	// disabled as invalid hub config should prevent providing help
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

	// plugin config is optional
	if pluginName != "" && pluginConfig != nil {
		pluginConfigFile := path.Join(configFolder, pluginName+".yaml")
		err := LoadConfig(pluginConfigFile, pluginConfig, substituteMap)
		if err != nil {
			logrus.Infof("Plugin configuration file %s not found. Ignored", pluginConfigFile)
		}
	}

	return nil
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

	loggingFolder := path.Dir(config.LogFile)
	if _, err := os.Stat(loggingFolder); os.IsNotExist(err) {
		logrus.Errorf("Logging folder '%s' not found\n", loggingFolder)
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

	// Address must exist
	if config.MqttAddress == "" {
		err := fmt.Errorf("Message bus address not provided\n")
		logrus.Error(err)
		return err
	}

	return nil
}
