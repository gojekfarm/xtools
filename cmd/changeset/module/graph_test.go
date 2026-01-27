package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestGraph creates a test graph with the structure:
//
//	root (no deps)
//	├── a (no deps)
//	├── b (depends on a)
//	└── c (depends on a, b)
func newTestGraph() *Graph {
	root := &Module{
		Name:      "github.com/foo/bar",
		ShortName: "",
		Path:      "/test/root",
	}
	a := &Module{
		Name:         "github.com/foo/bar/a",
		ShortName:    "a",
		Path:         "/test/root/a",
		Dependencies: []string{},
	}
	b := &Module{
		Name:         "github.com/foo/bar/b",
		ShortName:    "b",
		Path:         "/test/root/b",
		Dependencies: []string{"a"},
	}
	c := &Module{
		Name:         "github.com/foo/bar/c",
		ShortName:    "c",
		Path:         "/test/root/c",
		Dependencies: []string{"a", "b"},
	}

	return &Graph{
		Root: root,
		Modules: map[string]*Module{
			"":  root,
			"a": a,
			"b": b,
			"c": c,
		},
	}
}

func TestFindModule(t *testing.T) {
	g := newTestGraph()

	tests := []struct {
		name      string
		shortName string
		wantNil   bool
		wantName  string
	}{
		{"find root", "", false, "github.com/foo/bar"},
		{"find a", "a", false, "github.com/foo/bar/a"},
		{"find b", "b", false, "github.com/foo/bar/b"},
		{"find c", "c", false, "github.com/foo/bar/c"},
		{"find nonexistent", "nonexistent", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := g.FindModule(tt.shortName)
			if tt.wantNil {
				assert.Nil(t, mod)
			} else {
				require.NotNil(t, mod)
				assert.Equal(t, tt.wantName, mod.Name)
			}
		})
	}
}

func TestDependents(t *testing.T) {
	g := newTestGraph()

	tests := []struct {
		name      string
		shortName string
		wantNames []string
	}{
		{"dependents of a", "a", []string{"b", "c"}},
		{"dependents of b", "b", []string{"c"}},
		{"dependents of c", "c", nil},
		{"dependents of root", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := g.Dependents(tt.shortName)
			var gotNames []string
			for _, d := range deps {
				gotNames = append(gotNames, d.ShortName)
			}
			if tt.wantNames == nil {
				assert.Empty(t, gotNames)
			} else {
				assert.Equal(t, tt.wantNames, gotNames)
			}
		})
	}
}

func TestTopologicalSort(t *testing.T) {
	g := newTestGraph()

	sorted := g.TopologicalSort()

	// Convert to short names for easier checking
	var names []string
	for _, m := range sorted {
		names = append(names, m.ShortName)
	}

	// Verify all modules are present
	assert.Len(t, names, 4)

	// Find indices
	indexOf := func(name string) int {
		for i, n := range names {
			if n == name {
				return i
			}
		}
		return -1
	}

	// c must come after a and b
	assert.Greater(t, indexOf("c"), indexOf("a"), "c should come after a")
	assert.Greater(t, indexOf("c"), indexOf("b"), "c should come after b")

	// b must come after a
	assert.Greater(t, indexOf("b"), indexOf("a"), "b should come after a")
}

func TestIsInternal(t *testing.T) {
	g := newTestGraph()

	tests := []struct {
		name       string
		modulePath string
		want       bool
	}{
		{"root module", "github.com/foo/bar", true},
		{"submodule a", "github.com/foo/bar/a", true},
		{"nested path", "github.com/foo/bar/deep/nested", true},
		{"external module", "github.com/other/pkg", false},
		{"similar prefix", "github.com/foo/barbaz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.IsInternal(tt.modulePath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllModules(t *testing.T) {
	g := newTestGraph()

	mods := g.AllModules()

	assert.Len(t, mods, 4)

	// Should be sorted by short name (empty string first)
	expectedOrder := []string{"", "a", "b", "c"}
	for i, mod := range mods {
		assert.Equal(t, expectedOrder[i], mod.ShortName, "module at index %d", i)
	}
}
