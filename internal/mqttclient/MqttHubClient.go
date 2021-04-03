package mqttclient

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/api"
)

/* Client library with the MQTT API to the Hub using (tbd):
A: paho-mqtt
B: https://github.com/emqx/emqx
*/

// MqttHubClient is a wrapper around the generic MQTT client with convenience methods for use
// by plugins, Things and consumers to subscribe to Thing information and publish configuration,
// events and actions.
// This implements the IHubConnection API
type MqttHubClient struct {
	mqttClient         *MqttClient
	clientID           string
	timeoutSec         int
	senderVerification bool
}

// Start the client connection
func (client *MqttHubClient) Start(senderVerification bool) error {
	logrus.Infof("Starting MQTT Hub client. Connecting to '%s'. CaCertFile '%s'",
		client.mqttClient.hostPort, client.mqttClient.tlsCACertFile)
	client.senderVerification = senderVerification
	err := client.mqttClient.Connect(client.clientID, client.timeoutSec)
	return err
}

// End the client connection
func (client *MqttHubClient) Stop() {
	client.mqttClient.Disconnect()
}

// PublishAction publish a Thing action request to the Hub
func (client *MqttHubClient) PublishAction(thingID string, name string, params map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicAction, "{id}", thingID)
	actions := map[string]interface{}{name: params}
	message, err := json.Marshal(actions)
	err = client.mqttClient.Publish(topic, message)
	return err
}

// PublishConfig publish a Thing configuration request to the Hub
func (client *MqttHubClient) PublishConfigRequest(thingID string, values map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicSetConfig, "{id}", thingID)
	message, err := json.Marshal(values)
	err = client.mqttClient.Publish(topic, message)
	return err
}

// PublishEvent publish a Thing event to the WoST hub
// Intended to by used by a Thing
func (client *MqttHubClient) PublishEvent(thingID string, event map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicThingEvent, "{id}", thingID)
	message, _ := json.Marshal(event)
	err := client.mqttClient.Publish(topic, message)
	return err
}

// PublishPropertyValues publish a Thing property values to the WoST hub
// Intended to by used by a Thing to publish updates of property values
func (client *MqttHubClient) PublishPropertyValues(thingID string, values map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicThingPropertyValues, "{id}", thingID)
	message, _ := json.Marshal(values)
	err := client.mqttClient.Publish(topic, message)
	return err
}

// PublishTD publish a Thing description to the WoST hub
// Intended to by used by a Thing to publish an update to its TD
func (client *MqttHubClient) PublishTD(thingID string, td api.ThingTD) error {
	topic := strings.ReplaceAll(api.TopicThingTD, "{id}", thingID)
	message, _ := json.Marshal(td)
	err := client.mqttClient.Publish(topic, message)
	return err
}

// Subscribe subscribes to messages from Things
func (client *MqttHubClient) Subscribe(
	thingID string,
	handler func(thingID string, msgType string, raw []byte, senderID string)) {

	if thingID == "" {
		thingID = "+"
	}
	subscribedTopic := fmt.Sprintf("%s/%s/#", api.TopicRoot, thingID)
	subscribedHandler := handler
	client.mqttClient.Subscribe(subscribedTopic, func(topic string, payload []byte) {
		// FIXME: determine sender and format for td message
		sender := ""
		parts := strings.Split(topic, "/")
		if len(parts) > 2 {
			tid := parts[1] // thing ID
			msgType := parts[2]
			subscribedHandler(tid, msgType, payload, sender)
		}
	})
}

// SubscribeToAction subscribes a handler to requested actions.
func (client *MqttHubClient) SubscribeToActions(
	thingID string,
	handler func(thingID string, name string, params map[string]interface{}, senderID string)) {

	topic := strings.ReplaceAll(api.TopicAction, "{id}", thingID)
	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler

	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		// FIXME: determine sender and format for action message
		sender := ""
		actions := make(map[string]interface{})
		err := json.Unmarshal(message, &actions)
		if err == nil {
			for name, params := range actions {
				actionParam := params.(map[string]interface{})
				subscribedHandler(subscribedThingID, name, actionParam, sender)
			}
		} else {
			logrus.Warningf("Message on topic '%s' not JSON", topic)
		}
	})
}

// SubscribeToConfig subscribes a handler to the request for configuration updates.
func (client *MqttHubClient) SubscribeToConfig(
	thingID string, handler func(thingID string, config map[string]interface{}, senderID string)) {

	topic := strings.ReplaceAll(api.TopicSetConfig, "{id}", thingID)
	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		// FIXME: determine sender and format for event message
		sender := ""
		config := make(map[string]interface{})
		err := json.Unmarshal(message, &config)
		if err == nil {
			subscribedHandler(subscribedThingID, config, sender)
		}
	})
}

// SubscribeEvents receives Thing events from the WoST hub.
func (client *MqttHubClient) SubscribeToEvents(
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

// SubscribePropertyValues receives updates to Thing property values from the WoST Hub
func (client *MqttHubClient) SubscribeToPropertyValues(
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

// SubscribeTD subscribes to receive updates to TDs from the WoST Hub
//  thingID is the full ID of a thing, or "" to subscribe to all thingIDs
func (client *MqttHubClient) SubscribeToTD(
	thingID string, handler func(thingID string, thingTD api.ThingTD, senderID string)) {

	if thingID == "" {
		thingID = "+"
	}
	topic := strings.ReplaceAll(api.TopicThingTD, "{id}", thingID)
	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		// FIXME: determine sender and format for td message
		sender := ""
		// TODO: support for topics where thingID isn't the second part
		addressParts := strings.Split(address, "/")
		_ = subscribedThingID
		rxThingID := addressParts[1]
		td := make(map[string]interface{})
		err := json.Unmarshal(message, &td)
		if err != nil {
			logrus.Errorf("Received message on topic '%s' but unmarshal failed: %s", topic, err)
		} else {
			subscribedHandler(rxThingID, td, sender)
		}
	})
}

// Unsubscribe removes thing subscription
func (client *MqttHubClient) Unsubscribe(thingID string) {
	if thingID == "" {
		thingID = "+"
	}
	topic := api.TopicRoot + "/" + thingID + "/#"
	client.mqttClient.Unsubscribe(topic)
}

// NewMqttHubClient creates a new hub connection for Thing consumers
//   hostPort address and port to connect to
//   certFolder containing server and client certificates for TLS connections
//   clientID to identify as. Leave empty to use hostname-timestamp
//   credentials with secret to verify the identity
func NewMqttHubClient(hostPort string, caCertFile string, clientID string, credentials string) api.IHubClient {
	client := &MqttHubClient{
		timeoutSec: 3,
		clientID:   clientID,
		mqttClient: NewMqttClient(hostPort, caCertFile),
	}
	return client
}
