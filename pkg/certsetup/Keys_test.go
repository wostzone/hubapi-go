package certsetup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
)

const privKeyPemFile = "../../test/certs/privKey.pem"

func TestSaveLoadPrivKey(t *testing.T) {
	privKey := certsetup.CreateECDSAKeys()
	err := certsetup.SavePrivateKeyToPEM(privKey, privKeyPemFile)
	assert.NoError(t, err)

	privKey2, err := certsetup.LoadPrivateKeyFromPEM(privKeyPemFile)
	assert.NoError(t, err)
	assert.NotNil(t, privKey2)
}

func TestSaveLoadPrivKeyNotFound(t *testing.T) {
	privKey := certsetup.CreateECDSAKeys()
	// no access
	err := certsetup.SavePrivateKeyToPEM(privKey, "/root")
	assert.Error(t, err)

	//
	privKey2, err := certsetup.LoadPrivateKeyFromPEM("/root")
	assert.Error(t, err)
	assert.Nil(t, privKey2)
}

func TestPublicKeyPEM(t *testing.T) {
	privKey := certsetup.CreateECDSAKeys()

	pem, err := certsetup.PublicKeyToPEM(&privKey.PublicKey)

	assert.NoError(t, err)
	assert.NotEmpty(t, pem)

	pubKey, err := certsetup.PublicKeyFromPEM(pem)
	assert.NoError(t, err)
	assert.NotNil(t, pubKey)

	isEqual := privKey.PublicKey.Equal(pubKey)
	assert.True(t, isEqual)
}

func TestPrivateKeyPEM(t *testing.T) {
	privKey := certsetup.CreateECDSAKeys()

	pem, err := certsetup.PrivateKeyToPEM(privKey)

	assert.NoError(t, err)
	assert.NotEmpty(t, pem)

	privKey2, err := certsetup.PrivateKeyFromPEM(pem)
	assert.NoError(t, err)
	assert.NotNil(t, privKey2)

	isEqual := privKey.Equal(privKey2)
	assert.True(t, isEqual)
}

func TestInvalidPEM(t *testing.T) {
	privKey, err := certsetup.PrivateKeyFromPEM("PRIVATE KEY")
	assert.Error(t, err)
	assert.Nil(t, privKey)

	pubKey, err := certsetup.PublicKeyFromPEM("PUBLIC KEY")
	assert.Error(t, err)
	assert.Nil(t, pubKey)
}
