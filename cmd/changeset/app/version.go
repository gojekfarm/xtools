package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/git"
	"github.com/gojekfarm/xtools/cmd/changeset/module"
)

// Version consumes changesets and updates versions.
func Version(ctx context.Context, cmd *cli.Command) error {
	dir := "."
	ignoreList := cmd.StringSlice("ignore")
	snapshot := cmd.Bool("snapshot")

	// Check if initialized
	if !changeset.ChangesetDirExists(dir) {
		return cli.Exit("Changeset not initialized. Run 'changeset init' first.", 1)
	}

	// Skip uncommitted changes check for snapshot mode
	if !snapshot {
		hasChanges, err := git.HasUncommittedChanges(dir)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to check git status: %v", err), 1)
		}
		if hasChanges {
			return cli.Exit("You have uncommitted changes. Please commit or stash them first.", 1)
		}
	}

	// Read config
	cfg, err := changeset.ReadConfig(dir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read config: %v", err), 1)
	}

	// Combine ignore lists
	allIgnore := append(cfg.Ignore, ignoreList...)

	// Read changesets
	changesets, err := changeset.ReadChangesets(dir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read changesets: %v", err), 1)
	}

	if len(changesets) == 0 {
		fmt.Println("No changesets found. Nothing to release.")
		return nil
	}

	// Discover modules
	graph, err := module.Discover(dir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to discover modules: %v", err), 1)
	}

	// Get current versions from git tags
	tags, err := git.GetLatestVersions(dir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get git tags: %v", err), 1)
	}

	// Compute releases
	releases, err := changeset.ComputeReleases(changesets, graph, tags, cfg)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to compute releases: %v", err), 1)
	}

	// Filter ignored modules
	releases = changeset.FilterIgnored(releases, allIgnore)

	if len(releases) == 0 {
		fmt.Println("No releases to process after filtering.")
		return nil
	}

	// Apply snapshot suffix if requested
	if snapshot {
		snapshotSuffix := fmt.Sprintf("-snapshot.%s", time.Now().Format("20060102150405"))
		for i := range releases {
			releases[i].Version = releases[i].Version + snapshotSuffix
		}
		fmt.Printf("Processing %d snapshot release(s)...\n\n", len(releases))
	} else {
		fmt.Printf("Processing %d release(s)...\n\n", len(releases))
	}

	// Build version map for go.mod updates
	versions := make(map[string]string)
	for _, r := range releases {
		versions[r.Module] = r.Version
	}

	// Update go.mod files for each module that has internal dependencies
	fmt.Println("Updating go.mod files...")
	for _, mod := range graph.AllModules() {
		if len(mod.Dependencies) == 0 {
			continue
		}

		gomodPath := filepath.Join(mod.Path, "go.mod")
		if err := module.UpdateGoMod(gomodPath, cfg.Root, versions); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to update %s: %v", gomodPath, err), 1)
		}
		fmt.Printf("  Updated: %s\n", gomodPath)
	}

	// Update changelogs for each released module (skip for snapshots)
	if !snapshot {
		fmt.Println("\nUpdating changelogs...")
		for _, r := range releases {
			mod := graph.FindModule(r.Module)
			if mod == nil {
				continue
			}

			// Generate changelog entry for this module
			entry := changeset.GenerateChangelog([]changeset.Release{r}, changesets)
			if entry == nil {
				continue
			}

			changelogPath := filepath.Join(mod.Path, "CHANGELOG.md")
			if err := changeset.UpdateChangelog(changelogPath, entry); err != nil {
				return cli.Exit(fmt.Sprintf("Failed to update changelog for %s: %v", r.Module, err), 1)
			}

			modDisplay := r.Module
			if modDisplay == "" {
				modDisplay = "(root)"
			}
			fmt.Printf("  Updated: %s (%s)\n", changelogPath, modDisplay)
		}

		// Delete consumed changesets (skip for snapshots)
		fmt.Println("\nDeleting consumed changesets...")
		for _, cs := range changesets {
			if err := changeset.DeleteChangeset(cs); err != nil {
				fmt.Printf("  Warning: Failed to delete %s: %v\n", cs.ID, err)
			} else {
				fmt.Printf("  Deleted: %s\n", cs.ID)
			}
		}
	} else {
		fmt.Println("\nSnapshot mode: Skipping changelog updates and changeset deletion.")
	}

	// Write release manifest
	fmt.Println("\nWriting release manifest...")
	manifest := &changeset.Manifest{Releases: releases}
	if err := changeset.WriteManifest(dir, manifest); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to write manifest: %v", err), 1)
	}

	// Print summary
	summaryTitle := "Release summary:"
	if snapshot {
		summaryTitle = "Snapshot release summary:"
	}
	fmt.Printf("\n%s\n", summaryTitle)
	for _, r := range releases {
		modDisplay := r.Module
		if modDisplay == "" {
			modDisplay = "(root)"
		}
		reason := ""
		if r.Reason != "" {
			reason = fmt.Sprintf(" (%s)", r.Reason)
		}
		fmt.Printf("  %s: %s -> %s%s\n", modDisplay, r.PreviousVersion, r.Version, reason)
	}

	if snapshot {
		fmt.Println("\nSnapshot next steps:")
		fmt.Println("  1. Run 'changeset publish --no-push' to create snapshot tags locally")
		fmt.Println("  2. Test the snapshot versions")
		fmt.Println("  3. Delete snapshot tags when done: git tag -d <tag>")
	} else {
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Review the changes")
		fmt.Println("  2. Commit the updated files")
		fmt.Println("  3. Run 'changeset publish' to create and push git tags")
	}

	return nil
}
