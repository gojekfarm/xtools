package errors_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/errors"
)

func TestWrap(t *testing.T) {
	err := errors.New("test error")

	wrapped := errors.Wrap(err, "foo", "bar")

	assert.Equal(t, "test error [foo=bar]", wrapped.Error())
	assert.ErrorIs(t, wrapped, err)
	assert.EqualError(t, errors.Unwrap(wrapped), err.Error())

	var data *errors.ErrorData
	assert.True(t, errors.As(wrapped, &data))
	assert.EqualValues(t, map[string]string{"foo": "bar"}, data.Data())
}

func TestWrap_PanicsOnOddNumberOfAttrs(t *testing.T) {
	assert.Panics(t, func() {
		errors.Wrap(errors.New("test error"), "foo")
	})
}

func TestWrap_ReturnsNilIfErrIsNil(t *testing.T) {
	assert.Nil(t, errors.Wrap(nil, "foo", "bar"))
}

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
	var errData *errors.ErrorData

	errors.As(wrapped, &errData)

	fmt.Println(errData.Data())
}
