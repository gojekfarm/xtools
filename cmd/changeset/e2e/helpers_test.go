//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/app"
	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/testutil"
)

// buildCLI returns the CLI command tree for testing.
func buildCLI() *cli.Command {
	return app.BuildCLI("test")
}

// runCLI is a convenience wrapper around testutil.RunCLI.
func runCLI(t *testing.T, dir string, args ...string) testutil.CLIResult {
	t.Helper()
	return testutil.RunCLI(t, dir, buildCLI(), args)
}

// createChangeset creates a changeset file programmatically.
func createChangeset(t *testing.T, dir string, modules map[string]string, summary string) string {
	t.Helper()

	cs := &changeset.Changeset{
		Modules: make(map[string]changeset.Bump),
		Summary: summary,
	}
	for mod, bump := range modules {
		cs.Modules[mod] = changeset.Bump(bump)
	}

	if err := changeset.WriteChangeset(dir, cs); err != nil {
		t.Fatalf("failed to create changeset: %v", err)
	}

	return cs.ID
}

// setupTestRepoUninitialized creates a repo WITHOUT .changeset directory.
func setupTestRepoUninitialized(t *testing.T) string {
	t.Helper()
	dir := testutil.SetupTestRepo(t)

	// Remove .changeset directory
	changesetDir := filepath.Join(dir, ".changeset")
	if err := os.RemoveAll(changesetDir); err != nil {
		t.Fatalf("failed to remove .changeset: %v", err)
	}

	return dir
}

// setupTestRepoWithManifest creates a repo with a release manifest.
func setupTestRepoWithManifest(t *testing.T, releases []changeset.Release) string {
	t.Helper()
	dir := testutil.SetupTestRepo(t)

	manifest := &changeset.Manifest{Releases: releases}
	if err := changeset.WriteManifest(dir, manifest); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	return dir
}

// removeChangesets removes all changeset files (except README.md) from the repo.
func removeChangesets(t *testing.T, dir string) {
	t.Helper()

	changesetDir := filepath.Join(dir, ".changeset")
	files, err := filepath.Glob(filepath.Join(changesetDir, "*.md"))
	if err != nil {
		t.Fatalf("failed to glob changesets: %v", err)
	}

	for _, f := range files {
		if filepath.Base(f) != "README.md" {
			if err := os.Remove(f); err != nil {
				t.Fatalf("failed to remove changeset %s: %v", f, err)
			}
		}
	}
}
