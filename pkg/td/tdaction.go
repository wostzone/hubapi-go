// Package td with TD action creation
package td

import "github.com/wostzone/hubapi-go/api"

// CreateTDAction creates a new TD action description
//  title for presentation
//  description optional extra description of what the action does
// Returns an action object
func CreateTDAction(title string, description string) map[string]interface{} {
	action := make(map[string]interface{}, 0)
	action[api.WoTTitle] = title
	action[api.WoTDescription] = description

	return action
}

// CreateActionRequest creates a new message for requesting an action from a Thing
//  name contains the name of the action to request, as defined in the TD action section
//  params contains the corresponding parameters as defined in the TD action section
//
// This returns a message that can be published with IHubClient.PublishConfigRequest()
func CreateActionRequest(name string, params map[string]interface{}) map[string]interface{} {
	action := make(map[string]interface{}, 0)
	action[name] = params
	return action
}

// SetTDActionInput sets the an input section of the action
//  action to add input to
//  inputDataType "object", "string", "number", "int"
//  properties property definitions to be provided, created with CreateTDProperty
//  requiredProperties list of property names that must be provided
func SetTDActionInput(action map[string]interface{},
	inputDataType string,
	properties map[string]interface{},
	requiredProperties []string) {

	input := make(map[string]interface{}, 0)
	input[api.WoTDataType] = inputDataType
	input[api.WoTProperties] = properties
	input[api.WoTRequired] = requiredProperties
	action[api.WoTInput] = input
}

// SetTDActionForms sets the forms section of the action, if needed
// NOTE: In WoST actions are always routed via the Hub using the Hub's protocol binding.
// Under normal circumstances forms are therefore not needed.
//  action to add form to
//  forms with list of forms to add. See also CreateForm to create a single form
func SetTDActionForms(action map[string]interface{}, forms []map[string]interface{}) {
	action[api.WoTForms] = forms
}

// SetTDActionOutput sets the output section of the action
// ??? what is the purpose of this?
//  action to add output to
//  outputType "object", "string", "number", "int"
func SetTDActionOutput(action map[string]interface{}, outputType string) {
	output := make(map[string]interface{}, 0)
	output[api.WoTDataType] = outputType
	action[api.WoTOutput] = output
}
