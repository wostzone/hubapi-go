package wostmqtt

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/api"
)

// MqttThingClient is a wrapper around the generic MQTT client with convenience methods for use
// by devices to publish Thing information and subscribe to Thing configuration and action requests.
// This implements the IThingClient API
type MqttThingClient struct {
	mqttClient         *MqttClient
	clientID           string
	timeoutSec         int
	senderVerification bool
}

// Start the client connection
func (client *MqttThingClient) Start(senderVerification bool) error {
	client.senderVerification = senderVerification
	err := client.mqttClient.Connect(client.clientID, client.timeoutSec)
	return err
}

// End the client connection
func (client *MqttThingClient) Stop() {
	client.mqttClient.Disconnect()
}

// PublishEvent publish a Thing event to the WoST hub
// Intended to by used by a Thing
func (client *MqttThingClient) PublishEvent(thingID string, event map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicThingEvent, "{id}", thingID)
	message, err := json.Marshal(event)
	if err != nil {
		logrus.Error(err)
		return err
	}
	err = client.mqttClient.Publish(topic, message)
	return err
}

// PublishPropertyValues publish a Thing property values to the WoST hub
// Intended to by used by a Thing to publish updates of property values
func (client *MqttThingClient) PublishPropertyValues(thingID string, values map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicThingPropertyValues, "{id}", thingID)
	message, err := json.Marshal(values)
	if err != nil {
		logrus.Error(err)
		return err
	}
	err = client.mqttClient.Publish(topic, message)
	return err
}

// PublishTD publish a Thing description to the WoST hub
// Intended to by used by a Thing to publish an update to its TD
func (client *MqttThingClient) PublishTD(thingID string, td map[string]interface{}) error {
	topic := strings.ReplaceAll(api.TopicThingTD, "{id}", thingID)
	message, err := json.Marshal(td)
	if err != nil {
		logrus.Error(err)
		return err
	}
	err = client.mqttClient.Publish(topic, message)
	return err
}

// SubscribeToAction subscribes a handler to requested actions.
func (client *MqttThingClient) SubscribeToAction(
	thingID string, handler func(thingID string, action map[string]interface{}, senderID string)) {

	topic := strings.ReplaceAll(api.TopicAction, "{id}", thingID)
	// local copy of arguments
	subscribedThingID := thingID
	subscribedHandler := handler
	client.mqttClient.Subscribe(topic, func(address string, message []byte) {
		// FIXME: determine sender and format for action message
		sender := ""
		action := make(map[string]interface{})
		err := json.Unmarshal(message, &action)
		if err == nil {
			subscribedHandler(subscribedThingID, action, sender)
		}
	})
}

// SubscribeToConfig subscribes a handler to the request for configuration updates.
func (client *MqttThingClient) SubscribeToConfig(
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

// Create a new instance of the Thing Client API
//   hostPort address and port to connect to
//   certFolder containing server and client certificates for TLS connections
//   clientID to identify as
//   credentials with secret to verify the identity
func NewThingClient(hostPort string, caCertFile string, clientID string, credentials string) api.IThingClient {
	client := &MqttThingClient{
		timeoutSec: 3,
		clientID:   clientID,
		mqttClient: NewMqttClient(hostPort, caCertFile),
	}
	return client
}
