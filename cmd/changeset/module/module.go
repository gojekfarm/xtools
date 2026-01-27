package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Module represents a Go module in the repository.
type Module struct {
	Name         string   // Full module path (e.g., "github.com/gojekfarm/xtools/xkafka")
	ShortName    string   // Relative to root (e.g., "xkafka"), empty for root module
	Path         string   // Filesystem path to module directory
	Dependencies []string // Internal module dependencies (short names)
}

// Graph represents the module dependency graph.
type Graph struct {
	Root    *Module
	Modules map[string]*Module // Short name -> Module (empty string key for root)
}

// Discover finds all Go modules in a directory tree.
// Returns a Graph containing the root module and all submodules.
func Discover(dir string) (*Graph, error) {
	// Find and parse root module
	rootModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(rootModPath); err != nil {
		return nil, fmt.Errorf("no go.mod found in %s: %w", dir, err)
	}

	rootName, rootDeps, err := ParseGoMod(rootModPath)
	if err != nil {
		return nil, fmt.Errorf("parsing root go.mod: %w", err)
	}

	root := &Module{
		Name:      rootName,
		ShortName: "",
		Path:      dir,
	}

	graph := &Graph{
		Root:    root,
		Modules: make(map[string]*Module),
	}
	graph.Modules[""] = root

	// Walk directory tree to find submodules
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-directories
		if !info.IsDir() {
			return nil
		}

		// Skip the root directory (already processed)
		if path == dir {
			return nil
		}

		// Skip hidden directories and common non-module directories
		name := info.Name()
		if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" {
			return filepath.SkipDir
		}

		// Check for go.mod
		modPath := filepath.Join(path, "go.mod")
		if _, err := os.Stat(modPath); os.IsNotExist(err) {
			return nil
		}

		moduleName, deps, err := ParseGoMod(modPath)
		if err != nil {
			return fmt.Errorf("parsing go.mod at %s: %w", modPath, err)
		}

		// Calculate short name relative to root
		shortName := strings.TrimPrefix(moduleName, rootName+"/")
		if shortName == moduleName {
			// Module path doesn't start with root, unusual but handle it
			shortName = moduleName
		}

		mod := &Module{
			Name:      moduleName,
			ShortName: shortName,
			Path:      path,
		}

		// Filter dependencies to only internal ones
		for _, dep := range deps {
			if strings.HasPrefix(dep, rootName+"/") {
				depShortName := strings.TrimPrefix(dep, rootName+"/")
				mod.Dependencies = append(mod.Dependencies, depShortName)
			} else if dep == rootName {
				// Depends on root module
				mod.Dependencies = append(mod.Dependencies, "")
			}
		}

		graph.Modules[shortName] = mod
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Filter root module dependencies to internal ones
	for _, dep := range rootDeps {
		if strings.HasPrefix(dep, rootName+"/") {
			depShortName := strings.TrimPrefix(dep, rootName+"/")
			root.Dependencies = append(root.Dependencies, depShortName)
		}
	}

	return graph, nil
}
