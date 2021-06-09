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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hubapi-go/api"
	"github.com/wostzone/hubapi-go/pkg/certsetup"
	"github.com/wostzone/hubapi-go/pkg/signing"
)

var homeFolder string
var certFolder string

// TestMain clears the certs folder for clean testing
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	certFolder = path.Join(homeFolder, "certs")

	m.Run()
	os.Exit(0)
}
func TestTLSCertificateGeneration(t *testing.T) {
	hostname := "127.0.0.1"

	// test creating ca and server certificates
	caCertPEM, caKeyPEM := certsetup.CreateHubCA()
	require.NotNilf(t, caCertPEM, "Failed creating CA certificate")
	caCert, err := tls.X509KeyPair(caCertPEM, caKeyPEM)
	_ = caCert
	require.NoErrorf(t, err, "Failed parsing CA certificate")

	clientKey := signing.CreateECDSAKeys()
	clientKeyPEM := signing.PrivateKeyToPem(clientKey)
	clientPubPEM := signing.PublicKeyToPem(&clientKey.PublicKey)
	clientCertPEM, err := certsetup.CreateClientCert(hostname, api.OUClient, clientPubPEM, caCertPEM, caKeyPEM)
	require.NoErrorf(t, err, "Creating certificates failed:")
	require.NotNilf(t, clientCertPEM, "Failed creating client certificate")
	require.NotNilf(t, clientKeyPEM, "Failed creating client key")

	serverKey := signing.CreateECDSAKeys()
	serverKeyPEM := signing.PrivateKeyToPem(serverKey)
	serverPubPEM := signing.PublicKeyToPem(&serverKey.PublicKey)
	serverCertPEM, err := certsetup.CreateHubCert(hostname, serverPubPEM, caCertPEM, caKeyPEM)
	require.NoErrorf(t, err, "Failed creating server certificate")
	// serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	require.NoErrorf(t, err, "Failed creating server certificate")
	require.NotNilf(t, serverCertPEM, "Failed creating server certificate")
	require.NotNilf(t, serverKeyPEM, "Failed creating server private key")

	// verify the certificate
	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM(caCertPEM)
	require.True(t, ok, "Failed parsing CA certificate")

	serverBlock, _ := pem.Decode(serverCertPEM)
	require.NotNil(t, serverBlock, "Failed decoding server certificate PEM")

	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	require.NoError(t, err, "ParseCertificate for server failed")

	opts := x509.VerifyOptions{
		Roots:   certpool,
		DNSName: hostname,
		// DNSName:       "127.0.0.1",
		Intermediates: x509.NewCertPool(),
	}
	_, err = serverCert.Verify(opts)
	require.NoError(t, err, "Verify for server certificate failed")
}

func TestBadCert(t *testing.T) {
	hostname := "127.0.0.1"
	caCertPEM, caKeyPEM := certsetup.CreateHubCA()
	// caCertPEM = pem.Encode( )[]byte{1, 2, 3}

	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "",
		Bytes: []byte{1, 2, 3},
	})
	caCertPEM = certPEMBuffer.Bytes()

	clientKey := signing.CreateECDSAKeys()
	clientKeyPEM := signing.PrivateKeyToPem(clientKey)
	clientPubPEM := signing.PublicKeyToPem(&clientKey.PublicKey)
	clientCertPEM, err := certsetup.CreateClientCert(hostname, api.OUClient, clientPubPEM, caCertPEM, caKeyPEM)

	assert.NotNilf(t, clientKeyPEM, "Missing client key")
	assert.Errorf(t, err, "Creating certificates should fail")
	assert.Nilf(t, clientCertPEM, "Created client certificate")
}

func TestCreateCerts(t *testing.T) {
	hostname := "localhost"
	out, err := exec.Command("sh", "-c", "rm -f "+path.Join(certFolder, "*.pem")).Output()
	assert.NoError(t, err, out)
	err = certsetup.CreateCertificateBundle(hostname, certFolder)
	assert.NoError(t, err, out)
}
