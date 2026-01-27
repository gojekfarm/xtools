package git

import (
	"fmt"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/mod/semver"
)

// GetTags returns all version tags in the repository.
func GetTags(dir string) ([]Tag, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("opening repository: %w", err)
	}

	refs, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}

	var tags []Tag
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		module, version, ok := ParseTag(name)
		if !ok {
			// Skip non-version tags
			return nil
		}
		tags = append(tags, Tag{
			Name:    name,
			Module:  module,
			Version: version,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("iterating tags: %w", err)
	}

	return tags, nil
}

// GetLatestVersions returns the latest version for each module.
// Returns a map of module short name -> version (e.g., "xkafka" -> "v0.10.0").
// Root module uses empty string as key.
func GetLatestVersions(dir string) (map[string]string, error) {
	tags, err := GetTags(dir)
	if err != nil {
		return nil, err
	}

	// Group versions by module
	moduleVersions := make(map[string][]string)
	for _, tag := range tags {
		moduleVersions[tag.Module] = append(moduleVersions[tag.Module], tag.Version)
	}

	// Find the latest version for each module
	latest := make(map[string]string)
	for module, versions := range moduleVersions {
		semver.Sort(versions)
		latest[module] = versions[len(versions)-1]
	}

	return latest, nil
}

// CreateTag creates a git tag for a module version.
// For submodules, uses format "module/version".
// For root, uses format "version".
func CreateTag(dir string, module string, version string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("opening repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("getting HEAD: %w", err)
	}

	tagName := FormatTag(module, version)
	_, err = repo.CreateTag(tagName, head.Hash(), nil)
	if err != nil {
		return fmt.Errorf("creating tag %s: %w", tagName, err)
	}

	return nil
}

// CreateTags creates multiple git tags atomically.
func CreateTags(dir string, tags []Tag) error {
	for _, tag := range tags {
		if err := CreateTag(dir, tag.Module, tag.Version); err != nil {
			return err
		}
	}
	return nil
}

// PushTags pushes tags to the remote.
func PushTags(dir string, tags []Tag, remote string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("opening repository: %w", err)
	}

	// Build refspecs for the tags
	refSpecs := make([]config.RefSpec, len(tags))
	for i, tag := range tags {
		refSpecs[i] = config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag.Name, tag.Name))
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: remote,
		RefSpecs:   refSpecs,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("pushing tags: %w", err)
	}

	return nil
}

// HasUncommittedChanges returns true if there are uncommitted changes.
func HasUncommittedChanges(dir string) (bool, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return false, fmt.Errorf("opening repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("getting worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return false, fmt.Errorf("getting status: %w", err)
	}

	return !status.IsClean(), nil
}

// GetAllTags returns all tags grouped by module.
// Returns a map of module short name -> list of versions.
func GetAllTags(dir string) (map[string][]string, error) {
	tags, err := GetTags(dir)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, tag := range tags {
		result[tag.Module] = append(result[tag.Module], tag.Version)
	}

	// Sort versions for each module
	for module := range result {
		sort.Slice(result[module], func(i, j int) bool {
			return semver.Compare(result[module][i], result[module][j]) < 0
		})
	}

	return result, nil
}

// GetChangedFiles returns files changed since a given ref (branch/tag).
func GetChangedFiles(dir string, sinceRef string) ([]string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("opening repository: %w", err)
	}

	// Resolve the reference
	refHash, err := repo.ResolveRevision(plumbing.Revision(sinceRef))
	if err != nil {
		return nil, fmt.Errorf("resolving ref %s: %w", sinceRef, err)
	}

	refCommit, err := repo.CommitObject(*refHash)
	if err != nil {
		return nil, fmt.Errorf("getting commit for ref: %w", err)
	}

	// Get HEAD
	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("getting HEAD: %w", err)
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("getting HEAD commit: %w", err)
	}

	// Get trees for comparison
	refTree, err := refCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("getting ref tree: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("getting HEAD tree: %w", err)
	}

	// Get diff
	changes, err := refTree.Diff(headTree)
	if err != nil {
		return nil, fmt.Errorf("computing diff: %w", err)
	}

	// Collect changed file paths
	fileSet := make(map[string]bool)
	for _, change := range changes {
		if change.From.Name != "" {
			fileSet[change.From.Name] = true
		}
		if change.To.Name != "" {
			fileSet[change.To.Name] = true
		}
	}

	var files []string
	for f := range fileSet {
		files = append(files, f)
	}
	sort.Strings(files)

	return files, nil
}
