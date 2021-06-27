# WoST Web Of Things Library

This repository provides a library with definitions and methods to connect to the WoST Hub and exchange information between Things and consumers. It contains the type definitions and methods for devices to publish 'Thing' information and for users to subscribe to updates. 

## Project Status

Status: The status of this library is Alpha. It is functional but breaking changes must be expected.

This API repository is usable:
- The hubconfig can be used to read configuration for use in plugins
- The certsetup can be used to manage self signed certificates (see Hub's installation)
- The hubclient package can be used to create a MQTT client connection to the Hub message bus and publish/subscribe messages
- The td package can be used to construct a WoT TD with properties, events and actions

What is not yet functional:
- signing of messages is under consideration. This needs a protocol definition to follow. Presumably using JWS
- encryption of messages is under consideration. This needs a protocol definition to follow. Presumably using JWE
- the generated TD is basic and might not conform to the WoT standard that is itself in flux (June 2021)
- the client does not use the TD forms. This is under consideration.



## Audience

This project is aimed at WoST Thing and Hub Plugin developers that share concerns about the security and privacy risk of running a server on every WoT Thing. WoST developers choose to not run servers on Things and instead use a hub and spokes model.


## Summary

This library provides functions and a vocabulary to exchange messages with the WoST Hub in a WoT compatible manner. IoT developers can use it to provision their IoT device, publish their Thing Description, send events and receive actions. Plugin developers can use it to receive Thing Descriptions, events and property values, and to request Thing actions.

A Python and Javascript version is planned for the future.

## Dependencies

This requires the use of a WoST compatible Hub or Gateway.

Supported hubs and gateways:
- [WoST Hub](https://github.com/wostzone/hub)


## Usage

This module is intended to be used as a library by Things, Consumers and Hub Plugin developers. It features support for building WoT compliant Thing Description documents, reading Hub and plugin configuration files, and create client connections to the Hub over HTTP/TLS and MQTT/TLS.

### hubconfig

The hubconfig package contains the library to read the Hub and plugin configuration files and setup logging. It is intended for plugin developers that need Hub configuration. 

```golang
	import "github.com/wostzone/wostlib-go/pkg/hubconfig"
  ...
	hc, err := hubconfig.LoadHubConfig(homeFolder)
  ...
```


### hubclient

The hubclient package contains the client code to connect to the Hub using MQTT. This package wraps the Paho mqtt client and adds automatic reconnect and resubscribes in case connections get lost. 

See the [MQTT API documentation](docs/mqtt-api.md) for details. This client does not yet use [the Form method](https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#protocol-bindings).  This will be added in the near future.

Note that the above WoT specification talks about interaction between consumer and Thing. In the case of WoST this interaction takes place via the Hub's message bus.

The api/IThingClient.go package contains the interface on using this library.

Example:
```golang
	import "github.com/wostzone/wostlib-go/pkg/hubclient"
  ...
  mqttHostPort := "localhost:33101"
  caCertFile := ...
  clientCertFile := ...
  clientKeyFile := ...
 	deviceClient := hubclient.NewMqttHubDeviceClient(deviceID, mqttHostPort, 
   caCertFile, clientCertFile, clientKeyFile)
  err := deviceClient.Start()

```

### certsetup

The certsetup package provides functions for creating, saving and loading self signed certificates include a self signed Certificate Authority (CA). These are used for verifying authenticity of server and clients of the message bus.



### signing

>### This section is subject to change
The signing package provides functions to JWS sign and JWE encrypt messages. This is used to verify the authenticity of the sender of the message.

Signing support is built into the HTTP and MQTT protocol binding client and server. 
Well, soon anyways... 

Signing and sender verification is key to guarantee that the information has not been tampered with and originated from the sender. The WoT spec does not (?) have a place for this, so it might become a WoST extension.

### td

The td package provides methods to construct a WoT compliant Thing Description. 

```json
	import "github.com/wostzone/wostlib-go/pkg/td"
	import  "github.com/wostzone/wostlib-go/pkg/vocab"

  ...
	thing1 := td.CreateTD(deviceID, vocab.DeviceTypeSensor)
 	prop1 := td.CreateProperty("Prop1", "First property", vocab.PropertyTypeSensor)
 	td.AddTDProperty(thing1, "prop1", prop)
  ...
```

### testenv

The 'testenv' package includes a mosquitto launcher and test environment for testing plugins.

### tlsclient

Wrapper around HTTP/TLS connections. 
Used by the IDProv protocol binding.

### tlsserver

Server of HTTP/TLS connections. 
Used by the IDProv protocol server.

### vocab

Vocabulary of property and attributes names for building and using TDs.

# Contributing

Contributions to WoST projects are always welcome. There are many areas where help is needed, especially with documentation and building plugins for IoT and other devices. See [CONTRIBUTING](https://github.com/wostzone/hub/docs/CONTRIBUTING.md) for guidelines.


# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/
