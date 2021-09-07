package tlsserver

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubclient-go/pkg/tlsclient"
)

const JWTIssuer = "tlsserver.JWTAuthenticator"
const JwtRefreshCookieName = "authtoken"

// this is temporary while figuring things out
type JwtClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// LoginCredentials
type JWTLoginCredentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// JWTAuthenticator creates and verifies JWT tokens when a valid login/pw is provided
// This authentication implementation uses a JWT access and refresh token pair for access and refresh
// of authentication tokens.
//
// The login step verifies the given credentials and issues an access token and refresh token.
// The refresh token is also stored in a secure client cookie to save the hassle for the client storage.
//
// This adapter keeps the JWT signing secret in memory. As a result all tokens will be invalidated
// after a restart which in turn means that the user must log in again. This is intentional.
//
// In order to use JWT authentication, the client must do the following:
// 1. On first use, login through the login endpoint. This returns the access and refresh tokens
// 2. Replace the access token in the authorization header
// 3. If the access token expires invoke the refresh endpoint and receive a new set of tokens.
// 4. Replace the access token in the authorization header again. Same as pt 2.
// 5. If access or refresh fails, prompt the user to log in again with credentials
// 6. If the user logs out, invoke the logout endpoint to remove the refresh token.
//
// With the use of login, refresh and logout comes the need for an endpoint for each purpose. Each
// endpoint must be attached to the handler via the router. For example:
//  > router.HandleFunc("/login", .HandleJWTLogin)  body=JwtAuthLogin{}
//  > router.HandleFunc("/logout", .HandleJWTLogout)  cookie=refresh token
//  > router.HandleFunc("/refresh", .HandleJWTRefresh)  cookie=refresh token
//
// This service is a protocol adapter, not an authentication service. As such the handler
// for credential verification must be provided.
//
// *A note on oauth2*:
//
// This adapter does not implement oauth2 as oauth2 separates the identity provider from the
// resource provider. The primary purpose of this adapter is to authenticate requests made to the
// resource provider, not to validate credentials. For the sake of simplicity this adapter passes
// credentials provided through the /login endpoint to a separate credential verification callback.
// When that callback validates the credentials a set of tokens is returned.
//
// In oauth2 the identity provider is completely separate and the resource provider never sees
// the credentials. Instead it is provided tokens that it verifies with the identity provider.
// Since WoST hubs often do not have a fixed DNS, the use of login with google/fb/etc is not possible
// as these services rely on a redirect and a fixed configured URL. It would also rely on internet
// access which in case of WoST Hubs is optional.
//
// If there is a good use-case to extend WoST with a separate identity service, this adapter will
// be updated and might include OAuth support for those users with internet access.
//
// The application must use .AuthenticateRequest() to authenticate the incoming request using the
// access token.
//
type JWTAuthenticator struct {
	// the secrets verification handler
	verifyUsernamePassword func(username, password string) bool
	jwtKey                 []byte // secret for signing key

	accessTokenValidity  time.Duration
	refreshTokenValidity time.Duration

	// optional callback when an expired token is used
	// expiredTokenAlert func(claims *JwtClaims)
}

// AuthenticateRequest validates the access token
// The access token is provided in the Authorization field as the bearer token.
// Returns the authenticated user and true if there is a match, of false if authentication failed
func (jauth *JWTAuthenticator) AuthenticateRequest(resp http.ResponseWriter, req *http.Request) (userID string, match bool) {

	accessTokenString, err := jauth.GetBearerToken(req)
	if err != nil {
		// this just means JWT is not used
		logrus.Debugf("JWTAuthenticator: No bearer token in request %s '%s' from %s", req.Method, req.RequestURI, req.RemoteAddr)
		return "", false
	}
	// 	// try the cookie -> refresh
	// 	cookie, err := req.Cookie(JwtRefreshCookieName)
	// 	if err == nil && cookie != nil {
	// 		_ = cookie.Value
	// 		accessTokenString = cookie.Value
	// 	}
	// }
	jwtToken, claims, err := jauth.DecodeToken(accessTokenString)
	_ = claims
	if err != nil {
		logrus.Infof("JWTAuthenticator: Invalid access token in request %s '%s' from %s",
			req.Method, req.RequestURI, req.RemoteAddr)
		return "", false
	}
	// hoora
	logrus.Infof("JWTAuthenticator. Request by %s authenticated with valid JWT token", jwtToken.Header)
	return claims.Username, true
}

// CreateJWTTokens creates a new access and refresh token pair containing the username.
// The result is written to the response and a refresh token is set securely in a client cookie.
func (jauth *JWTAuthenticator) CreateJWTTokens(userID string, expTime time.Time) (accessToken string, refreshToken string, err error) {
	logrus.Infof("CreateJWTTokens for user '%s'", userID)
	accessExpTime := time.Now().Add(jauth.accessTokenValidity)
	// refreshExpTime := time.Now().Add(jauth.refreshTokenValidity)
	refreshExpTime := expTime

	// Create the JWT claims, which includes the username and expiry time
	accessClaims := &JwtClaims{
		Username: userID,
		StandardClaims: jwt.StandardClaims{
			Id:      userID,
			Issuer:  JWTIssuer,
			Subject: "accessToken",
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: accessExpTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	// Declare the token with the algorithm used for signing, and the claims
	jwtAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = jwtAccessToken.SignedString(jauth.jwtKey)
	if err != nil {
		return
	}

	// same for refresh token
	refreshClaims := &JwtClaims{
		Username: userID,
		StandardClaims: jwt.StandardClaims{
			Id:      userID,
			Issuer:  JWTIssuer,
			Subject: "refreshToken",
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: refreshExpTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	// Create the JWT string
	jwtRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = jwtRefreshToken.SignedString(jauth.jwtKey)
	return accessToken, refreshToken, err
}

// DecodeToken and return its claims
// Set error if token not valid
func (jauth *JWTAuthenticator) DecodeToken(tokenString string) (
	jwtToken *jwt.Token, claims *JwtClaims, err error) {

	claims = &JwtClaims{}
	jwtToken, err = jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			return jauth.jwtKey, nil
		})
	if err != nil || jwtToken == nil || !jwtToken.Valid {
		return nil, nil, fmt.Errorf("invalid JWT token. Err=%s", err)
	}
	err = jwtToken.Claims.Valid()
	if err != nil {
		return jwtToken, nil, fmt.Errorf("invalid JWT claims: err=%s", err)
	}
	claims = jwtToken.Claims.(*JwtClaims)

	return jwtToken, claims, nil
}

// GetBearerToken returns the bearer token from the Authorization header
// Returns an error if no token present or token isn't a bearer token
func (jauth *JWTAuthenticator) GetBearerToken(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("JWTAuthenticator: no Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("JWTAuthenticator: invalid Authorization header")
	}
	authType := strings.ToLower(parts[0])
	authTokenString := parts[1]
	if authType != "bearer" {
		return "", fmt.Errorf("JWTAuthenticator: not a bearer token")
	}
	return authTokenString, nil
}

// Handle a JWT login POST request.
// Attach this method to the router with the login route. For example:
//  > router.HandleFunc("/login", HandleJWTLogin)
// The body contains provided userID and password
// This:
//  1. returns a JWT access and refresh token pair
//  2. sets a secure, httpOnly, sameSite refresh cookie with the name 'JwtRefreshCookieName'
func (jauth *JWTAuthenticator) HandleJWTLogin(resp http.ResponseWriter, req *http.Request) {
	logrus.Infof("HttpAuthenticator.HandleJWTLogin")

	loginCred := JWTLoginCredentials{}
	err := json.NewDecoder(req.Body).Decode(&loginCred)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	// this is not an authentication provider. Use a callback for actual authentication
	match := jauth.verifyUsernamePassword(loginCred.Username, loginCred.Password)
	if !match {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	refreshExpTime := time.Now().Add(jauth.refreshTokenValidity)
	accessToken, refreshToken, err := jauth.CreateJWTTokens(loginCred.Username, refreshExpTime)

	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		logrus.Errorf("HttpAuthenticator.HandleJWTLogin: error %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	jauth.WriteJWTTokens(accessToken, refreshToken, refreshExpTime, resp)
}

// Handle a JWT refresh POST request.
// Attach this method to the router with the refresh route. For example:
//  > router.HandleFunc("/refresh", HandleJWTRefresh)
//
// A valid refresh token must be provided in the client cookie or set in the authorization header
//
// This:
//  1. Return unauthorized if no valid refresh token was found
//  2. returns a JWT access and refresh token pair if the refresh token was valid
//  3. sets a secure, httpOnly, sameSite refresh cookie with the name 'JwtRefreshCookieName'
func (jauth *JWTAuthenticator) HandleJWTRefresh(resp http.ResponseWriter, req *http.Request) {
	logrus.Infof("HttpAuthenticator.HandleJWTRefresh")
	var refreshTokenString string

	// validate the provided refresh token
	cookie, err := req.Cookie(JwtRefreshCookieName)
	if err == nil {
		refreshTokenString = cookie.Value
	} else {
		refreshTokenString, err = jauth.GetBearerToken(req)
	}
	// no refresh token found
	if err != nil || refreshTokenString == "" {
		resp.WriteHeader(http.StatusUnauthorized)
	}

	// is the token valid?
	_, claims, err := jauth.DecodeToken(refreshTokenString)
	if err != nil || claims.Id == "" {
		// refresh token is invalid. Authorization refused
		resp.WriteHeader(http.StatusUnauthorized)
	}

	refreshExpTime := time.Now().Add(jauth.refreshTokenValidity)
	accessToken, refreshToken, err := jauth.CreateJWTTokens(claims.Id, refreshExpTime)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		logrus.Errorf("HttpAuthenticator.HandleJWTLogin: error %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	jauth.WriteJWTTokens(accessToken, refreshToken, refreshExpTime, resp)

}

// WriteJWTTokens writes the access and refresh tokens as response message and in a
// secure client cookie. The cookieExpTime should be set to the refresh token expiration time.
func (jauth *JWTAuthenticator) WriteJWTTokens(
	accessToken string, refreshToken string, cookieExpTime time.Time, resp http.ResponseWriter) error {

	// Set a client cookie for refresh "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(resp, &http.Cookie{
		Name:     JwtRefreshCookieName,
		Value:    refreshToken,
		Expires:  cookieExpTime,
		HttpOnly: true, // prevent XSS attack (client cant read value)
		Secure:   true, //
		// assume that the client service/website runs on the same server to use cookies
		SameSite: http.SameSiteStrictMode,
	})

	response := tlsclient.JwtAuthResponse{AccessToken: accessToken, RefreshToken: refreshToken}
	responseMsg, _ := json.Marshal(response)
	_, err := resp.Write(responseMsg)
	return err
}

// Create a new JWT authenticator adapter.
//
//  secret for generating tokens, or nil to generate a random 64 byte secret
//  verifyUsernamePassword is the handler that validates the loginID and secret
func NewJWTAuthenticator(
	secret []byte, verifyUsernamePassword func(loginID, secret string) bool) *JWTAuthenticator {
	if secret == nil {
		secret = make([]byte, 64)
		rand.Read(secret)
	}
	ja := &JWTAuthenticator{
		verifyUsernamePassword: verifyUsernamePassword,
		jwtKey:                 secret,
		accessTokenValidity:    15 * time.Minute,
		refreshTokenValidity:   10 * 24 * time.Hour,
	}
	return ja
}
