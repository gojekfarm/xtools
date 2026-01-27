package changeset

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"

	"github.com/gojekfarm/xtools/cmd/changeset/module"
)

// ComputeReleases calculates version bumps from changesets.
//
// Algorithm:
//  1. Aggregate bumps per module (highest bump wins)
//  2. Get current versions from git tags
//  3. Cascade bumps to dependent modules
//  4. Compute next versions
//
// Parameters:
//   - changesets: All pending changesets
//   - graph: Module dependency graph
//   - tags: Current git tags (module short name -> version)
//   - cfg: Config for dependent bump behavior
//
// Returns releases sorted by module name.
func ComputeReleases(
	changesets []*Changeset,
	graph *module.Graph,
	tags map[string]string,
	cfg *Config,
) ([]Release, error) {
	// Step 1: Aggregate explicit bumps (highest bump wins)
	bumps := make(map[string]Bump)
	for _, cs := range changesets {
		for mod, bump := range cs.Modules {
			existing, ok := bumps[mod]
			if !ok || bump.Compare(existing) > 0 {
				bumps[mod] = bump
			}
		}
	}

	// Track reasons for bumps
	reasons := make(map[string]string)

	// Step 2: Topological sort and cascade to dependents
	sorted := graph.TopologicalSort()

	for _, mod := range sorted {
		bump, hasBump := bumps[mod.ShortName]
		if !hasBump {
			continue
		}

		// Cascade to modules that depend on this one
		for _, dependent := range graph.Dependents(mod.ShortName) {
			if _, alreadyBumped := bumps[dependent.ShortName]; !alreadyBumped {
				bumps[dependent.ShortName] = cfg.DependentBump
				reasons[dependent.ShortName] = "dependency"
			}
		}

		// Also mark this module as explicitly bumped if needed
		if _, hasReason := reasons[mod.ShortName]; !hasReason && bump != "" {
			// Explicit bump, no special reason
		}
	}

	// Step 3: Compute versions
	var releases []Release
	for shortName, bump := range bumps {
		// Skip if module not in graph (might be ignored or invalid)
		if graph.FindModule(shortName) == nil {
			continue
		}

		// Get current version from tags
		currentVersion := tags[shortName]
		if currentVersion == "" {
			currentVersion = "v0.0.0"
		}

		// Compute next version
		nextVersion, err := IncrementVersion(currentVersion, bump)
		if err != nil {
			return nil, fmt.Errorf("computing version for %s: %w", shortName, err)
		}

		releases = append(releases, Release{
			Module:          shortName,
			Version:         nextVersion,
			PreviousVersion: currentVersion,
			Bump:            bump,
			Reason:          reasons[shortName],
		})
	}

	// Sort by module name
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Module < releases[j].Module
	})

	return releases, nil
}

// IncrementVersion bumps a semantic version by the given bump type.
func IncrementVersion(current string, bump Bump) (string, error) {
	v, err := semver.NewVersion(current)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidVersion, err)
	}

	var next semver.Version
	switch bump {
	case BumpMajor:
		next = v.IncMajor()
	case BumpMinor:
		next = v.IncMinor()
	case BumpPatch:
		next = v.IncPatch()
	default:
		return "", fmt.Errorf("%w: unknown bump type %q", ErrInvalidVersion, bump)
	}

	return "v" + next.String(), nil
}

// FilterIgnored removes ignored modules from the releases list.
func FilterIgnored(releases []Release, ignore []string) []Release {
	ignoreSet := make(map[string]bool)
	for _, mod := range ignore {
		ignoreSet[mod] = true
	}

	var filtered []Release
	for _, r := range releases {
		if !ignoreSet[r.Module] {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
