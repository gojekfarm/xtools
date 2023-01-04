package generic

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSet(t *testing.T) {
	assert.NotNil(t, NewSet([]string{"a", "b"}...))
	assert.NotNil(t, NewSet([]int{1, 2}...))
}

func TestSet_Len(t *testing.T) {
	tests := []struct {
		name  string
		elems []string
		len   int
	}{
		{
			name: "ZeroLen",
		},
		{
			name:  "NonZeroLen",
			elems: []string{"a", "b", "c"},
			len:   3,
		},
		{
			name:  "NonZeroLenWithSameElements",
			elems: []string{"a", "b", "c", "c"},
			len:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.len, NewSet(tt.elems...).Len())
		})
	}
}

func TestSet_Add(t *testing.T) {
	tests := []struct {
		name     string
		elems    []string
		expected []string
	}{
		{
			name:     "EmptySet",
			expected: []string{},
		},
		{
			name:     "ElementsAddedToSet",
			elems:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "DuplicateElementsAddedToSet",
			elems:    []string{"a", "a", "b", "c", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSet[string]()
			s.Add(tt.elems...)
			sort.Strings(tt.expected)
			actual := s.Values()
			sort.Strings(actual)
			assert.EqualValues(t, tt.expected, actual)
		})
	}
}

func TestSet_Delete(t *testing.T) {
	tests := []struct {
		name     string
		elems    []string
		delete   []string
		expected []string
	}{
		{
			name:     "EmptySet",
			delete:   []string{"a"},
			expected: []string{},
		},
		{
			name:     "ElementsToDeletePresent",
			elems:    []string{"a", "b", "c"},
			delete:   []string{"a", "b"},
			expected: []string{"c"},
		},
		{
			name:     "ElementsToDeleteNotPresent",
			elems:    []string{"a", "b", "c", "d"},
			delete:   []string{"e", "f"},
			expected: []string{"a", "b", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSet(tt.elems...)
			s.Delete(tt.delete...)
			sort.Strings(tt.expected)
			actual := s.Values()
			sort.Strings(actual)
			assert.EqualValues(t, tt.expected, actual)
		})
	}
}

func TestSet_Has(t *testing.T) {
	tests := []struct {
		name  string
		elems []string
		elem  string
		want  bool
	}{
		{
			name: "EmptySetMatch",
			elem: "a",
		},
		{
			name:  "SetHasNoMatch",
			elems: []string{"a", "b", "c"},
			elem:  "d",
		},
		{
			name:  "SetHasAMatch",
			elems: []string{"a", "b", "c"},
			elem:  "c",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewSet(tt.elems...).Has(tt.elem))
		})
	}
}

func TestSet_HasAll(t *testing.T) {
	tests := []struct {
		name   string
		elems  []string
		toFind []string
		want   bool
	}{
		{
			name:   "EmptySetHasAll",
			toFind: []string{"a"},
		},
		{
			name:   "SetHasAllNoMatch",
			elems:  []string{"a", "b", "c", "d"},
			toFind: []string{"d", "f"},
		},
		{
			name:   "SetHasAllMatch",
			elems:  []string{"a", "b", "c", "d"},
			toFind: []string{"c", "d"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewSet(tt.elems...).HasAll(tt.toFind...))
		})
	}
}

func TestSet_HasAny(t *testing.T) {
	tests := []struct {
		name   string
		elems  []string
		toFind []string
		want   bool
	}{
		{
			name:   "EmptySetHasAny",
			toFind: []string{"a"},
		},
		{
			name:   "SetHasAnyNoMatch",
			elems:  []string{"a", "b", "c", "d"},
			toFind: []string{"e", "f"},
		},
		{
			name:   "SetHasAnyMatch",
			elems:  []string{"a", "b", "c", "d"},
			toFind: []string{"d", "e"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewSet(tt.elems...).HasAny(tt.toFind...))
		})
	}
}

func TestSet_Iterate(t *testing.T) {
	s1 := NewSet(1, 2, 3, 4, 5, 6, 7)
	count := 0
	s1.Iterate(func(i int) {
		count += i
	})
	assert.Equal(t, 28, count)

	s2 := NewSet(1, 2, 3, 4, 5, 6, 7)
	product := 1
	s2.Iterate(func(i int) {
		product *= i
	})
	assert.Equal(t, 5040, product)
}

func TestSet_Values(t *testing.T) {
	tests := []struct {
		name     string
		elems    []int
		expected []int
	}{
		{
			name:     "EmptySet",
			expected: []int{},
		},
		{
			name:     "NonEmptySet",
			elems:    []int{1, 2, 3, 4},
			expected: []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSet(tt.elems...)
			e := tt.expected
			sort.Ints(e)
			v := s.Values()
			sort.Ints(v)
			assert.EqualValues(t, e, v)
		})
	}
}

func TestSet_Clone(t *testing.T) {
	s1 := NewSet(1, 2, 3)
	s2 := s1.Clone()

	v1 := s1.Values()
	sort.Ints(v1)
	v2 := s2.Values()
	sort.Ints(v2)

	assert.NotEqual(t, fmt.Sprintf("%p", &s1), fmt.Sprintf("%p", &s2))
	assert.Equal(t, s1.Len(), s2.Len())
	assert.EqualValues(t, v1, v2)
}

func TestSet_Union(t *testing.T) {
	tests := []struct {
		name     string
		set1     Set[int]
		set2     Set[int]
		expected Set[int]
	}{
		{
			name:     "EmptySets",
			set1:     NewSet[int](),
			set2:     NewSet[int](),
			expected: NewSet[int](),
		},
		{
			name:     "DisjointSets",
			set1:     NewSet(1, 2),
			set2:     NewSet(3, 4),
			expected: NewSet(1, 2, 3, 4),
		},
		{
			name:     "OverlappingSets",
			set1:     NewSet(1, 2),
			set2:     NewSet(2, 3),
			expected: NewSet(1, 2, 3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.expected, tt.set1.Union(tt.set2))
		})
	}
}

func TestSet_Intersection(t *testing.T) {
	tests := []struct {
		name     string
		set1     Set[int]
		set2     Set[int]
		expected Set[int]
	}{
		{
			name:     "EmptySets",
			set1:     NewSet[int](),
			set2:     NewSet[int](),
			expected: NewSet[int](),
		},
		{
			name:     "DisjointSets",
			set1:     NewSet(1, 2),
			set2:     NewSet(3, 4),
			expected: NewSet[int](),
		},
		{
			name:     "OverlappingSets",
			set1:     NewSet(1, 2, 3),
			set2:     NewSet(2, 3, 4),
			expected: NewSet(2, 3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualValues(t, tt.expected, tt.set1.Intersection(tt.set2))
		})
	}
}

func TestSet_String(t *testing.T) {
	s := NewSet(1, 2, 3, 4)
	assert.Regexp(t, `^set\[.*\]$`, s.String())
}
