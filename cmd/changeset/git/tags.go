// Package git provides git operations for tags and repository state.
package git

import (
	"regexp"
	"strings"
)

// Tag represents a parsed git tag.
type Tag struct {
	Name    string // Full tag name (e.g., "xkafka/v0.10.0")
	Module  string // Module short name (e.g., "xkafka", or "" for root)
	Version string // Semantic version (e.g., "v0.10.0")
}

// semverRegex matches semantic versions like v0.10.0, v1.2.3-beta.1, etc.
var semverRegex = regexp.MustCompile(`^v\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)

// ParseTag parses a git tag name into module and version.
// Examples:
//
//	"v0.10.0"                    -> ("", "v0.10.0", true)
//	"xkafka/v0.10.0"             -> ("xkafka", "v0.10.0", true)
//	"xkafka/middleware/v0.10.0"  -> ("xkafka/middleware", "v0.10.0", true)
//	"not-a-version"              -> ("", "", false)
func ParseTag(tagName string) (module string, version string, ok bool) {
	// Handle root module tags (just version)
	if semverRegex.MatchString(tagName) {
		return "", tagName, true
	}

	// Find the last path segment that looks like a version
	parts := strings.Split(tagName, "/")
	if len(parts) < 2 {
		return "", "", false
	}

	lastPart := parts[len(parts)-1]
	if !semverRegex.MatchString(lastPart) {
		return "", "", false
	}

	modulePath := strings.Join(parts[:len(parts)-1], "/")
	return modulePath, lastPart, true
}

// FormatTag creates a tag name from module and version.
// For root module (empty module name), returns just the version.
// For submodules, returns "module/version".
func FormatTag(module string, version string) string {
	if module == "" {
		return version
	}
	return module + "/" + version
}
