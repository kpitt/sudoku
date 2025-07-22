package set

// Set represents a collection of unique elements of type T.
type Set[T comparable] struct {
	elements map[T]struct{}
}

// NewSet initializes a new generic set with the given elements.
func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{
		elements: make(map[T]struct{}),
	}
	if len(items) != 0 {
		s.Add(items...)
	}
	return s
}

// Add inserts new elements into the set.
func (s *Set[T]) Add(items ...T) {
	for _, item := range items {
		s.elements[item] = struct{}{}
	}
}

// Remove deletes an element from the set.
func (s *Set[T]) Remove(item T) {
	delete(s.elements, item)
}

// Contains checks for the existence of an element.
func (s *Set[T]) Contains(item T) bool {
	_, exists := s.elements[item]
	return exists
}

// Size returns the number of elements in the set.
func (s *Set[T]) Size() int {
	return len(s.elements)
}

// Clear removes all elements from the set.
func (s *Set[T]) Clear() {
	clear(s.elements)
}

// Values retrieves the values of all elements as a slice.
func (s *Set[T]) Values() []T {
	values := make([]T, 0, len(s.elements))
	for k := range s.elements {
		values = append(values, k)
	}
	return values
}

// Union updates set s to be the union of s with set a.
// Note that this function modifies s in place.  To return the union as a new
// set, use the `set.Union` function instead.
func (s *Set[T]) Union(a *Set[T]) {
	for k := range a.elements {
		s.Add(k)
	}
}

// Union returns a new set containing the union of sets a and b.
func Union[T comparable](a *Set[T], b *Set[T]) *Set[T] {
	u := NewSet[T]()
	u.Union(a)
	u.Union(b)
	return u
}
