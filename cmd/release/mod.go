package main

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// Module represents a Go module.
type Module struct {
	Name         string
	ShortName    string
	Version      string
	Path         string
	Dependencies []string
}

// FindModules finds all Go modules in the given directory.
func FindModules(dir string) ([]*Module, error) {
	var mods []*Module

	root, err := ParseGoModule(dir, nil)
	if err != nil {
		return nil, err
	}

	mods = append(mods, root)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != dir {
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				m, err := ParseGoModule(path, root)
				if err != nil {
					return err
				}

				mods = append(mods, m)
			}
		}

		return nil
	})

	return mods, err
}

func ParseGoModule(dir string, root *Module) (*Module, error) {
	modPath := filepath.Join(dir, "go.mod")

	contentBytes, err := os.ReadFile(modPath)
	if err != nil {
		return nil, err
	}

	goMod, err := modfile.Parse(modPath, contentBytes, nil)
	if err != nil {
		return nil, err
	}

	m := &Module{
		Name:         goMod.Module.Mod.Path,
		ShortName:    goMod.Module.Mod.Path,
		Version:      goMod.Module.Mod.Version,
		Path:         dir,
		Dependencies: []string{},
	}

	if root != nil {
		m.ShortName = strings.TrimPrefix(m.Name, root.Name+"/")
	}

	for _, r := range goMod.Require {
		if r.Indirect {
			continue
		}

		m.Dependencies = append(m.Dependencies, r.Mod.Path)
	}

	return m, nil
}
