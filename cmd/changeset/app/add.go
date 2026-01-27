package app

import (
	"context"
	"fmt"
	"log/slog"
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
	modules := cmd.StringSlice("module")
	summary := cmd.String("summary")

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
	} else if len(modules) > 0 {
		// Non-interactive mode
		var err error
		cs, err = createChangesetNonInteractive(modules, summary)
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to create changeset: %v", err), 1)
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

	slog.Info("Created changeset", "id", cs.ID, "file", cs.FilePath)

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
			slog.Warn("Failed to open editor", "error", err)
		}
	}

	return nil
}

func createChangesetNonInteractive(modules []string, summary string) (*changeset.Changeset, error) {
	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules specified")
	}

	moduleBumps := make(map[string]changeset.Bump)
	for _, m := range modules {
		parts := splitModuleBump(m)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid module:bump format: %q (expected 'module:bump')", m)
		}
		modName := parts[0]
		bumpType := parts[1]

		if !isValidBump(bumpType) {
			return nil, fmt.Errorf("invalid bump type %q for module %q (must be patch, minor, or major)", bumpType, modName)
		}

		moduleBumps[modName] = changeset.Bump(bumpType)
	}

	if summary == "" {
		summary = "No summary provided."
	}

	return &changeset.Changeset{
		Modules: moduleBumps,
		Summary: summary,
	}, nil
}

func splitModuleBump(s string) []string {
	idx := lastIndex(s, ':')
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+1:]}
}

func lastIndex(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func isValidBump(bump string) bool {
	return bump == "patch" || bump == "minor" || bump == "major"
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
