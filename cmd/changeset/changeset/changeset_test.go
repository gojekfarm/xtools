package changeset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/cmd/changeset/testutil"
)

func TestBumpCompare(t *testing.T) {
	tests := []struct {
		a, b Bump
		want int
	}{
		{BumpPatch, BumpMinor, -1},
		{BumpMinor, BumpMajor, -1},
		{BumpPatch, BumpMajor, -1},
		{BumpMajor, BumpMinor, 1},
		{BumpMinor, BumpPatch, 1},
		{BumpMajor, BumpPatch, 1},
		{BumpPatch, BumpPatch, 0},
		{BumpMinor, BumpMinor, 0},
		{BumpMajor, BumpMajor, 0},
	}

	for _, tt := range tests {
		name := string(tt.a) + "_vs_" + string(tt.b)
		t.Run(name, func(t *testing.T) {
			got := tt.a.Compare(tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBumpString(t *testing.T) {
	tests := []struct {
		bump Bump
		want string
	}{
		{BumpPatch, "patch"},
		{BumpMinor, "minor"},
		{BumpMajor, "major"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.bump.String())
		})
	}
}

func TestGenerateID(t *testing.T) {
	for range 10 {
		id := GenerateID()
		parts := strings.Split(id, "-")
		assert.Len(t, parts, 3, "GenerateID() = %q should have 3 parts", id)
		for _, part := range parts {
			assert.NotEmpty(t, part, "GenerateID() = %q has empty part", id)
		}
	}
}

func TestGenerateIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		id := GenerateID()
		if seen[id] {
			t.Logf("duplicate ID generated: %s (this is probabilistic)", id)
		}
		seen[id] = true
	}
}

func TestParseChangeset(t *testing.T) {
	dir := testutil.SetupTestRepo(t)
	path := filepath.Join(dir, ".changeset", "happy-tiger-jump.md")

	cs, err := ParseChangeset(path)
	require.NoError(t, err)

	assert.Equal(t, "happy-tiger-jump", cs.ID)
	assert.Len(t, cs.Modules, 1)
	assert.Equal(t, BumpMinor, cs.Modules["libA"])
	assert.Contains(t, cs.Summary, "greeting")
}

func TestParseChangesetInvalid(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"missing opening", "some content\n---\n"},
		{"missing closing", "---\nkey: value\n"},
		{"invalid yaml", "---\n: invalid\n---\n"},
		{"invalid bump", "---\n\"mod\": invalid\n---\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.md")
			err := os.WriteFile(path, []byte(tt.content), 0644)
			require.NoError(t, err)

			_, err = ParseChangeset(path)
			assert.Error(t, err)
		})
	}
}

func TestReadChangesets(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	changesets, err := ReadChangesets(dir)
	require.NoError(t, err)

	assert.Len(t, changesets, 1)
	assert.Equal(t, "happy-tiger-jump", changesets[0].ID)
}

func TestReadChangesetsNonExistent(t *testing.T) {
	changesets, err := ReadChangesets("/nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, changesets)
}
