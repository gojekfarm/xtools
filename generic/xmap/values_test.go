package xmap

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntValues(t *testing.T) {
	vals := Values(map[string]int{
		"1": 1,
		"2": 2,
		"3": 3,
		"4": 4,
		"5": 5,
		"6": 6,
	})

	sort.Ints(vals)
	assert.EqualValues(t, []int{1, 2, 3, 4, 5, 6}, vals)
}

func TestStructValues(t *testing.T) {
	type testStruct struct {
		ID    int
		Valid bool
	}

	vals := Values(map[string]testStruct{
		"1": {ID: 1, Valid: true},
		"2": {ID: 2},
		"3": {ID: 3, Valid: true},
		"4": {ID: 4},
	})

	sort.Slice(vals, func(i, j int) bool {
		return vals[i].ID < vals[j].ID
	})

	assert.EqualValues(t, []testStruct{{ID: 1, Valid: true}, {ID: 2}, {ID: 3, Valid: true}, {ID: 4}}, vals)
}
