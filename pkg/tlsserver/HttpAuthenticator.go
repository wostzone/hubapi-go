package tlsserver

import (
	"net/http"
)

const AuthTypeBasic = "basic"
const AuthTypeDigest = "digest"
const AuthTypeJWT = "jwt"
const AuthTypeCert = "cert"

// HttpAuthenticator chains the selected authenticators
type HttpAuthenticator struct {
	BasicAuth *BasicAuthenticator
	CertAuth  *CertAuthenticator
	JwtAuth   *JWTAuthenticator
}

// AuthenticateRequest
// Checks in order: client certificate, JWT bearer, Basic
// Returns the authenticated userID or an error if authentication failed
func (hauth *HttpAuthenticator) AuthenticateRequest(resp http.ResponseWriter, req *http.Request) (userID string, match bool) {
	if hauth.CertAuth != nil {
		// FIXME: how to differentiate between cert auth and other for authorization
		// workaround: plugin certificates do not have a name
		userID, match = hauth.CertAuth.AuthenticateRequest(resp, req)
		if match {
			return userID, match
		}
	}
	if hauth.JwtAuth != nil {
		userID, match = hauth.JwtAuth.AuthenticateRequest(resp, req)
		if match {
			return userID, match
		}
	}
	if hauth.BasicAuth != nil {
		userID, match = hauth.BasicAuth.AuthenticateRequest(resp, req)
		if match {
			return userID, match
		}
	}
	return userID, false
}

// Create a new HTTP authenticator
// Use .AuthenticateRequest() to authenticate the incoming request
//  verifyUsernamePassword is the handler that validates the loginID and secret
func NewHttpAuthenticator(
	verifyUsernamePassword func(loginID, secret string) bool) *HttpAuthenticator {
	ha := &HttpAuthenticator{
		BasicAuth: NewBasicAuthenticator(verifyUsernamePassword),
		JwtAuth:   NewJWTAuthenticator(nil, verifyUsernamePassword),
		CertAuth:  NewCertAuthenticator(),
	}
	return ha
}
