// Package api with group based authorization definition
// Group based authorization is managed centrally by the Hub and implemented by protocol bindings
// These definitions are intended for use by protocol bindings that implement authorization for their protocol
package api

// RoleNone means the user has no role
const RoleNone = "none"

// RoleViewer lets a user subscribe to Thing TD and events
const RoleViewer = "viewer"

// RoleUser lets a user subscribe to Thing TD, events and publish actions
const RoleUser = "user"

// RoleAdmin lets a user subscribe to Thing TD, events and publish actions and configuration
const RoleAdmin = "admin"

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
