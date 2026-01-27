package app

import (
	"context"
	"fmt"

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
		fmt.Println("No releases in manifest.")
		return nil
	}

	// Create tags
	fmt.Println("Creating git tags...")
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

		modDisplay := r.Module
		if modDisplay == "" {
			modDisplay = "(root)"
		}
		fmt.Printf("  Created: %s (%s %s)\n", tag.Name, modDisplay, r.Version)
	}

	fmt.Printf("\nCreated %d tag(s).\n", len(tags))
	fmt.Println("\nTo push tags, run: changeset publish")
	fmt.Println("Or manually: git push origin --tags")

	return nil
}
