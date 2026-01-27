package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTag(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantModule  string
		wantVersion string
		wantOK      bool
	}{
		{"root version tag", "v0.10.0", "", "v0.10.0", true},
		{"simple submodule", "xkafka/v0.10.0", "xkafka", "v0.10.0", true},
		{"nested submodule", "xkafka/middleware/v0.10.0", "xkafka/middleware", "v0.10.0", true},
		{"invalid tag - not a version", "not-a-version", "", "", false},
		{"partial version", "v1.0", "", "", false},
		{"missing v prefix", "1.0.0", "", "", false},
		{"prerelease version", "v1.0.0-beta.1", "", "v1.0.0-beta.1", true},
		{"prerelease with module", "pkg/v2.0.0-rc.1", "pkg", "v2.0.0-rc.1", true},
		{"build metadata", "v1.0.0+build.123", "", "v1.0.0+build.123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, version, ok := ParseTag(tt.input)
			assert.Equal(t, tt.wantOK, ok, "ParseTag(%q) ok", tt.input)
			if tt.wantOK {
				assert.Equal(t, tt.wantModule, module, "ParseTag(%q) module", tt.input)
				assert.Equal(t, tt.wantVersion, version, "ParseTag(%q) version", tt.input)
			}
		})
	}
}

func TestFormatTag(t *testing.T) {
	tests := []struct {
		name    string
		module  string
		version string
		want    string
	}{
		{"root module", "", "v1.0.0", "v1.0.0"},
		{"simple submodule", "xkafka", "v0.10.0", "xkafka/v0.10.0"},
		{"nested submodule", "a/b", "v1.0.0", "a/b/v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTag(tt.module, tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseTagFormatTagRoundTrip(t *testing.T) {
	tests := []struct {
		module  string
		version string
	}{
		{"", "v1.0.0"},
		{"pkg", "v0.1.0"},
		{"a/b/c", "v2.0.0-beta.1"},
	}

	for _, tt := range tests {
		tag := FormatTag(tt.module, tt.version)
		module, version, ok := ParseTag(tag)
		assert.True(t, ok, "ParseTag(FormatTag(%q, %q)) should succeed", tt.module, tt.version)
		assert.Equal(t, tt.module, module, "round trip module")
		assert.Equal(t, tt.version, version, "round trip version")
	}
}
