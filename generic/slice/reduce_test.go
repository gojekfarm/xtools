package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReduce(t *testing.T) {
	type testStruct struct {
		Value float64
	}

	type testStructSlice []testStruct

	tests := []struct {
		name        string
		elems       []any
		accumulator Accumulator[any, any]
		want        any
	}{
		{
			name:  "SumAllInts",
			elems: []any{1, 2, 3, 4, 5, 6},
			accumulator: func(sum any, i any) any {
				// nil check necessary because of any type having `nil` value as zero value
				if sum == nil {
					return i.(int)
				}
				return sum.(int) + i.(int)
			},
			want: 21,
		},
		{
			name:  "SumAllFloats",
			elems: []any{1.02, 2.01, 3.14, 4.2, 5.3, 6.9},
			accumulator: func(sum any, i any) any {
				// nil check necessary because of any type having `nil` value as zero value
				if sum == nil {
					return i.(float64)
				}
				return sum.(float64) + i.(float64)
			},
			want: 22.57,
		},
		{
			name: "FindStructWithMaxValue",
			elems: []any{
				testStruct{Value: 1.02},
				testStruct{Value: 4.2},
				testStruct{Value: 2.01},
				testStruct{Value: 6.91},
				testStruct{Value: 3.14},
				testStruct{Value: 5.3},
			},
			accumulator: func(in any, e any) any {
				// nil check necessary because of any type having `nil` value as zero value
				if in == nil {
					return e.(testStruct)
				}

				if in.(testStruct).Value > e.(testStruct).Value {
					return in.(testStruct)
				}

				return e.(testStruct)
			},
			want: testStruct{Value: 6.91},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Reduce(tt.elems, tt.accumulator))
		})
	}

	t.Run("FindStructWithMaxValueFromTypedSlice", func(t *testing.T) {
		assert.EqualValues(t, testStruct{Value: 6.91}, Reduce(testStructSlice{
			testStruct{Value: 1.02},
			testStruct{Value: 4.2},
			testStruct{Value: 2.01},
			testStruct{Value: 6.91},
			testStruct{Value: 3.14},
			testStruct{Value: 5.3},
		}, func(in testStruct, e testStruct) testStruct {
			if in.Value > e.Value {
				return in
			}

			return e
		}))
	})
}

func TestReduceWithInitialValue(t *testing.T) {
	type testStruct struct {
		Value float64
	}

	type testStructSlice []testStruct

	tests := []struct {
		name        string
		elems       []any
		initial     any
		accumulator Accumulator[any, any]
		want        any
	}{
		{
			name:        "SumAllInts",
			initial:     0,
			elems:       []any{1, 2, 3, 4, 5, 6},
			accumulator: func(sum any, i any) any { return sum.(int) + i.(int) },
			want:        21,
		},
		{
			name:        "SumAllFloats",
			initial:     0.0,
			elems:       []any{1.02, 2.01, 3.14, 4.2, 5.3, 6.9},
			accumulator: func(sum any, i any) any { return sum.(float64) + i.(float64) },
			want:        22.57,
		},
		{
			name:    "FindStructWithMaxValue",
			initial: testStruct{Value: 0.0},
			elems: []any{
				testStruct{Value: 1.02},
				testStruct{Value: 4.2},
				testStruct{Value: 2.01},
				testStruct{Value: 6.91},
				testStruct{Value: 3.14},
				testStruct{Value: 5.3},
			},
			accumulator: func(in any, e any) any {
				if in.(testStruct).Value > e.(testStruct).Value {
					return in.(testStruct)
				}
				return e.(testStruct)
			},
			want: testStruct{Value: 6.91},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ReduceWithInitialValue(tt.elems, tt.initial, tt.accumulator))
		})
	}

	t.Run("FindStructWithMaxValueFromTypedSlice", func(t *testing.T) {
		assert.EqualValues(t, testStruct{Value: 6.91}, ReduceWithInitialValue(testStructSlice{
			testStruct{Value: 1.02},
			testStruct{Value: 4.2},
			testStruct{Value: 2.01},
			testStruct{Value: 6.91},
			testStruct{Value: 3.14},
			testStruct{Value: 5.3},
		}, testStruct{Value: 0.0}, func(in testStruct, e testStruct) testStruct {
			if in.Value > e.Value {
				return in
			}

			return e
		}))
	})
}
