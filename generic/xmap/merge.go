package xmap

// Merge returns a new map with all key-value pairs from the given input maps.
// Note: If the same key exists in multiple input maps, the value from the last
// input map will be used.
func Merge[M ~map[K]V, K comparable, V any](maps ...M) M {
	output := make(M)

	for _, m := range maps {
		for k, v := range m {
			output[k] = v
		}
	}

	return output
}
