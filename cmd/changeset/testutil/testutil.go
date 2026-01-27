package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// SetupTestRepo creates a test repository in t.TempDir() with:
// - Multi-module Go structure (root + libA, libB, libC)
// - Initialized git repository with initial commit
// - Tags: v0.1.0, libA/v0.1.0, libB/v0.1.0, libC/v0.1.0
// - .changeset directory with config and sample changeset
func SetupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create directory structure
	dirs := []string{
		".changeset",
		"libA",
		"libB",
		"libC",
		"pkg/core",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", d, err)
		}
	}

	// File contents matching testdata/fakerepo
	files := map[string]string{
		"go.mod":      "module github.com/test/fakerepo\n\ngo 1.21\n",
		"libA/go.mod": "module github.com/test/fakerepo/libA\n\ngo 1.21\n",
		"libA/liba.go": `package libA

// Hello returns a greeting.
func Hello() string {
	return "Hello from libA"
}
`,
		"libB/go.mod": `module github.com/test/fakerepo/libB

go 1.21

require github.com/test/fakerepo/libA v0.1.0
`,
		"libB/libb.go": `package libB

// Greeting returns a greeting using libA.
func Greeting() string {
	return "Hello from libB"
}
`,
		"libC/go.mod": `module github.com/test/fakerepo/libC

go 1.21

require (
	github.com/test/fakerepo/libA v0.1.0
	github.com/test/fakerepo/libB v0.1.0
)
`,
		"libC/libc.go": `package libC

// Combined returns a combined greeting.
func Combined() string {
	return "Hello from libC"
}
`,
		"pkg/core/core.go": `package core

// Version is the version of the core package.
const Version = "1.0.0"
`,
		".changeset/config.json": `{
  "root": "github.com/test/fakerepo",
  "baseBranch": "main",
  "ignore": [],
  "ignorePaths": [],
  "dependentBump": "patch"
}
`,
		".changeset/README.md": "# Changesets\n\nThis directory contains changeset files for testing.\n",
		".changeset/happy-tiger-jump.md": `---
"libA": minor
---

Add new greeting function to libA.
`,
	}

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", name, err)
		}
	}

	// Initialize git repository
	gitCommands := [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
		{"add", "."},
		{"commit", "-m", "initial"},
		{"tag", "v0.1.0"},
		{"tag", "libA/v0.1.0"},
		{"tag", "libB/v0.1.0"},
		{"tag", "libC/v0.1.0"},
	}

	for _, args := range gitCommands {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	return dir
}
