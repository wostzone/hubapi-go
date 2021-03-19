package td

// Thing property creation

// CreateProperty creates a new property instance
//  title propery title for presentation
//  description optional extra description of what the property does
//  writable property is a configuration value and is writable
func CreateProperty(title string,
	description string,
	writable bool) map[string]interface{} {

	prop := make(map[string]interface{}, 0)
	prop["type"] = "null"
	prop["title"] = title
	prop["description"] = description
	prop["writable"] = writable
	// see https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#example-29
	prop["readOnly"] = !writable
	prop["writeOnly"] = writable

	return prop
}

func SetPropertyEnum(prop map[string]interface{}, enumValues ...interface{}) {
	prop["enum"] = enumValues
}

func SetPropertyUnit(prop map[string]interface{}, unit string) {
	prop["unit"] = unit
}

func SetPropertyTypeArray(prop map[string]interface{}, minItems uint, maxItems uint) {
	prop["type"] = "array"
	if minItems > 0 {
		prop["minItems"] = minItems
	}
	if maxItems > 0 {
		prop["maxItems"] = maxItems
	}
}

func SetPropertyTypeInteger(prop map[string]interface{}, min int, max int) {
	prop["type"] = "integer"
	if min > 0 {
		prop["minimum"] = min
	}
	if min < 0 || max != 0 {
		prop["maximum"] = max
	}
}

func SetPropertyTypeNumber(prop map[string]interface{}, min float64, max float64) {
	prop["type"] = "number"
	if min > 0 {
		prop["minimum"] = min
	}
	if min < 0 || max != 0 {
		prop["maximum"] = max
	}
}

func SetPropertyTypeObject(prop map[string]interface{}, object interface{}) {
	prop["type"] = "object"
	prop["object"] = object
}

//
func SetPropertyTypeString(prop map[string]interface{}, minLength int, maxLength int) {
	prop["type"] = "string"
	if minLength != 0 {
		prop["minLength"] = minLength
	}
	if maxLength != 0 {
		prop["maxLength"] = maxLength
	}
}
