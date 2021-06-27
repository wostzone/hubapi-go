// Package servetls with TLS server for use by plugins and testing
package tlsserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
)

// Simple TLS Server
type TLSServer struct {
	listenAddress string
	certFolder    string
	httpServer    *http.Server
	router        *mux.Router
}

// AddHandler adds a new handler for a path.
// Start must be called first
// The server is configured to verify provided client certificate but does not require
// that the client uses one. It is up to the application to decide which paths can be used
// without client certificate and which paths do require a client certificate.
// See the http.Request object to determine if a client cert is provided.
//
//  path to listen on. This supports wildcards
//  handler to invoke with the request
func (srv *TLSServer) AddHandler(path string, handler func(http.ResponseWriter, *http.Request)) {
	if srv.router == nil {
		logrus.Errorf("TLSServer.AddHandler Error. Start has to be called first")
		return
	}
	srv.router.HandleFunc(path, handler)
}

// Start the TLS server using CA and server certificates from the certfolder
// A client certificate is requested but not required.
func (srv *TLSServer) Start() error {
	logrus.Infof("TLSServer.Start: Starting TLS server on address: %s", srv.listenAddress)
	srv.router = mux.NewRouter()

	caCertPath := path.Join(srv.certFolder, certsetup.CaCertFile)
	_, err := os.Stat(caCertPath)
	if os.IsNotExist(err) {
		logrus.Errorf("TLSServer.Start: Missing CA certificate %s", caCertPath)
		return err
	}
	serverCertPath := path.Join(srv.certFolder, certsetup.ServerCertFile)
	serverKeyPath := path.Join(srv.certFolder, certsetup.ServerKeyFile)
	serverCertPEM, err := ioutil.ReadFile(serverCertPath)
	serverKeyPEM, err2 := ioutil.ReadFile(serverKeyPath)
	serverCert, err3 := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil || err2 != nil || err3 != nil {
		logrus.Errorf("TLSServer.Start: Server certificate pair not found")
		return err
	}
	caCertPEM, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)

	serverTLSConf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		// ClientAuth: tls.RequireAnyClientCert, // Require CA signed cert
		// ClientAuth: tls.RequestClientCert, //optional
		ClientAuth: tls.VerifyClientCertIfGiven,
		// ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:          caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
		// VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// 	logrus.Infof("***TLS server VerifyPeerCertificate called")
		// 	return nil
		// },
	}

	srv.httpServer = &http.Server{
		Addr: srv.listenAddress,
		// ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		// WriteTimeout: 10 * time.Second,
		Handler:   srv.router,
		TLSConfig: serverTLSConf,
	}

	go func() {
		err2 := srv.httpServer.ListenAndServeTLS("", "")
		// err := cs.httpServer.ListenAndServeTLS(serverCertFile, serverKeyFile)
		if err2 != nil && err2 != http.ErrServerClosed {
			err = fmt.Errorf("TLSServer.Start: ListenAndServeTLS: %s", err2)
			logrus.Error(err)
			// logrus.Fatalf("ServeMsgBus.Start: ListenAndServeTLS error: %s", err)
		}
	}()
	// Make sure the server is listening before continuing
	// Not pretty but it handles it
	time.Sleep(time.Second)
	return nil
}

// Stop the TLS server and close all connections
func (srv *TLSServer) Stop() {
	// cs.updateMutex.Lock()
	// defer cs.updateMutex.Unlock()
	logrus.Infof("TLSServer.Stop: Stopping TLS server")

	if srv.httpServer != nil {
		srv.httpServer.Shutdown(context.Background())
	}
}

// Create a new TLS Server instance. Use Start/Stop to run and close connections
//  listenAddress listening address
//  certFolder folder with ca, server certs and key (see certsetup for certificate creation
// and standard naming.
//
// returns TLS server for handling requests
func NewTLSServer(listenAddress string, certFolder string) *TLSServer {
	srv := &TLSServer{}
	// get the certificates ready
	if certFolder == "" {
		certFolder = "./certs"
	}
	srv.certFolder = certFolder
	srv.listenAddress = listenAddress
	return srv
}
