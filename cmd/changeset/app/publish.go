package app

import (
	"context"
	"fmt"

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
		fmt.Println("No releases in manifest.")
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
	fmt.Println("Creating git tags...")
	for _, tag := range tags {
		if err := git.CreateTag(dir, tag.Module, tag.Version); err != nil {
			// Tag might already exist from previous 'changeset tag' run
			fmt.Printf("  %s (may already exist: %v)\n", tag.Name, err)
		} else {
			modDisplay := tag.Module
			if modDisplay == "" {
				modDisplay = "(root)"
			}
			fmt.Printf("  Created: %s (%s %s)\n", tag.Name, modDisplay, tag.Version)
		}
	}

	// Push tags
	if !noPush {
		fmt.Println("\nPushing tags to origin...")
		if err := git.PushTags(dir, tags, "origin"); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to push tags: %v", err), 1)
		}
		fmt.Println("Tags pushed successfully!")

		// Delete manifest after successful push
		if err := changeset.DeleteManifest(dir); err != nil {
			fmt.Printf("Warning: Failed to delete manifest: %v\n", err)
		} else {
			fmt.Println("Release manifest cleaned up.")
		}
	} else {
		fmt.Println("\nTags created locally (--no-push specified).")
		fmt.Println("To push manually: git push origin --tags")
	}

	// Print summary
	fmt.Printf("\nPublished %d release(s):\n", len(tags))
	for _, r := range manifest.Releases {
		modDisplay := r.Module
		if modDisplay == "" {
			modDisplay = "(root)"
		}
		fmt.Printf("  %s: %s\n", modDisplay, r.Version)
	}

	return nil
}
