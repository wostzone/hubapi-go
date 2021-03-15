package wostmqtt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/api/pkg/wostmqtt"
)

const certFolder = "/etc/mosquitto/certs"
const clientID1 = "client1"
const hostname = "localhost:8883"

// !!! THIS REQUIRES A RUNNING MQTT SERVER ON LOCALHOST !!!

func TestPublishTD(t *testing.T) {
	thingID := "thing1"
	td1 := []byte("{this is a td}")
	wmc := wostmqtt.NewWostMqtt(certFolder)

	err := wmc.Start(hostname, clientID1)
	assert.NoError(t, err)

	err = wmc.PublishTD(thingID, td1)
	assert.NoError(t, err)

	// TODO, check if it was received by a consumer

	wmc.Stop()
}

func TestPublishPropertyValues(t *testing.T) {
	thingID := "thing1"
	propValues := []byte("{property values}")
	wmc := wostmqtt.NewWostMqtt(certFolder)

	err := wmc.Start(hostname, clientID1)
	assert.NoError(t, err)

	err = wmc.PublishPropertyValues(thingID, propValues)
	assert.NoError(t, err)

	// TODO, check if it was received by a consumer

	wmc.Stop()
}

func TestPublishEvent(t *testing.T) {
	thingID := "thing1"
	event1 := []byte("{event}")
	wmc := wostmqtt.NewWostMqtt(certFolder)

	err := wmc.Start(hostname, clientID1)
	assert.NoError(t, err)

	err = wmc.PublishEvent(thingID, event1)
	assert.NoError(t, err)

	// TODO, check if it was received by a consumer

	wmc.Stop()
}

func TestPublishAction(t *testing.T) {
	thingID := "thing1"
	action1 := []byte("{action}")
	wmc := wostmqtt.NewWostMqtt(certFolder)

	err := wmc.Start(hostname, clientID1)
	assert.NoError(t, err)

	err = wmc.PublishAction(thingID, action1)
	assert.NoError(t, err)

	// TODO, check if it was received by the Thing

	wmc.Stop()
}
