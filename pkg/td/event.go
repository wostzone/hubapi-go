// Package td with TD event creation
package td

// Thing event definition
// Credit: https://github.com/dravenk/webthing-go/blob/master/event.go

// CreateEvent creates a new TD event description
//  title title for presentation
//  description optional extra description of what the event represents
func CreateEvent(title string, description string) map[string]interface{} {
	event := make(map[string]interface{}, 0)
	event["title"] = title
	event["description"] = description

	return event
}
