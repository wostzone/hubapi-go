package tlsserver_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hubclient-go/pkg/certs"
	"github.com/wostzone/hubclient-go/pkg/config"
	"github.com/wostzone/hubclient-go/pkg/tlsclient"
	"github.com/wostzone/hubserve-go/pkg/certsetup"
	"github.com/wostzone/hubserve-go/pkg/tlsserver"
)

var serverAddress string
var serverPort uint = 4444
var clientHostPort string
var caCert *x509.Certificate
var pluginCert *tls.Certificate

// These are set in TestMain
var homeFolder string
var serverCertFolder string

var caCertPath string
var caKeyPath string
var pluginCertPath string
var pluginKeyPath string
var serverCertPath string
var serverKeyPath string

// TestMain runs a http server
// Used for all test cases in this package
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of httpauthhandler ------")
	// serverAddress = hubnet.GetOutboundIP("").String()
	// use the localhost interface for testing
	serverAddress = "127.0.0.1"
	hostnames := []string{serverAddress}
	clientHostPort = fmt.Sprintf("%s:%d", serverAddress, serverPort)

	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	serverCertFolder = path.Join(homeFolder, "certs")

	certsetup.CreateCertificateBundle(hostnames, serverCertFolder)
	caCertPath = path.Join(serverCertFolder, config.DefaultCaCertFile)
	caKeyPath = path.Join(serverCertFolder, config.DefaultCaKeyFile)
	// caKey, _ = certs.LoadKeysFromPEM(caKeyPath)
	caCert, _ = certs.LoadX509CertFromPEM(caCertPath)

	serverCertPath = path.Join(serverCertFolder, config.DefaultServerCertFile)
	serverKeyPath = path.Join(serverCertFolder, config.DefaultServerKeyFile)

	pluginCertPath = path.Join(serverCertFolder, config.DefaultPluginCertFile)
	pluginKeyPath = path.Join(serverCertFolder, config.DefaultPluginKeyFile)
	pluginCert, _ = certs.LoadTLSCertFromPEM(pluginCertPath, pluginKeyPath)
	res := m.Run()

	time.Sleep(time.Second)
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, caCertPath, nil)
	err := srv.Start()
	assert.NoError(t, err)
	srv.Stop()
}

func TestNoCA(t *testing.T) {
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, "", nil)
	err := srv.Start()
	assert.Error(t, err)
	srv.Stop()
}
func TestBadCert(t *testing.T) {
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverCertPath, caCertPath, nil)
	err := srv.Start()
	assert.Error(t, err)
	srv.Stop()
}

// Connect without authentication
func TestNoAuth(t *testing.T) {
	path1 := "/hello"
	path1Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, caCertPath, nil)

	// handler can be added any time
	srv.AddHandler(path1, func(string, http.ResponseWriter, *http.Request) {
		logrus.Infof("TestAuthCert: path1 hit")
		path1Hit++
	})
	err := srv.Start()
	assert.NoError(t, err)

	cl := tlsclient.NewTLSClient(clientHostPort, nil)
	require.NoError(t, err)
	cl.ConnectWithClientCert(nil)
	_, err = cl.Get(path1)
	assert.NoError(t, err)
	assert.Equal(t, 1, path1Hit)

	cl.Close()
	srv.Stop()
}

// Test with invalid login authentication
func TestUnauthorized(t *testing.T) {
	path1 := "/test1"
	loginID1 := "user1"
	password1 := "user1pass"

	// setup server and client environment
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, caCertPath,
		func(userID string, password string) bool {
			assert.Fail(t, "Did not expect the auth method to be invoked")
			return false
		})
	err := srv.Start()
	assert.NoError(t, err)
	//
	srv.AddHandler(path1, func(string, http.ResponseWriter, *http.Request) {
		logrus.Infof("TestNoAuth: path1 hit")
		assert.Fail(t, "did not expect the request to pass")
	})
	//
	cl := tlsclient.NewTLSClient(clientHostPort, caCert)
	assert.NoError(t, err)
	// AuthMethodNone will always succeed
	err = cl.ConnectWithLoginID(loginID1, password1, "", tlsclient.AuthMethodNone)
	assert.NoError(t, err)

	// request should fail as login failed
	_, err = cl.Get(path1)
	assert.Error(t, err)

	cl.Close()
	srv.Stop()
}

func TestCertAuth(t *testing.T) {
	path1 := "/hello"
	path1Hit := 0
	loginHit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, caCertPath, func(loginID1, password string) bool {
			loginHit++
			assert.Fail(t, "did not expect login check with cert auth")
			return false
		})
	err := srv.Start()
	assert.NoError(t, err)
	// handler can be added any time
	srv.AddHandler(path1, func(string, http.ResponseWriter, *http.Request) {
		logrus.Infof("TestAuthCert: path1 hit")
		path1Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, caCert)
	require.NoError(t, err)
	err = cl.ConnectWithClientCert(pluginCert)
	assert.NoError(t, err)
	_, err = cl.Get(path1)
	assert.NoError(t, err)
	assert.Equal(t, 1, path1Hit)

	cl.Close()
	srv.Stop()
}

// Test valid authentication using JWT
func TestJWTLogin(t *testing.T) {
	user1 := "user1"
	user1Pass := "pass1"
	loginHit := 0
	path2 := "/hello"
	path2Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, caCertPath, func(loginID1, password string) bool {
			loginHit++
			return loginID1 == user1 && password == user1Pass
		})
	err := srv.Start()
	assert.NoError(t, err)
	//
	srv.AddHandler(path2, func(userID string, resp http.ResponseWriter, req *http.Request) {
		path2Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, caCert)
	require.NoError(t, err)

	// first show that an incorrect password fails
	err = cl.ConnectWithLoginID(user1, "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, 1, loginHit)
	// this request should be unauthorized
	_, err = cl.Get(path2)
	assert.Error(t, err)
	assert.Equal(t, 0, path2Hit) // should not increase
	cl.Close()

	// try again with the correct password
	err = cl.ConnectWithLoginID(user1, user1Pass)
	assert.NoError(t, err)
	assert.Equal(t, 2, loginHit)

	// use access token
	_, err = cl.Get(path2)
	require.NoError(t, err)
	assert.Equal(t, 1, path2Hit)

	cl.Close()
	srv.Stop()
}

func TestJWTRefresh(t *testing.T) {
	user1 := "user1"
	user1Pass := "pass1"
	loginHit := 0
	path2 := "/hello"
	path2Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort, serverCertPath, serverKeyPath, caCertPath,
		func(loginID string, password string) bool {
			loginHit++
			return loginID == user1 && password == user1Pass
		})
	err := srv.Start()
	assert.NoError(t, err)
	//
	srv.AddHandler(path2, func(userID string, resp http.ResponseWriter, req *http.Request) {
		path2Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, caCert)
	require.NoError(t, err)

	err = cl.ConnectWithLoginID(user1, user1Pass)
	assert.NoError(t, err)
	assert.Equal(t, 1, loginHit)

	_, err = cl.RefreshJWTTokens("")
	assert.NoError(t, err)

	// use access token
	_, err = cl.Get(path2)
	require.NoError(t, err)
	assert.Equal(t, 1, path2Hit)
	srv.Stop()

}

func TestQueryParams(t *testing.T) {
	path2 := "/hello"
	path2Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, caCertPath, nil)
	err := srv.Start()
	assert.NoError(t, err)
	srv.AddHandler(path2, func(userID string, resp http.ResponseWriter, req *http.Request) {
		// query string
		q1 := srv.GetQueryString(req, "query1", "")
		assert.Equal(t, "bob", q1)
		// fail not a number
		_, err := srv.GetQueryInt(req, "query1", 0) // not a number
		assert.Error(t, err)
		// query of number
		q2, _ := srv.GetQueryInt(req, "query2", 0)
		assert.Equal(t, 3, q2)
		// default should work
		q3 := srv.GetQueryString(req, "query3", "default")
		assert.Equal(t, "default", q3)
		// multiple parameters fail
		_, err = srv.GetQueryInt(req, "multi", 0)
		assert.Error(t, err)
		path2Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, caCert)
	require.NoError(t, err)
	err = cl.ConnectWithClientCert(pluginCert)
	assert.NoError(t, err)

	_, err = cl.Get(fmt.Sprintf("%s?query1=bob&query2=3&multi=a&multi=b", path2))
	assert.NoError(t, err)
	assert.Equal(t, 1, path2Hit)

	cl.Close()
	srv.Stop()
}

func TestWriteResponse(t *testing.T) {
	path2 := "/hello"
	path2Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, caCertPath, nil)
	err := srv.Start()
	assert.NoError(t, err)
	srv.AddHandler(path2, func(userID string, resp http.ResponseWriter, req *http.Request) {
		srv.WriteBadRequest(resp, "bad request")
		srv.WriteInternalError(resp, "internal error")
		srv.WriteNotFound(resp, "not found")
		srv.WriteNotImplemented(resp, "not implemented")
		srv.WriteUnauthorized(resp, "unauthorized")
		path2Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, caCert)
	require.NoError(t, err)
	err = cl.ConnectWithClientCert(pluginCert)
	assert.NoError(t, err)

	_, err = cl.Get(path2)
	assert.Error(t, err)
	assert.Equal(t, 1, path2Hit)

	cl.Close()
	srv.Stop()
}

func TestBadPort(t *testing.T) {
	srv := tlsserver.NewTLSServer(serverAddress, 1, // bad port
		serverCertPath, serverKeyPath, caCertPath, nil)

	err := srv.Start()
	assert.Error(t, err)
}

// Test BASIC authentication
func TestBasicAuth(t *testing.T) {
	path1 := "/test1"
	path1Hit := 0
	loginID1 := "user1"
	password1 := "user1pass"

	// setup server and client environment
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		serverCertPath, serverKeyPath, caCertPath,
		func(userID, password string) bool {
			path1Hit++
			return userID == loginID1 && password == password1
		})
	err := srv.Start()
	assert.NoError(t, err)
	//
	srv.AddHandler(path1, func(string, http.ResponseWriter, *http.Request) {
		logrus.Infof("TestBasicAuth: path1 hit")
		path1Hit++
	})
	//
	cl := tlsclient.NewTLSClient(clientHostPort, caCert)
	assert.NoError(t, err)
	err = cl.ConnectWithLoginID(loginID1, password1, "", tlsclient.AuthMethodBasic)
	assert.NoError(t, err)

	// test the auth with a GET request
	_, err = cl.Get(path1)
	assert.NoError(t, err)
	assert.Equal(t, 2, path1Hit)

	// test a failed login
	cl.Close()
	err = cl.ConnectWithLoginID(loginID1, "wrongpassword", "", tlsclient.AuthMethodBasic)
	assert.NoError(t, err)
	_, err = cl.Get(path1)
	assert.Error(t, err)
	assert.Equal(t, 3, path1Hit) // should not increase

	cl.Close()
	srv.Stop()
}
