package xmap_test

import (
	"fmt"
	"sort"

	"github.com/gojekfarm/xtools/generic/xmap"
)

func ExampleFilter() {
	keyLenIsTwo := func(k, _ string) bool { return len(k) == 2 }
	input := map[string]string{"1": "1", "10": "10", "20": "20", "50": "50", "100": "100"}
	keyLenTwoMap := xmap.Filter(input, keyLenIsTwo)
	fmt.Println(keyLenTwoMap)
	// Output: map[10:10 20:20 50:50]
}

func ExampleKeys() {
	input := map[string]string{"1": "1", "10": "10", "20": "20", "50": "50", "100": "100"}
	keys := xmap.Keys(input)
	sort.Strings(keys)
	fmt.Println(keys)
	// Output: [1 10 100 20 50]
}
