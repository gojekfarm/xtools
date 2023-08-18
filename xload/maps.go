package xload

import (
	"context"

	"github.com/spf13/cast"
)

// MapLoader loads values from a map.
//
// Can be used with xload.FlattenMap as an intermediate format
// when loading from various sources.
type MapLoader map[string]string

// Load fetches the value from the map.
func (m MapLoader) Load(_ context.Context, key string) (string, error) {
	value, ok := m[key]
	if !ok {
		return "", nil
	}

	return value, nil
}

// FlattenMap flattens a map[string]interface{} into a map[string]string.
// Nested maps are flattened using given separator.
func FlattenMap(m map[string]interface{}, sep string) map[string]string {
	return flatten(m, "", sep)
}

func flatten(m map[string]interface{}, prefix string, sep string) map[string]string {
	flattened := make(map[string]string)

	for key, value := range m {
		switch value := value.(type) {
		case map[string]interface{}:
			for k, v := range flatten(value, key+sep, sep) {
				flattened[prefix+k] = v
			}
		default:
			flattened[prefix+key] = cast.ToString(value)
		}
	}

	return flattened
}
