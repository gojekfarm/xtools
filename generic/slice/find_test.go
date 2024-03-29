package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	type testStruct struct {
		ID int
	}

	type testStructSlice []testStruct

	tests := []struct {
		name      string
		elems     []any
		predicate func(any) bool
		want      bool
		wantElem  any
	}{
		{
			name: "StructWithID2",
			elems: []interface{}{
				testStruct{ID: 1},
				testStruct{ID: 2},
				testStruct{ID: 3},
				testStruct{ID: 4},
			},
			predicate: func(i any) bool {
				return i.(testStruct).ID == 2
			},
			want:     true,
			wantElem: testStruct{ID: 2},
		},
		{
			name: "PointerToStructWithID3",
			elems: []interface{}{
				&testStruct{ID: 1},
				&testStruct{ID: 2},
				&testStruct{ID: 3},
				&testStruct{ID: 4},
			},
			predicate: func(i any) bool {
				return i.(*testStruct).ID == 3
			},
			want:     true,
			wantElem: &testStruct{ID: 3},
		},
		{
			name: "NoStructWithMatchingID",
			elems: []interface{}{
				testStruct{ID: 1},
				testStruct{ID: 2},
				testStruct{ID: 3},
				testStruct{ID: 4},
			},
			predicate: func(i any) bool {
				return i.(testStruct).ID == 5
			},
		},
		{
			name: "NoPointerStructWithMatchingID",
			elems: []interface{}{
				&testStruct{ID: 1},
				&testStruct{ID: 2},
				&testStruct{ID: 4},
			},
			predicate: func(i any) bool {
				return i.(*testStruct).ID == 3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, index := Find(tt.elems, tt.predicate)
			if tt.want {
				assert.NotEqual(t, NotFound, index)
				assert.EqualValues(t, tt.wantElem, v)
			} else {
				assert.Equal(t, NotFound, index)
				assert.Nil(t, v)
			}
		})
	}

	t.Run("TypedSlices", func(t *testing.T) {
		t.Run("ElemFound", func(t *testing.T) {
			v, index := Find(testStructSlice{
				testStruct{ID: 1},
				testStruct{ID: 2},
				testStruct{ID: 3},
				testStruct{ID: 4},
			}, func(i testStruct) bool { return i.ID == 2 })
			assert.NotEqual(t, NotFound, index)
			assert.EqualValues(t, testStruct{ID: 2}, v)
		})

		t.Run("ElemNotFound", func(t *testing.T) {
			v, index := Find(testStructSlice{
				testStruct{ID: 1},
				testStruct{ID: 2},
				testStruct{ID: 3},
				testStruct{ID: 4},
			}, func(i testStruct) bool { return i.ID == 5 })
			assert.Equal(t, NotFound, index)
			assert.EqualValues(t, testStruct{}, v)
		})
	})
}
