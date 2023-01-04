package slice_test

import (
	"context"
	"fmt"
	"time"

	"github.com/gojekfarm/xtools/generic/slice"
)

func Example_mixed() {
	type Order struct {
		Value     int  `json:"value"`
		IsSpecial bool `json:"is_special"`
	}

	allOrders := []Order{
		{Value: 1000, IsSpecial: true},
		{Value: 1500, IsSpecial: false},
		{Value: 2000, IsSpecial: true},
		{Value: 3000, IsSpecial: false},
		{Value: 7000, IsSpecial: true},
		{Value: 9500, IsSpecial: true},
	}

	isSpecialPredicate := func(o Order) bool { return o.IsSpecial }
	orderValueMapper := func(o Order) int { return o.Value }
	totalSum := func(prevSum, currVal int) int { return prevSum + currVal }

	specialOrderTotalValue := slice.
		Reduce(slice.
			Map(slice.
				Filter(allOrders, isSpecialPredicate),
				orderValueMapper,
			),
			totalSum,
		)

	fmt.Println(specialOrderTotalValue)
	// Output: 19500
}

func ExampleFilter() {
	isEven := func(i int) bool { return i%2 == 0 }
	elems := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evenElems := slice.Filter(elems, isEven)
	fmt.Println(evenElems)
	// Output: [2 4 6 8 10]
}

func ExampleFind() {
	isEven := func(i int) bool { return i%2 == 0 }
	elems := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evenElem, index := slice.Find(elems, isEven)
	fmt.Println(evenElem, index, index != slice.NotFound)
	// Output: 2 1 true
}

func ExampleMap() {
	doSquare := func(i int) int { return i * i }
	elems := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	squared := slice.Map(elems, doSquare)
	fmt.Println(squared)
	// Output: [1 4 9 16 25 36 49 64 81 100]
}

func ExampleReduce() {
	sumFunc := func(sum, val int) int { return sum + val }
	elems := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	total := slice.Reduce(elems, sumFunc)
	fmt.Println(total)
	// Output: 55
}

func ExampleReduceWithInitialValue() {
	productFunc := func(prod, val int) int { return prod * val }
	elems := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	totalProduct := slice.ReduceWithInitialValue(elems, 1 /* initial */, productFunc)
	fmt.Println(totalProduct)
	// Output: 3628800
}

func ExampleMapConcurrent() {
	doSquare := func(i int) int {
		time.Sleep(time.Second)
		return i * i
	}

	elems := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	start := time.Now()
	squaredC := slice.MapConcurrent(elems, doSquare)
	fmt.Println(squaredC)
	fmt.Println("MapConcurrent took", int(time.Since(start).Seconds()), "Second")

	start = time.Now()
	squared := slice.Map(elems, doSquare)
	fmt.Println(squared)
	fmt.Println("Map took", int(time.Since(start).Seconds()), "Seconds")

	// Output:
	// [1 4 9 16 25 36 49 64 81 100]
	// MapConcurrent took 1 Second
	// [1 4 9 16 25 36 49 64 81 100]
	// Map took 10 Seconds
}

func ExampleMapConcurrentWithContext() {
	doSquare := func(ctx context.Context, i int) int {
		t := time.NewTimer(time.Duration(i*i) * time.Second)
		select {
		case <-ctx.Done():
			return 0 // Skipped
		case <-t.C:
			return i * i
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	elems := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	squaredC := slice.MapConcurrentWithContext(ctx, elems, doSquare)
	fmt.Println(squaredC)

	// Output: [1 4]
}
