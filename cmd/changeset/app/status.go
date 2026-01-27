package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/changeset"
	"github.com/gojekfarm/xtools/cmd/changeset/git"
	"github.com/gojekfarm/xtools/cmd/changeset/module"
)

// Status shows pending changesets and computed version bumps.
func Status(ctx context.Context, cmd *cli.Command) error {
	dir := "."
	verbose := cmd.Bool("verbose")
	outputFile := cmd.String("output")
	sinceRef := cmd.String("since")

	// Check if initialized
	if !changeset.ChangesetDirExists(dir) {
		return cli.Exit("Changeset not initialized. Run 'changeset init' first.", 1)
	}

	// Read config
	cfg, err := changeset.ReadConfig(dir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read config: %v", err), 1)
	}

	// Read changesets
	changesets, err := changeset.ReadChangesets(dir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to read changesets: %v", err), 1)
	}

	// If --since is provided, check for missing changesets
	if sinceRef != "" {
		return statusSince(dir, sinceRef, changesets, cfg, outputFile)
	}

	if len(changesets) == 0 {
		fmt.Println("No pending changesets.")
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
	releases = changeset.FilterIgnored(releases, cfg.Ignore)

	// Output JSON if requested
	if outputFile != "" {
		output := struct {
			Changesets []*changeset.Changeset `json:"changesets"`
			Releases   []changeset.Release    `json:"releases"`
		}{
			Changesets: changesets,
			Releases:   releases,
		}

		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to marshal JSON: %v", err), 1)
		}

		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to write output file: %v", err), 1)
		}

		fmt.Printf("Output written to %s\n", outputFile)
		return nil
	}

	// Print changesets
	fmt.Printf("Found %d changeset(s):\n\n", len(changesets))
	for _, cs := range changesets {
		fmt.Printf("  %s\n", cs.ID)
		if verbose {
			for mod, bump := range cs.Modules {
				modDisplay := mod
				if mod == "" {
					modDisplay = "(root)"
				}
				fmt.Printf("    - %s: %s\n", modDisplay, bump)
			}
			if cs.Summary != "" {
				fmt.Printf("    Summary: %s\n", cs.Summary)
			}
		}
	}

	// Print releases
	if len(releases) > 0 {
		fmt.Printf("\nPlanned releases:\n\n")
		for _, r := range releases {
			modDisplay := r.Module
			if r.Module == "" {
				modDisplay = "(root)"
			}
			reason := ""
			if r.Reason != "" {
				reason = fmt.Sprintf(" (%s)", r.Reason)
			}
			fmt.Printf("  %s: %s -> %s (%s)%s\n",
				modDisplay, r.PreviousVersion, r.Version, r.Bump, reason)
		}
	} else {
		fmt.Println("\nNo releases planned.")
	}

	return nil
}

// statusSince checks if modules with changes since a ref have corresponding changesets.
// This is used in CI to ensure PRs include changesets for changed code.
func statusSince(dir, sinceRef string, changesets []*changeset.Changeset, cfg *changeset.Config, outputFile string) error {
	// Discover modules
	graph, err := module.Discover(dir)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to discover modules: %v", err), 1)
	}

	// Get changed files since ref
	changedFiles, err := git.GetChangedFiles(dir, sinceRef)
	if err != nil {
		return cli.Exit(fmt.Sprintf("Failed to get changed files: %v", err), 1)
	}

	// Map files to modules
	changedModules := mapFilesToModules(changedFiles, graph, cfg)

	// Get modules covered by changesets
	coveredModules := make(map[string]bool)
	for _, cs := range changesets {
		for mod := range cs.Modules {
			coveredModules[mod] = true
		}
	}

	// Find modules with changes but no changeset
	var missingChangesets []string
	for mod := range changedModules {
		if !coveredModules[mod] {
			missingChangesets = append(missingChangesets, mod)
		}
	}

	// Output JSON if requested
	if outputFile != "" {
		output := struct {
			ChangedModules    []string `json:"changedModules"`
			CoveredModules    []string `json:"coveredModules"`
			MissingChangesets []string `json:"missingChangesets"`
		}{
			ChangedModules:    mapKeys(changedModules),
			CoveredModules:    mapKeys(coveredModules),
			MissingChangesets: missingChangesets,
		}

		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return cli.Exit(fmt.Sprintf("Failed to marshal JSON: %v", err), 1)
		}

		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to write output file: %v", err), 1)
		}

		fmt.Printf("Output written to %s\n", outputFile)
		if len(missingChangesets) > 0 {
			return cli.Exit("Missing changesets for changed modules", 1)
		}
		return nil
	}

	// Print results
	fmt.Printf("Changes since %s:\n\n", sinceRef)

	if len(changedModules) == 0 {
		fmt.Println("  No module changes detected.")
		return nil
	}

	fmt.Printf("  Changed modules: %d\n", len(changedModules))
	for mod := range changedModules {
		modDisplay := mod
		if mod == "" {
			modDisplay = "(root)"
		}
		status := "✓ has changeset"
		if !coveredModules[mod] {
			status = "✗ missing changeset"
		}
		fmt.Printf("    %s: %s\n", modDisplay, status)
	}

	if len(missingChangesets) > 0 {
		fmt.Printf("\nError: %d module(s) have changes but no changeset.\n", len(missingChangesets))
		fmt.Println("Run 'changeset add' to create a changeset for your changes.")
		return cli.Exit("Missing changesets", 1)
	}

	fmt.Println("\nAll changed modules have changesets.")
	return nil
}

// mapFilesToModules determines which modules are affected by the changed files.
func mapFilesToModules(files []string, graph *module.Graph, cfg *changeset.Config) map[string]bool {
	modules := make(map[string]bool)

	for _, file := range files {
		// Skip ignored paths
		if shouldIgnorePath(file, cfg.IgnorePaths) {
			continue
		}

		// Find which module this file belongs to
		mod := findModuleForFile(file, graph)
		if mod != nil {
			modules[mod.ShortName] = true
		}
	}

	// Remove ignored modules
	for _, ignored := range cfg.Ignore {
		delete(modules, ignored)
	}

	return modules
}

// findModuleForFile finds the module that contains the given file path.
func findModuleForFile(file string, graph *module.Graph) *module.Module {
	// Find the most specific module (longest path match)
	var bestMatch *module.Module
	bestMatchLen := -1

	for _, mod := range graph.AllModules() {
		// Get relative path from repo root to module
		modRelPath := mod.ShortName
		if modRelPath == "" {
			// Root module - matches everything not in a submodule
			if bestMatch == nil {
				bestMatch = mod
				bestMatchLen = 0
			}
			continue
		}

		// Check if file is under this module's path
		if strings.HasPrefix(file, modRelPath+"/") || file == modRelPath {
			if len(modRelPath) > bestMatchLen {
				bestMatch = mod
				bestMatchLen = len(modRelPath)
			}
		}
	}

	return bestMatch
}

// shouldIgnorePath checks if a file path matches any ignore patterns.
func shouldIgnorePath(file string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, file)
		if err == nil && matched {
			return true
		}
		// Also check if any path component matches
		matched, err = filepath.Match(pattern, filepath.Base(file))
		if err == nil && matched {
			return true
		}
	}
	return false
}

// mapKeys returns the keys of a map as a sorted slice.
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
