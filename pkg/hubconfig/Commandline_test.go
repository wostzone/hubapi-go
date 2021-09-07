package hubconfig_test

import (
	"flag"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hubserve-go/pkg/hubconfig"
)

// CustomConfig as example of how to extend the hub configuration
type CustomConfig struct {
	ExtraVariable string
}

var homeFolder string
var customConfig *CustomConfig

// Use the test folder during testing
func setup() {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	customConfig = &CustomConfig{}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	// os.Args = append(os.Args[0:1], strings.Split("", " ")...)
	// hubConfig, _ = hubconfig.SetupConfig(homeFolder, pluginID, customConfig)
}
func teardown() {
}
func TestSetupHubCommandline(t *testing.T) {
	setup()

	myArgs := strings.Split("--mqttAddress bob --logsFolder logs --logLevel debug", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig := hubconfig.CreateDefaultHubConfig(homeFolder)
	hubconfig.SetHubCommandlineArgs(hubConfig)
	// hubConfig, err := hubconfig.SetupConfig("", nil)

	flag.Parse()
	// assert.NoError(t, err)
	assert.Equal(t, "bob", hubConfig.MqttAddress)
	assert.Equal(t, "logs", hubConfig.LogsFolder)
	assert.Equal(t, "debug", hubConfig.Loglevel)
	// assert.Equal(t, "/etc/cert", hubConfig.Messenger.CertFolder)
}

func TestCommandlineWithError(t *testing.T) {
	setup()
	myArgs := strings.Split("--mqttAddress bob --badarg=bad", " ")
	// myArgs := strings.Split("--address bob", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	pluginConfig := make(map[string]interface{})
	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, "plugin1", &pluginConfig)
	// hubConfig.ConfigFolder = path.Join(homeFolder, "test")

	assert.Error(t, err, "Parse flag -badarg should fail")
	assert.Equal(t, "bob", hubConfig.MqttAddress)
	teardown()
}

// Test setup with extra commandline flag '--extra'
func TestSetupHubCommandlineWithExtendedConfig(t *testing.T) {
	setup()

	myArgs := strings.Split("-c ./config/hub.yaml --home ../../test --mqttAddress bob --extra value1", " ")
	// Remove testing package commandline arguments so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	// hubConfig := hubconfig.CreateDefaultHubConfig("")
	pluginConfig := CustomConfig{}
	// hubconfig.HubConfig = *hubConfig

	// hubconfig.SetHubCommandlineArgs(&config.HubConfig)
	flag.StringVar(&pluginConfig.ExtraVariable, "extra", "", "Extended extra configuration")

	// err := hubconfig.ParseCommandline(myArgs, &config)
	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, "plugin1", pluginConfig)

	assert.NoError(t, err)
	assert.Equal(t, "bob", hubConfig.MqttAddress)
	assert.Equal(t, "value1", pluginConfig.ExtraVariable)
}

// Test with a custom and bad config file
func TestSetupConfigBadConfigfile(t *testing.T) {
	setup()
	// The default directory is the project folder
	myArgs := strings.Split("-c ./config/hub-bad.yaml", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, "plugin1", nil)
	assert.Error(t, err)
	assert.Equal(t, "yaml: line 12", err.Error()[0:13], "Expected yaml parse error")
	assert.NotNil(t, hubConfig)
}

// Test with an invalid config file
func TestSetupConfigInvalidConfigfile(t *testing.T) {
	setup()
	myArgs := strings.Split("-c ./config/hub-invalid.yaml", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, "plugin1", nil)
	assert.Equal(t, "debug", hubConfig.Loglevel, "config file wasn't loaded")
	assert.Error(t, err, "Expected validation of config to fail")
	assert.NotNil(t, hubConfig)
}

func TestCommandlineMissingPluginID(t *testing.T) {
	setup()

	_, err := hubconfig.LoadCommandlineConfig(homeFolder, "", nil)
	assert.Error(t, err)
}

// TestSetupConfigNoConfig checks that setup still works if the plugin config doesn't exist
func TestSetupConfigNoConfig(t *testing.T) {
	setup()
	myArgs := strings.Split("", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	pluginConfig := CustomConfig{}
	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, "notaconfigfile", pluginConfig)
	assert.NoError(t, err)
	assert.NotNil(t, hubConfig)
}

func TestSetupLogging(t *testing.T) {
	setup()
	myArgs := strings.Split("--logLevel debug", " ")
	// Remove testing package created commandline and flags so we can test ours
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], myArgs...)

	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, "myplugin", nil)
	assert.NoError(t, err)
	require.NotNil(t, hubConfig)
	assert.Equal(t, "debug", hubConfig.Loglevel)
}
