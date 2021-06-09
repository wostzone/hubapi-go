// Package certsetup with creation of self signed certificate chain using ECDSA signing
// Credits: https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251
package certsetup

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi-go/api"
	"github.com/wostzone/hubapi-go/pkg/signing"
)

// const keySize = 2048 // 4096
const caDefaultValidityDuration = time.Hour * 24 * 364 * 10 // 10 years
const caTemporaryValidityDuration = time.Hour * 24 * 3      // 3 days

// const certDurationYears = 10
const DefaultCertDuration = time.Hour * 24 * 365
const TempCertDuration = time.Hour * 24 * 1

// Standard client and server certificate filenames all stored in PEM format
const (
	CaCertFile     = "caCert.pem" // CA that signed the server and client certificates
	CaKeyFile      = "caKey.pem"
	ServerCertFile = "hubCert.pem"
	ServerKeyFile  = "hubKey.pem"
	ClientCertFile = "clientCert.pem"
	ClientKeyFile  = "clientKey.pem"
)

// CreateCertificateBundle is a convenience function to create the Hub CA, server and (plugin) client
// certificates into the given folder. Intended for testing.
// This only creates missing certificates.
func CreateCertificateBundle(hostname string, certFolder string) error {
	var err error
	// create the CA if needed
	caCertPath := path.Join(certFolder, CaCertFile)
	caKeyPath := path.Join(certFolder, CaKeyFile)

	caCertPEM, _ := ioutil.ReadFile(caCertPath)
	caKeyPEM, _ := ioutil.ReadFile(caKeyPath)
	if caCertPEM == nil || caKeyPEM == nil {
		caCertPEM, caKeyPEM = CreateHubCA()
		err := ioutil.WriteFile(caKeyPath, caKeyPEM, 0600)
		if err != nil {
			logrus.Fatalf("CreateCertificates CA failed writing. Unable to continue: %s", err)
		}
		ioutil.WriteFile(caCertPath, caCertPEM, 0644)
	}

	// create the Server cert if needed
	serverCertPath := path.Join(certFolder, ServerCertFile)
	serverKeyPath := path.Join(certFolder, ServerKeyFile)
	serverCertPEM, _ := ioutil.ReadFile(serverCertPath)
	serverKeyPEM, _ := ioutil.ReadFile(serverKeyPath)
	if serverCertPEM == nil || serverKeyPEM == nil {
		serverKey := signing.CreateECDSAKeys()
		serverKeyPEM = signing.PrivateKeyToPem(serverKey)
		serverPubPEM := signing.PublicKeyToPem(&serverKey.PublicKey)
		serverCertPEM, err = CreateHubCert(hostname, serverPubPEM, caCertPEM, caKeyPEM)
		if err != nil {
			logrus.Fatalf("CreateCertificateBundle server failed: %s", err)
		}
		ioutil.WriteFile(serverKeyPath, serverKeyPEM, 0600)
		ioutil.WriteFile(serverCertPath, serverCertPEM, 0644)
	}
	// create the Client cert if needed
	clientCertPath := path.Join(certFolder, ClientCertFile)
	clientKeyPath := path.Join(certFolder, ClientKeyFile)
	clientCertPEM, _ := ioutil.ReadFile(clientCertPath)
	clientKeyPEM, _ := ioutil.ReadFile(clientKeyPath)
	if clientCertPEM == nil || clientKeyPEM == nil {

		clientKey := signing.CreateECDSAKeys()
		clientKeyPEM = signing.PrivateKeyToPem(clientKey)
		clientPubKeyPEM := signing.PublicKeyToPem(&clientKey.PublicKey)
		clientCertPEM, err = CreateClientCert(hostname, api.OUPlugin, clientPubKeyPEM, caCertPEM, caKeyPEM)
		if err != nil {
			logrus.Fatalf("CreateCertificateBundle client failed: %s", err)
		}
		ioutil.WriteFile(clientKeyPath, clientKeyPEM, 0600)
		ioutil.WriteFile(clientCertPath, clientCertPEM, 0644)
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
// Returns the signed certificate or error
func CreateClientCert(clientID string, ou string, clientPubKeyPEM, caCertPEM []byte, caKeyPEM []byte) (certPEM []byte, err error) {
	var certDuration = DefaultCertDuration

	caPrivKey, err := signing.PrivateKeyFromPem(string(caKeyPEM))
	if err != nil {
		return nil, err
	}
	caCert, err := CertFromPEM(caCertPEM)
	if err != nil {
		return nil, err
	}

	clientPubKey := signing.PublicKeyFromPem(clientPubKeyPEM)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:       []string{"WoST"},
			Locality:           []string{"WoST Zone"},
			CommonName:         clientID,
			OrganizationalUnit: []string{ou},
			Names:              make([]pkix.AttributeTypeAndValue, 0),
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Duration(certDuration)),

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},

		IsCA:                  false,
		BasicConstraintsValid: true,
	}

	derCertBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, clientPubKey, caPrivKey)
	certPEM = CertDerToPEM(derCertBytes)
	return certPEM, err
}

// CreateDeviceCSR creates a certificate signing request using the device's private key.
// The CSR is used in the provisioning process to request a certificate that is signed
// by the CA (certificate authority) of the server and contains the device's authentication ID.
//
//  privKey is the private key of the device.
//  deviceID contains the unique ThingID that represents the device.
// This returns the CSR in PEM format
// func CreateDeviceCSR(devicePrivKey *ecdsa.PrivateKey, deviceID string) (csrPEM []byte, err error) {
// 	// var oidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}
// 	subj := pkix.Name{
// 		// CommonName: deviceThingID,
// 		Country:  []string{"CA"},
// 		Province: []string{"BC"},
// 		Locality: []string{"WoSTZone"},
// 		// StreetAddress: []string{""},
// 		// PostalCode:    []string{""},
// 		Organization:       []string{"WoST"},
// 		OrganizationalUnit: []string{"Client"},
// 	}

// 	template := x509.CertificateRequest{
// 		Subject: subj,
// 		// SignatureAlgorithm: x509.SHA256WithRSA,
// 		SignatureAlgorithm: x509.ECDSAWithSHA256,
// 		ExtraExtensions: []pkix.Extension{
// 			SubjectAltName: "RID:" + deviceID,
// 		},
// 	}

// 	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, devicePrivKey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	csrPEMBuffer := new(bytes.Buffer)
// 	pem.Encode(csrPEMBuffer, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

// 	return csrPEMBuffer.Bytes(), err
// }

// CreateHubCA creates WoST Hub Root CA certificate and private key for signing server certificates
// Source: https://shaneutt.com/blog/golang-ca-and-signed-cert-go/
// This creates a CA certificate used for signing client and server certificates.
// CA is valid for 'caDurationYears'
//
//  temporary set to generate a temporary CA for one-off signing
func CreateHubCA() (certPEM []byte, keyPEM []byte) {
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
	privKeyPEM := signing.PrivateKeyToPem(privKey)

	// create the CA
	caCertDer, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &privKey.PublicKey, privKey)
	if err != nil {
		logrus.Errorf("CertSetup: Unable to create WoST Hub CA cert: %s", err)
		return nil, nil
	}

	caCertPEM := CertDerToPEM(caCertDer)
	return caCertPEM, []byte(privKeyPEM)
}

// CreateHubCert creates Wost server certificate
//  hosts contains one or more DNS or IP addresses to add tot he certificate. Localhost is always added
//  pubKey is the Hub public key in PEM format
//  caCertPEM is the CA to sign the server certificate
// returns the signed Hub certificate in PEM format
func CreateHubCert(hosts string, hubPublicKeyPEM []byte, caCertPEM []byte, caKeyPEM []byte) (certPEM []byte, err error) {
	// We need the CA key and certificate
	caPrivKey, err := signing.PrivateKeyFromPem(string(caKeyPEM))
	if err != nil {
		return nil, err
	}
	caCert, err := CertFromPEM(caCertPEM)
	if err != nil {
		return nil, err
	}

	hubPublicKey := signing.PublicKeyFromPem(hubPublicKeyPEM)

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
		NotAfter:  time.Now().Add(time.Duration(DefaultCertDuration)),

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		// ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:           false,
		MaxPathLenZero: true,
		// BasicConstraintsValid: true,
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}
	// determine the hosts for this hub
	hostList := strings.Split(hosts, ",")
	for _, h := range hostList {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	certDer, err := x509.CreateCertificate(rand.Reader, template, caCert, hubPublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}
	certPEM = CertDerToPEM(certDer)

	return certPEM, nil
}

// Convert certificate DER encoding to PEM
//  derBytes is the output of x509.CreateCertificate
func CertDerToPEM(derCertBytes []byte) []byte {
	// pem encode certificate
	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derCertBytes,
	})
	return certPEMBuffer.Bytes()
}

// Convert a PEM certificate to x509 instance
func CertFromPEM(certPEM []byte) (*x509.Certificate, error) {
	caCertBlock, _ := pem.Decode(certPEM)
	if caCertBlock == nil {
		return nil, errors.New("CertFromPEM pem.Decode failed")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	return caCert, err
}

// SignCSR generates a certificate from a Certificate Signing Request, signed by the CA
// Intended to generate a client certificate from a CA for use in authenticiation
// Thanks to https://stackoverflow.com/questions/42643048/signing-certificate-request-with-certificate-authority
//  csrPEM contains the PEM formatted certificate signing request
//  caCert contains the CA certificate for signing
//  caPrivKey contains the CA private key for signing
//  duration is the validity duration of the certificate
// func SignCSR(csrPEM []byte, caCert *x509.Certificate, caPrivKey *ecdsa.PrivateKey, duration time.Duration,
// ) (signedCertPEM []byte, err error) {

// 	csrBlock, _ := pem.Decode(csrPEM)
// 	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// check if the signature on the CSR is valid
// 	// Not sure what this means exactly
// 	err = csr.CheckSignature()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// create client certificate template using the CSR
// 	clientCRTTemplate := x509.Certificate{
// 		Signature:          csr.Signature,
// 		SignatureAlgorithm: csr.SignatureAlgorithm,
// 		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
// 		PublicKey:          csr.PublicKey,

// 		SerialNumber: big.NewInt(2),
// 		Issuer:       caCert.Subject,
// 		Subject:      csr.Subject,
// 		NotBefore:    time.Now().Add(-10 * time.Second),
// 		NotAfter:     time.Now().Add(duration),
// 		KeyUsage:     x509.KeyUsageDigitalSignature,
// 		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
// 	}

// 	// create client certificate from template and CA public key
// 	certDer, err := x509.CreateCertificate(rand.Reader, &clientCRTTemplate, caCert, csr.PublicKey, caPrivKey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	signedCertPEM = CertDerToPEM(certDer)
// 	return signedCertPEM, nil
// }
