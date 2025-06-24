package main

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/urfave/cli/v3"
)

func CreateRelease(ctx context.Context, cmd *cli.Command) error {
	version := cmd.String("version")
	major := cmd.Bool("major")
	minor := cmd.Bool("minor")
	patch := cmd.Bool("patch")

	if version == "" && !major && !minor && !patch {
		return cli.Exit("You must specify --version or one of --major, --minor, or --patch", 1)
	}

	if version != "" && (major || minor || patch) {
		return cli.Exit("Cannot use --version with --major, --minor, or --patch", 1)
	}

	latestTag, err := GetLatestTag(".")
	if err != nil {
		return cli.Exit("Failed to get latest tag: "+err.Error(), 1)
	}

	fmt.Println("Latest tag:", latestTag)

	var nextVersion string
	if version != "" {
		nextVersion = version
	} else {
		nextVersion, err = getNextVersion(latestTag, major, minor, patch)
		if err != nil {
			return cli.Exit("Failed to get next version: "+err.Error(), 1)
		}
	}

	fmt.Println("Next version:", nextVersion)

	root, mods, err := FindModules(".")
	if err != nil {
		return cli.Exit("Failed to find modules: "+err.Error(), 1)
	}

	fmt.Println("Updating dependencies...")

	err = UpdateDependencies(root, root, nextVersion)
	if err != nil {
		return cli.Exit("Failed to update dependencies for root module: "+err.Error(), 1)
	}

	for _, mod := range mods {
		err := UpdateDependencies(root, mod, nextVersion)
		if err != nil {
			return cli.Exit("Failed to update dependencies for module "+mod.Name+": "+err.Error(), 1)
		}
	}

	return nil
}

func getNextVersion(latestTag string, major bool, minor bool, patch bool) (string, error) {
	v, err := semver.NewVersion(latestTag)
	if err != nil {
		return "", err
	}

	if major {
		return "v" + v.IncMajor().String(), nil
	} else if minor {
		return "v" + v.IncMinor().String(), nil
	} else if patch {
		return "v" + v.IncPatch().String(), nil
	}

	return "", fmt.Errorf("no version increment specified")
}
