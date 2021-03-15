# WoST API

Golang client API library for building WoST based IoT devices and applications. 

* Hub configuration for use by plugins
* TD - Thing Description builder library
* MQTT API for connecting to the WoST Hub via MQTT
* HTTP API for connecting to the WoST Hub via HTTP
* WebSocket API for connecting to the WoST Hub via WebSocket (tenative tbd)

## Project Status

Status: In development, not ready for use

## Audience

This project is aimed at WoST developers that share concerns about the security and privacy risk of running a server on every WoT Thing. WoST developers choose to not run servers on Things and instead use a hub and spokes model.


## Summary

This library provides API's to connect to the WoST Hub. IoT developers can use it to provision their IoT device, publish their Thing Description, send events and receive actions.
This library is also used by WoST Hub protocol binding plugins to publish 3rd party and legacy IoT devices.

A Python and Javascript version is planned for the future.

## Dependencies

This requires the use of a WoST compatible Hub or Gateway and support the WoST protocol bindings for one of these APIs.

Supported hubs and gateways:
- [WoST Hub](https://github.com/wostzone/hub))

## Configuration

No configuration are used. All configuration is done through the API.

## Usage

This is a golang library that needs to be imported by application developers.

For example:
```go
  package myiotdevice
  import (
    "github.com/wostzone/api-go/pkg/td
    "github.com/wostzone/api-go/pkg/wosthttp
  )
  myTD := td.CreateTD(id)
  connection := wosthttp.CreateConnection(hostname, "myiotdevice")
  connection.PublishTD(myTD)
```

# Contributing

Contributions to WoST projects are always welcome. There are many areas where help is needed, especially with documentation and building plugins for IoT and other devices. See [CONTRIBUTING](https://github.com/wostzone/hub/docs/CONTRIBUTING.md) for guidelines.


# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/
