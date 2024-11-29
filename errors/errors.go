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

// ErrorData is an error that has key-value metadata attached to it.
type ErrorData struct {
	err  error
	data map[string]string
}

// Error returns the error message and any attached metadata.
func (e *ErrorData) Error() string {
	md := []string{}

	for k, v := range e.data {
		md = append(md, k+"="+v)
	}

	return e.err.Error() + " [" + strings.Join(md, ", ") + "]"
}

// Unwrap returns the underlying error.
func (e *ErrorData) Unwrap() error {
	return e.err
}

// Is returns true if the error is the same as the target error.
func (e *ErrorData) Is(target error) bool {
	return Is(e.err, target)
}

// Data returns the attached metadata.
func (e *ErrorData) Data() map[string]string {
	return e.data
}

// Wrap attaches additional metadata to an error.
func Wrap(err error, attrs ...string) error {
	if err == nil {
		return nil
	}

	if len(attrs)%2 != 0 {
		panic("[xtools/errors] attrs must be key/value pairs")
	}

	e := &ErrorData{
		err:  err,
		data: map[string]string{},
	}

	for i := 0; i < len(attrs); i += 2 {
		e.data[attrs[i]] = attrs[i+1]
	}

	return e
}
