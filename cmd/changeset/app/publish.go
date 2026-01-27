package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/git"
)

// Publish creates git tags and pushes them.
func Publish(ctx context.Context, cmd *cli.Command) error {
	dir := "."
	noPush := cmd.Bool("no-push")

	// Read release manifest
	manifest, err := changeset.ReadManifest(dir)
	if err != nil {
		if err == changeset.ErrNoManifest {
			return cli.Exit("No release manifest found. Run 'changeset version' first.", 1)
		}
		return cli.Exit(fmt.Sprintf("Failed to read manifest: %v", err), 1)
	}

	if len(manifest.Releases) == 0 {
		slog.Info("No releases in manifest")
		return nil
	}

	// Build tags list
	var tags []git.Tag
	for _, r := range manifest.Releases {
		tags = append(tags, git.Tag{
			Name:    git.FormatTag(r.Module, r.Version),
			Module:  r.Module,
			Version: r.Version,
		})
	}

	// Create tags
	slog.Info("Creating git tags")
	for _, tag := range tags {
		if err := git.CreateTag(dir, tag.Module, tag.Version); err != nil {
			// Tag might already exist from previous 'changeset tag' run
			slog.Warn("Tag may already exist", "tag", tag.Name, "error", err)
		} else {
			slog.Info("Created tag", "tag", tag.Name, "module", displayModule(tag.Module), "version", tag.Version)
		}
	}

	// Push tags
	if !noPush {
		slog.Info("Pushing tags to origin")
		if err := git.PushTags(dir, tags, "origin"); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to push tags: %v", err), 1)
		}
		slog.Info("Tags pushed successfully")

		// Delete manifest after successful push
		if err := changeset.DeleteManifest(dir); err != nil {
			slog.Warn("Failed to delete manifest", "error", err)
		} else {
			slog.Info("Release manifest cleaned up")
		}
	} else {
		slog.Info("Tags created locally",
			"hint", "To push manually: git push origin --tags",
		)
	}

	// Print summary
	slog.Info("Published releases", "count", len(tags))
	for _, r := range manifest.Releases {
		slog.Info("Published", "module", displayModule(r.Module), "version", r.Version)
	}

	return nil
}
