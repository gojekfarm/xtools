package xtools

import (
	"regexp"
	"testing"
)

func TestVersionString(t *testing.T) {
	semver := regexp.MustCompile(`^(?P<major>\d+\.)?(?P<minor>\d+\.)?(?P<patch>\*|\d+)(?:-(?P<mod>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?$`)
	if !semver.MatchString(Version()) {
		t.Fatalf("Version() did not return a valid semantic versioned string")
	}
}
