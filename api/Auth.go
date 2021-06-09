// Package api with group based authorization definition
// Group based authorization is managed centrally by the Hub and implemented by protocol bindings
// These definitions are intended for use by protocol bindings that implement authorization for their protocol
package api

// Group roles set permissions for accessing Things that are members of the same group
const (
	// GroupRoleNone indicates that the client has no particular role. It can not do anything until
	// the role is upgraded to viewer or better.
	// Subscribe permissions: none
	// Publish permissions: none
	GroupRoleNone = "none"

	// GroupRoleViewer lets a client subscribe to Thing TD and Thing Events
	// Subscribe permissions: TD, Events
	// Publish permissions: none
	GroupRoleViewer = "viewer"

	// GroupRoleUser lets a client subscribe to Thing TD, events and publish actions
	// Subscribe permissions: TD, Events
	// Publish permissions: Actions
	GroupRoleEditor = "editor"

	// GroupRoleManager lets a client subscribe to Thing TD, events, publish actions and update configuration
	// Subscribe permissions: TD, Events
	// Publish permissions: Actions, Configuration
	GroupRoleManager = "manager"

	// GroupRoleThing indicates the client is a IoT device that can publish and subscribe
	// to Thing topics.
	// Things should only publish events and updates for Things it published the TD for.
	// Publish permissions: TD, Events
	// Subscribe permissions: Actions, Configuration
	GroupRoleThing = "thing"
)

// Organization Unit for client authorization are stored in the client certificate OU field
const (
	// Default OU with no API access permissions
	OUNone = ""

	// OUClient lets a client connect to the message bus
	OUClient = "client"

	// OUIoTDevice indicates the client is a IoT device that can connect to the message bus
	// perform discovery and request provisioning.
	// Provision API permissions: GetDirectory, ProvisionRequest, GetStatus
	OUIoTDevice = "iotdevice"

	//OUAdmin lets a client approve thing provisioning (postOOB), add and remove users
	// Provision API permissions: GetDirectory, ProvisionRequest, GetStatus, PostOOB
	OUAdmin = "admin"

	// OUPlugin marks a client as a plugin.
	// By default, plugins have full permission to all APIs
	// Provision API permissions: Any
	OUPlugin = "plugin"
)

// AuthGroup defines a group with Thing and Users
// The permission is determined by taking the thing permission and user permission and
// return the lowest of the two.
// Eg an admin role can do anything only if the thing allows it
//
// This allows for Things to be shared with other groups with viewing rights only, even though
// there are user or admins in that group.
type AuthGroup struct {
	// The name of the group
	GroupName string
	// The members (thingIDs and userIDs) and their role: [memberid]role
	MemberRoles map[string]string
}
