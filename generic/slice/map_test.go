package slice

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStructA struct {
	Value int
}

type testStructASlice []testStructA

type testStructB struct {
	Value int
}

func TestMap(t *testing.T) {
	tests := []struct {
		name   string
		elems  []any
		mapper Mapper[any, any]
		want   []any
	}{
		{
			name:  "DoubleIntegers",
			elems: []interface{}{1, 2, 3, 4, 5, 6},
			mapper: func(i any) any {
				return i.(int) * 2
			},
			want: []any{2, 4, 6, 8, 10, 12},
		},
		{
			name:  "MapStructs",
			elems: []interface{}{testStructA{Value: 2}, testStructA{Value: 4}, testStructA{Value: 6}},
			mapper: func(i any) any {
				return testStructB{Value: i.(testStructA).Value * 2}
			},
			want: []any{testStructB{Value: 4}, testStructB{Value: 8}, testStructB{Value: 12}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.want, Map(tt.elems, tt.mapper))
		})
	}

	t.Run("MapTypedSlice", func(t *testing.T) {
		assert.EqualValues(t, []testStructB{{Value: 4}, {Value: 8}, {Value: 12}}, Map(testStructASlice{
			testStructA{Value: 2},
			testStructA{Value: 4},
			testStructA{Value: 6},
		}, func(i testStructA) testStructB { return testStructB{Value: i.Value * 2} }))
	})
}

func TestMapConcurrentWithContext(t *testing.T) {
	tCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cCtx, cancelNow := context.WithCancel(context.Background())
	cancelNow()

	tests := []struct {
		name   string
		ctx    context.Context
		elems  []any
		mapper MapperWithContext[any, any]
		want   []any
	}{
		{
			name:  "DoubleIntegers",
			ctx:   context.Background(),
			elems: []interface{}{1, 2, 3, 4, 5, 6},
			mapper: func(_ context.Context, i any) any {
				return i.(int) * 2
			},
			want: []any{2, 4, 6, 8, 10, 12},
		},
		{
			name:  "MapStructs",
			ctx:   context.Background(),
			elems: []interface{}{testStructA{Value: 2}, testStructA{Value: 4}, testStructA{Value: 6}},
			mapper: func(_ context.Context, i any) any {
				return testStructB{Value: i.(testStructA).Value * 2}
			},
			want: []any{testStructB{Value: 4}, testStructB{Value: 8}, testStructB{Value: 12}},
		},
		{
			name:  "SlowMapper",
			ctx:   tCtx,
			elems: []interface{}{1, 2, 4, 8, 16, 32},
			mapper: func(_ context.Context, i any) any {
				ii := i.(int)
				time.Sleep(time.Duration(ii) * time.Second)
				return ii * 2
			},
			want: []any{2, 4, 8},
		},
		{
			name:  "AlreadyCancelledContext",
			ctx:   cCtx,
			elems: []interface{}{1, 2, 4, 8, 16, 32},
			mapper: func(_ context.Context, i any) any {
				return i.(int) * 2
			},
			want: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.want, MapConcurrentWithContext(tt.ctx, tt.elems, tt.mapper))
		})
	}
}

func TestMapConcurrent(t *testing.T) {
	tests := []struct {
		name   string
		elems  []any
		mapper Mapper[any, any]
		want   []any
	}{
		{
			name:  "DoubleIntegers",
			elems: []interface{}{1, 2, 3, 4, 5, 6},
			mapper: func(i any) any {
				return i.(int) * 2
			},
			want: []any{2, 4, 6, 8, 10, 12},
		},
		{
			name:  "MapStructs",
			elems: []interface{}{testStructA{Value: 2}, testStructA{Value: 4}, testStructA{Value: 6}},
			mapper: func(i any) any {
				return testStructB{Value: i.(testStructA).Value * 2}
			},
			want: []any{testStructB{Value: 4}, testStructB{Value: 8}, testStructB{Value: 12}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.want, MapConcurrent(tt.elems, tt.mapper))
		})
	}
}

func BenchmarkMap(b *testing.B) {
	b.StopTimer()
	input := make([]int64, b.N)
	b.StartTimer()
	Map(input, func(v int64) int64 {
		return v
	})
}

func BenchmarkMapConcurrentWithContext(b *testing.B) {
	b.StopTimer()
	input := make([]int64, b.N)
	b.StartTimer()
	MapConcurrentWithContext(context.Background(), input, func(_ context.Context, v int64) int64 { return v })
}

func BenchmarkMapConcurrent(b *testing.B) {
	b.StopTimer()
	input := make([]int64, b.N)
	b.StartTimer()
	MapConcurrent(input, func(v int64) int64 { return v })
}
