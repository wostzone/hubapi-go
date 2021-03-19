// Package api with Thing client interface definition.
package api

// IThingClient is intended for use by WoST Things to publish their information and subscribe
// to configuration and action requests.
type IThingClient interface {
	// Start the client connection with the Hub using the hostname, clientID and credentials that were
	// provided with the client creation.
	//  senderVerification includes a sender clientID and signature with the message to verify integrity
	// and authenticity of the sender using the Hub public certificate.
	Start(senderVerification bool) error

	// End the client connection
	Stop()

	// PublishTD publish a WoT Thing Description to the WoST hub
	// This is intended for use by WoST Things to notify consumers of the existance of a Thing
	//  thingID is the unique ID of the Thing whose TD is published
	//  td is an object containing the Thing Description
	PublishTD(thingID string, td map[string]interface{}) error

	// PublishPropertyValues publish a Thing property values to the WoST hub.
	// This is intended for use by WoST Things to notify consumers of changes to properties
	//
	//  thingID is the unique ID of the Thing whose TD is published
	//  values is an object containing the property values: { "propname": "value", ...}
	PublishPropertyValues(thingID string, values map[string]interface{}) error

	// PublishEvent publish a Thing event to the WoST hub.
	// This is intended for use by WoST Things to notify consumers of an event.
	//  thingID is the unique ID of the Thing whose TD is published
	//  event is an object containing the event
	PublishEvent(thingID string, event map[string]interface{}) error

	// SubscribeToAction subscribes a handler to requested actions.
	// This is intended for Things that have actions defined in their TD
	//  thingID is the unique ID of the Thing that is subscribing.
	//  handler is the code that handles the action request
	//     thingID whose action to initiate
	//     action is an unmarshalled JSON document containing the action requests
	//     senderID is the authenticated client ID of the sender
	SubscribeToAction(thingID string, handler func(thingID string, action map[string]interface{}, senderID string))

	// SubscribeToConfig subscribes a handler to the request for configuration updates.
	// This is intended for Things that have writable configuration defined in their TD
	//  thingID is the unique ID of the Thing that is subscribing.
	//  handler is the code that handles the configuration request
	//     thingID whose configuration to set
	//     config is an unmarshalled JSON document of the configuration request values
	//     senderID is the authenticated client ID of the sender
	SubscribeToConfig(thingID string, handler func(thingID string, config map[string]interface{}, senderID string))
}
