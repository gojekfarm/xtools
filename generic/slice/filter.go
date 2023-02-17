package slice

// Predicate is a predicate function that returns true iff all intended conditions are satisfied by given input T.
type Predicate[T any] func(T) bool

// Filter returns a new slice with all elements from the given elems slice for which the Predicate is satisfied.
func Filter[S ~[]T, T any](elems S, predicate Predicate[T]) S {
	var output S

	for _, v := range elems {
		if predicate(v) {
			output = append(output, v)
		}
	}

	return output
}
