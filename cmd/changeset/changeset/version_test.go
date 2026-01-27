package changeset

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		bump    Bump
		want    string
		wantErr bool
	}{
		{"patch from zero", "v0.0.0", BumpPatch, "v0.0.1", false},
		{"minor from zero", "v0.0.0", BumpMinor, "v0.1.0", false},
		{"major from zero", "v0.0.0", BumpMajor, "v1.0.0", false},
		{"patch", "v1.2.3", BumpPatch, "v1.2.4", false},
		{"minor", "v1.2.3", BumpMinor, "v1.3.0", false},
		{"major", "v1.2.3", BumpMajor, "v2.0.0", false},
		{"patch double digit", "v0.10.5", BumpPatch, "v0.10.6", false},
		{"minor resets patch", "v1.2.3", BumpMinor, "v1.3.0", false},
		{"major resets all", "v1.2.3", BumpMajor, "v2.0.0", false},
		{"invalid version", "invalid", BumpPatch, "", true},
		// Note: semver library accepts versions without v prefix and normalizes output
		{"without v prefix", "1.0.0", BumpPatch, "v1.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IncrementVersion(tt.current, tt.bump)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidVersion)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterIgnored(t *testing.T) {
	releases := []Release{
		{Module: "a", Version: "v1.0.0"},
		{Module: "b", Version: "v1.0.0"},
		{Module: "c", Version: "v1.0.0"},
	}

	tests := []struct {
		name        string
		ignore      []string
		wantModules []string
	}{
		{"no ignore", []string{}, []string{"a", "b", "c"}},
		{"ignore one", []string{"a"}, []string{"b", "c"}},
		{"ignore two", []string{"a", "c"}, []string{"b"}},
		{"ignore nonexistent", []string{"d"}, []string{"a", "b", "c"}},
		{"ignore all", []string{"a", "b", "c"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterIgnored(releases, tt.ignore)

			var gotModules []string
			for _, r := range got {
				gotModules = append(gotModules, r.Module)
			}

			if tt.wantModules == nil {
				assert.Empty(t, gotModules)
			} else {
				assert.Equal(t, tt.wantModules, gotModules)
			}
		})
	}
}

func TestFilterIgnoredPreservesOrder(t *testing.T) {
	releases := []Release{
		{Module: "z", Version: "v1.0.0"},
		{Module: "a", Version: "v1.0.0"},
		{Module: "m", Version: "v1.0.0"},
	}

	got := FilterIgnored(releases, []string{})

	expected := []string{"z", "a", "m"}
	for i, r := range got {
		assert.Equal(t, expected[i], r.Module, "order should be preserved at index %d", i)
	}
}
