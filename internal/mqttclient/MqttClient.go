package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// ConnectionTimeoutSec constant with connection and reconnection timeouts
const ConnectionTimeoutSec = 20

// TLSPort is the default secure port to connect to mqtt
const TLSPort = 8883

// MqttClient client wrapper around pahoClient
// This addresses problems with reconnect and auto resubscribe while using clean session
type MqttClient struct {
	clientID string // unique ID of the client (used for logging)
	hostPort string // host:port of server to connect to
	pubQos   byte
	subQos   byte
	timeout  int // connection timeout in seconds before giving up. 0 is never
	//
	isRunning           bool                          // listen for messages while running
	pahoClient          pahomqtt.Client               // Paho MQTT Client
	subscriptions       map[string]*TopicSubscription // map of TopicSubscription for re-subscribing after reconnect
	tlsVerifyServerCert bool                          // verify the server certificate, this requires a Root CA signed cert
	tlsCACertFile       string                        // path to CA certificate
	// clientCertFile      string                        // optional client certificate file
	// clientKeyFile       string                        // optional client cert key file
	updateMutex *sync.Mutex // mutex for async updating of subscriptions
}

// TopicSubscription holds subscriptions to restore after disconnect
type TopicSubscription struct {
	topic     string
	handler   func(address string, message []byte)
	handlerID reflect.Value
	// token     pahomqtt.Token // for debugging
	// client *MqttClient //
}

// Connect to the MQTT broker
// If a previous connection exists then it is disconnected first. If no connection is possible
// this keeps retrying until the timeout is expired. With each retry a backoff period
// is increased until 120 seconds.
// The clientID is generated from the hostname and current timestamp.
//  clientCert to authenticate with client certificate. Use nil to authenticate with username/password
//  userName to authenticate with. Use "" to ignore
//  password to authenticate with. Use "" to ignore
func (mqttClient *MqttClient) Connect(clientCert *tls.Certificate, userName string, password string) error {
	// set config defaults

	// close existing connection
	if mqttClient.pahoClient != nil && mqttClient.pahoClient.IsConnected() {
		mqttClient.pahoClient.Disconnect(10 * ConnectionTimeoutSec)
	}

	brokerURL := fmt.Sprintf("tls://%s/", mqttClient.hostPort) // tcp://host:1883 ws://host:1883 tls://host:8883, tcps://awshost:8883/mqtt
	// brokerURL := fmt.Sprintf("tls://mqtt.eclipse.org:8883/")
	opts := pahomqtt.NewClientOptions()

	opts.AddBroker(brokerURL)
	opts.SetClientID(mqttClient.clientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetMaxReconnectInterval(60 * time.Second) // max wait 1 minute for a reconnect
	// Do not use MQTT persistence as not all brokers support it, and it causes problems on the broker if the client ID is
	// randomly generated. CleanSession disables persistence.
	opts.SetCleanSession(true)
	opts.SetKeepAlive(ConnectionTimeoutSec * time.Second) // pings to detect a disconnect. Use same as reconnect interval
	//opts.SetKeepAlive(60) // keepalive causes deadlock in v1.1.0. See github issue #126

	opts.SetOnConnectHandler(func(client pahomqtt.Client) {
		logrus.Warningf("MqttClient.onConnect: Connected to server at %s. Connected=%v. ClientId=%s",
			brokerURL, client.IsConnected(), mqttClient.clientID)
		// Subscribe to addresss already registered by the app on connect or reconnect
		mqttClient.resubscribe()
	})
	opts.SetConnectionLostHandler(func(client pahomqtt.Client, err error) {
		logrus.Warningf("MqttClient.onConnectionLost: Disconnected from server %s. Error %s, ClientId=%s",
			brokerURL, err, mqttClient.clientID)
	})
	// if lastWillAddress != "" {
	// 	opts.SetWill(lastWillAddress, lastWillValue, 1, false)
	// }
	// Use TLS if a CA certificate is given
	var rootCA *x509.CertPool
	if mqttClient.tlsCACertFile != "" {
		rootCA = x509.NewCertPool()
		caCertPEM, err := ioutil.ReadFile(mqttClient.tlsCACertFile)
		if err != nil {
			logrus.Errorf("MqttClient.Connect: Unable to read CA certificate chain: %s. Ignored.", err)
		}
		rootCA.AppendCertsFromPEM([]byte(caCertPEM))

	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !mqttClient.tlsVerifyServerCert,
		RootCAs:            rootCA, // include the CA cert in the host root ca set
		// https://opium.io/blog/mqtt-in-go/
		ServerName: "", // hostname on the server certificate. How to get this?
	}
	if clientCert != nil {
		tlsConfig.Certificates = []tls.Certificate{*clientCert}
	}
	opts.SetTLSConfig(tlsConfig)
	opts.Username = userName
	opts.Password = password

	logrus.Infof("MqttClient.Connect: Connecting to MQTT server: %s with clientID %s"+
		" AutoReconnect and CleanSession are set.",
		brokerURL, mqttClient.clientID)

	// FIXME: PahoMqtt disconnects when sending a lot of messages, like on startup of some adapters.
	mqttClient.pahoClient = pahomqtt.NewClient(opts)

	// start listening for messages
	mqttClient.isRunning = true
	//go messenger.messageChanLoop()

	// Auto reconnect doesn't work for initial attempt: https://github.com/eclipse/paho.mqtt.golang/issues/77
	retryDelaySec := 1
	retryDuration := 0
	var err error
	for mqttClient.timeout == 0 || retryDuration < mqttClient.timeout {
		token := mqttClient.pahoClient.Connect()
		token.Wait()
		// Wait to give connection time to settle. Sending a lot of messages causes the connection to fail. Bug?
		time.Sleep(1000 * time.Millisecond)
		err = token.Error()
		if err == nil {
			break
		}
		retryDuration++

		logrus.Errorf("MqttClient.Connect: Connecting to broker on %s failed: %s. retrying in %d seconds.",
			brokerURL, token.Error(), retryDelaySec)
		sleepDuration := time.Duration(retryDelaySec)
		retryDuration += int(sleepDuration)
		time.Sleep(sleepDuration * time.Second)
		// slowly increment wait time
		if retryDelaySec < 120 {
			retryDelaySec++
		}
	}
	return err
}

// Connect to the MQTT broker using password authentication
// If a previous connection exists then it is disconnected first. If no connection is possible
// this keeps retrying until the timeout is expired. With each retry a backoff period
// is increased until 120 seconds.
//  userName to identify as
//  password credentials to identify with
func (mqttClient *MqttClient) ConnectWithPassword(userName string, password string) error {
	err := mqttClient.Connect(nil, userName, password)
	return err
}

// Connect to the MQTT broker using client certificate authentication
//  pluginID to connect as (no pass required)
//  clientCertFile optional client certificate to authenticate the client with the broker
//  clientKeyFile  optional client key to authenticate the client with the broker
func (mqttClient *MqttClient) ConnectWithClientCert(pluginID string, clientCertFile string, clientKeyFile string) error {
	if clientCertFile == "" || clientKeyFile == "" {
		err := mqttClient.Connect(nil, "", "")
		return err
	}
	clientCert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		logrus.Errorf("ConnectWithClientCert: Failed to connect. Error loading certificates: %s", err)
		return err
	}
	err = mqttClient.Connect(&clientCert, pluginID, "")
	return err

}

// Disconnect from the MQTT broker and unsubscribe from all addresss and set
// device state to disconnected
func (mqttClient *MqttClient) Disconnect() {
	mqttClient.updateMutex.Lock()
	mqttClient.isRunning = false
	mqttClient.updateMutex.Unlock()

	if mqttClient.pahoClient != nil {
		logrus.Warningf("MqttClient.Disconnect: Client %s", mqttClient.clientID)
		//messenger.publish("$state", "disconnected")
		time.Sleep(time.Second / 10) // Disconnect doesn't seem to wait for all messages. A small delay ahead helps
		mqttClient.pahoClient.Disconnect(10 * ConnectionTimeoutSec * 1000)
		mqttClient.pahoClient = nil

		mqttClient.subscriptions = make(map[string]*TopicSubscription, 0)
		//close(messenger.messageChannel)     // end the message handler loop
	}
}

// Wrapper for message handling to support multiple subscribers to one topic
// func (mqttClient *MqttClient) onMessage(c pahomqtt.Client, msg pahomqtt.Message) {
// 	topic := msg.Topic()
// 	payload := msg.Payload()

// 	logrus.Infof("MqttClient.onMessage. address=%s", topic)
// 	subscription := mqttClient.subscriptions[topic]
// 	if subscription == nil {
// 		logrus.Errorf("onMessage: no subscription for topic %s", topic)
// 		return
// 	}
// 	subscription.handler(topic, payload)
// }

// Publish a message to a topic address
func (mqttClient *MqttClient) Publish(topic string, message []byte) error {
	var err error

	if mqttClient.pahoClient == nil || !mqttClient.pahoClient.IsConnected() {
		logrus.Warnf("MqttClient.Publish: Unable to publish. No connection with server.")
		return errors.New("no connection with server")
	}
	logrus.Infof("MqttClient.Publish []byte: topic=%s, qos=%d",
		topic, mqttClient.pubQos)
	token := mqttClient.pahoClient.Publish(topic, mqttClient.pubQos, false, message)

	err = token.Error()
	if err != nil {
		// TODO: confirm that with qos=1 the message is sent after reconnect
		logrus.Warnf("MqttClient.Publish: Error during publish on address %s: %v", topic, err)
		//return err
	}
	return err
}

// subscribe to addresss after establishing connection
// The application can already subscribe to addresss before the connection is established. If connection is lost then
// this will re-subscribe to those addresss as PahoMqtt drops the subscriptions after disconnect.
//
func (mqttClient *MqttClient) resubscribe() {
	// prevent simultaneous access to subscriptions
	mqttClient.updateMutex.Lock()
	defer mqttClient.updateMutex.Unlock()

	logrus.Infof("MqttClient.resubscribe to %d addresess", len(mqttClient.subscriptions))
	for topic, subscription := range mqttClient.subscriptions {
		// clear existing subscription in case it is still there
		mqttClient.pahoClient.Unsubscribe(topic)

		logrus.Debugf("MqttClient.resubscribe: address %s", topic)
		// create a new variable to hold the subscription in the closure
		mqttClient.pahoClient.Subscribe(
			topic, mqttClient.pubQos,
			func(c pahomqtt.Client, msg pahomqtt.Message) {
				topic := msg.Topic()
				payload := msg.Payload()

				logrus.Infof("MqttClient.resubscribe.onMessage. address=%s", topic)
				subscription.handler(topic, payload)
			})

		// token := messenger.pahoClient.Subscribe(newSubscr.topic, messenger.pubQos, newSubscr.onMessage)
		//token := messenger.pahoClient.Subscribe(newSubscr.address, newSubscr.qos, func (c pahomqtt.Client, msg pahomqtt.Message) {
		//logrus.Infof("mqtt.resubscribe.onMessage: address %s, subscription %s", msg.Topic(), newSubscr.address)
		//newSubscr.onMessage(c, msg)
		//})
		// newSubscr.token = token
	}
	logrus.Debugf("MqttClient.resubscribe complete")
}

// Set the connection timeout for the client.
// Must be invoked before the connect() call
func (MqttClient *MqttClient) SetTimeout(sec int) {
	MqttClient.timeout = sec
}

// Subscribe to a address
// If a subscription already exists, it is replaced.
// topic: address to subscribe to. This supports mqtt wildcards such as + and #
// handler: callback handler.
func (mqttClient *MqttClient) Subscribe(
	topic string, handler func(address string, message []byte)) {
	handlerID := reflect.ValueOf(handler)
	subscription := &TopicSubscription{
		topic:     topic,
		handler:   handler,
		handlerID: handlerID,
	}
	logrus.Infof("MqttClient.Subscribe: topic %s, qos %d", topic, mqttClient.subQos)

	mqttClient.updateMutex.Lock()
	defer mqttClient.updateMutex.Unlock()

	// if mqttClient.subscriptions[topic] != nil {
	// 	logrus.Warningf("Subscribe: Existing subscription to %s is replaced", topic)
	// } else {
	if mqttClient.pahoClient != nil {
		// save handler on the stack
		subscribedHandler := handler
		mqttClient.pahoClient.Subscribe(topic, mqttClient.subQos,
			func(c pahomqtt.Client, msg pahomqtt.Message) {
				topic := msg.Topic()
				payload := msg.Payload()

				logrus.Infof("MqttClient.onMessage. address=%s", topic)
				subscribedHandler(topic, payload)
			})
	}
	// }
	mqttClient.subscriptions[topic] = subscription
}

// Unsubscribe a topic and handler
// if handler is nil then all subscribers to the topic are removed
func (mqttClient *MqttClient) Unsubscribe(topic string) {
	logrus.Infof("MqttClient.Unsubscribe: topic %s", topic)

	// messenger.publishMutex.Lock()

	subscription := mqttClient.subscriptions[topic]
	if subscription == nil {
		// nothing to unsubscribe
		logrus.Warningf("Unsubscribe: Subscription on topic %s didn't exist. Ignored", topic)
		return
	}

	if mqttClient.pahoClient != nil {
		mqttClient.pahoClient.Unsubscribe(topic)
		mqttClient.subscriptions[topic] = nil
	}
}

// NewMqttClient creates a new MQTT messenger instance
// The clientCertFile and clientKeyFile are optional. If provided then they must be signed
// by the CA used by the broker, so that the broker can authenticate the client. Leave empty when
// not using client certificates.
//  hostPort to connect to
//  caCertFile must contain the server CA certificate filename
func NewMqttClient(hostPort string, caCertFile string) *MqttClient {

	// ClientID defaults to hostname-millisecondsSinceEpoc
	hostName, _ := os.Hostname()
	timeStamp := time.Now().UnixNano() / 1000000
	clientID := fmt.Sprintf("%s-%d", hostName, timeStamp)

	messenger := &MqttClient{
		clientID:      clientID,
		hostPort:      hostPort,
		pubQos:        1,
		subQos:        1,
		pahoClient:    nil,
		subscriptions: make(map[string]*TopicSubscription, 0),
		//messageChannel: make(chan *IncomingMessage),
		tlsCACertFile:       caCertFile,
		tlsVerifyServerCert: true,
		updateMutex:         &sync.Mutex{},
	}
	// guarantee unique ID ... okay this is ugly
	time.Sleep(time.Millisecond)
	return messenger
}
