# WoST Web Of Things API

This repository defines the API's to use to connect to the WoST Hub and exchange information between Things and consumers. It contains the type definitions and client libraries to exchange messages via the hub as well as constructing a *Thing Description* document. This API aims to adhere to the [WoT Specfications](https://www.w3.org/WoT/).

## Project Status

Status: The status of this library is Alpha. It is functional but breaking changes must be expected.

This library provides the Golang implementation of the MQTT API of the WoST Hub. MQTT is the most direct
method to talk to the Hub. There is no need for Golang developers to use HTTP or Websockets.

This API repository is usable:
- The hubconfig can be used to read configuration for use in plugins
- The certsetup can be used to generate self signed certificates (see Hub's installation)
- The hubclient package can be used to create a MQTT client connection to the Hub and publish/subscribe messages
- The td package can be used to construct a TD with properties, events and actions

What is not yet functional:
- The API for Thing and consumer provisioning.
- signing of messages is under consideration. This needs a protocol definition to follow. Presumably using JWS
- encryption of messages is under consideration. This needs a protocol definition to follow. Presumably using JWE
- the generated TD is basic and might not yet conform to the standard
- the methods in this API are preliminary and can still change
- the client does not use the TD forms. This is under consideration.



## Audience

This project is aimed at WoST Thing and Hub Plugin developers that share concerns about the security and privacy risk of running a server on every WoT Thing. WoST developers choose to not run servers on Things and instead use a hub and spokes model.


## Summary

This library provides API's to exchange messages with the WoST Hub in a WoT compatible manner. IoT developers can use it to provision their IoT device, publish their Thing Description, send events and receive actions. Plugin developers can use it to receive Thing Descriptions, events and property values, and to request Thing actions.

A Python and Javascript version is planned for the future.

## Dependencies

This requires the use of a WoST compatible Hub or Gateway.

Supported hubs and gateways:
- [WoST Hub](https://github.com/wostzone/hub)


## Usage

This module is intended to be used as a library by Things, Consumers and Hub Plugin developers. It features support for building WoT compliant Thing Description documents, reading Hub and plugin configuration files, client connections to the Hub over HTTP/TLS and MQTT/TLS.

The plugin test files contain examples for clients.

### hubconfig

>### Todo: include simple examples

The hubconfig package contains the library to read the Hub and plugin configuration files and setup logging. It is intended for plugin developers that need Hub configuration. 


### hubclient

The hubclient package contains the client code to connect to the Hub. Currently this supports the MQTT protocol. This package is for convenience. It wraps the Paho mqtt client and adds automatic reconnect and resubscribes in case connections get lost. 

See the [MQTT API documentation](docs/mqtt-api.md) for details. This client does not yet use [the Form method](https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#protocol-bindings).  This will be added in the near future.

Note that the above WoT specification talks about interaction between consumer and Thing. In the case of WoST this interaction takes place via the Hub's message bus.

The api/IThingClient.go package contains the interface on using this library.

### certsetup

>### This section is subject to change

The certsetup package provides functions for creating self signed certificates include a self signed Certificate Authority (CA). These can be used for verifying authenticity of server and clients of the message bus.

Most clients will use the existing CA certificate to connect to the MQTT message bus.
For this to work, the client must load the hubconfig using hubconfig.SetHubCommandlineArgs() to load the hub configuration including the certificate folder. The hubclient.NewThingClient and NewConsumerClient expect a CA certificate file. This can be determined with path.Join(hubconfig.CertFolder, certsetup.CaCertFile), or use the NewPluginClient() method which takes the hubconfig configuration and does this for you.


### signing

>### This section is subject to change
The signing package provides functions to JWS sign and JWE encrypt messages. This is used to verify the authenticity of the sender of the message.

Signing support is built into the HTTP and MQTT protocol binding client and server. 
Well, soon anyways... 

Signing and sender verification is key to guarantee that the information has not been tampered with and originated from the sender. The WoT spec does not (?) have a place for this, so it might become a WoST extension.

### td
>### Todo: include simple examples

The td package provides methods to construct a WoT compliant Thing Description. 


# Contributing

Contributions to WoST projects are always welcome. There are many areas where help is needed, especially with documentation and building plugins for IoT and other devices. See [CONTRIBUTING](https://github.com/wostzone/hub/docs/CONTRIBUTING.md) for guidelines.


# Credits

This project builds on the Web of Things (WoT) standardization by the W3C.org standards organization. For more information https://www.w3.org/WoT/
