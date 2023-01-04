package xmap

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntKeys(t *testing.T) {
	keys := Keys(map[int]bool{
		1: true,
		2: true,
		3: true,
		4: true,
		5: true,
		6: true,
	})

	sort.Ints(keys)
	assert.EqualValues(t, []int{1, 2, 3, 4, 5, 6}, keys)
}

func TestStructKeys(t *testing.T) {
	type testStruct struct {
		ID    int
		Valid bool
	}

	keys := Keys(map[testStruct]bool{
		{ID: 1, Valid: true}: true,
		{ID: 2}:              true,
		{ID: 3, Valid: true}: true,
		{ID: 4}:              true,
	})

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].ID < keys[j].ID
	})

	assert.EqualValues(t, []testStruct{{ID: 1, Valid: true}, {ID: 2}, {ID: 3, Valid: true}, {ID: 4}}, keys)
}
