//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/testutil"
)

func TestE2E_Status_WithChangesets(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir, "status")

	require.Equal(t, 0, result.ExitCode, "status should succeed")
	require.NoError(t, result.Err)

	// Verify changesets exist
	changesets, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)
	assert.Len(t, changesets, 1)
	assert.Equal(t, "happy-tiger-jump", changesets[0].ID)
}

func TestE2E_Status_NoChangesets(t *testing.T) {
	dir := testutil.SetupTestRepo(t)
	removeChangesets(t, dir)

	result := runCLI(t, dir, "status")

	require.Equal(t, 0, result.ExitCode, "status should succeed")
	require.NoError(t, result.Err)

	// Verify no changesets exist
	changesets, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)
	assert.Empty(t, changesets)
}

func TestE2E_Status_Verbose(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir, "status", "--verbose")

	require.Equal(t, 0, result.ExitCode, "status --verbose should succeed")
	require.NoError(t, result.Err)
}

func TestE2E_Status_NotInitialized(t *testing.T) {
	dir := setupTestRepoUninitialized(t)

	result := runCLI(t, dir, "status")

	assert.Equal(t, 1, result.ExitCode, "status should fail when not initialized")
	if assert.NotNil(t, result.Err, "should have an error") {
		assert.Contains(t, result.Err.Error(), "not initialized")
	}
}
