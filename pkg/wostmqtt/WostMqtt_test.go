package wostmqtt_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hubapi/pkg/td"
	"github.com/wostzone/hubapi/pkg/wostmqtt"
)

// const caCertFile = "/etc/mosquitto/certs/zcas_ca.crt"
const clientID2 = "client2"

// !!! THIS REQUIRES A RUNNING MQTT SERVER ON LOCALHOST !!!

func TestPublishTD(t *testing.T) {
	credentials := ""
	thingID := "thing1"
	td1 := td.CreateTD(thingID)
	var rxTd td.DynamicThingDescription
	var clientID1 = fmt.Sprintf("client1-%d", rand.Int())

	thingClient := wostmqtt.NewThingClient(mqttServerHostPort, caCertFile, clientID1, credentials)
	err := thingClient.Start(false)
	assert.NoError(t, err)
	consumerClient := wostmqtt.NewConsumerClient(mqttServerHostPort, caCertFile, clientID2, credentials)
	err = consumerClient.Start(false)
	assert.NoError(t, err)
	consumerClient.SubscribeTD(thingID, func(thingID string, thing map[string]interface{}, sender string) {
		logrus.Infof("TestPublishTD: Received TD of Thing %s from client %s", thingID, sender)
		rxTd = thing
	})
	time.Sleep(time.Millisecond)

	err = thingClient.PublishTD(thingID, td1)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond)

	assert.Equal(t, td1["ID"], rxTd["id"])

	// TODO, check if it was received by a consumer using a consumer client
	thingClient.Stop()
	consumerClient.Stop()
}

func TestPublishPropertyValues(t *testing.T) {
	credentials := ""
	thingID := "thing1"
	var clientID1 = fmt.Sprintf("client1-%d", rand.Int())
	propValues := map[string]interface{}{"propname": "value"}
	var rx map[string]interface{}

	thingClient := wostmqtt.NewThingClient(mqttServerHostPort, caCertFile, clientID1, credentials)
	err := thingClient.Start(false)
	assert.NoError(t, err)
	consumerClient := wostmqtt.NewConsumerClient(mqttServerHostPort, caCertFile, clientID2, credentials)
	err = consumerClient.Start(false)
	assert.NoError(t, err)
	consumerClient.SubscribePropertyValues(thingID, func(thingID string, values map[string]interface{}, sender string) {
		logrus.Infof("TestPublishPropertyValues: Received values of Thing %s from client %s", thingID, sender)
		rx = values
	})

	time.Sleep(time.Millisecond)
	err = thingClient.PublishPropertyValues(thingID, propValues)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 2)
	assert.Equal(t, propValues["propname"], rx["propname"])

	thingClient.Stop()
	consumerClient.Stop()
}

func TestPublishEvent(t *testing.T) {
	credentials := ""
	thingID := "thing1"
	var clientID1 = fmt.Sprintf("client1-%d", rand.Int())
	event1 := map[string]interface{}{"eventName": "eventValue"}
	var rx map[string]interface{}

	consumerClient := wostmqtt.NewConsumerClient(mqttServerHostPort, caCertFile, clientID1, credentials)
	thingClient := wostmqtt.NewThingClient(mqttServerHostPort, caCertFile, clientID2, credentials)

	err := thingClient.Start(false)
	assert.NoError(t, err)
	err = consumerClient.Start(false)
	assert.NoError(t, err)
	consumerClient.SubscribeEvent(thingID, func(thingID string, event map[string]interface{}, sender string) {
		logrus.Infof("TestPublishEvent: Received event of Thing %s from client %s", thingID, sender)
		rx = event
	})

	time.Sleep(time.Millisecond)
	err = thingClient.PublishEvent(thingID, event1)
	assert.NoError(t, err)

	// TODO, check if it was received by a consumer

	time.Sleep(time.Millisecond)
	assert.Equal(t, event1["eventName"], rx["eventName"])

	thingClient.Stop()
	consumerClient.Stop()
}

func TestPublishAction(t *testing.T) {
	credentials := ""
	thingID := "thing1"
	var clientID1 = fmt.Sprintf("client1-%d", rand.Int())
	var rx map[string]interface{}

	action1 := map[string]interface{}{"actionName": "actionValue"}
	consumerClient := wostmqtt.NewConsumerClient(mqttServerHostPort, caCertFile, clientID1, credentials)
	thingClient := wostmqtt.NewThingClient(mqttServerHostPort, caCertFile, clientID2, credentials)
	thingClient.SubscribeToAction(thingID, func(thingID string, action map[string]interface{}, sender string) {
		logrus.Infof("TestPublishAction: Received action of Thing %s from client %s", thingID, sender)
		rx = action
	})

	err := consumerClient.Start(false)
	err = thingClient.Start(false)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond)

	err = consumerClient.PublishAction(thingID, action1)
	assert.NoError(t, err)

	// TODO, check if it was received by the Thing
	time.Sleep(time.Millisecond)
	assert.Equal(t, action1["actionName"], rx["actionName"])

	thingClient.Stop()
	consumerClient.Stop()
}

func TestPublishConfig(t *testing.T) {
	credentials := ""
	thingID := "thing1"
	var clientID1 = fmt.Sprintf("client1-%d", rand.Int())
	var rx map[string]interface{}

	config1 := map[string]interface{}{"prop1": "value1"}
	consumerClient := wostmqtt.NewConsumerClient(mqttServerHostPort, caCertFile, clientID1, credentials)
	thingClient := wostmqtt.NewThingClient(mqttServerHostPort, caCertFile, clientID2, credentials)
	thingClient.SubscribeToConfig(thingID, func(thingID string, config map[string]interface{}, sender string) {
		logrus.Infof("TestPublishConfig: Received config of Thing %s from client %s", thingID, sender)
		rx = config
	})

	err := consumerClient.Start(false)
	err = thingClient.Start(false)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond)

	err = consumerClient.PublishConfig(thingID, config1)
	assert.NoError(t, err)

	// TODO, check if it was received by the Thing
	time.Sleep(time.Millisecond)
	assert.Equal(t, config1["prop1"], rx["prop1"])

	thingClient.Stop()
	consumerClient.Stop()
}
