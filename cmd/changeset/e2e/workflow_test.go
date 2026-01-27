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

func TestE2E_FullReleaseWorkflow(t *testing.T) {
	// Setup: fresh repo with one changeset (everything is already committed)
	dir := testutil.SetupTestRepo(t)

	// Step 1: Check status - should show pending changeset
	statusResult := runCLI(t, dir, "status")
	require.Equal(t, 0, statusResult.ExitCode, "status should succeed")
	require.NoError(t, statusResult.Err)

	// Verify changesets exist before version command
	changesetsBefore, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)
	require.Len(t, changesetsBefore, 1, "should have one changeset")

	// Step 2: Run version command
	versionResult := runCLI(t, dir, "version")
	require.Equal(t, 0, versionResult.ExitCode, "version should succeed")
	require.NoError(t, versionResult.Err)

	// Verify manifest was created
	testutil.AssertFileExists(t, filepath.Join(dir, ".changeset", "release-manifest.json"))

	// Verify changeset was consumed
	changesetsAfter, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)
	assert.Empty(t, changesetsAfter, "changesets should be consumed")

	// Verify CHANGELOG was updated
	testutil.AssertFileContains(t,
		filepath.Join(dir, "libA", "CHANGELOG.md"),
		"0.2.0",
	)

	// Step 3: Create tags (without push since no remote)
	tagResult := runCLI(t, dir, "tag")
	require.Equal(t, 0, tagResult.ExitCode, "tag should succeed")
	require.NoError(t, tagResult.Err)

	// Verify tags created
	testutil.AssertGitTag(t, dir, "libA/v0.2.0")
}

func TestE2E_DependencyCascade(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	// Remove existing changeset
	removeChangesets(t, dir)

	// Create changeset for libA only (libB and libC depend on it)
	createChangeset(t, dir,
		map[string]string{"libA": "minor"},
		"Add new feature to libA",
	)

	// Commit the changeset change
	testutil.CommitChanges(t, dir, "add libA changeset")

	// Run version
	result := runCLI(t, dir, "version")
	require.Equal(t, 0, result.ExitCode, "version should succeed")
	require.NoError(t, result.Err)

	// Verify cascade: libB and libC should also be bumped (patch due to dependency)
	manifest, err := changeset.ReadManifest(dir)
	require.NoError(t, err)

	// Should have releases for libA, libB, and libC
	moduleNames := make([]string, len(manifest.Releases))
	for i, rel := range manifest.Releases {
		moduleNames[i] = rel.Module
	}

	assert.Contains(t, moduleNames, "libA", "libA should be released")
	assert.Contains(t, moduleNames, "libB", "libB should be released (depends on libA)")
	assert.Contains(t, moduleNames, "libC", "libC should be released (depends on libA)")
}

func TestE2E_AddThenRelease(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	// Remove existing changesets
	removeChangesets(t, dir)
	testutil.CommitChanges(t, dir, "remove existing changesets")

	// Add a new changeset using non-interactive mode
	addResult := runCLI(t, dir,
		"add",
		"--module", "libB:patch",
		"--summary", "Fix bug in libB",
	)
	require.Equal(t, 0, addResult.ExitCode, "add should succeed")
	require.NoError(t, addResult.Err)

	// Commit the new changeset
	testutil.CommitChanges(t, dir, "add libB changeset")

	// Check status - verify changeset exists
	changesets, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)
	require.Len(t, changesets, 1, "should have one changeset")

	// Verify the changeset includes libB
	cs := changesets[0]
	_, hasLibB := cs.Modules["libB"]
	assert.True(t, hasLibB, "changeset should include libB")

	// Run version
	versionResult := runCLI(t, dir, "version")
	require.Equal(t, 0, versionResult.ExitCode, "version should succeed")
	require.NoError(t, versionResult.Err)

	// Verify manifest
	manifest, err := changeset.ReadManifest(dir)
	require.NoError(t, err)

	// Find libB release
	var libBRelease *changeset.Release
	for i := range manifest.Releases {
		if manifest.Releases[i].Module == "libB" {
			libBRelease = &manifest.Releases[i]
			break
		}
	}

	require.NotNil(t, libBRelease, "libB should have a release")
	assert.Equal(t, "v0.1.1", libBRelease.Version, "libB should be bumped to patch version")
}

func TestE2E_InitThenAdd(t *testing.T) {
	// Start with uninitialized repo
	dir := setupTestRepoUninitialized(t)

	// Initialize
	initResult := runCLI(t, dir, "init")
	require.Equal(t, 0, initResult.ExitCode, "init should succeed")

	// Add empty changeset
	addResult := runCLI(t, dir, "add", "--empty")
	require.Equal(t, 0, addResult.ExitCode, "add should succeed after init")

	// Verify changeset exists
	changesets, err := changeset.ReadChangesets(dir)
	require.NoError(t, err)
	assert.Len(t, changesets, 1)
}
