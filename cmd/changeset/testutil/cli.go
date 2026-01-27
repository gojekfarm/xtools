package testutil

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// CLIResult captures the result of running a CLI command.
type CLIResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Err      error
}

// RunCLI executes a CLI command in the given directory.
// Note: stdout is not captured to avoid test framework issues.
// Use Err field to check error messages.
func RunCLI(t *testing.T, dir string, cmd *cli.Command, args []string) CLIResult {
	t.Helper()

	// Override OsExiter to prevent os.Exit during tests
	origExiter := cli.OsExiter
	var capturedExitCode int
	cli.OsExiter = func(code int) {
		capturedExitCode = code
	}
	defer func() {
		cli.OsExiter = origExiter
	}()

	// Save and restore working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to directory %s: %v", dir, err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	// Run command
	ctx := context.Background()
	fullArgs := append([]string{"changeset"}, args...)
	runErr := cmd.Run(ctx, fullArgs)

	result := CLIResult{
		Err: runErr,
	}

	// Extract exit code from error or captured exit
	if runErr == nil {
		result.ExitCode = capturedExitCode
	} else if exitErr, ok := runErr.(cli.ExitCoder); ok {
		result.ExitCode = exitErr.ExitCode()
	} else {
		result.ExitCode = 1
	}

	// If we captured an exit code but no error, set a default exit code
	if result.ExitCode == 0 && capturedExitCode != 0 {
		result.ExitCode = capturedExitCode
	}

	return result
}

// CommitChanges commits all changes in a test repo.
func CommitChanges(t *testing.T, dir, message string) {
	t.Helper()

	cmds := [][]string{
		{"add", "."},
		{"commit", "-m", message},
	}
	for _, args := range cmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
}

// AssertFileContains checks that a file contains expected content.
func AssertFileContains(t *testing.T, path string, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err, "reading file %s", path)
	assert.Contains(t, string(content), expected, "file %s should contain %q", path, expected)
}

// AssertFileExists checks that a file exists.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, "file should exist: %s", path)
}

// AssertFileNotExists checks that a file does not exist.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), "file should not exist: %s", path)
}

// AssertGitTag checks that a git tag exists.
func AssertGitTag(t *testing.T, dir, tag string) {
	t.Helper()

	cmd := exec.Command("git", "tag", "-l", tag)
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), tag, "tag %s should exist", tag)
}
