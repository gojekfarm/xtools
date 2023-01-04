package xmap

// Predicate is a predicate function that returns true,
// iff all intended conditions are satisfied
// by given input key T1 and value T2.
type Predicate[T1 comparable, T2 any] func(T1, T2) bool

// Filter returns a new map with all elements from the given input map for which the Predicate is satisfied.
func Filter[T1 comparable, T2 any](input map[T1]T2, predicate Predicate[T1, T2]) map[T1]T2 {
	output := make(map[T1]T2)

	for k, v := range input {
		if predicate(k, v) {
			output[k] = v
		}
	}

	return output
}
