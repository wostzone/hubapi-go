// Package servetls with TLS server for use by plugins and testing
package servetls

import (
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
	"github.com/wostzone/hubapi-go/pkg/certsetup"
)

// StartTLSServer starts a HTTP TLS server using CA and server certificates from the certfolder
//  listenAddress listening address
//  certFolder folder with ca, server certs and key (see cersetup for standard names)
// returns TLS server for handling requests
func StartTLSServer(listenAddress string, certFolder string) (tlsServer *http.Server, err error) {
	router := mux.NewRouter()

	// get the certificates ready
	if certFolder == "" {
		certFolder = "./certs"
	}
	_, err = os.Stat(certFolder)
	if os.IsNotExist(err) {
		logrus.Errorf("Missing certificate folder %s", certFolder)
		return nil, err
	}
	caCertPath := path.Join(certFolder, certsetup.CaCertFile)
	serverCertPath := path.Join(certFolder, certsetup.ServerCertFile)
	serverKeyPath := path.Join(certFolder, certsetup.ServerKeyFile)
	serverCertPEM, err := ioutil.ReadFile(serverCertPath)
	serverKeyPEM, err2 := ioutil.ReadFile(serverKeyPath)
	serverCert, err3 := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil || err2 != nil || err3 != nil {
		logrus.Errorf("Server certificate pair not found")
		return nil, err
	}
	caCertPEM, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)

	serverTLSConf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAnyClientCert,
		ClientCAs:    caCertPool,
	}

	tlsServer = &http.Server{
		Addr: listenAddress,
		// ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		// WriteTimeout: 10 * time.Second,
		Handler:   router,
		TLSConfig: serverTLSConf,
	}

	go func() {
		err2 := tlsServer.ListenAndServeTLS("", "")
		// err := cs.httpServer.ListenAndServeTLS(serverCertFile, serverKeyFile)
		if err2 != nil && err2 != http.ErrServerClosed {
			err = fmt.Errorf("ListenAndServeTLS: %s", err2)
			logrus.Error(err)
			// logrus.Fatalf("ServeMsgBus.Start: ListenAndServeTLS error: %s", err)
		}
	}()
	// Make sure the server is listening before continuing
	// Not pretty but it handles it
	time.Sleep(time.Second)
	return tlsServer, nil
}
