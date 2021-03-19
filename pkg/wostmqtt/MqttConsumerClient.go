package wostmqtt

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/api"
)

/* Client library with the MQTT API to the Hub using (tbd):
A: paho-mqtt
B: https://github.com/emqx/emqx
*/

// MqttConsumerClient is a wrapper around the generic MQTT client with convenience methods for use
// by consumers to subscribe to Thing information and publish configuration and actions.
// This implements the IConsumerClient API
type MqttConsumerClient struct {
	mqttClient         *MqttClient
	clientID           string
	timeoutSec         int
	senderVerification bool
}

// Start the client connection
func (client *MqttConsumerClient) Start(senderVerification bool) error {
	client.senderVerification = senderVerification
	err := client.mqttClient.Connect(client.clientID, client.timeoutSec)
	return err
}

// End the client connection
func (client *MqttConsumerClient) Stop() {
	client.mqttClient.Disconnect()
}

// PublishAction publish a Thing action request to the Hub
func (client *MqttConsumerClient) PublishAction(thingID string, action map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicAction, "{id}", thingID)
	message, err := json.Marshal(action)
	if err != nil {
		return err
	}
	err = client.mqttClient.Publish(topic, message)
	return err
}

// PublishConfig publish a Thing configuration request to the Hub
func (client *MqttConsumerClient) PublishConfig(thingID string, values map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicSetConfig, "{id}", thingID)
	message, err := json.Marshal(values)
	if err != nil {
		return err
	}
	err = client.mqttClient.Publish(topic, message)
	return err
}

// SubscribeTD subscribes to receive updates to TDs from the WoST Hub
func (client *MqttConsumerClient) SubscribeTD(
	thingID string, handler func(thingID string, td map[string]interface{}, senderID string)) {
	topic := strings.ReplaceAll(api.TopicThingTD, "{id}", thingID)
	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		// FIXME: determine sender and format for td message
		sender := ""
		td := make(map[string]interface{})
		err := json.Unmarshal(message, &td)
		if err != nil {
			logrus.Errorf("Received message on topic '%s' but unmarshal failed: %s", topic, err)
		} else {
			subscribedHandler(subscribedThingID, td, sender)
		}
	})
}

// SubscribePropertyValues receives updates to Thing property values from the WoST Hub
func (client *MqttConsumerClient) SubscribePropertyValues(
	thingID string, handler func(thingID string, values map[string]interface{}, senderID string)) {
	topic := strings.ReplaceAll(api.TopicThingPropertyValues, "{id}", thingID)

	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		// FIXME: determine sender and format for property values message
		sender := ""
		values := make(map[string]interface{})
		err := json.Unmarshal(message, &values)
		if err == nil {
			subscribedHandler(subscribedThingID, values, sender)
		}
	})
}

// SubscribeEvents receives Thing events from the WoST hub.
func (client *MqttConsumerClient) SubscribeEvent(
	thingID string, handler func(thingID string, event map[string]interface{}, senderID string)) {
	topic := strings.ReplaceAll(api.TopicThingEvent, "{id}", thingID)

	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		sender := ""
		// FIXME: determine sender and format for event message
		event := make(map[string]interface{})
		err := json.Unmarshal(message, &event)
		if err == nil {
			subscribedHandler(subscribedThingID, event, sender)
		}
	})
}

// Create a new instance of the WoST MQTT client
// This implements the Api interface
func NewConsumerClient(hostPort string, caCertFile string, clientID string, credentials string) api.IConsumerClient {
	client := &MqttConsumerClient{
		timeoutSec: 3,
		clientID:   clientID,
		mqttClient: NewMqttClient(hostPort, caCertFile),
	}
	return client
}
