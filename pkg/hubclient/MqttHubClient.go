package hubclient

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

/* Client library with the MQTT API to the Hub using (tbd):
A: paho-mqtt
B: https://github.com/emqx/emqx
*/

// Message types to receive
const (
	MessageTypeAction = "action"
	MessageTypeConfig = "config" // update property config
	MessageTypeEvent  = "event"
	MessageTypeTD     = "td"
	MessageTypeValues = "values" // receive property values
)

// MqttHubClient is a wrapper around the generic MQTT client with convenience methods for use
// by plugins, Things and consumers to subscribe to Thing information and publish configuration,
// events and actions.
// This implements the IHubConnection API
type MqttHubClient struct {
	// authentication using client cert
	clientCertFile string
	clientKeyFile  string
	// authentication using username/password
	userName string
	password string
	//
	mqttClient *MqttClient
	// senderVerification bool
}

// Close the client connection
func (client *MqttHubClient) Close() {
	logrus.Warningf("MqttHubClient.Close")
	if client.mqttClient != nil {
		client.mqttClient.Close()
	}
}

// Connect the client connection
func (client *MqttHubClient) Connect() error {
	logrus.Warningf("MqttHubClient.Connect: Username='%s', connecting to '%s'. CaCertFile '%s'",
		client.userName, client.mqttClient.hostPort, client.mqttClient.tlsCACertFile)
	// client.senderVerification = senderVerification
	if client.clientCertFile != "" && client.clientKeyFile != "" {
		return client.mqttClient.ConnectWithClientCert(client.userName, client.clientCertFile, client.clientKeyFile)
	} else {
		return client.mqttClient.ConnectWithPassword(client.userName, client.password)
	}
}

// PublishAction publish a Thing action request to the Hub
func (client *MqttHubClient) PublishAction(thingID string, name string, params map[string]interface{}) error {
	topic := strings.ReplaceAll(TopicAction, "{id}", thingID)
	actions := map[string]interface{}{name: params}
	message, err := json.Marshal(actions)
	if err != nil {
		return err
	}
	err = client.mqttClient.Publish(topic, message)
	return err
}

// PublishConfig publish a Thing configuration request to the Hub
func (client *MqttHubClient) PublishConfigRequest(thingID string, values map[string]interface{}) error {
	topic := strings.ReplaceAll(TopicSetConfig, "{id}", thingID)
	message, err := json.Marshal(values)
	if err != nil {
		return err
	}
	err = client.mqttClient.Publish(topic, message)
	return err
}

// PublishEvent publish a Thing event to the WoST hub
// Intended to by used by a Thing
func (client *MqttHubClient) PublishEvent(thingID string, event map[string]interface{}) error {
	topic := strings.ReplaceAll(TopicThingEvent, "{id}", thingID)
	message, _ := json.Marshal(event)
	err := client.mqttClient.Publish(topic, message)
	return err
}

// PublishPropertyValues publish a Thing property values to the WoST hub
// Intended to by used by a Thing to publish updates of property values
func (client *MqttHubClient) PublishPropertyValues(thingID string, values map[string]interface{}) error {
	topic := strings.ReplaceAll(TopicThingPropertyValues, "{id}", thingID)
	message, _ := json.Marshal(values)
	err := client.mqttClient.Publish(topic, message)
	return err
}

// PublishTD publish a Thing description to the WoST hub
// Intended to by used by a Thing to publish an update to its TD
func (client *MqttHubClient) PublishTD(thingID string, td map[string]interface{}) error {
	topic := strings.ReplaceAll(TopicThingTD, "{id}", thingID)
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
	subscribedTopic := fmt.Sprintf("%s/%s/#", TopicRoot, thingID)
	subscribedHandler := handler
	client.mqttClient.Subscribe(subscribedTopic, func(topic string, payload []byte) {
		// FIXME: determine sender and format for td message
		sender := ""
		parts := strings.Split(topic, "/")
		if len(parts) > 2 {
			// Topic format is things/thingID/messageType
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

	topic := strings.ReplaceAll(TopicAction, "{id}", thingID)
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

	topic := strings.ReplaceAll(TopicSetConfig, "{id}", thingID)
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
	topic := strings.ReplaceAll(TopicThingEvent, "{id}", thingID)

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

	topic := strings.ReplaceAll(TopicThingPropertyValues, "{id}", thingID)

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
	thingID string, handler func(thingID string, thingTD map[string]interface{}, senderID string)) {

	if thingID == "" {
		thingID = "+"
	}
	topic := strings.ReplaceAll(TopicThingTD, "{id}", thingID)
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
	topic := TopicRoot + "/" + thingID + "/#"
	client.mqttClient.Unsubscribe(topic)
}

// NewMqttHubClient creates a new hub connection for consumers
// Consumers use a login name and password to authenticate
//   hostPort address and port to connect to
//   caCertFile containing the broker CA certificate for TLS connections
//   userName of user that is connecting, or thingID of device
//   password credentials with secret to verify the identity, eg password or Shared Key
func NewMqttHubClient(hostPort string, caCertFile string, userName string, password string) *MqttHubClient {

	client := &MqttHubClient{
		mqttClient: NewMqttClient(hostPort, caCertFile, DefaultTimeoutSec),
		userName:   userName,
		password:   password,
	}
	return client
}

// NewMqttHubClient creates a new hub connection for use by Plugins
// Plugins use a client certificate to authenticate.
//  pluginID is the instance ID of the plugin to identify as
//  hubConfig with mqtt listening ports and certificate location
func NewMqttHubPluginClient(pluginID string, hubConfig *hubconfig.HubConfig) *MqttHubClient {

	caCertPath := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	pluginCertPath := path.Join(hubConfig.CertsFolder, certsetup.PluginCertFile)
	pluginKeyPath := path.Join(hubConfig.CertsFolder, certsetup.PluginKeyFile)
	hostPort := fmt.Sprintf("%s:%d", hubConfig.MqttAddress, hubConfig.MqttCertPort)
	client := &MqttHubClient{
		clientCertFile: pluginCertPath,
		clientKeyFile:  pluginKeyPath,
		userName:       pluginID,
		mqttClient:     NewMqttClient(hostPort, caCertPath, hubConfig.MqttTimeout),
	}
	return client
}

// NewMqttHubDeviceClient creates a new hub mqtt connection for devices that publish Things.
// devices authenticate with a client certificate assigned during provisioning.
//   deviceID instance ID of the device connecting
//   hostPort address and port to connect to. This must use the mqtt cert port
//   caCertFile CA certificate for verifying the TLS connections
//   clientCertFile client certificate to identify the device
//   clientKeyFile for certificate authentication
func NewMqttHubDeviceClient(deviceID string, hostPort string,
	caCertFile string, deviceCertFile string, deviceKeyFile string) *MqttHubClient {

	client := &MqttHubClient{
		clientCertFile: deviceCertFile,
		clientKeyFile:  deviceKeyFile,
		userName:       deviceID,
		mqttClient:     NewMqttClient(hostPort, caCertFile, DefaultTimeoutSec),
	}
	return client
}
