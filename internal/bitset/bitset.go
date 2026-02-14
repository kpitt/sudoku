package bitset

import (
	"iter"
	"math/bits"
)

// BitSet16 is a very efficient set implementation for small integers in the
// range [0, 15], based on a `uint16` bitmask.  For performance reasons, it
// does not validate inserted elements, so using integers outside of the [0, 15]
// range will produce incorrect results.
type BitSet16 uint16

// FromValues16 initializes a new 16-bit bitset with the given elements.
func FromValues16(items ...int) BitSet16 {
	s := BitSet16(0)
	for _, item := range items {
		s |= BitSet16(1 << item)
	}
	return s
}

// Add inserts a new element into the bitset.
func (s *BitSet16) Add(item int) {
	*s |= BitSet16(1 << item)
}

// Remove deletes an element from the bitset.
func (s *BitSet16) Remove(item int) {
	*s &= ^BitSet16(1 << item)
}

// Contains checks for the existence of an element.
func (s BitSet16) Contains(item int) bool {
	return s&BitSet16(1<<item) != 0
}

// Size returns the number of elements in the bitset.
func (s BitSet16) Size() int {
	return bits.OnesCount16(uint16(s))
}

// Empty returns true if the bitset is empty.
func (s BitSet16) Empty() bool {
	return s == 0
}

// Equal returns true if this set and set a contain exactly the same elements.
func (s BitSet16) Equal(a BitSet16) bool {
	return s == a
}

// Clear removes all elements from the bitset.
func (s *BitSet16) Clear() {
	*s = 0
}

// Values retrieves the values of all elements as a slice.
// The elements are returned in increasing numerical order.
func (s BitSet16) Values() []int {
	values := make([]int, 0, s.Size())
	for k := range s.All() {
		values = append(values, k)
	}
	return values
}

// Value retrieves the "first" element in the bitset.  This is intended to be
// used when the bitset is known to contain exactly one element.
// Returns a value of 16 if the bitset is empty.
func (s BitSet16) Value() int {
	return bits.TrailingZeros16(uint16(s))
}

// All returns an iterator over all elements in the bitset.
// The elements are returned in increasing numerical order.
func (s BitSet16) All() iter.Seq[int] {
	return func(yield func(int) bool) {
		// Capture the set by value so that it doesn't escape the curried
		// function.
		s1 := s
		for s1 != 0 {
			// Get the position of the least significant bit
			lsb := bits.TrailingZeros16(uint16(s1))
			if !yield(lsb) {
				return
			}
			// Clear the least significant bit
			s1 &= s1 - 1
		}
	}
}

// Union returns a new set that is the union of s and a.
func (s BitSet16) Union(a BitSet16) BitSet16 {
	return s | a
}

// Union returns a new set containing the union of specified sets.
func Union(sets ...BitSet16) BitSet16 {
	u := BitSet16(0)
	for _, s := range sets {
		u |= s
	}
	return u
}

// Intersection returns a new set containing elements common to both sets.
func (s BitSet16) Intersection(other BitSet16) BitSet16 {
	return s & other
}

// Difference returns a new set containing elements in s that are not in other.
func (s BitSet16) Difference(other BitSet16) BitSet16 {
	return s &^ other
}

// Intersects returns true if the sets share at least one element.
func (s BitSet16) Intersects(other BitSet16) bool {
	return s&other != 0
}

// IsSubsetOf returns true if all elements of s are also in other.
func (s BitSet16) IsSubsetOf(other BitSet16) bool {
	return s&other == s
}

// String returns a string representation of the set, e.g. "{1, 5, 9}".
func (s BitSet16) String() string {
	b := []byte{'{'}
	first := true
	for val := range s.All() {
		if !first {
			b = append(b, ',', ' ')
		}
		// Appending small integers: handling 1 or 2 digits (max 16)
		if val >= 10 {
			b = append(b, byte('0'+val/10))
		}
		b = append(b, byte('0'+val%10))
		first = false
	}
	b = append(b, '}')
	return string(b)
}
