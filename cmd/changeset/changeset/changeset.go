// Package changeset provides core types and operations for managing changesets.
package changeset

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Bump represents a semantic version bump type.
type Bump string

const (
	BumpPatch Bump = "patch"
	BumpMinor Bump = "minor"
	BumpMajor Bump = "major"
)

// Compare returns:
//
//	-1 if a < b
//	 0 if a == b
//	 1 if a > b
//
// Ordering: patch < minor < major
func (a Bump) Compare(b Bump) int {
	order := map[Bump]int{
		BumpPatch: 0,
		BumpMinor: 1,
		BumpMajor: 2,
	}
	aVal, bVal := order[a], order[b]
	if aVal < bVal {
		return -1
	}
	if aVal > bVal {
		return 1
	}
	return 0
}

// String returns the string representation of the bump.
func (a Bump) String() string {
	return string(a)
}

// Changeset represents a single changeset file.
type Changeset struct {
	ID       string          // Filename without extension (e.g., "hungry-tiger-jump")
	Modules  map[string]Bump // Module short name -> bump type
	Summary  string          // Markdown description
	FilePath string          // Full path to the file
}

// ParseChangeset reads and parses a single changeset file.
// Returns error if file format is invalid.
func ParseChangeset(path string) (*Changeset, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening changeset: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// First line must be ---
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return nil, fmt.Errorf("%w: missing opening ---", ErrInvalidChangeset)
	}

	// Read YAML frontmatter until closing ---
	var yamlLines []string
	foundClose := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundClose = true
			break
		}
		yamlLines = append(yamlLines, line)
	}

	if !foundClose {
		return nil, fmt.Errorf("%w: missing closing ---", ErrInvalidChangeset)
	}

	// Parse YAML
	yamlContent := strings.Join(yamlLines, "\n")
	modules := make(map[string]Bump)
	if yamlContent != "" {
		var rawModules map[string]string
		if err := yaml.Unmarshal([]byte(yamlContent), &rawModules); err != nil {
			return nil, fmt.Errorf("%w: invalid YAML: %v", ErrInvalidChangeset, err)
		}

		for mod, bump := range rawModules {
			switch Bump(bump) {
			case BumpPatch, BumpMinor, BumpMajor:
				modules[mod] = Bump(bump)
			default:
				return nil, fmt.Errorf("%w: invalid bump type %q for module %q", ErrInvalidChangeset, bump, mod)
			}
		}
	}

	// Read summary (everything after the closing ---)
	var summaryLines []string
	for scanner.Scan() {
		summaryLines = append(summaryLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading changeset: %w", err)
	}

	summary := strings.TrimSpace(strings.Join(summaryLines, "\n"))

	// Extract ID from filename
	id := strings.TrimSuffix(filepath.Base(path), ".md")

	return &Changeset{
		ID:       id,
		Modules:  modules,
		Summary:  summary,
		FilePath: path,
	}, nil
}

// ReadChangesets reads all changeset files from the .changeset directory.
// Skips config.json, README.md, and release-manifest.json.
func ReadChangesets(dir string) ([]*Changeset, error) {
	changesetDir := filepath.Join(dir, ".changeset")
	entries, err := os.ReadDir(changesetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading .changeset directory: %w", err)
	}

	var changesets []*Changeset
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip non-changeset files
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		if name == "README.md" {
			continue
		}

		path := filepath.Join(changesetDir, name)
		cs, err := ParseChangeset(path)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", name, err)
		}
		changesets = append(changesets, cs)
	}

	return changesets, nil
}

// WriteChangeset writes a changeset to the .changeset directory.
// Generates a random ID if cs.ID is empty.
func WriteChangeset(dir string, cs *Changeset) error {
	changesetDir := filepath.Join(dir, ".changeset")

	// Ensure directory exists
	if err := os.MkdirAll(changesetDir, 0755); err != nil {
		return fmt.Errorf("creating .changeset directory: %w", err)
	}

	// Generate ID if needed
	if cs.ID == "" {
		cs.ID = GenerateID()
	}

	// Build content
	var content strings.Builder
	content.WriteString("---\n")

	// Write modules as YAML
	if len(cs.Modules) > 0 {
		for mod, bump := range cs.Modules {
			content.WriteString(fmt.Sprintf("%q: %s\n", mod, bump))
		}
	}

	content.WriteString("---\n\n")
	content.WriteString(cs.Summary)
	content.WriteString("\n")

	// Write file
	path := filepath.Join(changesetDir, cs.ID+".md")
	if err := os.WriteFile(path, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("writing changeset: %w", err)
	}

	cs.FilePath = path
	return nil
}

// DeleteChangeset removes a changeset file.
func DeleteChangeset(cs *Changeset) error {
	if cs.FilePath == "" {
		return fmt.Errorf("changeset has no file path")
	}
	if err := os.Remove(cs.FilePath); err != nil {
		return fmt.Errorf("deleting changeset: %w", err)
	}
	return nil
}

// Word lists for ID generation (similar to changesets package)
var (
	adjectives = []string{
		"angry", "brave", "calm", "dark", "eager", "fair", "gentle", "happy",
		"icy", "jolly", "keen", "lively", "merry", "nice", "odd", "proud",
		"quick", "rare", "shy", "tall", "unique", "vast", "warm", "young",
		"bright", "clever", "fancy", "golden", "honest", "lucky", "mighty",
		"noble", "orange", "purple", "quiet", "rapid", "silent", "tender",
		"violet", "witty", "hungry", "fuzzy", "grumpy", "sleepy", "silly",
	}
	nouns = []string{
		"ant", "bee", "cat", "dog", "elk", "fox", "goat", "hawk", "ibis",
		"jay", "kite", "lion", "mouse", "newt", "owl", "panda", "quail",
		"rabbit", "snake", "tiger", "urchin", "viper", "wolf", "yak", "zebra",
		"bear", "crane", "dove", "eagle", "frog", "gecko", "horse", "iguana",
		"koala", "lemur", "moose", "otter", "parrot", "seal", "turtle", "whale",
	}
	verbs = []string{
		"jump", "run", "walk", "fly", "swim", "dance", "sing", "play",
		"read", "write", "think", "dream", "sleep", "wake", "eat", "drink",
		"laugh", "smile", "wave", "spin", "twist", "turn", "leap", "skip",
		"hop", "bounce", "glide", "soar", "dive", "climb", "crawl", "roll",
	}
)

// GenerateID returns a random changeset ID (e.g., "hungry-tiger-jump").
func GenerateID() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	verb := verbs[rand.Intn(len(verbs))]
	return fmt.Sprintf("%s-%s-%s", adj, noun, verb)
}
