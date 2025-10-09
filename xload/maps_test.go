package xload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenMap(t *testing.T) {
	input := map[string]any{
		"NAME":    "xload",
		"VERSION": 1.1,
		"AUTHOR": map[string]any{
			"NAME":  "gojek",
			"EMAIL": "test@gojek.com",
			"ADDRESS": map[string]any{
				"CITY": "Bombay",
			},
		},
	}

	want := map[string]string{
		"NAME":                "xload",
		"VERSION":             "1.1",
		"AUTHOR_NAME":         "gojek",
		"AUTHOR_EMAIL":        "test@gojek.com",
		"AUTHOR_ADDRESS_CITY": "Bombay",
	}

	got := FlattenMap(input, "_")
	assert.EqualValues(t, want, got)
}
