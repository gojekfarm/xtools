package errors_test

import (
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

	var tags *errors.ErrorTags
	assert.True(t, errors.As(wrapped, &tags))
	assert.EqualValues(t, map[string]string{"foo": "bar"}, tags.All())
}

func TestWrap_MultipleWraps(t *testing.T) {
	err := errors.New("test error")

	wrapped1 := errors.Wrap(err, "foo", "bar")
	wrapped2 := errors.Wrap(wrapped1, "baz", "qux")

	assert.Equal(t, "test error [baz=qux, foo=bar]", wrapped2.Error())
}

func TestWrap_PanicsOnOddNumberOfAttrs(t *testing.T) {
	assert.Panics(t, func() {
		errors.Wrap(errors.New("test error"), "foo")
	})
}

func TestWrap_ReturnsNilIfErrIsNil(t *testing.T) {
	assert.Nil(t, errors.Wrap(nil, "foo", "bar"))
}
