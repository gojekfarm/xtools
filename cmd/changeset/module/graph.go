package module

import (
	"sort"
	"strings"
)

// FindModule returns a module by short name.
// Use empty string to get the root module.
func (g *Graph) FindModule(shortName string) *Module {
	return g.Modules[shortName]
}

// Dependents returns modules that depend on the given module.
func (g *Graph) Dependents(shortName string) []*Module {
	var dependents []*Module
	for _, mod := range g.Modules {
		for _, dep := range mod.Dependencies {
			if dep == shortName {
				dependents = append(dependents, mod)
				break
			}
		}
	}

	// Sort for deterministic output
	sort.Slice(dependents, func(i, j int) bool {
		return dependents[i].ShortName < dependents[j].ShortName
	})

	return dependents
}

// TopologicalSort returns modules in dependency order (leaves first).
// Modules with no dependencies come first, then modules that only
// depend on those, and so on. This ensures that when processing
// modules in order, all dependencies are processed before dependents.
func (g *Graph) TopologicalSort() []*Module {
	// Kahn's algorithm
	// Calculate in-degree (number of internal dependencies) for each module
	inDegree := make(map[string]int)
	for shortName := range g.Modules {
		inDegree[shortName] = 0
	}

	// Count internal dependencies
	for _, mod := range g.Modules {
		for _, dep := range mod.Dependencies {
			if _, exists := g.Modules[dep]; exists {
				inDegree[mod.ShortName]++
			}
		}
	}

	// Start with modules that have no internal dependencies
	var queue []string
	for shortName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, shortName)
		}
	}

	// Sort queue for deterministic output
	sort.Strings(queue)

	var result []*Module
	for len(queue) > 0 {
		// Pop from queue
		shortName := queue[0]
		queue = queue[1:]

		mod := g.Modules[shortName]
		result = append(result, mod)

		// Decrease in-degree for all modules that depend on this one
		for _, dependent := range g.Dependents(shortName) {
			inDegree[dependent.ShortName]--
			if inDegree[dependent.ShortName] == 0 {
				queue = append(queue, dependent.ShortName)
				// Keep queue sorted
				sort.Strings(queue)
			}
		}
	}

	return result
}

// IsInternal returns true if the module path is internal to this repo.
func (g *Graph) IsInternal(modulePath string) bool {
	if modulePath == g.Root.Name {
		return true
	}
	return strings.HasPrefix(modulePath, g.Root.Name+"/")
}

// AllModules returns all modules as a slice, sorted by short name.
func (g *Graph) AllModules() []*Module {
	var modules []*Module
	for _, mod := range g.Modules {
		modules = append(modules, mod)
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].ShortName < modules[j].ShortName
	})
	return modules
}
