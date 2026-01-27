//go:build e2e

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/cmd/changeset/testutil"
)

func TestE2E_Init_Success(t *testing.T) {
	// Setup: repo without .changeset
	dir := setupTestRepoUninitialized(t)

	// Execute
	result := runCLI(t, dir, "init")

	// Verify
	require.Equal(t, 0, result.ExitCode, "init should succeed")
	require.NoError(t, result.Err)

	// Check files created
	testutil.AssertFileExists(t, filepath.Join(dir, ".changeset", "config.json"))
	testutil.AssertFileExists(t, filepath.Join(dir, ".changeset", "README.md"))

	// Check config content
	testutil.AssertFileContains(t,
		filepath.Join(dir, ".changeset", "config.json"),
		"github.com/test/fakerepo",
	)
}

func TestE2E_Init_AlreadyInitialized(t *testing.T) {
	// Setup: repo WITH .changeset already
	dir := testutil.SetupTestRepo(t)

	// Execute
	result := runCLI(t, dir, "init")

	// Verify - should fail because already initialized
	assert.Equal(t, 1, result.ExitCode, "init should fail when already initialized")
	if assert.NotNil(t, result.Err, "should have an error") {
		assert.Contains(t, result.Err.Error(), "already")
	}
}

func TestE2E_Init_NoGoMod(t *testing.T) {
	// Setup: empty temp directory without go.mod
	dir := t.TempDir()

	// Execute
	result := runCLI(t, dir, "init")

	// Verify - should fail without go.mod
	assert.Equal(t, 1, result.ExitCode, "init should fail without go.mod")
}
