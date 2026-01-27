package module

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	all := append([]string{dir, "..", "testdata"}, parts...)
	return filepath.Join(all...)
}

func TestDiscover(t *testing.T) {
	repoPath := testdataPath("fakerepo")

	graph, err := Discover(repoPath)
	require.NoError(t, err)
	require.NotNil(t, graph.Root)

	// Check root module
	assert.Equal(t, "github.com/test/fakerepo", graph.Root.Name)
	assert.Empty(t, graph.Root.ShortName)

	// Check we found all modules
	assert.Len(t, graph.Modules, 4)
	assert.Contains(t, graph.Modules, "")
	assert.Contains(t, graph.Modules, "libA")
	assert.Contains(t, graph.Modules, "libB")
	assert.Contains(t, graph.Modules, "libC")

	// Check libB dependencies
	libB := graph.FindModule("libB")
	require.NotNil(t, libB)
	assert.Contains(t, libB.Dependencies, "libA")

	// Check libC dependencies
	libC := graph.FindModule("libC")
	require.NotNil(t, libC)
	assert.Contains(t, libC.Dependencies, "libA")
	assert.Contains(t, libC.Dependencies, "libB")
}

func TestDiscoverNonExistent(t *testing.T) {
	_, err := Discover("/nonexistent/path")
	assert.Error(t, err)
}
