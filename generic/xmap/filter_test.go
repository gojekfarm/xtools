package xmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterInt(t *testing.T) {
	assert.EqualValues(t, map[int]bool{
		2: true,
		4: true,
		6: true,
	}, Filter(map[int]bool{
		1: true,
		2: true,
		3: true,
		4: true,
		5: true,
		6: true,
	}, func(k int, v bool) bool { return k%2 == 0 }))
}

func TestFilterStruct(t *testing.T) {
	type testStruct struct {
		ID    int
		Valid bool
	}

	assert.EqualValues(t, map[testStruct]bool{
		{ID: 1, Valid: true}: true,
		{ID: 3, Valid: true}: true,
	}, Filter(map[testStruct]bool{
		{ID: 1, Valid: true}: true,
		{ID: 2}:              true,
		{ID: 3, Valid: true}: true,
		{ID: 4}:              true,
	}, func(k testStruct, _ bool) bool { return k.Valid }))
}
