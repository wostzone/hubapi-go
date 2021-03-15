package wostapi

// WoST MQTT protocol definitions

// TopicRoot is the base of the topic
const TopicRoot = "things"

// TopicThingTD topic for thing publishing its TD
const TopicThingTD = TopicRoot + "/{id}/td"

// TopicThingPropertyValues topic for Thing publishing updates to property values
const TopicThingPropertyValues = TopicRoot + "/{id}/values"

// TopicThingEvent topic for thing publishing its Thing events
const TopicThingEvent = TopicRoot + "/{id}/event"

// TopicSetConfig topic request to update Thing configuration
const TopicSetConfig = TopicRoot + "/{id}/config"

// TopicAction topic request to start action
const TopicAction = TopicRoot + "/{id}/action"

//
