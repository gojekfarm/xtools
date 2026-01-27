//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/testutil"
)

func TestE2E_Tag_Success(t *testing.T) {
	// Setup with manifest
	releases := []changeset.Release{
		{
			Module:          "libA",
			Version:         "v0.2.0",
			PreviousVersion: "v0.1.0",
			Bump:            changeset.BumpMinor,
		},
	}
	dir := setupTestRepoWithManifest(t, releases)

	result := runCLI(t, dir, "tag")

	require.Equal(t, 0, result.ExitCode, "tag should succeed: %s", result.Stdout)

	// Verify tag was created
	testutil.AssertGitTag(t, dir, "libA/v0.2.0")
}

func TestE2E_Tag_NoManifest(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir, "tag")

	assert.Equal(t, 1, result.ExitCode, "tag should fail without manifest")
	if assert.NotNil(t, result.Err, "should have an error") {
		assert.Contains(t, result.Err.Error(), "manifest")
	}
}

func TestE2E_Publish_NoPush(t *testing.T) {
	// Setup with manifest
	releases := []changeset.Release{
		{
			Module:          "libB",
			Version:         "v0.2.0",
			PreviousVersion: "v0.1.0",
			Bump:            changeset.BumpMinor,
		},
	}
	dir := setupTestRepoWithManifest(t, releases)

	result := runCLI(t, dir, "publish", "--no-push")

	require.Equal(t, 0, result.ExitCode, "publish --no-push should succeed: %s", result.Stdout)

	// Verify tag was created
	testutil.AssertGitTag(t, dir, "libB/v0.2.0")
}

func TestE2E_Publish_NoManifest(t *testing.T) {
	dir := testutil.SetupTestRepo(t)

	result := runCLI(t, dir, "publish")

	assert.Equal(t, 1, result.ExitCode, "publish should fail without manifest")
}
