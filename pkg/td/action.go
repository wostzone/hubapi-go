// Package td with TD action creation
package td

// CreateAction creates a new TD event description
//  title for presentation
//  description optional extra description of what the action does
// Returns an action object
func CreateAction(title string, description string) map[string]interface{} {
	action := make(map[string]interface{}, 0)
	action["title"] = title
	action["description"] = description

	return action
}

// SetActionInput sets the an input section of the action
//  action to add input to
//  inputType "object", "string", "number", "int"
//  properties property definitions to be provided, created with CreateProperty
//  requiredProperties list of property names that must be provided
func SetActionInput(action map[string]interface{},
	inputType string,
	properties map[string]interface{},
	requiredProperties []string) {

	input := make(map[string]interface{}, 0)
	input["type"] = inputType
	input["properties"] = properties
	input["required"] = requiredProperties
	action["input"] = input
}

// SetActionForms sets the forms section of the action, if needed
// NOTE: In WoST actions are always routed via the Hub using the Hub's protocol binding.
// Under normal circumstances forms are therefore not needed.
//  action to add form to
//  forms with list of forms to add. See also CreateForm to create a single form
func SetActionForms(action map[string]interface{}, forms []map[string]interface{}) {
	action["forms"] = forms
}

// SetActionOutput sets the output section of the action
// ??? what is the purpose of this?
//  action to add output to
//  outputType "object", "string", "number", "int"
func SetActionOutput(action map[string]interface{}, outputType string) {
	output := make(map[string]interface{}, 0)
	output["type"] = outputType
	action["output"] = output
}
