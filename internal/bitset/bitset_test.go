package bitset

import (
	"slices"
	"testing"
)

func TestNewSet(t *testing.T) {
	var s BitSet16
	if s.Size() != 0 {
		t.Errorf("Default BitSet16 should be empty, got size %d", s.Size())
	}

	s2 := FromValues16(1, 2, 3)
	if s2.Size() != 3 {
		t.Errorf("FromValues16(1, 2, 3) should have size 3, got %d", s2.Size())
	}
	if !s2.Contains(1) || !s2.Contains(2) || !s2.Contains(3) {
		t.Errorf("FromValues16(1, 2, 3) missing elements")
	}
}

func TestAdd(t *testing.T) {
	var s BitSet16
	s.Add(1)
	if !s.Contains(1) {
		t.Errorf("Add(1) failed")
	}
	s.Add(2)
	s.Add(3)
	if s.Size() != 3 || !s.Contains(2) || !s.Contains(3) {
		t.Errorf("Add(2, 3) failed, size is %d", s.Size())
	}
	// Test duplicates
	s.Add(1)
	if s.Size() != 3 {
		t.Errorf("Add existing element should not increase size")
	}
}

func TestRemove(t *testing.T) {
	s := FromValues16(1, 2, 3)
	s.Remove(2)
	if s.Contains(2) {
		t.Errorf("Remove(2) failed")
	}
	if s.Size() != 2 {
		t.Errorf("Size should be 2 after removing one element, got %d", s.Size())
	}
	s.Remove(15) // Remove non-existent
	if s.Size() != 2 {
		t.Errorf("Remove non-existent element should not change size")
	}
}

func TestClear(t *testing.T) {
	s := FromValues16(1, 2, 3)
	s.Clear()
	if s.Size() != 0 {
		t.Errorf("Clear failed, size is %d", s.Size())
	}
}

func TestValues(t *testing.T) {
	s := FromValues16(1, 2, 3)
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
	s1 := FromValues16(1, 2)
	s2 := FromValues16(2, 3)

	// Test method Union (in-place)
	s1.Union(s2)
	if s1.Size() != 3 {
		t.Errorf("Union (method) failed size check: %d", s1.Size())
	}
	if !s1.Contains(1) || !s1.Contains(2) || !s1.Contains(3) {
		t.Errorf("Union (method) missing elements")
	}

	// Test function Union (new set)
	sA := FromValues16(4, 8)
	sB := FromValues16(8, 12)
	sC := Union(sA, sB)

	if sC.Size() != 3 {
		t.Errorf("Union (function) failed size check: %d", sC.Size())
	}
	if !sC.Contains(4) || !sC.Contains(8) || !sC.Contains(12) {
		t.Errorf("Union (function) missing elements")
	}

	// Ensure originals are untouched
	if sA.Size() != 2 || sB.Size() != 2 {
		t.Errorf("Union (function) modified original sets")
	}
}

func TestIntersection(t *testing.T) {
	s1 := FromValues16(1, 2)
	s3 := FromValues16(2, 4, 6)
	common := FromValues16(2)

	if i := s1.Intersection(s3); !i.Equal(common) {
		t.Errorf("Intersection = %s; want %s", i.String(), common.String())
	}
}

func TestDifference(t *testing.T) {
	s1 := FromValues16(1, 2)
	common := FromValues16(2)
	diff := FromValues16(1)

	if d := s1.Difference(common); !d.Equal(diff) {
		t.Errorf("Difference = %s; want %s", d.String(), diff.String())
	}
}

func TestIntersects(t *testing.T) {
	s1 := FromValues16(1, 2)
	s3 := FromValues16(2, 4, 6)

	if !s1.Intersects(s3) {
		t.Errorf("Expected s1 and s3 to intersect")
	}
	if s1.Intersects(FromValues16(4, 5)) {
		t.Errorf("Expected s1 and {4,5} NOT to intersect")
	}
}

func TestIsSubsetOf(t *testing.T) {
	small := FromValues16(1, 2)
	large := FromValues16(1, 2, 3)

	if !small.IsSubsetOf(large) {
		t.Errorf("Expected small to be subset of large")
	}
	if large.IsSubsetOf(small) {
		t.Errorf("Expected large NOT to be subset of small")
	}
}

func TestString(t *testing.T) {
	s1 := FromValues16(1, 2)
	str := s1.String()
	// {01, 02}
	if str != "{1, 2}" {
		t.Errorf("String() = %q; want \"{1, 2}\"", str)
	}
}
