package slice

// Accumulator is the function executed on each element of the slice in order,
// passing the return value of the previous function call on the preceding slice element.
type Accumulator[T1, T2 any] func(T2, T1) T2

// Reduce iterates over the slice and executes an Accumulator function on each element.
// The final result of running the Accumulator across all slice elements is returned.
func Reduce[S ~[]T1, T1, T2 any](elems S, accumulator Accumulator[T1, T2]) T2 {
	var initial T2

	return ReduceWithInitialValue(elems, initial, accumulator)
}

// ReduceWithInitialValue iterates over the slice and executes an Accumulator function on each element.
// Unlike Reduce, the initial value of the accumulator can be provided as an argument.
// The final result of running the Accumulator across all slice elements is returned.
func ReduceWithInitialValue[S ~[]T1, T1, T2 any](elems S, initial T2, accumulator Accumulator[T1, T2]) T2 {
	for _, v := range elems {
		initial = accumulator(initial, v)
	}

	return initial
}
