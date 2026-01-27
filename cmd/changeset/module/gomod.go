// Package module provides Go module discovery and dependency operations.
package module

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/mod/modfile"
)

// ParseGoMod reads and parses a go.mod file.
// Returns the module path and direct dependencies.
func ParseGoMod(path string) (modulePath string, deps []string, err error) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("reading go.mod: %w", err)
	}

	goMod, err := modfile.Parse(path, contentBytes, nil)
	if err != nil {
		return "", nil, fmt.Errorf("parsing go.mod: %w", err)
	}

	modulePath = goMod.Module.Mod.Path

	for _, r := range goMod.Require {
		if r.Indirect {
			continue
		}
		deps = append(deps, r.Mod.Path)
	}

	return modulePath, deps, nil
}

// UpdateGoMod updates internal dependency versions in a go.mod file.
//
// Parameters:
//   - path: Path to go.mod file
//   - root: Root module path (to identify internal deps)
//   - versions: Map of module short name -> new version
func UpdateGoMod(path string, root string, versions map[string]string) error {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}

	goMod, err := modfile.Parse(path, contentBytes, nil)
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}

	for _, r := range goMod.Require {
		// Check if this is an internal dependency
		if !strings.HasPrefix(r.Mod.Path, root) {
			continue
		}

		// Get the short name (relative to root)
		shortName := strings.TrimPrefix(r.Mod.Path, root+"/")
		if shortName == r.Mod.Path {
			// This is the root module itself
			shortName = ""
		}

		newVersion, ok := versions[shortName]
		if !ok {
			continue
		}

		if err := goMod.AddRequire(r.Mod.Path, newVersion); err != nil {
			return fmt.Errorf("updating dependency %s: %w", r.Mod.Path, err)
		}
	}

	goMod.Cleanup()

	content, err := goMod.Format()
	if err != nil {
		return fmt.Errorf("formatting go.mod: %w", err)
	}

	return os.WriteFile(path, content, 0644)
}
