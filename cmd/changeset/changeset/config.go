package changeset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents .changeset/config.json.
type Config struct {
	Root          string   `json:"root"`          // Root module path
	BaseBranch    string   `json:"baseBranch"`    // Default: "main"
	Ignore        []string `json:"ignore"`        // Modules to ignore
	IgnorePaths   []string `json:"ignorePaths"`   // File paths to ignore in CI check
	DependentBump Bump     `json:"dependentBump"` // How to bump dependents (default: "patch")
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		BaseBranch:    "main",
		DependentBump: BumpPatch,
	}
}

// ReadConfig reads .changeset/config.json.
// Returns DefaultConfig if file doesn't exist.
func ReadConfig(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ".changeset", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// WriteConfig writes config to .changeset/config.json.
func WriteConfig(dir string, cfg *Config) error {
	changesetDir := filepath.Join(dir, ".changeset")

	// Ensure directory exists
	if err := os.MkdirAll(changesetDir, 0755); err != nil {
		return fmt.Errorf("creating .changeset directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	configPath := filepath.Join(changesetDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

const readmeContent = `# Changesets

This directory contains changeset files that describe changes to the codebase.

## What is a changeset?

A changeset is a file that describes which packages should be released and how
(major, minor, or patch). When it's time to release, these changesets are
consumed to determine version bumps.

## Creating a changeset

Run:
` + "```" + `
changeset add
` + "```" + `

This will interactively create a changeset file.

## File format

` + "```" + `markdown
---
"package-name": minor
"other-package": patch
---

Description of the changes.
` + "```" + `
`

// InitChangeset creates the .changeset directory with config and README.
func InitChangeset(dir string, cfg *Config) error {
	changesetDir := filepath.Join(dir, ".changeset")

	// Check if already initialized
	if _, err := os.Stat(changesetDir); err == nil {
		return fmt.Errorf(".changeset directory already exists")
	}

	// Create directory
	if err := os.MkdirAll(changesetDir, 0755); err != nil {
		return fmt.Errorf("creating .changeset directory: %w", err)
	}

	// Write config
	if err := WriteConfig(dir, cfg); err != nil {
		return err
	}

	// Write README
	readmePath := filepath.Join(changesetDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("writing README: %w", err)
	}

	return nil
}

// ChangesetDirExists returns true if the .changeset directory exists.
func ChangesetDirExists(dir string) bool {
	changesetDir := filepath.Join(dir, ".changeset")
	_, err := os.Stat(changesetDir)
	return err == nil
}
