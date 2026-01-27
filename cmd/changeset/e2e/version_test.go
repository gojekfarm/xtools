//go:build e2e

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/testutil"
)

func TestE2E_Version_Success(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	// Everything is already committed by SetupTestRepo
	result := runCLI(t, dir, "version")

	require.Equal(t, 0, result.ExitCode, "version should succeed")
	require.NoError(t, result.Err)

	// Verify manifest was created
	testutil.AssertFileExists(t, filepath.Join(dir, ".changeset", "release-manifest.json"))

	// Verify changesets were consumed
	changesets, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)
	assert.Empty(t, changesets, "changesets should be consumed")
}

func TestE2E_Version_NoChangesets(t *testing.T) {
	dir := testutil.SetupTestRepo(t)
	removeChangesets(t, dir)
	testutil.CommitChanges(t, dir, "remove changesets")

	result := runCLI(t, dir, "version")

	// Should succeed but indicate no releases
	require.Equal(t, 0, result.ExitCode, "version should succeed")
	require.NoError(t, result.Err)

	// No manifest should be created when there are no changesets
	_, err := changeset.ReadManifest(dir)
	assert.Error(t, err, "no manifest should exist when there are no changesets")
}

func TestE2E_Version_Snapshot(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir, "version", "--snapshot")

	require.Equal(t, 0, result.ExitCode, "version --snapshot should succeed")
	require.NoError(t, result.Err)

	// Verify manifest was created with snapshot versions
	manifest, err := changeset.ReadManifest(dir)
	require.NoError(t, err)
	require.NotEmpty(t, manifest.Releases)

	// Snapshot versions should contain timestamp-like suffix
	for _, rel := range manifest.Releases {
		assert.Contains(t, rel.Version, "-", "snapshot version should have prerelease suffix")
	}
}

func TestE2E_Version_NotInitialized(t *testing.T) {
	dir := setupTestRepoUninitialized(t)

	result := runCLI(t, dir, "version")

	assert.Equal(t, 1, result.ExitCode, "version should fail when not initialized")
	if assert.NotNil(t, result.Err, "should have an error") {
		assert.Contains(t, result.Err.Error(), "not initialized")
	}
}
