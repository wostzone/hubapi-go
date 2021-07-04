// Package certsetup with creation of self signed certificate chain using ECDSA signing
// Credits: https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251
package certsetup

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/signing"
)

// Standard WoST client and server key/certificate filenames. All stored in PEM format.
const (
	CaCertFile     = "caCert.pem" // CA that signed the server and client certificates
	CaKeyFile      = "caKey.pem"
	HubCertFile    = "hubCert.pem"
	HubKeyFile     = "hubKey.pem"
	PluginCertFile = "pluginCert.pem"
	PluginKeyFile  = "pluginKey.pem"
	// ClientCertFile = "clientCert.pem"
	// ClientKeyFile  = "clientKey.pem"
)

// Organization Unit for client authorization are stored in the client certificate OU field
const (
	// Default OU with no API access permissions
	OUNone = ""

	// OUClient lets a client connect to the message bus
	OUClient = "client"

	// OUIoTDevice indicates the client is a IoT device that can connect to the message bus
	// perform discovery and request provisioning.
	// Provision API permissions: GetDirectory, ProvisionRequest, GetStatus
	OUIoTDevice = "iotdevice"

	//OUAdmin lets a client approve thing provisioning (postOOB), add and remove users
	// Provision API permissions: GetDirectory, ProvisionRequest, GetStatus, PostOOB
	OUAdmin = "admin"

	// OUPlugin marks a client as a plugin.
	// By default, plugins have full permission to all APIs
	// Provision API permissions: Any
	OUPlugin = "plugin"
)

// Plugin certificate ID
const pluginClientID = "plugin"

// const keySize = 2048 // 4096
const caDefaultValidityDuration = time.Hour * 24 * 364 * 10 // 10 years
const caTemporaryValidityDuration = time.Hour * 24 * 3      // 3 days

const DefaultCertDurationDays = 365
const TempCertDurationDays = 1

// CreateCertificateBundle is a convenience function to create the Hub CA, server and (plugin) client
// certificates into the given folder.
// The names contain the hostnames and ip addresses that are valid for the hub.
// This only creates missing certificates.
//  names contain the list of hostname and ip addresses the hub can be reached at. Used in hub cert.
func CreateCertificateBundle(names []string, certFolder string) error {
	var err error
	// create the CA if needed
	caCertPEM, _ := LoadPEM(certFolder, CaCertFile)
	caKeyPEM, _ := LoadPEM(certFolder, CaKeyFile)
	if caCertPEM == "" || caKeyPEM == "" {
		caCertPEM, caKeyPEM = CreateHubCA()
		err = SaveKeyToPEM(caKeyPEM, certFolder, CaKeyFile)
		if err != nil {
			logrus.Fatalf("CreateCertificates CA failed writing. Unable to continue: %s", err)
		}
		err = SaveCertToPEM(caCertPEM, certFolder, CaCertFile)
	}

	// create the Hub server cert if needed
	serverCertPEM, _ := LoadPEM(certFolder, HubCertFile)
	serverKeyPEM, _ := LoadPEM(certFolder, HubKeyFile)
	if serverCertPEM == "" || serverKeyPEM == "" {
		serverKey := signing.CreateECDSAKeys()
		serverKeyPEM, _ = signing.PrivateKeyToPEM(serverKey)
		serverPubPEM, err := signing.PublicKeyToPEM(&serverKey.PublicKey)
		if err != nil {
			logrus.Fatalf("CreateCertificateBundle server public key failed: %s", err)
		}
		serverCertPEM, err = CreateHubCert(names, serverPubPEM, caCertPEM, caKeyPEM)
		if err != nil {
			logrus.Fatalf("CreateCertificateBundle server failed: %s", err)
		}
		SaveKeyToPEM(serverKeyPEM, certFolder, HubKeyFile)
		SaveCertToPEM(serverCertPEM, certFolder, HubCertFile)
	}
	// create the Plugin certificate
	pluginCertPEM, _ := LoadPEM(certFolder, PluginCertFile)
	pluginKeyPEM, _ := LoadPEM(certFolder, PluginKeyFile)
	if pluginCertPEM == "" || pluginKeyPEM == "" {

		pluginKey := signing.CreateECDSAKeys()
		pluginKeyPEM, _ = signing.PrivateKeyToPEM(pluginKey)
		pluginPubKeyPEM, err := signing.PublicKeyToPEM(&pluginKey.PublicKey)
		if err != nil {
			logrus.Fatalf("CreateCertificateBundle plugin cert failed: %s", err)
		}
		// The plugin client cert uses the fixed common name 'plugin'
		pluginCertPEM, err = CreateClientCert(pluginClientID, OUPlugin, pluginPubKeyPEM,
			caCertPEM, caKeyPEM, time.Now(), DefaultCertDurationDays)
		if err != nil {
			logrus.Fatalf("CreateCertificateBundle client failed: %s", err)
		}
		SaveKeyToPEM(pluginKeyPEM, certFolder, PluginKeyFile)
		SaveCertToPEM(pluginCertPEM, certFolder, PluginCertFile)
	}
	return nil
}

// CreateClientCert creates a client side Hub certificate for mutual authentication from client's public key
// The client role is intended to indicate authorization by role. It is stored in the
// certificate OrganizationalUnit. See RoleXxx in api
//
// This generates a certificate using the client's public key in PEM format
//  clientID used as the CommonName
//  ou of the client, stored as the OrganizationalUnit
//  clientPubKeyPEM with the client's public key
//  caCertPEM CA's certificate in PEM format.
//  caKeyPEM CA's ECDSA key used in signing.
//  start time the certificate is first valid. Intended for testing. Use time.now()
//  durationDays nr of days the certificate will be valid
// Returns the signed certificate or error
func CreateClientCert(clientID string, ou string, clientPubKeyPEM, caCertPEM string,
	caKeyPEM string, start time.Time, durationDays int) (certPEM string, err error) {

	caPrivKey, err := signing.PrivateKeyFromPEM(caKeyPEM)
	if err != nil {
		return "", err
	}
	caCert, err := CertFromPEM(caCertPEM)
	if err != nil {
		return "", err
	}

	clientPubKey, err := signing.PublicKeyFromPEM(clientPubKeyPEM)
	if err != nil {
		return "", err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:       []string{"WoST"},
			Locality:           []string{"WoST Zone"},
			CommonName:         clientID,
			OrganizationalUnit: []string{ou},
			Names:              make([]pkix.AttributeTypeAndValue, 0),
		},
		NotBefore: start,
		NotAfter:  start.AddDate(0, 0, durationDays),

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},

		IsCA:                  false,
		BasicConstraintsValid: true,
	}
	derCertBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, clientPubKey, caPrivKey)
	certPEM = CertDerToPEM(derCertBytes)
	return certPEM, err
}

// CreateHubCA creates WoST Hub Root CA certificate and private key for signing server certificates
// Source: https://shaneutt.com/blog/golang-ca-and-signed-cert-go/
// This creates a CA certificate used for signing client and server certificates.
// CA is valid for 'caDurationYears'
//
//  temporary set to generate a temporary CA for one-off signing
func CreateHubCA() (certPEM string, keyPEM string) {
	validity := caDefaultValidityDuration

	// set up our CA certificate
	// see also: https://superuser.com/questions/738612/openssl-ca-keyusage-extension
	rootTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Country:      []string{"CA"},
			Organization: []string{"WoST"},
			Province:     []string{"BC"},
			Locality:     []string{"WoST Zone"},
			CommonName:   "WoST CA",
		},
		NotBefore: time.Now().Add(-10 * time.Second),
		NotAfter:  time.Now().Add(validity),
		// CA cert can be used to sign certificate and revocation lists
		KeyUsage:    x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},

		// This hub cert is the only CA. Don't allow intermediate CAs
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create the CA private key
	privKey := signing.CreateECDSAKeys()
	privKeyPEM, _ := signing.PrivateKeyToPEM(privKey)

	// create the CA
	caCertDer, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &privKey.PublicKey, privKey)
	if err != nil {
		logrus.Errorf("CertSetup: Unable to create WoST Hub CA cert: %s", err)
		return "", ""
	}

	caCertPEM := CertDerToPEM(caCertDer)
	return caCertPEM, privKeyPEM
}

// CreateHubCert creates Wost server certificate
// The Hub certificate is valid for the given names (domain name and IP addresses).
// This implies that the Hub must use a fixed IP. DNS names are not used for validation.
//  names contains one or more domain names or IP addresses the Hub can be reached on, to add to the certificate
//  pubKey is the Hub public key in PEM format
//  caCertPEM is the CA to sign the server certificate
// returns the signed Hub certificate in PEM format
func CreateHubCert(names []string, hubPublicKeyPEM string, caCertPEM string, caKeyPEM string) (certPEM string, err error) {
	// We need the CA key and certificate
	caPrivKey, err := signing.PrivateKeyFromPEM(caKeyPEM)
	if err != nil {
		return "", err
	}
	caCert, err := CertFromPEM(caCertPEM)
	if err != nil {
		return "", err
	}

	hubPublicKey, err := signing.PublicKeyFromPEM(hubPublicKeyPEM)
	if err != nil {
		return "", err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization: []string{"WoST"},
			Country:      []string{"CA"},
			Province:     []string{"BC"},
			Locality:     []string{"WoST Zone"},
			CommonName:   "WoST Hub",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(0, 0, DefaultCertDurationDays),

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		// ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:           false,
		MaxPathLenZero: true,
		// BasicConstraintsValid: true,
		// IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		IPAddresses: []net.IP{},
	}
	// determine the hosts for this hub

	for _, h := range names {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	certDer, err := x509.CreateCertificate(rand.Reader, template, caCert, hubPublicKey, caPrivKey)
	if err != nil {
		return "", err
	}
	certPEM = CertDerToPEM(certDer)

	return certPEM, nil
}

// Convert certificate DER encoding to PEM
//  derBytes is the output of x509.CreateCertificate
func CertDerToPEM(derCertBytes []byte) string {
	// pem encode certificate
	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derCertBytes,
	})
	return certPEMBuffer.String()
}

// Convert a PEM certificate to x509 instance
func CertFromPEM(certPEM string) (*x509.Certificate, error) {
	caCertBlock, _ := pem.Decode([]byte(certPEM))
	if caCertBlock == nil {
		return nil, errors.New("CertFromPEM pem.Decode failed")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	return caCert, err
}

// LoadOrCreateCertKey is a helper to load a public/private key pair for certificate management
// If the keys don't exist, they are created.
//  certFolder location where key file is stored
//  keyFile is the name of the key file, certsetup.ClientKeyFile, ServerKeyFile or CAKeyFile
// Returns ECDSA private key
func LoadOrCreateCertKey(certFolder string, keyFile string) (*ecdsa.PrivateKey, error) {

	pkPath := path.Join(certFolder, keyFile)
	privKey, err := signing.LoadPrivateKeyFromPEM(pkPath)

	if privKey == nil {
		privKey = signing.CreateECDSAKeys()
		err = signing.SavePrivateKeyToPEM(privKey, pkPath)
		if err != nil {
			logrus.Errorf("CreateClientKeys.Start, failed saving private key: %s", err)
			return nil, err
		}
	}
	return privKey, nil
}

// LoadPEM loads PEM file from certificate folder
// Return loaded PEM file as string
func LoadPEM(certFolder string, fileName string) (pem string, err error) {
	pemPath := path.Join(certFolder, fileName)
	pemData, err := ioutil.ReadFile(pemPath)
	return string(pemData), err
}

// SaveKeyToPEM saves the private key in PEM format to file in the certificate folder
// permissions will be 0600
// Return error
func SaveKeyToPEM(pem string, certFolder string, fileName string) error {
	pemPath := path.Join(certFolder, fileName)
	err := ioutil.WriteFile(pemPath, []byte(pem), 0600)
	return err
}

// SaveCertToPEM saves the certificate in pem format to file in the certificate folder
// permissions will be 0644
// Return error
func SaveCertToPEM(pem string, certFolder string, fileName string) error {
	pemPath := path.Join(certFolder, fileName)
	err := ioutil.WriteFile(pemPath, []byte(pem), 0644)
	return err
}
