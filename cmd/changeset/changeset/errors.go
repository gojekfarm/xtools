package changeset

import "errors"

// Sentinel errors for changeset operations.
var (
	ErrNoChangesets     = errors.New("no changesets found")
	ErrNoManifest       = errors.New("release manifest not found")
	ErrInvalidChangeset = errors.New("invalid changeset format")
	ErrInvalidVersion   = errors.New("invalid semantic version")
	ErrUncommitted      = errors.New("uncommitted changes exist")
)
