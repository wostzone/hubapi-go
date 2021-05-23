// Package api with hub client library interface definition
package api

// Message types to receive
const (
	MessageTypeAction    = "action"
	MessageTypeConfig    = "config" // update property config
	MessageTypeEvent     = "event"
	MessageTypeTD        = "td"
	MessageTypeValues    = "values"    // receive property values
	MessageTypeProvision = "provision" // receive a provision request message
)

// ThingTD contains the Thing Description document
type ThingTD map[string]interface{}

// IHubClient interface describing methods to connect to the Hub using one of
// the protocol bindings.
// Intended for Things, consumers and plugins to exchange messages.
type IHubClient interface {

	// PublishAction requests an action from a Thing.
	// This is intended for consumers that want to control a Thing. The consumer must have
	// sufficient authorization otherwise this request is ignored.
	//  thingID is the unique ID of the Thing whose action is published
	//  action is an object containing the action to request with optional parameters
	//
	PublishAction(thingID string, action string, data map[string]interface{}) error

	// PublishConfigRequest requests a configuration update from a Thing.
	// This is intended for consumers that want to configure a Thing. The consumer must have
	// sufficient authorization otherwise this request is ignored.
	//  thingID is the unique ID of the Thing to configure
	//  request is an object containing the configuration request values
	PublishConfigRequest(thingID string, request map[string]interface{}) error

	// PublishEvent publish a Thing event to the WoST hub.
	// This is intended for use by WoST Things and Hub plugins to notify consumers of an event.
	//  thingID is the unique ID of the Thing whose TD is published
	//  event is an object containing the event
	PublishEvent(thingID string, event map[string]interface{}) error

	// PublishPropertyValues publish a Thing property values to the WoST hub.
	// This is intended for use by WoST Things and Hub plugins to notify consumers of changes to properties
	//
	//  thingID is the unique ID of the Thing whose TD is published
	//  values is an object containing the property values: { "propname": "value", ...}
	PublishPropertyValues(thingID string, values map[string]interface{}) error

	// PublishTD publish a WoT Thing Description to the WoST hub
	// This is intended for use by WoST Things and Hub plugins to notify consumers of the existance of a Thing
	//  thingID is the unique ID of the Thing whose TD is published
	//  td is an map containing the Thing Description created with CreateTD
	PublishTD(thingID string, td ThingTD) error

	// RequestProvisioning sends a request for (re)provisioning.
	// The server processes the request and replies with an updated certificate if the request is approved.
	// Certificate renewals will receive an update right away. Requests that need admin approval
	// will receive a timeout. The server will cache the request until the administrator has approved
	// or denied the request. On the next request, the server will send the response.
	// The recommended approach is to issue this request:
	//  - when not provisioned: on power on, every hour for the first day, and once a day after
	//  - when provisioned: halfway through the certificate lifetime
	//
	// The CSR must contain the following fields:
	//
	//
	//  thingID of the device requesting to be provisioned
	//  csrPEM is the PEM encoded certificate signing request (see certsetup.CreateCSR)
	//  waitSec defines how many seconds to wait for a response
	// returns a PEM encoded signed certificate used for authentication, or a response 'timeout'.
	// If timeout is false and no certificate is received, the request was denied.
	RequestProvisioning(thingID string, csrPEM []byte, waitSec uint) (certPEM []byte, timeout bool)

	// Start the client connection with the Hub
	Start() error

	// End the hub connection
	Stop()

	// Subscribe subscribes a handler to all messages from a Thing
	//  thingID is the unique ID of the Thing that is be subscribed to. Use "" for all things (use with caution)
	//  messageTypes is array of message types: [td, value, event, action, config]
	//  handler is the code that handles the action request
	//     thingID the message belongs to
	//     msgType the type of message, td, event, action, ... see MessageTypeXxx
	//     message is an raw message byte array
	//     senderID is the authenticated client ID of the sender, "" if unauthenticated
	Subscribe(thingID string,
		handler func(thingID string, msgType string, message []byte, senderID string))

	// SubscribeToActions subscribes a handler to requested actions.
	// This is intended for Things and Hub plugins that have actions defined in their TD
	//  thingID is the unique ID of the Thing that is subscribing.
	//  handler is the code that handles the action request
	//     thingID whose action is received
	//     name is the name of the action
	//     params contains the action input parameters
	//     senderID is the authenticated client ID of the sender
	SubscribeToActions(thingID string,
		handler func(thingID string, name string, params map[string]interface{}, senderID string))

	// SubscribeToConfig subscribes a handler to receive the request for configuration updates.
	// This is intended for Things and Hub plugins that have writable configuration defined in their TD
	//  thingID is the unique ID of the Thing that is subscribing.
	//  handler is the code that handles the configuration request
	//     thingID whose configuration update request is receveived
	//     config is an unmarshalled JSON document of the configuration request values
	//     senderID is the authenticated client ID of the sender
	SubscribeToConfig(thingID string, handler func(thingID string, config map[string]interface{}, senderID string))

	// SubscribeToEvent receives Thing events from the WoST hub.
	//
	//  thingID is a the ID to subscribe to, use "" to receive events from all Things
	//  handler is the function that is invoked when the event is received
	//    thingID is the ID of the Thing whose event is received
	//    event is an unmarshalled JSON document with event: {...}
	//    senderID is the authenticated client ID of the sender
	SubscribeToEvents(thingID string, handler func(thingID string, event map[string]interface{}, senderID string))

	// SubscribeToPropertyValues receives updates to Thing property values from the WoST Hub
	//
	//  thingID is a the ID to subscribe to, use "" to receive values from all TDs
	//  handler is the function that is invoked when the values are received
	//    thingID is the ID of the Thing whose values are received
	//    values is an unmarshalled JSON document with property values: { "propname": "value", ...}
	//    senderID is the authenticated client ID of the sender
	SubscribeToPropertyValues(thingID string, handler func(thingID string, values map[string]interface{}, senderID string))

	// SubscribeToTD subscribes to receive updates to TDs from the WoST Hub
	//
	//  thingID is a the ID to subscribe to, use "" to receive all TDs
	//  handler is the function that is invoked when the TD is received
	//    thingID is the ID of the Thing whose TD is received
	//    td is an unmarshalled JSON document containing the received Thing Description
	//    senderID is the authenticated client ID of the sender
	SubscribeToTD(thingID string, handler func(thingID string, td ThingTD, senderID string))

	// Unsubscribe from all messages from thing with thingID
	Unsubscribe(thingID string)
}
