// Package hubconfig with Hub commandline configuration handling
package hubconfig

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// SetHubCommandlineArgs creates common hub commandline flag commands for parsing commandlines
//
// -c            /path/to/hub.yaml    optional alternative configuration, default is {home}/config/hub.yaml
// -home         /path/to/app/home    optional alternative application home folder/ Defa
// -certsFolder  /path/to/alt/certs   optional certificate folder, eg when using mqtt. Default is {home}/certs
// -configFolder /path/to/alt/config  optional alternative config, eg /etc/wost
// -address      localhost            optional message bus address
// -certPortMqtt 9883                 mqtt port for certificate authentication
// -unpwPortWS   9884                 websocket port for username/password authentication
// -logFile      /path/to/plugin.log  optional logfile override
// -logLevel warning                  for extra logging, default is hub loglevel
//
func SetHubCommandlineArgs(config *HubConfig) {
	// Flags -c and --home are handled separately in SetupConfig. It is added here to avoid flag parse error
	flag.String("c", "hub.yaml", "Set the hub configuration file ")
	flag.StringVar(&config.Home, "home", config.Home, "Application working `folder`")

	flag.StringVar(&config.CertsFolder, "certsFolder", config.CertsFolder, "Optional certificates directory for TLS")
	flag.StringVar(&config.ConfigFolder, "configFolder", config.ConfigFolder, "Plugin configuration `folder`")
	flag.StringVar(&config.MqttAddress, "mqttAddress", config.MqttAddress, "Message bus hostname or address")
	flag.IntVar(&config.MqttUnpwPortWS, "mqttUnpwPortWS", config.MqttUnpwPortWS, "Websocket TLS username/pw auth port")
	flag.IntVar(&config.MqttCertPort, "mqttCertPort", config.MqttCertPort, "MQTT TLS Client certificate auth port")
	flag.StringVar(&config.LogFile, "logFile", config.LogFile, "Log to file")
	// flag.StringVar(&config.PluginFolder, "pluginFolder", config.PluginFolder, "Alternate plugin `folder`. Empty to not load plugins.")
	flag.StringVar(&config.Loglevel, "logLevel", config.Loglevel, "Loglevel: {error|`warning`|info|debug}")
}

// LoadCommandlineConfig loads the hub and plugin configurations (See LoadPluginConfig)
// and applies commandline parameters to allow modifying this configuration from the
// commandline.
// Returns the hub configuration and error code in case of error
func LoadCommandlineConfig(homeFolder string, pluginID string, pluginConfig interface{}) (*HubConfig, error) {
	hubConfig, err := LoadHubConfig(homeFolder, pluginID)
	if err != nil {
		logrus.Errorf("LoadCommandlineConfig: Failed loading Hub configuration: %s", err)
		return hubConfig, err
	}
	err = LoadPluginConfig(hubConfig.ConfigFolder, pluginID, pluginConfig, nil)
	if err != nil {
		logrus.Errorf("LoadCommandlineConfig: Failed loading configuration for plugin '%s': %s", pluginID, err)
		return hubConfig, err
	}

	SetHubCommandlineArgs(hubConfig)

	// catch parsing errors, in case flag.ErrorHandling = flag.ContinueOnError
	err = flag.CommandLine.Parse(os.Args[1:])

	if err != nil {
		return hubConfig, err
	}

	// validation pass in case commandline argument messed up the config
	err = ValidateHubConfig(hubConfig)

	// It is up to the app to change to the home directory.
	// os.Chdir(hubConfig.HomeFolder)

	// Last set the hub/plugin logging
	if pluginID != "" {
		logFolder := path.Dir(hubConfig.LogFile)
		logFileName := path.Join(logFolder, pluginID+".log")
		SetLogging(hubConfig.Loglevel, logFileName)
	} else if hubConfig.LogFile != "" {
		SetLogging(hubConfig.Loglevel, hubConfig.LogFile)
	}
	return hubConfig, err
}
