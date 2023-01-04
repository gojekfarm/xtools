package xmap

// Keys returns a new slice with all keys from the given input map.
// Note: Returned keys slice is not sorted.
func Keys[M ~map[K]V, K comparable, V any](input M) []K {
	output := make([]K, 0, len(input))

	for k := range input {
		output = append(output, k)
	}

	return output
}
