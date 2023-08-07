package xmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntMerge(t *testing.T) {
	merged := Merge(
		map[int]bool{1: true},
		map[int]bool{2: true},
		map[int]bool{3: true},
		map[int]bool{4: true},
		map[int]bool{5: true},
		map[int]bool{6: true},
	)

	assert.EqualValues(t,
		map[int]bool{
			1: true, 2: true, 3: true,
			4: true, 5: true, 6: true,
		},
		merged,
	)
}

func TestIntMergeLastKeyOverride(t *testing.T) {
	merged := Merge(
		map[int]bool{1: true},
		map[int]bool{2: true},
		map[int]bool{3: true},
		map[int]bool{4: true},
		map[int]bool{5: true},
		map[int]bool{6: true},
		map[int]bool{5: false},
	)

	assert.EqualValues(t,
		map[int]bool{
			1: true, 2: true, 3: true,
			4: true, 5: false, 6: true,
		},
		merged,
	)
}

func TestStructMerge(t *testing.T) {
	type testStruct struct {
		ID    int
		Valid bool
	}

	merged := Merge(map[testStruct]bool{
		{ID: 1, Valid: true}: true,
		{ID: 2}:              true,
	}, map[testStruct]bool{
		{ID: 3, Valid: true}: true,
		{ID: 4}:              true,
	})

	assert.EqualValues(t,
		map[testStruct]bool{
			{ID: 1, Valid: true}: true,
			{ID: 2}:              true,
			{ID: 3, Valid: true}: true,
			{ID: 4}:              true,
		}, merged,
	)
}
