package td

import "github.com/sirupsen/logrus"

type DynamicThingDescription map[string]interface{}

// tbd json-ld parsers:
// Most popular; https://github.com/xeipuuv/gojsonschema
// Other:  https://github.com/piprate/json-gold

// AddAction adds or replaces an action in the TD
//  td is a TD created with 'CreateTD'
//  name of action to add
//  action created with 'CreateAction'
func AddTdAction(td map[string]interface{}, name string, action interface{}) {
	actions := td["actions"].(map[string]interface{})
	if action == nil {
		logrus.Errorf("Add action '%s' to TD. Action is nil", name)
	} else {
		actions[name] = action
	}
}

// AddEvent adds or replaces an event in the TD
//  td is a TD created with 'CreateTD'
//  name of action to add
//  event created with 'CreateEvent'
func AddTdEvent(td map[string]interface{}, name string, event interface{}) {
	events := td["events"].(map[string]interface{})
	if event == nil {
		logrus.Errorf("Add event '%s' to TD. Event is nil.", name)
	} else {
		events[name] = event
	}
}

// AddProperty adds or replaces a property in the TD
//  td is a TD created with 'CreateTD'
//  name of property to add
//  property created with 'CreateProperty'
func AddTdProperty(td map[string]interface{}, name string, property interface{}) {
	props := td["properties"].(map[string]interface{})
	if property == nil {
		logrus.Errorf("Add property '%s' to TD. Propery is nil.", name)
	} else {
		props[name] = property
	}
}

// SetTdVersion adds or replace TD version info
//  td is a TD created with 'CreateTD'
//  version with map of 'name: version'
func SetTdVersion(td map[string]interface{}, version map[string]string) {
	td["version"] = version
}

// SetTDForm sets the top level forms section of the TD
// NOTE: In WoST actions are always routed via the Hub using the Hub's protocol binding.
// Under normal circumstances forms are therefore not needed.
//  td to add form to
//  forms with list of forms to add. See also CreateForm to create a single form
func SetTdForms(td map[string]interface{}, formList []map[string]interface{}) {
	td["forms"] = formList
}

// CreateTD creates a new Thing Description document
func CreateTD(id string) DynamicThingDescription {
	td := make(DynamicThingDescription, 0)
	td["@context"] = "http://www.w3.org/ns/td"
	td["id"] = id
	td["properties"] = make(map[string]interface{})
	td["events"] = make(map[string]interface{})
	td["actions"] = make(map[string]interface{})
	return td
}
