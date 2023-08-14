module github.com/gojekfarm/xtools/examples/xload

go 1.20

replace (
	github.com/gojekfarm/xtools/xload => ../../xload
	github.com/gojekfarm/xtools/xload/providers/yaml => ../../xload/providers/yaml
)

require (
	github.com/gojekfarm/xtools/xload v0.4.0
	github.com/gojekfarm/xtools/xload/providers/yaml v0.4.0
)

require (
	github.com/spf13/cast v1.5.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
