package wostapi

// WostAPI for protocol client implementations
// Standardization on this API allows for changing protocols without many changes to the
// application or Thing.
type WostAPI interface {
	// Start the client connection
	Start(hostname string, clientID string) error

	// End the client connection
	Stop()

	// PublishTD publish a Thing description to the WoST hub
	PublishTD(thingID string, td []byte) error

	// PublishPropertyValues publish a Thing property values to the WoST hub
	PublishPropertyValues(thingID string, values []byte) error

	// PublishEvent publish a Thing event to the WoST hub
	PublishEvent(thingID string, event []byte) error

	// PublishAction requests an action from a Thing
	PublishAction(thingID string, action []byte) error
}
