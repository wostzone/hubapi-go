module github.com/wostzone/hubapi

go 1.14

require (
	github.com/eclipse/paho.mqtt.golang v1.3.2
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/yaml.v2 v2.4.0
)

// Until Hub is stable
replace github.com/wostzone/hubapi => ../hubapi-go
