// Package yaml provides a YAML loader for xload.
package yaml

import (
	"os"

	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"

	"github.com/gojekfarm/xtools/xload"
)

// NewFileLoader reads YAML file and returns a xload.Loader
func NewFileLoader(filename string, delim string) (xload.Loader, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return NewBytesLoader(b, delim)
}

// NewBytesLoader reads YAML bytes and returns a xload.Loader
func NewBytesLoader(b []byte, delim string) (xload.Loader, error) {
	var out map[string]interface{}

	err := yaml.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}

	return xload.MapLoader(flatten("", delim, out)), nil
}

func flatten(prefix, delim string, data map[string]interface{}) map[string]string {
	flattened := make(map[string]string)

	for key, value := range data {
		switch value := value.(type) {
		case map[string]interface{}:
			for k, v := range flatten(key+delim, delim, value) {
				flattened[prefix+k] = v
			}
		default:
			flattened[prefix+key] = cast.ToString(value)
		}
	}

	return flattened
}
