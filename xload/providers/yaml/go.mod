module github.com/gojekfarm/xtools/xload/providers/yaml

go 1.20

replace github.com/gojekfarm/xtools/xload => ../..

require (
	github.com/gojekfarm/xtools/xload v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/stretchr/testify v1.8.4

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
)
