package slice

// NotFound indicates that the element was not found in the elements slice passed to Find function.
const NotFound = -1

// Find returns the first element from the given elems slice for which the Predicate is satisfied.
// Returned can be either the zero value of type or nil(if slice of pointers is given) and the index,
// iff found otherwise -1 is returned.
func Find[T any](elems []T, predicate Predicate[T]) (T, int) {
	for i, v := range elems {
		if predicate(v) {
			return v, i
		}
	}

	return *new(T), NotFound
}
