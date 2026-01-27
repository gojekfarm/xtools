package changeset

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// ChangelogEntry represents a single changelog entry.
type ChangelogEntry struct {
	Version string
	Date    time.Time
	Changes map[Bump][]string // Bump type -> list of summaries
}

// GenerateChangelog creates a changelog entry from releases and changesets.
// It groups changes by bump type for better readability.
func GenerateChangelog(releases []Release, changesets []*Changeset) *ChangelogEntry {
	if len(releases) == 0 {
		return nil
	}

	// Use the first release version as the entry version
	// (assuming all releases in a batch have the same version pattern)
	version := releases[0].Version

	entry := &ChangelogEntry{
		Version: version,
		Date:    time.Now(),
		Changes: make(map[Bump][]string),
	}

	// Collect all unique summaries
	summarySet := make(map[string]bool)
	for _, cs := range changesets {
		if cs.Summary != "" && !summarySet[cs.Summary] {
			summarySet[cs.Summary] = true
			// Find the highest bump in this changeset to categorize it
			highestBump := BumpPatch
			for _, bump := range cs.Modules {
				if bump.Compare(highestBump) > 0 {
					highestBump = bump
				}
			}
			entry.Changes[highestBump] = append(entry.Changes[highestBump], cs.Summary)
		}
	}

	return entry
}

// FormatChangelogEntry formats a changelog entry as markdown.
func FormatChangelogEntry(entry *ChangelogEntry) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", entry.Version, entry.Date.Format("2006-01-02")))

	// Order: major, minor, patch
	bumpOrder := []Bump{BumpMajor, BumpMinor, BumpPatch}
	bumpHeaders := map[Bump]string{
		BumpMajor: "### Breaking Changes",
		BumpMinor: "### Features",
		BumpPatch: "### Bug Fixes",
	}

	for _, bump := range bumpOrder {
		changes := entry.Changes[bump]
		if len(changes) == 0 {
			continue
		}

		sb.WriteString(bumpHeaders[bump] + "\n\n")
		for _, change := range changes {
			// Handle multi-line summaries
			lines := strings.Split(change, "\n")
			for i, line := range lines {
				if i == 0 {
					sb.WriteString("- " + line + "\n")
				} else if strings.TrimSpace(line) != "" {
					sb.WriteString("  " + line + "\n")
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// UpdateChangelog prepends a new entry to an existing CHANGELOG.md.
// Creates the file if it doesn't exist.
func UpdateChangelog(path string, entry *ChangelogEntry) error {
	newContent := FormatChangelogEntry(entry)

	// Read existing content
	existingContent, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading changelog: %w", err)
	}

	var finalContent string
	if len(existingContent) == 0 {
		// New file - add header
		finalContent = "# Changelog\n\nAll notable changes to this project will be documented in this file.\n\n" + newContent
	} else {
		// Prepend to existing content after the header
		existing := string(existingContent)
		// Find where to insert (after the title and intro)
		insertPos := 0
		lines := strings.Split(existing, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "## ") {
				// Found first version entry
				insertPos = strings.Index(existing, line)
				break
			}
			if i == len(lines)-1 {
				// No version entries found, append to end
				insertPos = len(existing)
			}
		}

		if insertPos > 0 {
			finalContent = existing[:insertPos] + newContent + existing[insertPos:]
		} else {
			finalContent = existing + "\n" + newContent
		}
	}

	if err := os.WriteFile(path, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("writing changelog: %w", err)
	}

	return nil
}

// ParseChangelog reads an existing CHANGELOG.md.
// Returns entries sorted by version (newest first).
func ParseChangelog(path string) ([]ChangelogEntry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading changelog: %w", err)
	}

	var entries []ChangelogEntry
	var currentEntry *ChangelogEntry
	var currentBump Bump
	var currentChanges []string

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			// Save previous entry
			if currentEntry != nil {
				if len(currentChanges) > 0 && currentBump != "" {
					currentEntry.Changes[currentBump] = currentChanges
				}
				entries = append(entries, *currentEntry)
			}

			// Parse version header: "## v1.0.0 (2024-01-01)"
			header := strings.TrimPrefix(line, "## ")
			parts := strings.SplitN(header, " ", 2)
			version := parts[0]

			var date time.Time
			if len(parts) > 1 {
				dateStr := strings.Trim(parts[1], "()")
				date, _ = time.Parse("2006-01-02", dateStr)
			}

			currentEntry = &ChangelogEntry{
				Version: version,
				Date:    date,
				Changes: make(map[Bump][]string),
			}
			currentBump = ""
			currentChanges = nil
		} else if strings.HasPrefix(line, "### ") {
			// Save changes for previous bump type
			if currentEntry != nil && len(currentChanges) > 0 && currentBump != "" {
				currentEntry.Changes[currentBump] = currentChanges
				currentChanges = nil
			}

			// Parse bump type header
			header := strings.TrimPrefix(line, "### ")
			switch header {
			case "Breaking Changes":
				currentBump = BumpMajor
			case "Features":
				currentBump = BumpMinor
			case "Bug Fixes":
				currentBump = BumpPatch
			default:
				currentBump = ""
			}
		} else if strings.HasPrefix(line, "- ") && currentEntry != nil && currentBump != "" {
			currentChanges = append(currentChanges, strings.TrimPrefix(line, "- "))
		}
	}

	// Save last entry
	if currentEntry != nil {
		if len(currentChanges) > 0 && currentBump != "" {
			currentEntry.Changes[currentBump] = currentChanges
		}
		entries = append(entries, *currentEntry)
	}

	// Sort by version (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Version > entries[j].Version
	})

	return entries, nil
}
