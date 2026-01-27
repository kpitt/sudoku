package set

import (
	"slices"
	"testing"
)

func TestNewSet(t *testing.T) {
	s := New[int]()
	if s.Size() != 0 {
		t.Errorf("NewSet() should be empty, got size %d", s.Size())
	}

	s2 := New(1, 2, 3)
	if s2.Size() != 3 {
		t.Errorf("NewSet(1, 2, 3) should have size 3, got %d", s2.Size())
	}
	if !s2.Contains(1) || !s2.Contains(2) || !s2.Contains(3) {
		t.Errorf("NewSet(1, 2, 3) missing elements")
	}
}

func TestAdd(t *testing.T) {
	s := New[string]()
	s.Add("a")
	if !s.Contains("a") {
		t.Errorf("Add('a') failed")
	}
	s.Add("b", "c")
	if s.Size() != 3 {
		t.Errorf("Add('b', 'c') failed, size is %d", s.Size())
	}
	// Test duplicates
	s.Add("a")
	if s.Size() != 3 {
		t.Errorf("Add existing element should not increase size")
	}
}

func TestRemove(t *testing.T) {
	s := New(1, 2, 3)
	s.Remove(2)
	if s.Contains(2) {
		t.Errorf("Remove(2) failed")
	}
	if s.Size() != 2 {
		t.Errorf("Size should be 2 after removing one element, got %d", s.Size())
	}
	s.Remove(99) // Remove non-existent
	if s.Size() != 2 {
		t.Errorf("Remove non-existent element should not change size")
	}
}

func TestClear(t *testing.T) {
	s := New(1, 2, 3)
	s.Clear()
	if s.Size() != 0 {
		t.Errorf("Clear failed, size is %d", s.Size())
	}
}

func TestValues(t *testing.T) {
	s := New(1, 2, 3)
	vals := s.Values()
	if len(vals) != 3 {
		t.Errorf("Values() returned wrong length: %d", len(vals))
	}
	slices.Sort(vals)
	if !slices.Equal(vals, []int{1, 2, 3}) {
		t.Errorf("Values() returned incorrect elements: %v", vals)
	}
}

func TestUnion(t *testing.T) {
	s1 := New(1, 2)
	s2 := New(2, 3)

	// Test method Union (in-place)
	s1.Union(s2)
	if s1.Size() != 3 {
		t.Errorf("Union (method) failed size check: %d", s1.Size())
	}
	if !s1.Contains(1) || !s1.Contains(2) || !s1.Contains(3) {
		t.Errorf("Union (method) missing elements")
	}

	// Test function Union (new set)
	sA := New(10, 20)
	sB := New(20, 30)
	sC := Union(sA, sB)

	if sC.Size() != 3 {
		t.Errorf("Union (function) failed size check: %d", sC.Size())
	}
	if !sC.Contains(10) || !sC.Contains(20) || !sC.Contains(30) {
		t.Errorf("Union (function) missing elements")
	}

	// Ensure originals are untouched
	if sA.Size() != 2 || sB.Size() != 2 {
		t.Errorf("Union (function) modified original sets")
	}
}
