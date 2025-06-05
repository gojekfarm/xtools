package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
	"golang.org/x/mod/semver"
)

func ListModules(ctx context.Context, cmd *cli.Command) error {
	fmt.Println("Listing modules...")
	mods, err := FindModules(".")
	if err != nil {
		return err
	}

	versions, err := GetAllTags(".")
	if err != nil {
		return err
	}

	indexed := indexModules(mods, versions)

	for _, mod := range mods {
		fmt.Println(mod.ShortName, mod.Version)

		for _, dep := range mod.Dependencies {
			m, ok := indexed[dep]
			if !ok {
				continue
			}

			fmt.Println("  |-" + m.ShortName + " " + m.Version)
		}
	}

	return nil
}

func indexModules(mods []*Module, versions map[string][]string) map[string]*Module {
	indexed := make(map[string]*Module)

	for _, mod := range mods {
		versions := versions[mod.ShortName]
		semver.Sort(versions)

		if len(versions) > 0 {
			mod.Version = versions[len(versions)-1]
		}

		indexed[mod.Name] = mod
	}

	return indexed
}
