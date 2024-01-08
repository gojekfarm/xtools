package xload

import (
	"fmt"
	"reflect"
)

// ErrRequired is returned when a required key is missing.
type ErrRequired struct{ key string }

func (e *ErrRequired) Error() string { return "required key missing: " + e.key }

// ErrUnknownTagOption is returned when an unknown tag option is used.
type ErrUnknownTagOption struct {
	key string
	opt string
}

func (e *ErrUnknownTagOption) Error() string {
	if e.key == "" {
		return fmt.Sprintf("unknown tag option: %s", e.opt)
	}

	return fmt.Sprintf("`%s` key has unknown tag option: %s", e.key, e.opt)
}

// ErrUnknownFieldType is returned when the key type is not supported.
type ErrUnknownFieldType struct {
	field string
	kind  reflect.Kind
	key   string
}

func (e *ErrUnknownFieldType) Error() string {
	return fmt.Sprintf("`%s: %s` key=%s has an invalid value", e.field, e.kind, e.key)
}

// ErrInvalidMapValue is returned when the map value is invalid.
type ErrInvalidMapValue struct{ key string }

func (e *ErrInvalidMapValue) Error() string {
	return fmt.Sprintf("`%s` key has an invalid map value", e.key)
}

// ErrInvalidPrefix is returned when the prefix option is used on a non-struct key.
type ErrInvalidPrefix struct {
	field string
	kind  reflect.Kind
}

func (e *ErrInvalidPrefix) Error() string {
	return fmt.Sprintf("prefix is only valid on struct types, found `%s: %s`", e.field, e.kind)
}

// ErrInvalidPrefixAndKey is returned when the prefix option is used with a key.
type ErrInvalidPrefixAndKey struct {
	field string
	key   string
}

func (e *ErrInvalidPrefixAndKey) Error() string {
	return fmt.Sprintf("`%s` key=%s has both prefix and key", e.field, e.key)
}
