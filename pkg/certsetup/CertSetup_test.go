package certsetup_test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/signing"
)

var homeFolder string
var certFolder string

// removeCerts easy cleanup for existing device certificate
func removeServerCerts() {
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(certFolder, "*.pem")).Output()
}

// TestMain clears the certs folder for clean testing
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	certFolder = path.Join(homeFolder, "certs")

	res := m.Run()
	os.Exit(res)
}

func TestLoadCreateCertKey(t *testing.T) {
	removeServerCerts()
	privKey, err := certsetup.LoadOrCreateCertKey(certFolder, certsetup.PluginKeyFile)
	assert.NoError(t, err)
	assert.NotNil(t, privKey)
}

func TestTLSCertificateGeneration(t *testing.T) {
	hostnames := []string{"127.0.0.1"}
	clientID := "3rdparty-client"

	// test creating ca and server certificates
	caCertPEM, caKeyPEM := certsetup.CreateHubCA()
	require.NotEmptyf(t, caCertPEM, "Failed creating CA certificate")
	caCert, err := tls.X509KeyPair([]byte(caCertPEM), []byte(caKeyPEM))
	_ = caCert
	require.NoErrorf(t, err, "Failed parsing CA certificate")

	clientKey := signing.CreateECDSAKeys()
	clientKeyPEM, _ := signing.PrivateKeyToPEM(clientKey)
	clientPubPEM, err := signing.PublicKeyToPEM(&clientKey.PublicKey)
	require.NoError(t, err)
	clientCertPEM, err := certsetup.CreateClientCert(clientID, certsetup.OUClient,
		clientPubPEM, caCertPEM, caKeyPEM, time.Now(), certsetup.DefaultCertDurationDays)
	require.NoErrorf(t, err, "Creating certificates failed:")
	require.NotEmptyf(t, clientCertPEM, "Failed creating client certificate")
	require.NotEmptyf(t, clientKeyPEM, "Failed creating client key")

	serverKey := signing.CreateECDSAKeys()
	serverKeyPEM, _ := signing.PrivateKeyToPEM(serverKey)
	serverPubPEM, err := signing.PublicKeyToPEM(&serverKey.PublicKey)
	serverCertPEM, err := certsetup.CreateHubCert(hostnames, serverPubPEM, caCertPEM, caKeyPEM)
	require.NoErrorf(t, err, "Failed creating server certificate")
	// serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	require.NoErrorf(t, err, "Failed creating server certificate")
	require.NotEmptyf(t, serverCertPEM, "Failed creating server certificate")
	require.NotEmptyf(t, serverKeyPEM, "Failed creating server private key")

	// verify the certificate
	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM([]byte(caCertPEM))
	require.True(t, ok, "Failed parsing CA certificate")

	serverBlock, _ := pem.Decode([]byte(serverCertPEM))
	require.NotNil(t, serverBlock, "Failed decoding server certificate PEM")

	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	require.NoError(t, err, "ParseCertificate for server failed")

	opts := x509.VerifyOptions{
		Roots:   certpool,
		DNSName: hostnames[0],
		// DNSName:       "127.0.0.1",
		Intermediates: x509.NewCertPool(),
	}
	_, err = serverCert.Verify(opts)
	require.NoError(t, err, "Verify for server certificate failed")
}

func TestHubCert(t *testing.T) {
	hostnames := []string{"127.0.0.1"}

	// test creating hub certificate
	caCertPEM, caKeyPEM := certsetup.CreateHubCA()
	require.NotEmptyf(t, caCertPEM, "Failed creating CA certificate")
	caCert, err := tls.X509KeyPair([]byte(caCertPEM), []byte(caKeyPEM))
	_ = caCert
	require.NoErrorf(t, err, "Failed parsing CA certificate")
	hubKey := signing.CreateECDSAKeys()
	hubPubPEM, err := signing.PublicKeyToPEM(&hubKey.PublicKey)
	hubCertPEM, err := certsetup.CreateHubCert(hostnames, hubPubPEM, caCertPEM, caKeyPEM)
	require.NoErrorf(t, err, "Failed creating hub certificate")
	require.NotNil(t, hubCertPEM)
}

func TestHubCertBadCA(t *testing.T) {
	hostnames := []string{"127.0.0.1"}
	caCertPEM, caKeyPEM := certsetup.CreateHubCA()
	hubKey := signing.CreateECDSAKeys()
	hubPubPEM, err := signing.PublicKeyToPEM(&hubKey.PublicKey)
	//
	hubCertPEM, err := certsetup.CreateHubCert(hostnames, hubPubPEM, caCertPEM, "BadCAKey")
	require.Error(t, err)
	require.Empty(t, hubCertPEM)

	hubCertPEM, err = certsetup.CreateHubCert(hostnames, hubPubPEM, "BadCACert", caKeyPEM)
	require.Error(t, err)
	require.Empty(t, hubCertPEM)

	hubCertPEM, err = certsetup.CreateHubCert(hostnames, "BadHubPublicKey", caCertPEM, caKeyPEM)
	require.Error(t, err)
	require.Empty(t, hubCertPEM)
}

func TestClientCertBadCA(t *testing.T) {
	clientID := "client1"
	ou := certsetup.OUClient
	caCertPEM, caKeyPEM := certsetup.CreateHubCA()

	clientKey := signing.CreateECDSAKeys()
	// clientKeyPEM, _ := signing.PrivateKeyToPEM(clientKey)
	clientPubPEM, _ := signing.PublicKeyToPEM(&clientKey.PublicKey)

	//
	clientCertPEM, err := certsetup.CreateClientCert(clientID, ou, "bad pubkey",
		caCertPEM, caKeyPEM, time.Now(), certsetup.TempCertDurationDays)
	assert.Error(t, err)
	assert.Empty(t, clientCertPEM)

	clientCertPEM, err = certsetup.CreateClientCert(clientID, ou, clientPubPEM,
		"bad CAcert", caKeyPEM, time.Now(), certsetup.TempCertDurationDays)
	assert.Error(t, err)
	assert.Empty(t, clientCertPEM)

	clientCertPEM, err = certsetup.CreateClientCert(clientID, ou, clientPubPEM,
		caCertPEM, "bad CA Key", time.Now(), certsetup.TempCertDurationDays)
	assert.Error(t, err)
	assert.Empty(t, clientCertPEM)
}

func TestBadCert(t *testing.T) {
	// hostnames := []string{"127.0.0.1"}
	clientID := "3rdpartyID"
	caCertPEM, caKeyPEM := certsetup.CreateHubCA()
	// caCertPEM = pem.Encode( )[]byte{1, 2, 3}

	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "",
		Bytes: []byte{1, 2, 3},
	})
	caCertPEM = string(certPEMBuffer.Bytes())

	clientKey := signing.CreateECDSAKeys()
	clientKeyPEM, _ := signing.PrivateKeyToPEM(clientKey)
	clientPubPEM, _ := signing.PublicKeyToPEM(&clientKey.PublicKey)
	clientCertPEM, err := certsetup.CreateClientCert(clientID, certsetup.OUClient, clientPubPEM,
		caCertPEM, caKeyPEM, time.Now(), certsetup.TempCertDurationDays)

	assert.NotEmptyf(t, clientKeyPEM, "Missing client key")
	assert.Errorf(t, err, "Creating certificates should fail")
	assert.Emptyf(t, clientCertPEM, "Created client certificate")
}

func TestCreateCerts(t *testing.T) {
	hostnames := []string{"localhost"}

	// out, err := exec.Command("sh", "-c", "rm -f "+path.Join(certFolder, "*.pem")).Output()
	// require.NoError(t, err, out)
	removeServerCerts()

	err := certsetup.CreateCertificateBundle(hostnames, certFolder)
	require.NoError(t, err)
	// load the certs
	clientKeyPEM, err := certsetup.LoadPEM(certFolder, certsetup.PluginKeyFile)
	require.NoError(t, err)
	clientPrivKey, err := signing.PrivateKeyFromPEM(clientKeyPEM)
	require.NoError(t, err)
	pubKey := clientPrivKey.PublicKey
	clientPubKeyPEM, err := signing.PublicKeyToPEM(&pubKey)
	require.NoError(t, err)
	caCertPEM, err := certsetup.LoadPEM(certFolder, certsetup.CaCertFile)
	assert.NoError(t, err)
	caKeyPEM, err := certsetup.LoadPEM(certFolder, certsetup.CaKeyFile)
	assert.NoError(t, err)

	clientCertPEM, err := certsetup.LoadPEM(certFolder, certsetup.PluginCertFile)
	_, err = tls.X509KeyPair([]byte(clientCertPEM), []byte(clientKeyPEM))
	assert.NoError(t, err)

	// CA key/cert and pubkey must be usable for creating a cert
	cert, err := certsetup.CreateClientCert("client1", "ou1", clientPubKeyPEM,
		caCertPEM, caKeyPEM, time.Now(), certsetup.TempCertDurationDays)
	assert.NoError(t, err)
	assert.NotNil(t, cert)

}
