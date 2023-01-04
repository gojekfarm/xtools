package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	type testStruct struct {
		ID    int
		Valid bool
	}

	tests := []struct {
		name      string
		elems     []any
		predicate func(any) bool
		want      []any
	}{
		{
			name:  "EvenIntegers",
			elems: []interface{}{1, 2, 3, 4, 5, 6, 7},
			predicate: func(i any) bool {
				return i.(int)%2 == 0
			},
			want: []any{2, 4, 6},
		},
		{
			name: "ValidStructs",
			elems: []interface{}{
				testStruct{ID: 1, Valid: true},
				testStruct{ID: 2},
				testStruct{ID: 3, Valid: true},
				testStruct{ID: 4},
			},
			predicate: func(i any) bool {
				return i.(testStruct).Valid
			},
			want: []any{testStruct{ID: 1, Valid: true}, testStruct{ID: 3, Valid: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.want, Filter(tt.elems, tt.predicate))
		})
	}
}
