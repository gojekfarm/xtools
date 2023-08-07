// Package yaml provides a YAML loader for xload.
package yaml

import (
	"io"

	"gopkg.in/yaml.v3"

	"github.com/gojekfarm/xtools/xload"
)

// NewLoader reads YAML from the given io.Reader and returns a xload.Loader
// Nested keys are flattened using the given separator.
func NewLoader(r io.Reader, sep string) (xload.Loader, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}

	err = yaml.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}

	return xload.MapLoader(xload.FlattenMap(out, sep)), nil
}
