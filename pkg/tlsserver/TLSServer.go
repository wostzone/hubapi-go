// Package servetls with TLS server for use by plugins and testing
package tlsserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubclient-go/pkg/tlsclient"
)

// Simple TLS Server
type TLSServer struct {
	address           string
	port              uint
	caCert            *x509.Certificate
	serverCert        *tls.Certificate
	httpServer        *http.Server
	router            *mux.Router
	httpAuthenticator *HttpAuthenticator
}

// AddHandler adds a new handler for a path.
//
// The server authenticates the request before passing it to this handler.
// The handler's userID is that of the authenticated user, and is intended for authorization of the request.
// If authentication is not enabled then the userID is empty.
//
//  path to listen on. This supports wildcards
//  handler to invoke with the request. The userID is only provided when an authenticator is used
func (srv *TLSServer) AddHandler(path string,
	handler func(userID string, resp http.ResponseWriter, req *http.Request)) {

	// do we need a local copy of handler? not sure
	local_handler := handler
	if srv.httpAuthenticator != nil {
		// the internal authenticator performs certificate based, basic or jwt token authentication if needed
		srv.router.HandleFunc(path, func(resp http.ResponseWriter, req *http.Request) {
			// valid authentication without userID means a plugin certificate was used which is always authorized
			userID, match := srv.httpAuthenticator.AuthenticateRequest(resp, req)
			if !match {
				msg := fmt.Sprintf("TLSServer.HandleFunc %s: User '%s' from %s is unauthorized", path, userID, req.RemoteAddr)
				logrus.Infof("%s", msg)
				srv.WriteForbidden(resp, msg)
			} else {
				local_handler(userID, resp, req)
			}
		})
	} else {
		srv.router.HandleFunc(path, func(resp http.ResponseWriter, req *http.Request) {
			// no authenticator means we don't know who the user is
			local_handler("", resp, req)
		})
	}
}

// Start the TLS server using the provided CA and Server certificates.
// The server will request but not require a client certificate. If one is provided it must be valid.
func (srv *TLSServer) Start() error {
	var err error

	logrus.Infof("Starting TLS server on address: %s:%d", srv.address, srv.port)
	if srv.caCert == nil || srv.serverCert == nil {
		err := fmt.Errorf("missing CA or server certificate")
		logrus.Error(err)
		return err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(srv.caCert)

	serverTLSConf := &tls.Config{
		Certificates:       []tls.Certificate{*srv.serverCert},
		ClientAuth:         tls.VerifyClientCertIfGiven,
		ClientCAs:          caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	srv.httpServer = &http.Server{
		Addr: fmt.Sprintf("%s:%d", srv.address, srv.port),
		// ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		// WriteTimeout: 10 * time.Second,
		Handler:   srv.router,
		TLSConfig: serverTLSConf,
	}
	// mutex to capture error result in case startup in the background failed
	go func() {
		// serverTLSConf contains certificate and key
		err2 := srv.httpServer.ListenAndServeTLS("", "")
		if err2 != nil && err2 != http.ErrServerClosed {
			err = fmt.Errorf("TLSServer.Start: ListenAndServeTLS: %s", err2)
			logrus.Error(err)
		}
	}()
	// Make sure the server is listening before continuing
	time.Sleep(time.Second)
	return err
}

// Stop the TLS server and close all connections
func (srv *TLSServer) Stop() {
	logrus.Infof("TLSServer.Stop: Stopping TLS server")

	if srv.httpServer != nil {
		srv.httpServer.Shutdown(context.Background())
	}
}

// Create a new TLS Server instance. Use Start/Stop to run and close connections
// The authenticator is optional to authenticate and authorize each of the requests. It returns
// an error if auth fails, after it writes the error message to the ResponseWriter.
//
//  address          server listening address
//  port             listening port
//  caCertPath       CA certificate
//  serverCertPath   Server certificate of this server
//  serverKeyPath    Server key of this server
//  authenticator    optional, function to authenticate requests
//
// returns TLS server for handling requests
func NewTLSServer(address string, port uint,
	serverCert *tls.Certificate, caCert *x509.Certificate,
	authenticator func(userID, secret string) bool) *TLSServer {
	// for now the JWT login path is fixed. Once a use-case comes up that requires something configurable
	// this can be updated.
	jwtLoginPath := tlsclient.DefaultJWTLoginPath
	hwtRefreshPath := tlsclient.DefaultJWTRefreshPath

	srv := &TLSServer{
		router:     mux.NewRouter(),
		caCert:     caCert,
		serverCert: serverCert,
	}
	if authenticator != nil {
		srv.httpAuthenticator = NewHttpAuthenticator(authenticator)
		srv.router.HandleFunc(jwtLoginPath, srv.httpAuthenticator.JwtAuth.HandleJWTLogin)
		srv.router.HandleFunc(hwtRefreshPath, srv.httpAuthenticator.JwtAuth.HandleJWTRefresh)
	}
	srv.address = address
	srv.port = port
	return srv
}
