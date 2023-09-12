package exporter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xtel"
)

func TestNewSTDOut(t *testing.T) {
	p, err := xtel.NewProvider("test-service", NewSTDOut(STDOutOptions{PrettyPrint: true}))
	assert.NoError(t, err)
	assert.NotNil(t, p)
}
