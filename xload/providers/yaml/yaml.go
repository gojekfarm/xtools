// Package yaml provides a YAML loader for xload.
package yaml

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/gojekfarm/xtools/xload"
)

// NewFileLoader reads YAML from the given file and returns a xload.Loader
// Nested keys are flattened using the given separator.
//
// IMPORTANT: The separator must be consistent with prefix used in the struct
// tags.
func NewFileLoader(path, sep string) (_ xload.MapLoader, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = f.Close()
	}()

	return NewLoader(f, sep)
}

// NewLoader reads YAML from the given io.Reader and returns a xload.Loader
// Nested keys are flattened using the given separator.
//
// IMPORTANT: The separator must be consistent with prefix used in the struct
// tags.
func NewLoader(r io.Reader, sep string) (xload.MapLoader, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}

	err = yaml.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}

	return xload.FlattenMap(out, sep), nil
}
