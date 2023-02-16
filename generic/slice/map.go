package slice

import "context"

// Mapper is a function that maps/converts type T1 to T2.
type Mapper[T1, T2 any] func(T1) T2

// MapperWithContext is a function that maps/converts type T1 to T2 and also gets the context.
type MapperWithContext[T1, T2 any] func(context.Context, T1) T2

// Map returns a new slice populated with the result of calling the Mapper
// on every element in the given elems slice.
func Map[S ~[]T1, T1, T2 any](elems S, mapper Mapper[T1, T2]) []T2 {
	output := make([]T2, 0, len(elems))

	for _, e := range elems {
		output = append(output, mapper(e))
	}

	return output
}

// MapConcurrentWithContext does the same as Map, but concurrently, and receives a context.Context to be cancellable.
// Note: For simple map operations, Map is about 50x faster than MapConcurrentWithContext.
func MapConcurrentWithContext[S ~[]T1, T1, T2 any](
	ctx context.Context, elems S, mapper MapperWithContext[T1, T2]) []T2 {
	elemOrder := make(chan chan T2, len(elems))
	output := make([]T2, 0, len(elems))

	go func() {
		defer close(elemOrder)

		for _, v := range elems {
			elemC := make(chan T2, 1)
			select {
			case <-ctx.Done():
				return
			case elemOrder <- elemC:
			}

			go func(elemC chan<- T2, v T1) {
				select {
				case <-ctx.Done():
					return
				case elemC <- mapper(ctx, v):
				}
			}(elemC, v)
		}
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case elemC, ok := <-elemOrder:
			if !ok {
				break loop
			}
			select {
			case <-ctx.Done():
				break loop
			case elem := <-elemC:
				output = append(output, elem)
			}
		}
	}

	return output
}

// MapConcurrent does the same as Map, but concurrently.
// Note: For simple map operations, Map is about 50x faster than MapConcurrent.
func MapConcurrent[S ~[]T1, T1, T2 any](elems S, mapper Mapper[T1, T2]) []T2 {
	return MapConcurrentWithContext(context.Background(), elems, func(_ context.Context, e T1) T2 { return mapper(e) })
}
