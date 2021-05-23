package td

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi-go/api"
)

// tbd json-ld parsers:
// Most popular; https://github.com/xeipuuv/gojsonschema
// Other:  https://github.com/piprate/json-gold

// AddTDAction adds or replaces an action in the TD
//  td is a TD created with 'CreateTD'
//  name of action to add
//  action created with 'CreateAction'
func AddTDAction(td map[string]interface{}, name string, action interface{}) {
	actions := td[api.WoTActions].(map[string]interface{})
	if action == nil {
		logrus.Errorf("Add action '%s' to TD. Action is nil", name)
	} else {
		actions[name] = action
	}
}

// AddTDEvent adds or replaces an event in the TD
//  td is a TD created with 'CreateTD'
//  name of action to add
//  event created with 'CreateEvent'
func AddTDEvent(td map[string]interface{}, name string, event interface{}) {
	events := td[api.WoTEvents].(map[string]interface{})
	if event == nil {
		logrus.Errorf("Add event '%s' to TD. Event is nil.", name)
	} else {
		events[name] = event
	}
}

// AddTDProperty adds or replaces a property in the TD
//  td is a TD created with 'CreateTD'
//  name of property to add
//  property created with 'CreateProperty'
func AddTDProperty(td map[string]interface{}, name string, property interface{}) {
	props := td[api.WoTProperties].(map[string]interface{})
	if property == nil {
		logrus.Errorf("Add property '%s' to TD. Propery is nil.", name)
	} else {
		props[name] = property
	}
}

// RemoveTDProperty removes a property from the TD.
func RemoveTDProperty(td map[string]interface{}, name string) {
	props := td[api.WoTProperties].(map[string]interface{})
	if props == nil {
		logrus.Errorf("RemoveTDProperty: TD does not have any properties.")
		return
	}
	props[name] = nil

}

// SetThingVersion adds or replace Thing version info in the TD
//  td is a TD created with 'CreateTD'
//  version with map of 'name: version'
func SetThingVersion(td map[string]interface{}, version map[string]string) {
	td[api.WoTVersion] = version
}

// SetThingTitle sets the title and description of the Thing in the TD
//  td is a TD created with 'CreateTD'
//  title of the Thing
//  description of the Thing
func SetThingDescription(td map[string]interface{}, title string, description string) {
	td[api.WoTTitle] = title
	td[api.WoTDescription] = description
}

// SetThingErrorStatus sets the error status of a Thing
// This is set under the 'status' property, 'error' subproperty
//  td is a TD created with 'CreateTD'
//  status is a status tring
func SetThingErrorStatus(td map[string]interface{}, errorStatus string) {
	// FIXME:is this a property
	status := td["status"]
	if status == nil {
		status = make(map[string]interface{})
		td["status"] = status
	}
	status.(map[string]interface{})["error"] = errorStatus
}

// SetTDForm sets the top level forms section of the TD
// NOTE: In WoST actions are always routed via the Hub using the Hub's protocol binding.
// Under normal circumstances forms are therefore not needed.
//  td to add form to
//  forms with list of forms to add. See also CreateForm to create a single form
func SetTDForms(td map[string]interface{}, formList []map[string]interface{}) {
	td[api.WoTForms] = formList
}

// CreateThingID creates a ThingID from the zone it belongs to, the hardware device ID and device Type
// This creates a Thing ID: URN:zone:deviceID:deviceType
//  zone is the name of the zone the device is part of
//  deviceID is the ID of the device to use as part of the Thing ID
func CreateThingID(zone string, deviceID string, deviceType api.DeviceType) string {
	thingID := fmt.Sprintf("urn:%s:%s:%s", zone, deviceID, deviceType)
	return thingID
}

// CreatePublisherThingID creates a globally unique Thing ID that includes the zone and publisher
// name where the Thing originates from. The publisher is especially useful where protocol
// bindings create thing IDs. In this case the publisher is the gateway used by the protocol binding
// or the PB itself.
//
// This creates a Thing ID: URN:zone:publisher:deviceID:deviceType
//  zone is the name of the zone the device is part of
//  publisher is the name of the publisher that the thing originates from.
//  deviceID is the ID of the device to use as part of the Thing ID
func CreatePublisherThingID(zone string, publisher string, deviceID string, deviceType api.DeviceType) string {
	thingID := fmt.Sprintf("urn:%s:%s:%s:%s", zone, publisher, deviceID, deviceType)
	return thingID
}

// CreateTD creates a new Thing Description document with properties, events and actions
func CreateTD(thingID string, deviceType api.DeviceType) map[string]interface{} {
	td := make(map[string]interface{}, 0)
	td[api.WoTAtContext] = "http://www.w3.org/ns/td"
	td[api.WoTID] = thingID
	// TODO @type is a JSON-LD keyword to label using semantic tags, eg it needs a schema
	if deviceType != "" {
		td[api.WoTAtType] = deviceType
	}
	td[api.WoTCreated] = time.Now().Format(api.TimeFormat)
	td[api.WoTActions] = make(map[string]interface{})
	td[api.WoTEvents] = make(map[string]interface{})
	td[api.WoTProperties] = make(map[string]interface{})
	return td
}
