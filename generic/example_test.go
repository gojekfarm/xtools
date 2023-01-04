package generic_test

import (
	"fmt"
	"sort"

	"github.com/gojekfarm/xtools/generic"
)

func ExampleNewSet() {
	s1 := generic.NewSet("a", "b", "c", "b")
	fmt.Println(s1)

	s2 := generic.NewSet(1, 2, 2, 3, 4, 5, 6, 5)
	fmt.Println(s2)

	floats := []float64{1.1, 2.3, 3.4, 2.3}
	s3 := generic.NewSet(floats...)
	fmt.Println(s3)

	// Output:
	// set[a b c]
	// set[1 2 3 4 5 6]
	// set[1.1 2.3 3.4]
}

func ExampleNewSet_customType() {
	type A struct {
		Name string `json:"name"`
	}

	as := []A{{Name: "a"}, {Name: "b"}, {Name: "b"}}
	s := generic.NewSet(as...)
	fmt.Println(s)

	// Output:
	// set[{Name:a} {Name:b}]
}

func ExampleSet_Add() {
	s := generic.NewSet(1)
	s.Add(2, 3)
	fmt.Println(s)

	// Output:
	// set[1 2 3]
}

func ExampleSet_Delete() {
	s := generic.NewSet(1, 2, 3, 4)
	s.Delete(2, 3)
	fmt.Println(s)

	// Output:
	// set[1 4]
}

func ExampleSet_Clone() {
	s1 := generic.NewSet(1, 2, 3, 4)
	s2 := s1.Clone()

	fmt.Println(s1)
	fmt.Println(s2)

	// Output:
	// set[1 2 3 4]
	// set[1 2 3 4]
}

func ExampleSet_Len() {
	s := generic.NewSet(1, 2, 3, 4)
	fmt.Println(s.Len())

	// Output: 4
}

func ExampleSet_Has() {
	s := generic.NewSet(1, 2, 3, 4)
	fmt.Println(s.Has(2))
	fmt.Println(s.Has(5))

	// Output:
	// true
	// false
}

func ExampleSet_HasAny() {
	s := generic.NewSet(1, 2, 3, 4)
	fmt.Println(s.HasAny(4, 5, 6))
	fmt.Println(s.HasAny(5, 6, 7))

	// Output:
	// true
	// false
}

func ExampleSet_HasAll() {
	s := generic.NewSet(1, 2, 3, 4)
	fmt.Println(s.HasAll(2, 3, 4))
	fmt.Println(s.HasAll(3, 4, 5))

	// Output:
	// true
	// false
}

func ExampleSet_Values() {
	s := generic.NewSet(1, 2, 3, 4)

	arr := s.Values()
	sort.Ints(arr)

	fmt.Println(arr)

	for _, v := range arr {
		fmt.Println(v)
	}

	// Output:
	// [1 2 3 4]
	// 1
	// 2
	// 3
	// 4
}

func ExampleSet_Union() {
	s1 := generic.NewSet(1, 2, 3, 4)
	s2 := generic.NewSet(3, 4, 5, 6)
	fmt.Println(s1.Union(s2))

	// Output:
	// set[1 2 3 4 5 6]
}

func ExampleSet_Iterate() {
	s := generic.NewSet(1, 2, 3, 4)

	var elems []int
	s.Iterate(func(i int) {
		elems = append(elems, i)
	})

	sort.Ints(elems)

	fmt.Println(elems)

	// Output:
	// [1 2 3 4]
}

func ExampleSet_Intersection() {
	s1 := generic.NewSet(1, 2, 3, 4)
	s2 := generic.NewSet(3, 4, 5, 6)
	fmt.Println(s1.Intersection(s2))

	// Output:
	// set[3 4]
}
