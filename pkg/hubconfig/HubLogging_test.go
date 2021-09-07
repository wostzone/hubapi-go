package hubconfig_test

import (
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hubserve-go/pkg/hubconfig"
)

func TestLogging(t *testing.T) {
	wd, _ := os.Getwd()
	logFile := path.Join(wd, "../../test/logs/TestLogging.log")

	os.Remove(logFile)
	hubconfig.SetLogging("info", logFile)
	logrus.Info("Hello info")
	hubconfig.SetLogging("debug", logFile)
	logrus.Debug("Hello debug")
	hubconfig.SetLogging("warn", logFile)
	logrus.Warn("Hello warn")
	hubconfig.SetLogging("error", logFile)
	logrus.Error("Hello error")
	assert.FileExists(t, logFile)
	os.Remove(logFile)
}

func TestLoggingBadFile(t *testing.T) {
	logFile := "/root/cantloghere.log"

	err := hubconfig.SetLogging("info", logFile)
	assert.Error(t, err)
	os.Remove(logFile)
}
