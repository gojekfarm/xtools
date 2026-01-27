//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/testutil"
)

func TestE2E_Add_Empty(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir, "add", "--empty")

	require.Equal(t, 0, result.ExitCode, "add --empty should succeed")
	require.NoError(t, result.Err)

	// Verify changeset was created (should have 2 now: original + new empty)
	changesets, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)
	assert.Len(t, changesets, 2)
}

func TestE2E_Add_NonInteractive(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir,
		"add",
		"--module", "libA:minor",
		"--module", "libB:patch",
		"--summary", "Test changeset from e2e",
	)

	require.Equal(t, 0, result.ExitCode, "add with flags should succeed")
	require.NoError(t, result.Err)

	// Verify content
	changesets, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)

	// Find the new changeset (not the original happy-tiger-jump)
	var newCS *changeset.Changeset
	for _, cs := range changesets {
		if cs.ID != "happy-tiger-jump" {
			newCS = cs
			break
		}
	}

	require.NotNil(t, newCS, "should have created a new changeset")
	assert.Equal(t, changeset.BumpMinor, newCS.Modules["libA"])
	assert.Equal(t, changeset.BumpPatch, newCS.Modules["libB"])
	assert.Contains(t, newCS.Summary, "Test changeset from e2e")
}

func TestE2E_Add_NonInteractive_InvalidBump(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir,
		"add",
		"--module", "libA:invalid",
		"--summary", "Test",
	)

	assert.Equal(t, 1, result.ExitCode, "add with invalid bump should fail")
	if assert.NotNil(t, result.Err, "should have an error") {
		assert.Contains(t, result.Err.Error(), "invalid bump type")
	}
}

func TestE2E_Add_NonInteractive_InvalidFormat(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir,
		"add",
		"--module", "libA", // Missing :bump
		"--summary", "Test",
	)

	assert.Equal(t, 1, result.ExitCode, "add with invalid format should fail")
	if assert.NotNil(t, result.Err, "should have an error") {
		assert.Contains(t, result.Err.Error(), "invalid module:bump format")
	}
}

func TestE2E_Add_NotInitialized(t *testing.T) {
	dir := setupTestRepoUninitialized(t)

	result := runCLI(t, dir, "add", "--empty")

	assert.Equal(t, 1, result.ExitCode, "add should fail when not initialized")
	if assert.NotNil(t, result.Err, "should have an error") {
		assert.Contains(t, result.Err.Error(), "not initialized")
	}
}
