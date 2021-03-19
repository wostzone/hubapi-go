package td_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hubapi/pkg/td"
)

func TestCreateTD(t *testing.T) {
	thingID := "Thing1"
	thing := td.CreateTD(thingID)
	assert.NotNil(t, thing)

	// Set version
	versions := map[string]string{"Software": "v10.1", "Hardware": "v2.0"}
	td.SetTdVersion(thing, versions)

	// Define TD property
	prop := td.CreateProperty("Prop1", "First property", false)
	enumValues := make([]string, 0) //{"value1", "value2"}
	td.SetPropertyEnum(prop, enumValues)
	td.SetPropertyUnit(prop, "C")
	td.SetPropertyTypeInteger(prop, 1, 10)
	td.SetPropertyTypeNumber(prop, 1, 10)
	td.SetPropertyTypeString(prop, 1, 10)
	td.SetPropertyTypeObject(prop, nil)
	td.SetPropertyTypeArray(prop, 3, 10)
	td.AddTdProperty(thing, "prop1", prop)
	// invalid prop should not blow up
	td.AddTdProperty(thing, "prop2", nil)

	// Define event
	ev1 := td.CreateEvent("ev1", "First event")
	td.AddTdEvent(thing, "ev1", ev1)
	// invalid event should not blow up
	td.AddTdEvent(thing, "ev1", nil)

	// Define action
	ac1 := td.CreateAction("action1", "First action")
	actionProp := td.CreateAction("channel", "Select channel")
	required := []string{"channel"}
	td.SetActionInput(ac1, "string", actionProp, required)
	td.SetActionOutput(ac1, "string")

	td.AddTdAction(thing, "action1", ac1)
	// invalid action should not blow up
	td.AddTdAction(thing, "action1", nil)

	// Define form
	f1 := td.CreateForm("form1", "", "application/json", "GET")
	formList := make([]map[string]interface{}, 0)
	formList = append(formList, f1)
	td.SetTdForms(thing, formList)
	// invalid form should not blow up
	td.SetTdForms(thing, nil)
}
