// package hubclient for creating hub connections by things and consumers
package hubclient

import (
	"fmt"
	"path"

	"github.com/wostzone/hubapi/api"
	"github.com/wostzone/hubapi/internal/mqttclient"
	"github.com/wostzone/hubapi/pkg/certsetup"
	"github.com/wostzone/hubapi/pkg/hubconfig"
)

// NewHubClient creates a new hub connection for things, consumers and plugins
//   hostPort address and port to connect to
//   caCertFile CA certificate for verifying the TLS connections
//   clientID to identify as. Leave empty to use hostname-timestamp
//   credentials with secret to verify the identity
func NewConsumerClient(hostPort string, caCertFile string, clientID string, credentials string) api.IConsumerClient {
	client := mqttclient.NewMqttHubClient(hostPort, caCertFile, clientID, credentials)
	return client
}

// NewPluginClient creates a new hub connection for plugins.
// plugins are very similar to consumers, except they have access to the hubconfig file
//   hostPort address and port to connect to
//   caCertFile CA certificate for verifying the TLS connections
//   clientID to identify as. Leave empty to use hostname-timestamp
//   credentials with secret to verify the identity
func NewPluginClient(clientID string, hubConfig *hubconfig.HubConfig) api.IHubClient {
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Messenger.Address, hubConfig.Messenger.Port)
	caCertFile := path.Join(hubConfig.Messenger.CertsFolder, certsetup.CaCertFile)
	credentials := "" // todo
	client := mqttclient.NewMqttHubClient(hostPort, caCertFile, clientID, credentials)
	return client
}

// NewThingClient creates a new hub connection for Things
//   hostPort address and port to connect to
//   caCertFile CA certificate for verifying the TLS connections
//   clientID to identify as. Leave empty to use hostname-timestamp
//   credentials with secret to verify the identity
func NewThingClient(hostPort string, caCertFile string, clientID string, credentials string) api.IThingClient {
	client := mqttclient.NewMqttHubClient(hostPort, caCertFile, clientID, credentials)
	return client
}
