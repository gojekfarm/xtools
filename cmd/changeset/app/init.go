package app

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/module"
)

// Init initializes the .changeset directory with config and README.
func Init(ctx context.Context, cmd *cli.Command) error {
	dir := "."

	// Check if already initialized
	if changeset.ChangesetDirExists(dir) {
		return cli.Exit("Changeset already initialized. Run 'changeset status' to see pending changesets.", 1)
	}

	// Discover root module to populate config
	graph, err := module.Discover(dir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to discover modules: %v", err), 1)
	}

	// Create config with root module path
	cfg := changeset.DefaultConfig()
	cfg.Root = graph.Root.Name

	// Initialize changeset directory
	if err := changeset.InitChangeset(dir, cfg); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to initialize: %v", err), 1)
	}

	fmt.Println("Changeset initialized successfully!")
	fmt.Printf("  Root module: %s\n", cfg.Root)
	fmt.Printf("  Config: .changeset/config.json\n")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'changeset add' to create your first changeset")
	fmt.Println("  2. Run 'changeset status' to see pending releases")

	return nil
}
