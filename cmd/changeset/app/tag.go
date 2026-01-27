package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/git"
)

// Tag creates git tags without pushing.
func Tag(ctx context.Context, cmd *cli.Command) error {
	dir := "."

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

	// Create tags
	slog.Info("Creating git tags")
	var tags []git.Tag
	for _, r := range manifest.Releases {
		tag := git.Tag{
			Name:    git.FormatTag(r.Module, r.Version),
			Module:  r.Module,
			Version: r.Version,
		}
		tags = append(tags, tag)

		if err := git.CreateTag(dir, r.Module, r.Version); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to create tag %s: %v", tag.Name, err), 1)
		}

		slog.Info("Created tag", "tag", tag.Name, "module", displayModule(r.Module), "version", r.Version)
	}

	slog.Info("Tags created", "count", len(tags))
	slog.Info("To push tags",
		"option1", "changeset publish",
		"option2", "git push origin --tags",
	)

	return nil
}
