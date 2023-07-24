package xconfig

import "errors"

var (
	// ErrNotPointer is returned when the given config is not a pointer.
	ErrNotPointer = errors.New("xconfig: config must be a pointer")
	// ErrNotStruct is returned when the given config is not a struct.
	ErrNotStruct = errors.New("xconfig: config must be a struct")
	// ErrUnknownTagOption is returned when an unknown tag option is used.
	ErrUnknownTagOption = errors.New("xconfig: unknown tag option")
	// ErrRequired is returned when a required field is missing.
	ErrRequired = errors.New("xconfig: missing required value")
	// ErrUnknownFieldType is returned when the field type is not supported.
	ErrUnknownFieldType = errors.New("xconfig: unknown field type")
	// ErrInvalidMapValue is returned when the map value is invalid.
	ErrInvalidMapValue = errors.New("xconfig: invalid map value")
	// ErrMissingKey is returned when the key is missing from the tag.
	ErrMissingKey = errors.New("xconfig: missing key")
)
