// Package yaml provides a YAML loader for xload.
package yaml

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileLoader(t *testing.T) {
	l, err := NewFileLoader("data_test.yml", "_")
	assert.NoError(t, err)
	assert.NotNil(t, l)

	ctx := context.Background()

	name, err := l.Load(ctx, "NAME")
	assert.NoError(t, err)
	assert.Equal(t, "Bombay Bob", name)

	street, err := l.Load(ctx, "ADDRESS_STREET")
	assert.NoError(t, err)
	assert.Equal(t, "123 Main St", street)

	phone, err := l.Load(ctx, "PHONE")
	assert.NoError(t, err)
	assert.Equal(t, "", phone)
}

func TestFileLoader_InvalidPath(t *testing.T) {
	l, err := NewFileLoader("invalid_path.yml", "_")
	assert.Error(t, err)
	assert.Nil(t, l)
}

func TestNewLoader_InvalidYAML(t *testing.T) {
	yaml := `a
b`

	b := bytes.NewBufferString(yaml)

	l, err := NewLoader(b, "_")
	assert.Error(t, err)
	assert.Nil(t, l)
}

type errReader struct{}

func (e errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated error")
}

func TestNewLoader_ReaderError(t *testing.T) {
	b := errReader{}

	_, _ = io.ReadAll(b)

	l, err := NewLoader(b, "_")
	assert.Error(t, err)
	assert.Nil(t, l)
}
