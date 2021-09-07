module github.com/wostzone/hubserve-go

go 1.14

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/wostzone/hubclient-go v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/wostzone/hubclient-go => ../hubclient-go
