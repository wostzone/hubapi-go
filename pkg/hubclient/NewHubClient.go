// package hubclient for creating hub connections by things and consumers
package hubclient

import (
	"fmt"
	"path"

	"github.com/wostzone/hubapi-go/api"
	"github.com/wostzone/hubapi-go/internal/mqttclient"
	"github.com/wostzone/hubapi-go/pkg/certsetup"
	"github.com/wostzone/hubapi-go/pkg/hubconfig"
)

// NewHubClient creates a new hub connection for things, consumers and plugins
//   hostPort address and port to connect to
//   caCertFile CA certificate for verifying the TLS connections
//   clientID to identify as. Leave empty to use hostname-timestamp
//   credentials with secret to verify the identity
func NewHubClient(hostPort string, caCertFile string, clientID string, credentials string) api.IHubClient {
	client := mqttclient.NewMqttHubClient(hostPort, caCertFile, clientID, credentials)
	return client
}

// NewPluginClient creates a new hub mqtt connection for plugins.
// plugins must authenticate with a client certificate.
//   pluginID is the plugin instance ID used to connect with
//   hubConfig with Hub configuration include certificate location
func NewPluginClient(pluginID string, hubConfig *hubconfig.HubConfig) api.IHubClient {
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Messenger.Address, hubConfig.Messenger.PluginPortMqtt)
	caCertFile := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	clientCertFile := path.Join(hubConfig.CertsFolder, certsetup.ClientCertFile)
	clientKeyFile := path.Join(hubConfig.CertsFolder, certsetup.ClientKeyFile)

	client := mqttclient.NewMqttHubPluginClient(pluginID, hostPort, caCertFile, clientCertFile, clientKeyFile)
	return client
}

// NewDeviceClient creates a new hub mqtt connection for devices that publish Things.
// devices must authenticate with a client certificate assigned during provisioning.
//   deviceID thingID of the device connecting
//   hostPort address and port to connect to
//   caCertFile CA certificate for verifying the TLS connections
//   clientCertFile client certificate to identify the device
//   clientKeyFile for certificate authentication
func NewDeviceClient(deviceID string, hostPort string, caCertFile string, clientCertFile string, clientKeyFile string) api.IHubClient {

	client := mqttclient.NewMqttHubPluginClient(deviceID, hostPort, caCertFile, clientCertFile, clientKeyFile)
	return client
}
