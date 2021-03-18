package wostmqtt

import (
	"encoding/json"
	"strings"

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
	certFolder         string
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
func (client *MqttConsumerClient) PublishAction(thingID string, action interface{}) error {
	topic := strings.ReplaceAll(api.TopicAction, "{id}", thingID)
	message, err := json.Marshal(action)
	if err != nil {
		return err
	}
	err = client.mqttClient.Publish(topic, message)
	return err
}

// PublishConfig publish a Thing configuration request to the Hub
func (client *MqttConsumerClient) PublishConfig(thingID string, values interface{}) error {
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
	thingID string, handler func(thingID string, td interface{}, senderID string)) {
	topic := strings.ReplaceAll(api.TopicThingTD, "{id}", thingID)
	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		sender := "" // FIXME
		td := make(map[string]interface{})
		err := json.Unmarshal(message, td)
		if err == nil {
			subscribedHandler(subscribedThingID, td, sender)
		}
	})
}

// SubscribePropertyValues receives updates to Thing property values from the WoST Hub
func (client *MqttConsumerClient) SubscribePropertyValues(
	thingID string, handler func(thingID string, values interface{}, senderID string)) {
	topic := strings.ReplaceAll(api.TopicThingPropertyValues, "{id}", thingID)

	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		sender := "" // FIXME
		values := make(map[string]interface{})
		err := json.Unmarshal(message, values)
		if err == nil {
			subscribedHandler(subscribedThingID, values, sender)
		}
	})
}

// SubscribeEvents receives Thing events from the WoST hub.
func (client *MqttConsumerClient) SubscribeEvent(
	thingID string, handler func(thingID string, event interface{}, senderID string)) {
	topic := strings.ReplaceAll(api.TopicThingEvent, "{id}", thingID)

	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		sender := "" // FIXME
		event := make(map[string]interface{})
		err := json.Unmarshal(message, event)
		if err == nil {
			subscribedHandler(subscribedThingID, event, sender)
		}
	})
}

// Create a new instance of the WoST MQTT client
// This implements the Api interface
func NewConsumerClient(certFolder string) api.IConsumerClient {
	client := &MqttConsumerClient{
		certFolder: certFolder,
		timeoutSec: 3,
	}
	return client
}
