package wostmqtt_test

import (
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hubapi/pkg/wostmqtt"
)

const mqttServerHostPort = "localhost:8883"
const caCertFile = "/etc/mosquitto/certs/mqtt_srv.crt"
const clientID = "client5"
const TEST_TOPIC = "test"

// These tests require an MQTT TLS server on localhost with TLS support

func TestMqttConnect(t *testing.T) {
	client := wostmqtt.NewMqttClient(mqttServerHostPort, caCertFile)
	const timeout = 10
	err := client.Connect(clientID, timeout)
	assert.NoError(t, err)
	// reconnect
	err = client.Connect(clientID, timeout)
	assert.NoError(t, err)
	client.Disconnect()
}

func TestMqttNoConnect(t *testing.T) {
	invalidHost := "nohost:1111"
	client := wostmqtt.NewMqttClient(invalidHost, caCertFile)
	timeout := 5
	require.NotNil(t, client)
	err := client.Connect(clientID, timeout)
	assert.Error(t, err)
	client.Disconnect()
}

func TestMQTTPubSub(t *testing.T) {
	var rx string
	rxMutex := sync.Mutex{}
	var msg1 = "Hello world"
	// clientID := "test"
	const timeout = 10
	// certFolder := ""

	client := wostmqtt.NewMqttClient(mqttServerHostPort, caCertFile)
	err := client.Connect(clientID, timeout)
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
	client := wostmqtt.NewMqttClient(mqttServerHostPort, caCertFile)
	var rx1 string
	var rx2 string
	rxMutex := sync.Mutex{}
	var msg1 = "Hello world 1"
	var msg2 = "Hello world 2"
	// clientID := "test"
	const timeout = 10

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	err := client.Connect(clientID, timeout)
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
	assert.Equalf(t, "", rx1, "Did not receive the message on handler 1")
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

	client.Unsubscribe(TEST_TOPIC)
	err = client.Publish(TEST_TOPIC, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")
	rxMutex.Unlock()

	// when unsubscribing without handler, all handlers should be unsubscribed
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
	client := wostmqtt.NewMqttClient(mqttServerHostPort, caCertFile)
	const timeout = 10

	err := client.Connect(clientID, timeout)
	require.NoError(t, err)

	client.Unsubscribe(TEST_TOPIC)
	client.Disconnect()
}

func TestMQTTPubNoConnect(t *testing.T) {
	invalidHost := "localhost:1111"
	client := wostmqtt.NewMqttClient(invalidHost, caCertFile)
	const timeout = 10
	var msg1 = "Hello world 1"

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	err := client.Publish(TEST_TOPIC, []byte(msg1))
	require.Error(t, err)

	client.Disconnect()
}

func TestMQTTSubBeforeConnect(t *testing.T) {
	client := wostmqtt.NewMqttClient(mqttServerHostPort, caCertFile)
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

	err := client.Connect(clientID, timeout)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	err = client.Publish(TEST_TOPIC, []byte(msg))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	rxMutex.Lock()
	assert.Equal(t, msg, rx)
	rxMutex.Unlock()

	client.Disconnect()
}
