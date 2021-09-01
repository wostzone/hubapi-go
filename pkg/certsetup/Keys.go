package certsetup

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

//---------------------------------------------------------------------------------
// ECDSA Key management
//---------------------------------------------------------------------------------

// CreateECDSAKeys creates a asymmetric key set
// Returns a private key that contains its associated public key
func CreateECDSAKeys() *ecdsa.PrivateKey {
	rng := rand.Reader
	curve := elliptic.P256()
	privKey, _ := ecdsa.GenerateKey(curve, rng)
	return privKey
}

// Load ECDSA public/private key pair
//  path is the path to the PEM file
func LoadPrivateKeyFromPEM(path string) (privateKey *ecdsa.PrivateKey, err error) {
	pem, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	privKey, err := PrivateKeyFromPEM(string(pem))
	return privKey, err
}

// PrivateKeyFromPEM converts PEM encoded private keys into an ECDSA key object
// See also PrivateKeyToPem for the opposite.
// Returns nil if the encoded pem source isn't a pem format
func PrivateKeyFromPEM(pemEncodedPriv string) (privateKey *ecdsa.PrivateKey, err error) {
	block, _ := pem.Decode([]byte(pemEncodedPriv))
	if block == nil {
		return nil, errors.New("Not a valid PEM string")
	}
	x509Encoded := block.Bytes
	// rawPrivateKey, err := x509.ParsePKCS8PrivateKey(x509Encoded)
	rawPrivateKey, err := x509.ParsePKCS8PrivateKey(x509Encoded)
	if rawPrivateKey != nil {
		privateKey = rawPrivateKey.(*ecdsa.PrivateKey)
		if privateKey == nil {
			err = errors.New("PrivateKeyFromPem: PEM is not a ECDSA key format")
		}
	}
	return privateKey, err
}

// PrivateKeyToPEM converts a private key into their PEM encoded ascii format
//  privKey contains the private key to save
func PrivateKeyToPEM(privateKey *ecdsa.PrivateKey) (string, error) {
	x509Encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded), err
}

// PublicKeyFromPEM converts a PEM encoded public key into a ECDSA or RSA public key object
func PublicKeyFromPEM(pemEncodedPub string) (publicKey *ecdsa.PublicKey, err error) {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	if blockPub == nil {
		return nil, errors.New("Not a valid PEM string")
	}
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err == nil {
		publicKey = genericPublicKey.(*ecdsa.PublicKey)
		if publicKey == nil {
			err = errors.New("PublicKeyFromPEM: Not a ECDSA public key")
		}
	}

	return
}

// PublicKeyToPEM converts a public key into PEM encoded format.
// ECDSA and RSA keys are supported.
// See also PublicKeyFromPem for its counterpart
func PublicKeyToPEM(publicKey *ecdsa.PublicKey) (string, error) {
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})
	return string(pemEncodedPub), err
}

// SavePrivateKeyToPEM saves a public/private key pair to PEM file with 0600 permissions
//  privKey contains the private key to save.
//  path is the path to the PEM file
func SavePrivateKeyToPEM(privKey *ecdsa.PrivateKey, path string) error {
	pem, err := PrivateKeyToPEM(privKey)
	if err == nil {
		err = ioutil.WriteFile(path, []byte(pem), 0600)
	}
	return err
}
