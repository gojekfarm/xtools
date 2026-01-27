package changeset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Release represents a computed release for a module.
type Release struct {
	Module          string `json:"module"`
	Version         string `json:"version"`
	PreviousVersion string `json:"previousVersion"`
	Bump            Bump   `json:"bump"`
	Reason          string `json:"reason,omitempty"` // "dependency" if auto-bumped
}

// Manifest represents .changeset/release-manifest.json.
type Manifest struct {
	Releases []Release `json:"releases"`
}

const manifestFileName = "release-manifest.json"

// ReadManifest reads .changeset/release-manifest.json.
// Returns error if file doesn't exist.
func ReadManifest(dir string) (*Manifest, error) {
	manifestPath := filepath.Join(dir, ".changeset", manifestFileName)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoManifest
		}
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	return &manifest, nil
}

// WriteManifest writes the manifest to .changeset/release-manifest.json.
func WriteManifest(dir string, m *Manifest) error {
	changesetDir := filepath.Join(dir, ".changeset")

	// Ensure directory exists
	if err := os.MkdirAll(changesetDir, 0755); err != nil {
		return fmt.Errorf("creating .changeset directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	manifestPath := filepath.Join(changesetDir, manifestFileName)
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	return nil
}

// DeleteManifest removes .changeset/release-manifest.json.
func DeleteManifest(dir string) error {
	manifestPath := filepath.Join(dir, ".changeset", manifestFileName)
	if err := os.Remove(manifestPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("deleting manifest: %w", err)
	}
	return nil
}

// ManifestExists returns true if release-manifest.json exists.
func ManifestExists(dir string) bool {
	manifestPath := filepath.Join(dir, ".changeset", manifestFileName)
	_, err := os.Stat(manifestPath)
	return err == nil
}
