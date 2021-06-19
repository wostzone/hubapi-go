// Package tlsclient with a simple TLS client helper for mutual authentication
package tlsclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
)

// Simple TLS Client
type TLSClient struct {
	address    string
	certFolder string
	httpClient *http.Client
	timeout    time.Duration
}

// GetOutboundInterface Get preferred outbound network interface of this machine
// Credits: https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
// and https://qiita.com/shaching/items/4c2ee8fd2914cce8687c
func GetOutboundInterface(address string) (interfaceName string, macAddress string, ipAddr net.IP) {

	// This dial command doesn't actually create a connection
	conn, err := net.Dial("udp", address)
	if err != nil {
		logrus.Errorf("GetOutboundInterface: %s", err)
		return "", "", nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ipAddr = localAddr.IP

	// find the first interface for this address
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {

		if addrs, err := interf.Addrs(); err == nil {
			for index, addr := range addrs {
				logrus.Debug("[", index, "]", interf.Name, ">", addr)

				// only interested in the name with current IP address
				if strings.Contains(addr.String(), ipAddr.String()) {
					logrus.Debug("GetOutboundInterface: Use name : ", interf.Name)
					interfaceName = interf.Name
					macAddress = fmt.Sprint(interf.HardwareAddr)
					break
				}
			}
		}
	}
	// netInterface, err = net.InterfaceByName(interfaceName)
	// macAddress = netInterface.HardwareAddr
	fmt.Println("MAC: ", macAddress)
	return
}

// testing
func (cl *TLSClient) Post(path string, msg interface{}) ([]byte, error) {
	url := fmt.Sprintf("https://%s/%s", cl.address, path)

	bodyBytes, _ := json.Marshal(msg)
	body := bytes.NewReader(bodyBytes)
	resp, err := cl.httpClient.Post(url, "", body)
	data, err := ioutil.ReadAll(resp.Body)

	return data, err
}

// invoke a HTTPS method and read response
//  client is the http client to use
//  method: GET, PUT, POST, ...
//  addr the server to connect to
//  path to invoke
//  msg body to include
func (cl *TLSClient) Invoke(method string, path string, msg interface{}) ([]byte, error) {
	var body io.Reader
	var err error
	var req *http.Request

	if cl == nil || cl.httpClient == nil {
		logrus.Errorf("Invoke: '%s'. Client is not started", path)
		return nil, errors.New("Invoke: client is not started")
	}
	logrus.Infof("TLSClient.Invoke: %s: %s", method, path)

	// careful, a double // in the path causes a 301 and changes post to get
	url := fmt.Sprintf("https://%s%s", cl.address, path)
	if msg != nil {
		bodyBytes, _ := json.Marshal(msg)
		body = bytes.NewReader(bodyBytes)
	}
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	// cl.httpClient.
	resp, err := cl.httpClient.Do(req)
	if err != nil {
		logrus.Errorf("TLSClient.Invoke: %s %s: %s", method, path, err)
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		msg := fmt.Sprintf("%s: %s", resp.Status, respBody)
		if resp.Status == "" {
			msg = fmt.Sprintf("%d (%s): %s", resp.StatusCode, resp.Status, respBody)
		}
		err = errors.New(msg)
	}
	if err != nil {
		logrus.Errorf("TLSClient:Invoke: Error %s %s: %s", method, path, err)
		return nil, err
	}
	return respBody, err
}

// Start the client.
// 1. If a CA certificate is not available then insecure-skip-verify is used to allow
// connection to an unverified server (leap of faith)
// 2. Mutual TLS authentication is used when both CA and client certificates are available
func (cl *TLSClient) Start() (err error) {
	var clientCertList []tls.Certificate = []tls.Certificate{}
	var checkServerCert = false

	// Use CA certificate for server authentication if it exists
	caCertPEM, err := certsetup.LoadPEM(cl.certFolder, certsetup.CaCertFile)
	caCertPool := x509.NewCertPool()
	if err == nil {
		logrus.Infof("TLSClient.Start: Using CA certificate in '%s' for server verification", certsetup.CaCertFile)
		caCertPool.AppendCertsFromPEM([]byte(caCertPEM))
		checkServerCert = true
	} else {
		logrus.Infof("TLSClient.Start, No CA certificate at '%s/%s'. InsecureSkipVerify used", cl.certFolder, certsetup.CaCertFile)
	}

	// Use client certificate for mutual authentication with the server
	clientCertPEM, _ := certsetup.LoadPEM(cl.certFolder, certsetup.ClientCertFile)
	clientKeyPEM, _ := certsetup.LoadPEM(cl.certFolder, certsetup.ClientKeyFile)
	if clientCertPEM != "" && clientKeyPEM != "" {
		logrus.Infof("TLSClient.Start: Using client certificate from %s for mutual auth", certsetup.ClientCertFile)
		clientCert, err := tls.X509KeyPair([]byte(clientCertPEM), []byte(clientKeyPEM))
		if err != nil {
			logrus.Error("TLSClient.Start: Invalid client certificate or key: ", err)
			return err
		}
		clientCertList = append(clientCertList, clientCert)
	} else {
		logrus.Infof("TLSClient.Start, No client key/certificate in '%s/%s'. Mutual auth disabled.", cl.certFolder, certsetup.ClientKeyFile)
	}
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       clientCertList,
		InsecureSkipVerify: !checkServerCert,
	}
	// tlsTransport := http.Transport{
	// 	TLSClientConfig: tlsConfig,
	// }
	tlsTransport := http.DefaultTransport
	tlsTransport.(*http.Transport).TLSClientConfig = tlsConfig

	cl.httpClient = &http.Client{
		Transport: tlsTransport,
		Timeout:   cl.timeout,
	}
	return nil
}

// Stop the TLS client
func (cl *TLSClient) Stop() {
	// cs.updateMutex.Lock()
	// defer cs.updateMutex.Unlock()
	logrus.Infof("TLSClient.Stop: Stopping TLS client")

	if cl.httpClient != nil {
		cl.httpClient.CloseIdleConnections()
		cl.httpClient = nil
	}
}

// Create a new TLS Client instance.
// If the certFolder contains a CA certificate, then server authentication is used.
// If the certFolder also contains a client certificate and key then the client is
// configured for mutual authentication.
// Use Start/Stop to run and close connections
//  address address of the server
//  certFolder folder with ca, client certs and key. (see cersetup for standard names)
// returns TLS client for submitting requests
func NewTLSClient(address string, certFolder string) *TLSClient {
	cl := &TLSClient{
		address:    address,
		certFolder: certFolder,
		timeout:    time.Second,
	}
	return cl
}
