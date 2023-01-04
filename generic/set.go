package generic

import (
	"fmt"
	"strings"
)

// Set is a data-type which holds only unique values of any type.
type Set[T comparable] map[T]bool

// NewSet creates a Set with the given elements.
func NewSet[T comparable](elems ...T) Set[T] {
	s := make(Set[T])
	s.Add(elems...)

	return s
}

// Len returns the length of Set; i.e. the number of elements in the Set.
func (s Set[T]) Len() int {
	return len(map[T]bool(s))
}

// Has checks if this Set has the given element.
func (s Set[T]) Has(elem T) bool {
	_, ok := s[elem]

	return ok
}

// HasAll checks if this Set has all the given elements.
func (s Set[T]) HasAll(elems ...T) bool {
	for _, e := range elems {
		if !s.Has(e) {
			return false
		}
	}

	return true
}

// HasAny checks if this Set has any of the given elements.
func (s Set[T]) HasAny(elems ...T) bool {
	for _, e := range elems {
		if s.Has(e) {
			return true
		}
	}

	return false
}

// Add will add the elements to the Set.
func (s Set[T]) Add(elems ...T) {
	for _, elem := range elems {
		s[elem] = true
	}
}

// Delete will delete the elements from the Set.
func (s Set[T]) Delete(elems ...T) {
	for _, elem := range elems {
		delete(s, elem)
	}
}

// Iterate takes an iterator func of type `func(T)` and iterates through each element in the Set.
func (s Set[T]) Iterate(it func(T)) {
	for e := range s {
		it(e)
	}
}

// Values returns the elements of the Set as a slice.
func (s Set[T]) Values() []T {
	elems := make([]T, 0, s.Len())
	s.Iterate(func(elem T) {
		elems = append(elems, elem)
	})

	return elems
}

// Clone creates a new copy of the Set.
func (s Set[T]) Clone() Set[T] {
	return NewSet[T](s.Values()...)
}

// Union returns a Set which has all the elements from this Set and the other Set.
func (s Set[T]) Union(other Set[T]) Set[T] {
	set := s.Clone()
	set.Add(other.Values()...)

	return set
}

// Intersection returns a Set with common elements from this Set and the other Set.
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	set := make(Set[T])

	s.Iterate(func(elem T) {
		if other.Has(elem) {
			set.Add(elem)
		}
	})

	return set
}

// String returns the ''native'' string format for Set.
func (s Set[T]) String() string {
	ms := fmt.Sprintf("%+v", map[T]bool(s))[3:]

	return "set" + strings.ReplaceAll(ms, ":true", "")
}
