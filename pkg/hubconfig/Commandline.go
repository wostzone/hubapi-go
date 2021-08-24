// Package hubconfig with Hub commandline configuration handling
package hubconfig

import (
	"errors"
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// SetHubCommandlineArgs creates common hub commandline flag commands for parsing commandlines
//
//  -c            /path/to/hub.yaml    optional alternative configuration, default is {home}/config/hub.yaml
//  -home         /path/to/app/home    optional alternative application home folder/ Defa
//  -certsFolder  /path/to/alt/certs   optional certificate folder, eg when using mqtt. Default is {home}/certs
//  -configFolder /path/to/alt/config  optional alternative config, eg /etc/wost
//  -address      localhost            optional message bus address
//  -mqttPortUnpw 9883                 mqtt port for username/password authentication
//  -mqttPortCert 9884                 mqtt port for certificate authentication
//  -mqttPortWS   9001                 websocket port for username/password authentication
//  -logsFolder  /path/to/folder       optional logfile location
//  -logLevel warning                  for extra logging, default is hub loglevel
//
func SetHubCommandlineArgs(config *HubConfig) {
	// Flags -c and --home are handled separately in SetupConfig. It is added here to avoid flag parse error
	flag.String("c", "hub.yaml", "Set the hub configuration file ")
	flag.StringVar(&config.Home, "home", config.Home, "Application working `folder`")

	flag.StringVar(&config.CertsFolder, "certsFolder", config.CertsFolder, "Optional certificates directory for TLS")
	flag.StringVar(&config.ConfigFolder, "configFolder", config.ConfigFolder, "Plugin configuration `folder`")
	flag.StringVar(&config.MqttAddress, "mqttAddress", config.MqttAddress, "Message bus hostname or address")
	flag.IntVar(&config.MqttPortWS, "mqttPortWS", config.MqttPortWS, "Websocket TLS username/pw auth port")
	flag.IntVar(&config.MqttPortUnpw, "mqttPortUnpw", config.MqttPortUnpw, "MQTT TLS with username/pw auth port")
	flag.IntVar(&config.MqttPortCert, "mqttPortCert", config.MqttPortCert, "MQTT TLS Client certificate auth port")
	flag.StringVar(&config.LogsFolder, "logsFolder", config.LogsFolder, "Logging folder")
	flag.StringVar(&config.Loglevel, "logLevel", config.Loglevel, "Loglevel: {error|`warning`|info|debug}")
}

// LoadCommandlineConfig loads the hub and plugin configurations (See LoadPluginConfig)
// and applies commandline parameters to allow modifying this configuration from the
// commandline.
// Returns the hub configuration and error code in case of error
func LoadCommandlineConfig(homeFolder string, pluginID string, pluginConfig interface{}) (*HubConfig, error) {
	if pluginID == "" {
		err := errors.New("LoadCommandlineConfig: Missing plugin/hub ID")
		logrus.Errorf("%s", err)
		return nil, err
	}

	// Option --home overrides the default home folder.
	args := os.Args[1:]
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

	// Option -c overrides the default hub config file.
	hubConfigFile := ""
	for index, arg := range args {
		if arg == "-c" {
			hubConfigFile = args[index+1]
			// make relative paths absolute
			if !path.IsAbs(hubConfigFile) {
				hubConfigFile = path.Join(homeFolder, hubConfigFile)
			}
			logrus.Infof("Commandline option '-c %s' overrides defaulthub config file", hubConfigFile)
			break
		}
	}

	hubConfig, err := LoadHubConfig(hubConfigFile, homeFolder, pluginID)
	if err != nil {
		logrus.Errorf("LoadCommandlineConfig: Failed loading Hub configuration: %s", err)
		// return hubConfig, err
	}
	// non-fatal as config file is optional and defaults should work
	LoadPluginConfig(hubConfig.ConfigFolder, pluginID, pluginConfig, nil)

	SetHubCommandlineArgs(hubConfig)

	// catch parsing errors, in case flag.ErrorHandling = flag.ContinueOnError
	// parse commandline arguments before exiting on error
	err3 := flag.CommandLine.Parse(os.Args[1:])
	if err3 != nil {
		return hubConfig, err3
	} else if err != nil {
		return hubConfig, err
	}

	// validation pass in case commandline argument messed up the config
	err = ValidateHubConfig(hubConfig)

	// It is up to the app to change to the home directory.
	// os.Chdir(hubConfig.HomeFolder)

	// Last set the hub/plugin logging
	logFileName := path.Join(hubConfig.LogsFolder, pluginID+".log")
	SetLogging(hubConfig.Loglevel, logFileName)
	return hubConfig, err
}
