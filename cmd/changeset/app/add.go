package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/module"
)

// Add creates a new changeset interactively or with flags.
func Add(ctx context.Context, cmd *cli.Command) error {
	dir := "."
	empty := cmd.Bool("empty")
	openEditor := cmd.Bool("open")

	// Check if initialized
	if !changeset.ChangesetDirExists(dir) {
		return cli.Exit("Changeset not initialized. Run 'changeset init' first.", 1)
	}

	var cs *changeset.Changeset

	if empty {
		// Create empty changeset
		cs = &changeset.Changeset{
			Modules: make(map[string]changeset.Bump),
			Summary: "Empty changeset for changes that don't require a release.",
		}
	} else {
		// Interactive flow
		var err error
		cs, err = createChangesetInteractive(dir)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to create changeset: %v", err), 1)
		}
	}

	// Write changeset
	if err := changeset.WriteChangeset(dir, cs); err != nil {
		return cli.Exit(fmt.Sprintf("Failed to write changeset: %v", err), 1)
	}

	fmt.Printf("Created changeset: %s\n", cs.ID)
	fmt.Printf("  File: %s\n", cs.FilePath)

	// Open in editor if requested
	if openEditor && cs.FilePath != "" {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		editorCmd := exec.Command(editor, cs.FilePath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to open editor: %v\n", err)
		}
	}

	return nil
}

func createChangesetInteractive(dir string) (*changeset.Changeset, error) {
	// Discover modules
	graph, err := module.Discover(dir)
	if err != nil {
		return nil, fmt.Errorf("discovering modules: %w", err)
	}

	// Get list of modules
	modules := graph.AllModules()
	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules found")
	}

	// Build options for multi-select
	var options []huh.Option[string]
	for _, mod := range modules {
		label := mod.ShortName
		if label == "" {
			label = "(root) " + mod.Name
		}
		options = append(options, huh.NewOption(label, mod.ShortName))
	}

	// Sort options by label
	sort.Slice(options, func(i, j int) bool {
		return options[i].Key < options[j].Key
	})

	// Select modules
	var selectedModules []string
	selectForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Which packages have changes?").
				Description("Select all packages that should be released").
				Options(options...).
				Value(&selectedModules),
		),
	)

	if err := selectForm.Run(); err != nil {
		return nil, fmt.Errorf("module selection: %w", err)
	}

	if len(selectedModules) == 0 {
		return nil, fmt.Errorf("no modules selected")
	}

	// For each selected module, ask for bump type
	moduleBumps := make(map[string]changeset.Bump)
	for _, mod := range selectedModules {
		displayName := mod
		if mod == "" {
			displayName = "(root)"
		}

		var bumpType string
		bumpForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("What type of change is this for %s?", displayName)).
					Options(
						huh.NewOption("patch - Bug fixes, minor changes", "patch"),
						huh.NewOption("minor - New features, backwards compatible", "minor"),
						huh.NewOption("major - Breaking changes", "major"),
					).
					Value(&bumpType),
			),
		)

		if err := bumpForm.Run(); err != nil {
			return nil, fmt.Errorf("bump selection for %s: %w", displayName, err)
		}

		moduleBumps[mod] = changeset.Bump(bumpType)
	}

	// Get summary
	var summary string
	summaryForm := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Describe the changes").
				Description("This will appear in the changelog").
				CharLimit(1000).
				Value(&summary),
		),
	)

	if err := summaryForm.Run(); err != nil {
		return nil, fmt.Errorf("summary input: %w", err)
	}

	return &changeset.Changeset{
		Modules: moduleBumps,
		Summary: summary,
	}, nil
}
