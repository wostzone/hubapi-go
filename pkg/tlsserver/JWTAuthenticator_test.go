package tlsserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hubserve-go/pkg/tlsserver"
)

// JWT test cases for 100% coverage
func TestJWTToken(t *testing.T) {
	user1 := "user1"
	jauth := tlsserver.NewJWTAuthenticator(nil, func(login, pass string) bool {
		assert.Fail(t, "Should never reach here")
		return false
	})
	expTime := time.Now().Add(time.Second * 100)
	accessToken, refreshToken, err := jauth.CreateJWTTokens("user1", expTime)
	assert.NoError(t, err)

	jwtToken, claims, err := jauth.DecodeToken(accessToken)
	_ = jwtToken
	require.NoError(t, err)
	assert.Equal(t, user1, claims.Username)

	jwtToken, claims, err = jauth.DecodeToken(refreshToken)
	_ = jwtToken
	require.NoError(t, err)
	assert.Equal(t, user1, claims.Username)
}

func TestJWTNotWostToken(t *testing.T) {
	user1 := "user1"
	secret := []byte("notreallyasecret")

	jauth := tlsserver.NewJWTAuthenticator(secret, func(login, pass string) bool {
		assert.Fail(t, "Should never reach here")
		return false
	})

	claims1 := jwt.StandardClaims{Id: user1}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims1)
	// construct a token with the right secret but a different struct type
	signedToken, err := token.SignedString(secret)
	assert.NoError(t, err)
	jwtToken, claim2, err := jauth.DecodeToken(signedToken)
	_ = jwtToken
	// apparently this still works
	require.NoError(t, err)
	assert.Equal(t, user1, claim2.Id)
}

// start with a val

// additional JWT test cases for 100% coverage
func TestJWTBadToken(t *testing.T) {
	jauth := tlsserver.NewJWTAuthenticator(nil, func(login, pass string) bool {
		assert.Fail(t, "Should never reach here")
		return false
	})

	// start with a valid access token
	req, _ := http.NewRequest("GET", "badurl", nil)
	expTime := time.Now().Add(time.Second * 100)
	accessToken, refreshToken, err := jauth.CreateJWTTokens("user1", expTime)
	assert.NoError(t, err)
	assert.NotNil(t, accessToken)
	assert.NotNil(t, refreshToken)
	req.Header.Add("Authorization", "bearer "+accessToken)
	userID, match := jauth.AuthenticateRequest(nil, req)
	assert.True(t, match)
	assert.NotEmpty(t, userID)

	// missing auth token
	req = &http.Request{}
	_, match = jauth.AuthenticateRequest(nil, req)
	assert.False(t, match)

	// invalid auth header
	req, _ = http.NewRequest("GET", "badurl", nil)
	req.Header.Add("Authorization", "")
	_, match = jauth.AuthenticateRequest(nil, req)
	assert.False(t, match)

	// incomplete bearer token
	req, _ = http.NewRequest("GET", "badurl", nil)
	req.Header.Add("Authorization", "bearer")
	_, match = jauth.AuthenticateRequest(nil, req)
	assert.False(t, match)

	// invalid bearer token
	req, _ = http.NewRequest("GET", "badurl", nil)
	req.Header.Add("Authorization", "bearer invalidtoken")
	_, match = jauth.AuthenticateRequest(nil, req)
	assert.False(t, match)
}

func TestBadLogin(t *testing.T) {
	jauth := tlsserver.NewJWTAuthenticator(nil, func(login, pass string) bool {
		assert.Fail(t, "Should never reach here")
		return false
	})
	body := http.NoBody
	req, err := http.NewRequest("GET", "someurl", body)
	assert.NoError(t, err)
	resp := httptest.NewRecorder()
	jauth.HandleJWTLogin(resp, req)
}
