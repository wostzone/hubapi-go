package td

import "github.com/wostzone/hubapi/api"

// Thing property creation

// CreateTDProperty creates a new property instance
//  title propery title for presentation
//  description optional extra description of what the property does
//  propType provides @type value for a property
//  writable property is a configuration value and is writable
func CreateTDProperty(title string, description string, propType api.ThingPropType) map[string]interface{} {

	var writable = (propType == api.PropertyTypeConfig)
	prop := make(map[string]interface{}, 0)
	prop["@type"] = propType
	prop["title"] = title
	if description != "" {
		prop["description"] = description
	}
	// default is read-only
	if writable {
		prop["writable"] = writable
		// see https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#example-29
		prop["readOnly"] = !writable
		prop["writeOnly"] = writable
	}
	return prop
}

func SetTDPropertyEnum(prop map[string]interface{}, enumValues ...interface{}) {
	prop["enum"] = enumValues
}

func SetTDPropertyUnit(prop map[string]interface{}, unit string) {
	prop["unit"] = unit
}

// SetTDPropertyDataTypeArray sets the property data type as an array (of ?)
// If maxItems is 0, both minItems and maxItems are ignored
//  minItems is the minimum nr of items required
//  maxItems sets the maximum nr of items required
func SetTDPropertyDataTypeArray(prop map[string]interface{}, minItems uint, maxItems uint) {
	prop["type"] = api.DataTypeArray
	if maxItems > 0 {
		prop["minItems"] = minItems
		prop["maxItems"] = maxItems
	}
}

// SetPropertyTypeNumber sets the property data type as an integer
// If min and max are both 0, they are ignored
//  min is the minimum value
//  max sets the maximum value
func SetTDPropertyDataTypeInteger(prop map[string]interface{}, min int, max int) {
	prop["type"] = api.DataTypeInt
	if !(min == 0 && max == 0) {
		prop["minimum"] = min
		prop["maximum"] = max
	}
}

// SetTDPropertyDataTypeNumber sets the property data type as floating point number
// If min and max are both 0, they are ignored
//  min is the minimum value
//  max sets the maximum value
func SetTDPropertyDataTypeNumber(prop map[string]interface{}, min float64, max float64) {
	prop["type"] = api.DataTypeNumber
	if !(min == 0 && max == 0) {
		prop["minimum"] = min
		prop["maximum"] = max
	}
}

// SetTDPropertyDataTypeObject sets the property data type as an object
func SetTDPropertyDataTypeObject(prop map[string]interface{}, object interface{}) {
	prop["type"] = api.DataTypeObject
	prop["object"] = object
}

// SetTDPropertyDataTypeString sets the property data type as string
// If minLength and maxLength are both 0, they are ignored
//  minLength is the minimum value
//  maxLength sets the maximum value
func SetTDPropertyDataTypeString(prop map[string]interface{}, minLength int, maxLength int) {
	prop["type"] = api.DataTypeString
	if !(minLength == 0 && maxLength == 0) {
		prop["minLength"] = minLength
		prop["maxLength"] = maxLength
	}
}
