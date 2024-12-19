package errors

import (
	"errors"
	"strings"
)

// Aliases to the standard error types
var (
	New    = errors.New
	Is     = errors.Is
	As     = errors.As
	Unwrap = errors.Unwrap
	Join   = errors.Join
)

// ErrorTags is an error that has key-value tags attached to it.
type ErrorTags struct {
	err  error
	tags map[string]string
}

// Error returns the error message with the tags attached.
func (e *ErrorTags) Error() string {
	md := []string{}

	for k, v := range e.tags {
		md = append(md, k+"="+v)
	}

	return e.err.Error() + " [" + strings.Join(md, ", ") + "]"
}

// Unwrap returns the underlying error.
func (e *ErrorTags) Unwrap() error {
	return e.err
}

// Is returns true if the error is the same as the target error.
func (e *ErrorTags) Is(target error) bool {
	return Is(e.err, target)
}

// All returns all the tags attached to the error.
func (e *ErrorTags) All() map[string]string {
	return e.tags
}

// Wrap attaches additional tags to an error.
func Wrap(err error, attrs ...string) error {
	if err == nil {
		return nil
	}

	if len(attrs)%2 != 0 {
		panic("[xtools/errors] attrs must be key/value pairs")
	}

	var e *ErrorTags

	if !As(err, &e) {
		e = &ErrorTags{
			err:  err,
			tags: map[string]string{},
		}
	}

	for i := 0; i < len(attrs); i += 2 {
		e.tags[attrs[i]] = attrs[i+1]
	}

	return e
}
