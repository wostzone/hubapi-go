package wostmqtt

import (
	"strings"

	"github.com/wostzone/api/wostapi"
)

/* Client library with the MQTT API to the Hub using (tbd):
A: paho-mqtt
B: https://github.com/emqx/emqx
*/

type WostMqttClient struct {
	mqttClient *MqttClient
	certFolder string
	timeoutSec int
}

// Start the client connection
func (wmc *WostMqttClient) Start(hostname string, clientID string) error {
	wmc.mqttClient = NewMqttClient(wmc.certFolder, hostname)
	err := wmc.mqttClient.Connect(clientID, wmc.timeoutSec)

	return err
}

// End the client connection
func (wmc *WostMqttClient) Stop() {
	wmc.mqttClient.Disconnect()
}

// PublishTD publish a Thing description to the WoST hub
// Intended to by used by a Thing to publish an update to its TD
func (wmc *WostMqttClient) PublishTD(thingID string, td []byte) error {
	topic := strings.ReplaceAll(wostapi.TopicThingTD, "{id}", thingID)
	err := wmc.mqttClient.Publish(topic, td)
	return err
}

// PublishPropertyValues publish a Thing property values to the WoST hub
// Intended to by used by a Thing to publish updates of property values
func (wmc *WostMqttClient) PublishPropertyValues(thingID string, values []byte) error {
	topic := strings.ReplaceAll(wostapi.TopicThingPropertyValues, "{id}", thingID)
	err := wmc.mqttClient.Publish(topic, values)
	return err
}

// PublishEvent publish a Thing event to the WoST hub
// Intended to by used by a Thing
func (wmc *WostMqttClient) PublishEvent(thingID string, event []byte) error {
	topic := strings.ReplaceAll(wostapi.TopicThingEvent, "{id}", thingID)
	err := wmc.mqttClient.Publish(topic, event)
	return err
}

// PublishAction publish a Thing action to the WoST hub
// Intended to be used by consumers to request an action from a Thing
// Whether the action takes place depends on the user's permissions to publish
// this action.
func (wmc *WostMqttClient) PublishAction(thingID string, action []byte) error {
	topic := strings.ReplaceAll(wostapi.TopicThingEvent, "{id}", thingID)
	err := wmc.mqttClient.Publish(topic, action)
	return err
}

// Create a new instance of the WoST MQTT client
// This implements the WostAPI interface
func NewWostMqtt(certFolder string) *WostMqttClient {
	wm := &WostMqttClient{
		certFolder: certFolder,
		timeoutSec: 3,
	}
	return wm
}
