// Package certsetup with creation of self signed certificate chain
// Credits: https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251
package certsetup

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

const keySize = 2048 // 4096
const caDurationYears = 10

// const certDurationYears = 10
const DefaultCertDuration = time.Hour * 24 * 365

// Standard client and server certificate filenames
const (
	CaCertFile     = "ca.crt" // CA that signed the server and client certificates
	CaKeyFile      = "ca.key"
	ServerCertFile = "hub.crt"
	ServerKeyFile  = "hub.key"
	ClientCertFile = "client.crt"
	ClientKeyFile  = "client.key"
)

// func GenCARoot() (*x509.Certificate, []byte, *rsa.PrivateKey) {
// 	if _, err := os.Stat("someFile"); err == nil {
// 		//read PEM and cert from file
// 	}
// 	var rootTemplate = x509.Certificate{
// 		SerialNumber: big.NewInt(1),
// 		Subject: pkix.Name{
// 			Country:      []string{"SE"},
// 			Organization: []string{"Company Co."},
// 			CommonName:   "Root CA",
// 		},
// 		NotBefore:             time.Now().Add(-10 * time.Second),
// 		NotAfter:              time.Now().AddDate(10, 0, 0),
// 		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
// 		BasicConstraintsValid: true,
// 		IsCA:                  true,
// 		MaxPathLen:            2,
// 		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
// 	}
// 	priv, err := rsa.GenerateKey(rand.Reader, 2048)
// 	if err != nil {
// 		panic(err)
// 	}
// 	rootCert, rootPEM := genCert(&rootTemplate, &rootTemplate, &priv.PublicKey, priv)
// 	return rootCert, rootPEM, priv
// }

// CreateClientCert creates a client side certificate, signed by the CA
func CreateClientCert(caCertPEM []byte, caKeyPEM []byte, hostname string) (pkPEM []byte, certPEM []byte, err error) {
	// The device gets a new private key
	clientKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}
	// We need the CA private key and certificate
	caPrivKeyBlock, _ := pem.Decode(caKeyPEM)
	caPrivKey, err := x509.ParsePKCS1PrivateKey(caPrivKeyBlock.Bytes)
	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return nil, nil, err
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	// create a CSR for this client
	csrPEM, err := CreateCSR(clientKey, hostname)
	clientCertPEM, err := SignCertificate(csrPEM, caCert, caPrivKey, DefaultCertDuration)

	// // hostname = "localhost"
	// // set up our server certificate
	// template := &x509.Certificate{
	// 	SerialNumber: big.NewInt(2021),
	// 	Subject: pkix.Name{
	// 		Organization:  []string{"WoST"},
	// 		Country:       []string{"CA"},
	// 		Province:      []string{"BC"},
	// 		Locality:      []string{"WoST Client"},
	// 		StreetAddress: []string{""},
	// 		PostalCode:    []string{""},
	// 		CommonName:    hostname,
	// 	},
	// 	NotBefore:    time.Now(),
	// 	NotAfter:     time.Now().AddDate(certDurationYears, 0, 0),
	// 	SubjectKeyId: []byte{1, 2, 3, 4, 6},
	// 	ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	// 	KeyUsage:     x509.KeyUsageDigitalSignature,
	// }

	// clientCertBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, &clientKey.PublicKey, caPrivKey)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// clientCertPEMBuffer := new(bytes.Buffer)
	// pem.Encode(clientCertPEMBuffer, &pem.Block{
	// 	Type:  "CERTIFICATE",
	// 	Bytes: clientCertBytes,
	// })

	clientKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(clientKeyPEMBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
	})

	return clientCertPEM, clientKeyPEMBuffer.Bytes(), nil
}

// CreateCSR creates a certificate signing request using the device's private key.
// The CSR is used in the provisioning process to obtain a certificate that is signed
// by the CA (certificate authority) of the server. The signed certificate is used
// by the device to authenticate and identify itself with the server, for example a message bus.
//
// The same private key of the device must be used when connecting to the server with
// the certificate that is received for this CSR.
//  privKey is the private key of the device.
//  deviceThingID contains the ThingID that represents the device.
// This returns the CSR in PEM format
func CreateCSR(devicePrivKey *rsa.PrivateKey, deviceThingID string) (csrPEM []byte, err error) {
	// var oidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}
	subj := pkix.Name{
		CommonName: deviceThingID,
		Country:    []string{"CA"},
		Province:   []string{"BC"},
		Locality:   []string{"WoSTZone"},
		// StreetAddress: []string{""},
		// PostalCode:    []string{""},
		Organization:       []string{"WoST"},
		OrganizationalUnit: []string{"Client"},
		// extended attributes
		// ExtraNames: []pkix.AttributeTypeAndValue{
		// 	{
		// 		Type: oidEmailAddress,
		// 		Value: asn1.RawValue{
		// 			Tag:   asn1.TagIA5String,
		// 			Bytes: []byte(emailAddress),
		// 		},
		// 	},
		// },
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, devicePrivKey)
	if err != nil {
		return nil, err
	}
	csrPEMBuffer := new(bytes.Buffer)
	pem.Encode(csrPEMBuffer, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	return csrPEMBuffer.Bytes(), err
}

// Create the CA, server and client certificates into the given folder
func CreateCertificateBundle(hostname string, certFolder string) error {
	caCertPEM, caKeyPEM := CreateWoSTCA()
	serverCertPEM, serverKeyPEM, _ := CreateHubCert(caCertPEM, caKeyPEM, hostname)
	clientCertPEM, clientKeyPEM, _ := CreateClientCert(caCertPEM, caKeyPEM, hostname)

	caCertPath := path.Join(certFolder, CaCertFile)
	caKeyPath := path.Join(certFolder, CaKeyFile)
	serverCertPath := path.Join(certFolder, ServerCertFile)
	serverKeyPath := path.Join(certFolder, ServerKeyFile)
	clientCertPath := path.Join(certFolder, ClientCertFile)
	clientKeyPath := path.Join(certFolder, ClientKeyFile)

	err := ioutil.WriteFile(caKeyPath, caKeyPEM, 0600)
	if err != nil {
		logrus.Fatalf("CreateCertificates failed writing. Unable to continue: %s", err)
	}
	ioutil.WriteFile(caCertPath, caCertPEM, 0644)
	ioutil.WriteFile(serverKeyPath, serverKeyPEM, 0600)
	ioutil.WriteFile(serverCertPath, serverCertPEM, 0644)
	ioutil.WriteFile(clientKeyPath, clientKeyPEM, 0600)
	ioutil.WriteFile(clientCertPath, clientCertPEM, 0644)
	return nil
}

// CreateHubCert creates Wost message bus server key and certificate
// TODO: replace with create CSR and SignCertificate from CSR
func CreateHubCert(caCertPEM []byte, caKeyPEM []byte, hostname string) (pkPEM []byte, certPEM []byte, err error) {
	// We need the CA key and certificate
	caPrivKeyBlock, _ := pem.Decode(caKeyPEM)
	caPrivKey, err := x509.ParsePKCS1PrivateKey(caPrivKeyBlock.Bytes)
	certBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	if hostname == "" {
		hostname = "localhost"
	}
	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:  []string{"WoST Zone"},
			Country:       []string{"CA"},
			Province:      []string{"BC"},
			Locality:      []string{"WoST Hub"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    hostname,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Duration(DefaultCertDuration)),
		// SubjectKeyId: []byte{1, 2, 3, 4, 6},   // WTF is this???
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// If an IP address is given, then allow localhost
	ipAddr := net.ParseIP(hostname)
	if ipAddr != nil {
		logrus.Infof("CreateHubCert: hostname %s is an IP address. Setting as SAN", hostname)
		cert.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback, ipAddr}
	}

	privKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &privKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	privKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(privKeyPEMBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	return certPEMBuffer.Bytes(), privKeyPEMBuffer.Bytes(), nil
}

// CreateWoSTCA creates WoST CA and certificate and private key for signing server certificates
// Source: https://shaneutt.com/blog/golang-ca-and-signed-cert-go/
func CreateWoSTCA() (certPEM []byte, keyPEM []byte) {
	// set up our CA certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:  []string{"WoST Zone"},
			Country:       []string{"CA"},
			Province:      []string{"BC"},
			Locality:      []string{"Project"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "Root CA",
		},
		NotBefore:             time.Now().Add(-10 * time.Second),
		NotAfter:              time.Now().AddDate(caDurationYears, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		// MaxPathLen: 2,
		// 		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create the CA private key
	privKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		logrus.Errorf("CertSetup: Unable to create private key: %s", err)
		return nil, nil
	}

	// PEM encode private key
	privKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(privKeyPEMBuffer, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privKey.PublicKey, privKey)
	if err != nil {
		logrus.Errorf("CertSetup: Unable to create CA cert: %s", err)
		return nil, nil
	}

	// pem encode certificate
	certPEMBuffer := new(bytes.Buffer)
	pem.Encode(certPEMBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	return certPEMBuffer.Bytes(), privKeyPEMBuffer.Bytes()
}

// Sign a certificate from a Certificate Signing Request
// Intended to generate a client certificate from a CA for use in authenticiation
// Thanks to https://stackoverflow.com/questions/42643048/signing-certificate-request-with-certificate-authority
//  csrPEM contains the PEM formatted certificate signing request
//  caCert contains the CA certificate for signing
//  caPrivKey contains the CA private key for signing
//  duration is the validity duration of the certificate
func SignCertificate(csrPEM []byte, caCert *x509.Certificate, caPrivKey *rsa.PrivateKey, duration time.Duration,
) (clientCertPEM []byte, err error) {

	csrBlock, _ := pem.Decode(csrPEM)
	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// check if the signature on the CSR is valid
	// Not sure what this means exactly
	err = csr.CheckSignature()
	if err != nil {
		return nil, err
	}

	// create client certificate template using the CSR
	clientCRTTemplate := x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,

		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,

		SerialNumber: big.NewInt(2),
		Issuer:       caCert.Subject,
		Subject:      csr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(duration),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// create client certificate from template and CA public key
	certRaw, err := x509.CreateCertificate(rand.Reader, &clientCRTTemplate, caCert, csr.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}
	clientCertPEMBuffer := new(bytes.Buffer)
	pem.Encode(clientCertPEMBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: certRaw})
	return clientCertPEMBuffer.Bytes(), nil
}
