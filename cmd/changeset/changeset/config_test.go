package changeset

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/cmd/changeset/testutil"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "main", cfg.BaseBranch)
	assert.Equal(t, BumpPatch, cfg.DependentBump)
	assert.Empty(t, cfg.Root)
	assert.Empty(t, cfg.Ignore)
	assert.Empty(t, cfg.IgnorePaths)
}

func TestReadConfig(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	cfg, err := ReadConfig(dir)
	require.NoError(t, err)

	assert.Equal(t, "github.com/test/fakerepo", cfg.Root)
	assert.Equal(t, "main", cfg.BaseBranch)
	assert.Equal(t, BumpPatch, cfg.DependentBump)
}

func TestReadConfigNonExistent(t *testing.T) {
	cfg, err := ReadConfig("/nonexistent/path")
	require.NoError(t, err, "should return default config for non-existent directory")

	def := DefaultConfig()
	assert.Equal(t, def.BaseBranch, cfg.BaseBranch)
	assert.Equal(t, def.DependentBump, cfg.DependentBump)
}

func TestChangesetDirExists(t *testing.T) {
	repoDir := testutil.SetupTestRepo(t)

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"existing", repoDir, true},
		{"nonexistent", "/nonexistent/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChangesetDirExists(tt.dir)
			assert.Equal(t, tt.want, got)
		})
	}
}
