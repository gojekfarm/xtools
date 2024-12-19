package errors_test

import (
	"fmt"

	"github.com/gojekfarm/xtools/errors"
)

func ExampleWrap() {
	// Create a generic error
	err := errors.New("record not found")

	// Wrap the error with key-value pairs
	wrapped := errors.Wrap(
		err,
		"table", "users",
		"id", "123",
	)

	// Add more tags as the error propagates
	wrapped = errors.Wrap(
		wrapped,
		"experiment_id", "456",
	)

	// errors.Is will check for not found error
	fmt.Println(errors.Is(wrapped, err))

	// Use errors.As to read attached tags.
	var errTags *errors.ErrorTags

	errors.As(wrapped, &errTags)

	// Use the tags to construct detailed error messages,
	// log additional context, or return structured errors.
	fmt.Println(errTags.All())
}
