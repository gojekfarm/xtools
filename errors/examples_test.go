package errors_test

import (
	"fmt"

	"github.com/gojekfarm/xtools/errors"
)

func ExampleWrap() {
	err := errors.New("test error")

	// Wrap the error with some key-value pairs
	wrapped := errors.Wrap(
		err,
		"foo", "bar",
		"baz", "qux",
	)

	// errors.Is will check for the original error
	fmt.Println(errors.Is(wrapped, err))

	// Use errors.As to read attached metadata
	var errTags *errors.ErrorTags

	errors.As(wrapped, &errTags)

	fmt.Println(errTags.All())
}
