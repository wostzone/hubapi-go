package mqttclient_test

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hubapi-go/api"
	"github.com/wostzone/hubapi-go/internal/mqttclient"
	"github.com/wostzone/hubapi-go/pkg/certsetup"
	"github.com/wostzone/hubapi-go/pkg/td"
)

const zone = "test"
const mqttTestConsumerConnection = "localhost:33101"
const mqttTestThingConnection = "localhost:33101"

// THIS USES THE SETUP IN MqttClient_test.go

// Custom test to production
// func TestPublishCustom(t *testing.T) {
// 	logrus.Infof("--- TestPublishCustom ---")
// 	thingID := "urn:TestPublishCustom"

// 	// thingClient := mqttclient.NewMqttHubClient("localhost:8883", "/home/henk/bin/wost/certs/ca.crt", "", "")
// 	thingClient := mqttclient.NewMqttHubPluginClient("henksplugin",
// 		"localhost:9883", "/home/henk/bin/wost/certs/ca.crt",
// 		"/home/henk/bin/wost/certs/client.crt", "/home/henk/bin/wost/certs/client.key")

// 	err := thingClient.Start()
// 	assert.NoError(t, err)
// 	thingTD := td.CreateTD(thingID, api.DeviceTypeService)
// 	thingClient.PublishTD(thingID, thingTD)
// 	time.Sleep(time.Second)
// 	thingClient.Stop()
// }

func TestPublishAction(t *testing.T) {
	logrus.Infof("--- TestPublishAction ---")

	thingID := "thing1"
	var rxName string
	var rxParams map[string]interface{}
	actionName := "action1"
	actionInput := map[string]interface{}{"input1": "inputValue"}
	consumerClient := mqttclient.NewMqttHubClient(mqttTestConsumerConnection, mqttTestCaCertFile, "", "")
	thingClient := mqttclient.NewMqttHubClient(mqttTestThingConnection, mqttTestCaCertFile, "", "")
	thingClient.SubscribeToActions(thingID, func(thingID string, name string, params map[string]interface{}, sender string) {
		logrus.Infof("TestPublishAction: Received action of Thing %s from client %s", thingID, sender)
		rxName = name
		rxParams = params
	})

	err := consumerClient.Start()
	err = thingClient.Start()
	assert.NoError(t, err)

	time.Sleep(time.Millisecond)

	err = consumerClient.PublishAction(thingID, actionName, actionInput)
	assert.NoError(t, err)

	// TODO, check if it was received by the Thing
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, actionName, rxName)
	assert.Equal(t, actionInput, rxParams)

	thingClient.Stop()
	consumerClient.Stop()
	// make sure it doest reconnect
	time.Sleep(1 * time.Second)
}

func TestPublishConfig(t *testing.T) {
	logrus.Infof("--- TestPublishConfig ---")

	credentials := ""
	thingID := "thing1"
	var rx map[string]interface{}
	var rxID string

	config1 := map[string]interface{}{"prop1": "value1"}
	consumerClient := mqttclient.NewMqttHubClient(mqttTestConsumerConnection, mqttTestCaCertFile, "", credentials)
	thingClient := mqttclient.NewMqttHubClient(mqttTestThingConnection, mqttTestCaCertFile, "", credentials)
	thingClient.SubscribeToConfig(thingID, func(thingID string, config map[string]interface{}, sender string) {
		logrus.Infof("TestPublishConfig: Received config of Thing %s from client %s", thingID, sender)
		rx = config
		rxID = thingID
	})

	err := consumerClient.Start()
	err = thingClient.Start()
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = consumerClient.PublishConfigRequest(thingID, config1)
	assert.NoError(t, err)

	// TODO, check if it was received by the Thing
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, config1["prop1"], rx["prop1"])
	assert.Equal(t, thingID, rxID)
	thingClient.Stop()
	consumerClient.Stop()
}

func TestPublishEvent(t *testing.T) {
	logrus.Infof("--- TestPublishEvent ---")

	credentials := ""
	thingID := "thing1"
	event1 := map[string]interface{}{"eventName": "eventValue"}
	var rx map[string]interface{}

	consumerClient := mqttclient.NewMqttHubClient(mqttTestConsumerConnection, mqttTestCaCertFile, "", credentials)
	thingClient := mqttclient.NewMqttHubClient(mqttTestThingConnection, mqttTestCaCertFile, "", credentials)

	err := thingClient.Start()
	assert.NoError(t, err)
	err = consumerClient.Start()
	assert.NoError(t, err)
	consumerClient.SubscribeToEvents(thingID, func(thingID string, event map[string]interface{}, sender string) {
		logrus.Infof("TestPublishEvent: Received event of Thing %s from client %s", thingID, sender)
		rx = event
	})

	time.Sleep(time.Millisecond)
	err = thingClient.PublishEvent(thingID, event1)
	assert.NoError(t, err)

	// TODO, check if it was received by a consumer

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, event1["eventName"], rx["eventName"])

	thingClient.Stop()
	consumerClient.Stop()
}

func TestPublishPropertyValues(t *testing.T) {
	logrus.Infof("--- TestPublishPropertyValues ---")
	credentials := ""
	thingID := "thing1"
	propValues := map[string]interface{}{"propname": "value"}
	var rx map[string]interface{}

	thingClient := mqttclient.NewMqttHubClient(mqttTestConsumerConnection, mqttTestCaCertFile, "", credentials)
	err := thingClient.Start()
	assert.NoError(t, err)
	consumerClient := mqttclient.NewMqttHubClient(mqttTestThingConnection, mqttTestCaCertFile, "", credentials)
	err = consumerClient.Start()
	assert.NoError(t, err)
	consumerClient.SubscribeToPropertyValues(thingID, func(thingID string, values map[string]interface{}, sender string) {
		logrus.Infof("TestPublishPropertyValues: Received values of Thing %s from client %s", thingID, sender)
		rx = values
	})

	time.Sleep(time.Millisecond)
	err = thingClient.PublishPropertyValues(thingID, propValues)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, propValues["propname"], rx["propname"])

	thingClient.Stop()
	consumerClient.Stop()
}
func TestPublishTD(t *testing.T) {
	logrus.Infof("--- TestPublishTD ---")
	credentials := ""
	deviceID := "thing1"
	thingID := td.CreateThingID(zone, deviceID, api.DeviceTypeSensor)
	td1 := td.CreateTD(thingID, api.DeviceTypeSensor)
	var rxTd map[string]interface{}

	thingClient := mqttclient.NewMqttHubClient(mqttTestConsumerConnection, mqttTestCaCertFile, "", credentials)
	err := thingClient.Start()
	assert.NoError(t, err)
	consumerClient := mqttclient.NewMqttHubClient(mqttTestThingConnection, mqttTestCaCertFile, "", credentials)
	err = consumerClient.Start()
	assert.NoError(t, err)
	consumerClient.SubscribeToTD(thingID, func(thingID string, thing api.ThingTD, sender string) {
		logrus.Infof("TestPublishTD: Received TD of Thing %s from client %s", thingID, sender)
		rxTd = thing
	})
	time.Sleep(time.Millisecond * 100)

	err = thingClient.PublishTD(thingID, td1)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, td1["id"], rxTd["id"])

	// TODO, check if it was received by a consumer using a consumer client
	thingClient.Stop()
	consumerClient.Stop()
}

// subscribe to all things
func TestSubscribeAll(t *testing.T) {
	logrus.Infof("--- TestSubscribeAll ---")
	credentials := ""
	deviceID := "thing1"
	thingID := td.CreateThingID(zone, deviceID, api.DeviceTypeSensor)
	td1 := td.CreateTD(thingID, api.DeviceTypeSensor)
	txTd, _ := json.Marshal(td1)
	var rxTd []byte
	var rxThing string

	pluginClient := mqttclient.NewMqttHubClient(mqttTestConsumerConnection, mqttTestCaCertFile, "", credentials)
	err := pluginClient.Start()
	assert.NoError(t, err)
	thingClient := mqttclient.NewMqttHubClient(mqttTestThingConnection, mqttTestCaCertFile, "", credentials)
	err = thingClient.Start()
	assert.NoError(t, err)
	pluginClient.Subscribe("", func(thingID string, msgType string, raw []byte, sender string) {
		logrus.Infof("TestSubscribe: Received msg %s of Thing %s from client %s", msgType, thingID, sender)
		rxTd = raw
		rxThing = thingID
	})
	time.Sleep(time.Millisecond * 100)

	err = thingClient.PublishTD(thingID, td1)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, string(txTd), string(rxTd))
	assert.Equal(t, thingID, rxThing)

	// after unsubscribe there should be no more messages
	pluginClient.Unsubscribe("")
	time.Sleep(100 * time.Millisecond)
	err = thingClient.PublishTD(thingID, td1)
	rxTd = nil
	time.Sleep(100 * time.Millisecond)
	assert.NotEqual(t, td1, rxTd)

	// TODO, check if it was received by a consumer using a consumer client
	thingClient.Stop()
	pluginClient.Stop()
}

func TestRequestProvisioning(t *testing.T) {
	pluginID := "plugin1"
	deviceID := "thing1"
	thingID := td.CreateThingID(zone, deviceID, api.DeviceTypeSensor)

	// setup a provisioning server
	pluginClient := mqttclient.NewMqttHubPluginClient(
		pluginID, mqttTestConsumerConnection, mqttTestCaCertFile, mqttTestClientCertFile, mqttTestClientKeyFile)
	err := pluginClient.Start()
	require.NoError(t, err)

	caCertPEM, _ := ioutil.ReadFile(mqttTestCaCertFile)
	caCertBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	caKeyPEM, _ := ioutil.ReadFile(mqttTestCaKeyFile)
	caKeyBlock, _ := pem.Decode(caKeyPEM)
	caPrivKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)

	pluginClient.SubscribeToProvisionRequest(func(thingID string, csrPEM []byte, sender string) {
		certPEM, err := certsetup.SignCertificate(csrPEM, caCert, caPrivKey, time.Second)
		assert.NoError(t, err)
		pluginClient.PublishProvisionResponse(thingID, certPEM)
	})

	// create a provisioning request for a thing
	thingClient := mqttclient.NewMqttHubClient(mqttTestThingConnection, mqttTestCaCertFile, "", "")
	err = thingClient.Start()
	assert.NoError(t, err)

	thingKeyPEM, _ := ioutil.ReadFile(mqttTestClientKeyFile)
	thingKeyBlock, _ := pem.Decode(thingKeyPEM)
	thingPrivKey, err := x509.ParsePKCS1PrivateKey(thingKeyBlock.Bytes)
	csrPEM, err := certsetup.CreateCSR(thingPrivKey, thingID)
	assert.NoError(t, err)
	assert.NotNil(t, csrPEM)
	err = thingClient.PublishProvisionRequest(thingID, csrPEM)
	assert.NoError(t, err)

	thingClient.Stop()
	pluginClient.Stop()
}
