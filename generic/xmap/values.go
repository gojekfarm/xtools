package xmap

// Values returns a new slice with all values from the given input map.
// Note: Returned values slice is not sorted.
func Values[M ~map[K]V, K comparable, V any](input M) []V {
	output := make([]V, 0, len(input))

	for _, v := range input {
		output = append(output, v)
	}

	return output
}
