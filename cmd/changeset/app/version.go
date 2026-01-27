package app

import (
	"context"
	"fmt"
	"log/slog"
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
		slog.Info("No changesets found, nothing to release")
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
		slog.Info("No releases to process after filtering")
		return nil
	}

	// Apply snapshot suffix if requested
	if snapshot {
		snapshotSuffix := fmt.Sprintf("-snapshot.%s", time.Now().Format("20060102150405"))
		for i := range releases {
			releases[i].Version = releases[i].Version + snapshotSuffix
		}
		slog.Info("Processing snapshot releases", "count", len(releases))
	} else {
		slog.Info("Processing releases", "count", len(releases))
	}

	// Build version map for go.mod updates
	versions := make(map[string]string)
	for _, r := range releases {
		versions[r.Module] = r.Version
	}

	// Update go.mod files for each module that has internal dependencies
	slog.Info("Updating go.mod files")
	for _, mod := range graph.AllModules() {
		if len(mod.Dependencies) == 0 {
			continue
		}

		gomodPath := filepath.Join(mod.Path, "go.mod")
		if err := module.UpdateGoMod(gomodPath, cfg.Root, versions); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to update %s: %v", gomodPath, err), 1)
		}
		slog.Info("Updated go.mod", "path", gomodPath)
	}

	// Update changelogs for each released module (skip for snapshots)
	if !snapshot {
		slog.Info("Updating changelogs")
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

			slog.Info("Updated changelog", "path", changelogPath, "module", displayModule(r.Module))
		}

		// Delete consumed changesets (skip for snapshots)
		slog.Info("Deleting consumed changesets")
		for _, cs := range changesets {
			if err := changeset.DeleteChangeset(cs); err != nil {
				slog.Warn("Failed to delete changeset", "id", cs.ID, "error", err)
			} else {
				slog.Info("Deleted changeset", "id", cs.ID)
			}
		}
	} else {
		slog.Info("Snapshot mode: Skipping changelog updates and changeset deletion")
	}

	// Write release manifest
	slog.Info("Writing release manifest")
	manifest := &changeset.Manifest{Releases: releases}
	if err := changeset.WriteManifest(dir, manifest); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to write manifest: %v", err), 1)
	}

	// Print summary
	summaryTitle := "Release summary"
	if snapshot {
		summaryTitle = "Snapshot release summary"
	}
	slog.Info(summaryTitle)
	for _, r := range releases {
		slog.Info("Release",
			"module", displayModule(r.Module),
			"from", r.PreviousVersion,
			"to", r.Version,
			"reason", r.Reason,
		)
	}

	if snapshot {
		slog.Info("Snapshot next steps",
			"1", "Run 'changeset publish --no-push' to create snapshot tags locally",
			"2", "Test the snapshot versions",
			"3", "Delete snapshot tags when done: git tag -d <tag>",
		)
	} else {
		slog.Info("Next steps",
			"1", "Review the changes",
			"2", "Commit the updated files",
			"3", "Run 'changeset publish' to create and push git tags",
		)
	}

	return nil
}
