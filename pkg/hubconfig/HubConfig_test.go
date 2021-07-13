package hubconfig_test

import (
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

type ConfigType1 struct {
	C1 string
	c2 string
}

func TestDefaultConfigNoHome(t *testing.T) {
	// This result is unpredictable as it depends on where the binary lives.
	// This changes depends on whether to run as debug, coverage or F5 run
	hc := hubconfig.CreateDefaultHubConfig("")
	require.NotNil(t, hc)
	err := hubconfig.ValidateHubConfig(hc)
	_ = err // unpredictable outcome
	// assert.NoError(t, err)
	hc = hubconfig.CreateDefaultHubConfig("./")
	require.NotNil(t, hc)
	_ = err // unpredictable outcome
	// assert.NoError(t, err)

}
func TestDefaultConfigWithHome(t *testing.T) {
	// vscode debug and test runs use different binary folder.
	// Use current dir instead to determine where home is.
	wd, _ := os.Getwd()
	home := path.Join(wd, "../../test")
	logrus.Infof("TestDefaultConfig: Current folder is %s", wd)
	hc := hubconfig.CreateDefaultHubConfig(home)
	require.NotNil(t, hc)
	err := hubconfig.ValidateHubConfig(hc)
	assert.NoError(t, err)
}

func TestLoadHubConfig(t *testing.T) {
	wd, _ := os.Getwd()
	homeFolder = path.Join(wd, "../../test")
	// hc := hubconfig.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	// require.NotNil(t, hc)

	// configFile := path.Join(hc.ConfigFolder, "hub.yaml")
	// err := hubconfig.LoadHubConfig(configFile, hc)
	hc, err := hubconfig.LoadHubConfig(homeFolder, "plugin1")
	assert.NoError(t, err)
	err = hubconfig.ValidateHubConfig(hc)
	assert.NoError(t, err)
	assert.Equal(t, "info", hc.Loglevel)
}

func TestSubstitute(t *testing.T) {
	substMap := make(map[string]string, 0)
	substMap["pluginID"] = "plugin1"
	hc := hubconfig.HubConfig{}
	wd, _ := os.Getwd()
	templateFile := path.Join(wd, "../../test/config/hub-template.yaml")
	err := hubconfig.LoadConfig(templateFile, &hc, substMap)
	assert.NoError(t, err)
	// from the template file
	assert.Equal(t, "/var/log/plugin1.log", hc.LogFile)
}

func TestLoadHubConfigNotFound(t *testing.T) {
	wd, _ := os.Getwd()
	hc := hubconfig.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	require.NotNil(t, hc)
	configFile := path.Join(hc.ConfigFolder, "hub-notfound.yaml")
	err := hubconfig.LoadConfig(configFile, hc, nil)
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadHubConfigYamlError(t *testing.T) {
	wd, _ := os.Getwd()
	hc := hubconfig.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	require.NotNil(t, hc)

	configFile := path.Join(hc.ConfigFolder, "hub-bad.yaml")
	err := hubconfig.LoadConfig(configFile, hc, nil)
	// Error should contain info on bad file
	errTxt := err.Error()
	assert.Equal(t, "yaml: line 12", errTxt[:13], "Expected line 12 to be bad")
	assert.Error(t, err, "Configfile should not be found")
}

func TestLoadHubConfigBadFolders(t *testing.T) {

	wd, _ := os.Getwd()
	hc := hubconfig.CreateDefaultHubConfig(path.Join(wd, "../../test"))
	err := hubconfig.ValidateHubConfig(hc)
	assert.NoError(t, err, "Default config should be okay")

	gc2 := *hc
	gc2.Home = "/not/a/home/folder"
	err = hubconfig.ValidateHubConfig(&gc2)
	assert.Error(t, err)
	gc2 = *hc
	gc2.ConfigFolder = "./doesntexist"
	err = hubconfig.ValidateHubConfig(&gc2)
	assert.Error(t, err)
	gc2 = *hc
	gc2.LogFile = "/this/path/doesntexist"
	err = hubconfig.ValidateHubConfig(&gc2)
	assert.Error(t, err)
	gc2 = *hc
	gc2.CertsFolder = "./doesntexist"
	err = hubconfig.ValidateHubConfig(&gc2)
	assert.Error(t, err)
	gc2 = *hc
	// gc2.PluginFolder = "./doesntexist"
	// err = hubconfig.ValidateConfig(&gc2)
	// assert.Error(t, err)
}
