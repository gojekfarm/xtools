package changeset

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatChangelogEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    *ChangelogEntry
		contains []string
		excludes []string
	}{
		{
			name: "patch only",
			entry: &ChangelogEntry{
				Version: "v1.0.0",
				Date:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Changes: map[Bump][]string{
					BumpPatch: {"Fixed a bug"},
				},
			},
			contains: []string{"## v1.0.0", "2024-01-15", "### Bug Fixes", "Fixed a bug"},
			excludes: []string{"### Features", "### Breaking Changes"},
		},
		{
			name: "minor only",
			entry: &ChangelogEntry{
				Version: "v1.1.0",
				Date:    time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
				Changes: map[Bump][]string{
					BumpMinor: {"Added new feature"},
				},
			},
			contains: []string{"## v1.1.0", "### Features", "Added new feature"},
			excludes: []string{"### Bug Fixes", "### Breaking Changes"},
		},
		{
			name: "major only",
			entry: &ChangelogEntry{
				Version: "v2.0.0",
				Date:    time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
				Changes: map[Bump][]string{
					BumpMajor: {"Breaking API change"},
				},
			},
			contains: []string{"## v2.0.0", "### Breaking Changes", "Breaking API change"},
			excludes: []string{"### Features", "### Bug Fixes"},
		},
		{
			name: "mixed changes",
			entry: &ChangelogEntry{
				Version: "v2.1.1",
				Date:    time.Date(2024, 4, 5, 0, 0, 0, 0, time.UTC),
				Changes: map[Bump][]string{
					BumpMajor: {"Breaking change"},
					BumpMinor: {"New feature"},
					BumpPatch: {"Bug fix"},
				},
			},
			contains: []string{
				"### Breaking Changes",
				"### Features",
				"### Bug Fixes",
				"Breaking change",
				"New feature",
				"Bug fix",
			},
		},
		{
			name: "multiple changes per type",
			entry: &ChangelogEntry{
				Version: "v1.0.1",
				Date:    time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
				Changes: map[Bump][]string{
					BumpPatch: {"Fix one", "Fix two", "Fix three"},
				},
			},
			contains: []string{"- Fix one", "- Fix two", "- Fix three"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatChangelogEntry(tt.entry)

			for _, want := range tt.contains {
				assert.Contains(t, got, want)
			}

			for _, exclude := range tt.excludes {
				assert.NotContains(t, got, exclude)
			}
		})
	}
}

func TestFormatChangelogEntryOrder(t *testing.T) {
	entry := &ChangelogEntry{
		Version: "v1.0.0",
		Date:    time.Now(),
		Changes: map[Bump][]string{
			BumpMajor: {"Major change"},
			BumpMinor: {"Minor change"},
			BumpPatch: {"Patch change"},
		},
	}

	got := FormatChangelogEntry(entry)

	majorIdx := strings.Index(got, "### Breaking Changes")
	minorIdx := strings.Index(got, "### Features")
	patchIdx := strings.Index(got, "### Bug Fixes")

	require.NotEqual(t, -1, majorIdx, "missing Breaking Changes header")
	require.NotEqual(t, -1, minorIdx, "missing Features header")
	require.NotEqual(t, -1, patchIdx, "missing Bug Fixes header")

	assert.Less(t, majorIdx, minorIdx, "Breaking Changes should come before Features")
	assert.Less(t, minorIdx, patchIdx, "Features should come before Bug Fixes")
}

func TestGenerateChangelog(t *testing.T) {
	releases := []Release{
		{Module: "pkg", Version: "v1.0.0", Bump: BumpMinor},
	}
	changesets := []*Changeset{
		{
			Modules: map[string]Bump{"pkg": BumpMinor},
			Summary: "Added new feature",
		},
	}

	entry := GenerateChangelog(releases, changesets)

	require.NotNil(t, entry)
	assert.Equal(t, "v1.0.0", entry.Version)
	assert.Len(t, entry.Changes[BumpMinor], 1)
}

func TestGenerateChangelogEmpty(t *testing.T) {
	entry := GenerateChangelog([]Release{}, []*Changeset{})
	assert.Nil(t, entry)
}

func TestGenerateChangelogDeduplicates(t *testing.T) {
	releases := []Release{
		{Module: "pkg", Version: "v1.0.0", Bump: BumpMinor},
	}
	changesets := []*Changeset{
		{Modules: map[string]Bump{"pkg": BumpMinor}, Summary: "Same change"},
		{Modules: map[string]Bump{"pkg": BumpMinor}, Summary: "Same change"},
	}

	entry := GenerateChangelog(releases, changesets)

	require.NotNil(t, entry)

	total := 0
	for _, changes := range entry.Changes {
		total += len(changes)
	}
	assert.Equal(t, 1, total, "should deduplicate identical summaries")
}
