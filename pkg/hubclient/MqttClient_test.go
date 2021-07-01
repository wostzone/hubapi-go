package hubclient_test

import (
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/testenv"
)

// Use test/mosquitto-test.conf
const mqttTestPluginConnection = "localhost:33100"

var mqttTestCaCertFile string
var mqttTestCaKeyFile string
var mqttTestClientCertFile string
var mqttTestClientKeyFile string

const TEST_TOPIC = "test"

// For running mosquitto in test
const mosquittoConfigFile = "mosquitto-test.conf"
const testPluginID = "test-user"

// var mosquittoCmd *exec.Cmd

// setup - launch mosquitto
func TestMain(m *testing.M) {
	hostnames := []string{"localhost"}
	cwd, _ := os.Getwd()
	home := path.Join(cwd, "../../test")
	os.Chdir(home)
	certsFolder := path.Join(home, "certs")

	mqttTestCaCertFile = path.Join(certsFolder, certsetup.CaCertFile)
	mqttTestCaKeyFile = path.Join(certsFolder, certsetup.CaKeyFile)
	mqttTestClientCertFile = path.Join(certsFolder, certsetup.PluginCertFile)
	mqttTestClientKeyFile = path.Join(certsFolder, certsetup.PluginKeyFile)

	configFolder := path.Join(home, "config")
	certsetup.CreateCertificateBundle(hostnames, certsFolder)
	mosqConfigPath := path.Join(configFolder, mosquittoConfigFile)

	mosquittoCmd, err := testenv.Launch(mosqConfigPath)
	if err != nil {
		logrus.Fatalf("Unable to setup mosquitto: %s", err)
	}

	result := m.Run()
	mosquittoCmd.Process.Kill()

	os.Exit(result)
}

func TestMqttConnect(t *testing.T) {
	logrus.Infof("--- TestMqttConnect ---")

	client := hubclient.NewMqttClient(mqttTestPluginConnection, mqttTestCaCertFile)
	client.SetTimeout(10)
	err := client.ConnectWithClientCert(testPluginID, mqttTestClientCertFile, mqttTestClientKeyFile)
	assert.NoError(t, err)
	// reconnect
	err = client.ConnectWithClientCert(testPluginID, mqttTestClientCertFile, mqttTestClientKeyFile)
	assert.NoError(t, err)
	client.Disconnect()
}

func TestMqttNoConnect(t *testing.T) {
	logrus.Infof("--- TestMqttNoConnect ---")

	invalidHost := "nohost:1111"
	client := hubclient.NewMqttClient(invalidHost, mqttTestCaCertFile)
	client.SetTimeout(5)
	require.NotNil(t, client)
	err := client.ConnectWithClientCert(testPluginID, mqttTestClientCertFile, mqttTestClientKeyFile)
	assert.Error(t, err)
	client.Disconnect()
}

func TestMQTTPubSub(t *testing.T) {
	logrus.Infof("--- TestMQTTPubSub ---")

	var rx string
	rxMutex := sync.Mutex{}
	var msg1 = "Hello world"
	// clientID := "test"
	const timeout = 10
	// certFolder := ""

	client := hubclient.NewMqttClient(mqttTestPluginConnection, mqttTestCaCertFile)
	client.SetTimeout(5)
	err := client.ConnectWithClientCert(testPluginID, mqttTestClientCertFile, mqttTestClientKeyFile)
	require.NoError(t, err)

	client.Subscribe(TEST_TOPIC, func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx = string(msg)
		logrus.Infof("Received message: %s", msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", TEST_TOPIC)

	err = client.Publish(TEST_TOPIC, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")

	// allow time to receive
	time.Sleep(1000 * time.Millisecond)
	rxMutex.Lock()
	defer rxMutex.Unlock()
	require.Equalf(t, msg1, rx, "Did not receive the message")

	client.Disconnect()
}

func TestMQTTMultipleSubscriptions(t *testing.T) {
	logrus.Infof("--- TestMQTTMultipleSubscriptions ---")

	client := hubclient.NewMqttClient(mqttTestPluginConnection, mqttTestCaCertFile)
	var rx1 string
	var rx2 string
	rxMutex := sync.Mutex{}
	var msg1 = "Hello world 1"
	var msg2 = "Hello world 2"
	// clientID := "test"
	const timeout = 10

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	client.SetTimeout(5)
	err := client.ConnectWithClientCert(testPluginID, mqttTestClientCertFile, mqttTestClientKeyFile)
	require.NoError(t, err)
	handler1 := func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx1 = string(msg)
		logrus.Infof("Received message on handler 1: %s", msg)
	}
	handler2 := func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx2 = string(msg)
		logrus.Infof("Received message on handler 2: %s", msg)
	}
	_ = handler2
	client.Subscribe(TEST_TOPIC, handler1)
	client.Subscribe(TEST_TOPIC, handler2)
	err = client.Publish(TEST_TOPIC, []byte(msg1))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	// tbd
	assert.Equalf(t, "", rx1, "Did not expect a message on handler 1")
	assert.Equalf(t, msg1, rx2, "Did not receive the message on handler 2")
	// after unsubscribe no message should be received by handler 1
	rx1 = ""
	rx2 = ""
	rxMutex.Unlock()
	client.Unsubscribe(TEST_TOPIC)
	err = client.Publish(TEST_TOPIC, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Received a message on handler 2 after unsubscribe")
	rx1 = ""
	rx2 = ""
	rxMutex.Unlock()

	client.Subscribe(TEST_TOPIC, handler1)
	err = client.Publish(TEST_TOPIC, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, msg2, rx1, "Did not receive a message on handler 1 after subscribe")
	assert.Equalf(t, "", rx2, "Receive the message on handler 2")
	rxMutex.Unlock()

	// when unsubscribing without handler, all handlers should be unsubscribed
	rx1 = ""
	rx2 = ""
	client.Subscribe(TEST_TOPIC, handler1)
	client.Subscribe(TEST_TOPIC, handler2)
	client.Unsubscribe(TEST_TOPIC)
	err = client.Publish(TEST_TOPIC, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")
	rxMutex.Unlock()

	client.Disconnect()
}

func TestMQTTBadUnsubscribe(t *testing.T) {
	logrus.Infof("--- TestMQTTBadUnsubscribe ---")

	client := hubclient.NewMqttClient(mqttTestPluginConnection, mqttTestCaCertFile)
	client.SetTimeout(10)
	err := client.ConnectWithClientCert(testPluginID, mqttTestClientCertFile, mqttTestClientKeyFile)
	require.NoError(t, err)

	client.Unsubscribe(TEST_TOPIC)
	client.Disconnect()
}

func TestMQTTPubNoConnect(t *testing.T) {
	logrus.Infof("--- TestMQTTPubNoConnect ---")

	invalidHost := "localhost:1111"
	client := hubclient.NewMqttClient(invalidHost, mqttTestCaCertFile)
	var msg1 = "Hello world 1"

	err := client.Publish(TEST_TOPIC, []byte(msg1))
	require.Error(t, err)

	client.Disconnect()
}

func TestMQTTSubBeforeConnect(t *testing.T) {
	logrus.Infof("--- TestMQTTSubBeforeConnect ---")

	client := hubclient.NewMqttClient(mqttTestPluginConnection, mqttTestCaCertFile)
	const timeout = 10
	const msg = "hello 1"
	var rx string
	rxMutex := sync.Mutex{}
	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("Received message on handler 1: %s", msg)
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx = string(msg)
	}
	client.Subscribe(TEST_TOPIC, handler1)

	client.SetTimeout(timeout)
	err := client.ConnectWithClientCert(testPluginID, mqttTestClientCertFile, mqttTestClientKeyFile)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	err = client.Publish(TEST_TOPIC, []byte(msg))
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equal(t, msg, rx)
	rxMutex.Unlock()

	client.Disconnect()
}

func TestSubscribeWildcard(t *testing.T) {
	logrus.Infof("--- TestSubscribeWildcard ---")
	const testTopic1 = "test/1/5"
	const wildcardTopic = "test/+/#"

	client := hubclient.NewMqttClient(mqttTestPluginConnection, mqttTestCaCertFile)
	const timeout = 10
	const msg = "hello 1"
	var rx string
	rxMutex := sync.Mutex{}
	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("Received message on handler 1: %s", msg)
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx = string(msg)
	}
	client.Subscribe(wildcardTopic, handler1)

	// connect after subscribe uses resubscribe
	client.SetTimeout(timeout)
	err := client.ConnectWithClientCert(testPluginID, mqttTestClientCertFile, mqttTestClientKeyFile)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	err = client.Publish(testTopic1, []byte(msg))
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equal(t, msg, rx)
	rxMutex.Unlock()

	client.Disconnect()
}
