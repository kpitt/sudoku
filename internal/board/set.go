package board

// Set represents a collection of unique elements of type T.
type Set[T comparable] struct {
	m map[T]struct{}
}

// NewSet initializes a new generic set with the given elements.
func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{
		m: make(map[T]struct{}),
	}
	s.Add(items...)
	return s
}

// Add inserts new elements into the set.
func (s *Set[T]) Add(items ...T) {
	for _, item := range items {
		s.m[item] = struct{}{}
	}
}

// Remove deletes an element from the set.
func (s *Set[T]) Remove(item T) {
	delete(s.m, item)
}

// Contains checks for the existence of an element.
func (s *Set[T]) Contains(item T) bool {
	_, exists := s.m[item]
	return exists
}

// Size returns the number of elements in the set.
func (s *Set[T]) Size() int {
	return len(s.m)
}

// Clear removes all elements from the set.
func (s *Set[T]) Clear() {
	s.m = make(map[T]struct{})
}

// Elements retrieves all elements as a slice.
func (s *Set[T]) Elements() []T {
	elements := make([]T, 0, len(s.m))
	for k := range s.m {
		elements = append(elements, k)
	}
	return elements
}
