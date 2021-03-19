// Package api with consumer client library interface definition
package api

// ConsumerAPI that connects to the WoST Hub using one of the protocol bindings.
// Intended for Thing consumers and plugins to receive updates from Things.
type IConsumerClient interface {
	// Start the client connection with the Hub using the hostname, clientID and credentials that were
	// provided with the client creation.
	//  senderVerification includes a sender clientID and signature with the message to verify integrity
	// and authenticity of the sender using the Hub public certificate.
	Start(senderVerification bool) error

	// End the client connection
	Stop()

	// PublishAction requests an action from a Thing.
	// This is intended for consumers that want to control a Thing. The consumer must have
	// sufficient authorization otherwise this request is ignored.
	//  thingID is the unique ID of the Thing whose action is published
	//  action is an object containing the action to request
	//    senderID is the authenticated client ID of the sender
	PublishAction(thingID string, action map[string]interface{}) error

	// PublishConfig requests a configuration update from a Thing.
	// This is intended for consumers that want to configure a Thing. The consumer must have
	// sufficient authorization otherwise this request is ignored.
	//  thingID is the unique ID of the Thing to configure
	//  config is an object containing the configuration request values
	PublishConfig(thingID string, config map[string]interface{}) error

	// SubscribeTD subscribes to receive updates to TDs from the WoST Hub
	//
	//  thingID is a the ID to subscribe to, use "" to receive all TDs
	//  handler is the function that is invoked when the TD is received
	//    thingID is the ID of the Thing whose TD is received
	//    td is an unmarshalled JSON document containing the received Thing Description
	//    senderID is the authenticated client ID of the sender
	SubscribeTD(thingID string, handler func(thingID string, td map[string]interface{}, senderID string))

	// SubscribePropertyValues receives updates to Thing property values from the WoST Hub
	//
	//  thingID is a the ID to subscribe to, use "" to receive events from all TDs
	//  handler is the function that is invoked when the values are received
	//    thingID is the ID of the Thing whose values are received
	//    values is an unmarshalled JSON document with property values: { "propname": "value", ...}
	//    senderID is the authenticated client ID of the sender
	SubscribePropertyValues(thingID string, handler func(thingID string, values map[string]interface{}, senderID string))

	// SubscribeEvents receives Thing events from the WoST hub.
	//
	//  thingID is a the ID to subscribe to, use "" to receive events from all Things
	//  handler is the function that is invoked when the event is received
	//    thingID is the ID of the Thing whose event is received
	//    event is an unmarshalled JSON document with event: {...}
	//    senderID is the authenticated client ID of the sender
	SubscribeEvent(thingID string, handler func(thingID string, event map[string]interface{}, senderID string))
}
