package solver

import (
	"testing"

	"github.com/kpitt/sudoku/internal/puzzle"
)

func TestFindSkyscraper(t *testing.T) {
	// Create an empty puzzle and solver
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)

	// We'll test with digit 1.
	val := 1

	// Setup Skyscraper pattern on Rows.
	// Base Lines: Row 1 and Row 4.
	// Common Column: Col 1.
	// Tops: (1, 7) and (4, 8).
	// Target to eliminate: (3, 7).
	// (3, 7) sees Top 1 (1, 7) via Column 7.
	// (3, 7) sees Top 2 (4, 8) via Box 5 (both in Box 5).

	// Helper to set candidate
	setCandidate := func(r, c int) {
		p.Get(r, c).Candidates.Add(val)
		s.rows[r].Unsolved[val].Add(c)
		s.columns[c].Unsolved[val].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[val].Add(boxLoc)
	}

	// Prepare state:
	// Clear all candidates for val 1 first
	for r := 0; r < 9; r++ {
		s.rows[r].Unsolved[val].Clear()
		for c := 0; c < 9; c++ {
			p.Get(r, c).RemoveCandidate(val)
		}
	}
	// Also clear nums for cols/boxes for consistency
	for c := 0; c < 9; c++ {
		s.columns[c].Unsolved[val].Clear()
	}
	for b := 0; b < 9; b++ {
		s.boxes[b].Unsolved[val].Clear()
	}

	// Set up the pattern
	setCandidate(1, 1) // Base 1, Common
	setCandidate(1, 7) // Base 1, Top 1

	setCandidate(4, 1) // Base 2, Common
	setCandidate(4, 8) // Base 2, Top 2

	// Set up target
	setCandidate(3, 7) // Target (Row 3, Col 7)

	// Run Skyscraper
	found := s.findSkyscraper()
	if !found {
		t.Errorf("Skyscraper technique should have been found")
	}

	// Verify eliminations
	if p.Get(3, 7).HasCandidate(val) {
		t.Errorf("Target (3,7) should have eliminated candidate %d", val)
	}
}

func TestFindTwoStringKite(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)
	val := 1

	setCandidate := func(r, c int) {
		p.Get(r, c).Candidates.Add(val)
		s.rows[r].Unsolved[val].Add(c)
		s.columns[c].Unsolved[val].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[val].Add(boxLoc)
	}

	// Clear candidates for val 1
	for r := 0; r < 9; r++ {
		s.rows[r].Unsolved[val].Clear()
		for c := 0; c < 9; c++ {
			p.Get(r, c).RemoveCandidate(val)
		}
	}
	for c := 0; c < 9; c++ {
		s.columns[c].Unsolved[val].Clear()
	}
	for b := 0; b < 9; b++ {
		s.boxes[b].Unsolved[val].Clear()
	}

	setCandidate(0, 0)
	setCandidate(0, 4)
	setCandidate(1, 2)
	setCandidate(5, 2)
	setCandidate(5, 4) // Target

	found := s.findTwoStringKite()
	if !found {
		t.Errorf("2-String Kite should have been found")
	}

	if p.Get(5, 4).HasCandidate(val) {
		t.Errorf("Candidate 1 should be eliminated from (5,4)")
	}
}

func TestFindXYWing(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)

	setCandidate := func(r, c, v int) {
		p.Get(r, c).Candidates.Add(v)
		s.rows[r].Unsolved[v].Add(c)
		s.columns[c].Unsolved[v].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[v].Add(boxLoc)
	}

	// Clear all candidates
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			p.Get(r, c).Candidates.Clear()
			for v := 1; v <= 9; v++ {
				s.rows[r].Unsolved[v].Clear()
				s.columns[c].Unsolved[v].Clear()
				s.boxes[p.Get(r, c).Box()].Unsolved[v].Clear()
			}
		}
	}

	// XY-Wing Setup
	// Pivot: (4,4) with candidates {1, 2}
	setCandidate(4, 4, 1)
	setCandidate(4, 4, 2)

	// Pincer 1: (4,0) with candidates {1, 3} (Same Row)
	setCandidate(4, 0, 1)
	setCandidate(4, 0, 3)

	// Pincer 2: (0,4) with candidates {2, 3} (Same Col)
	setCandidate(0, 4, 2)
	setCandidate(0, 4, 3)

	// Target: (0,0) with candidate {3} (Sees both pincers)
	setCandidate(0, 0, 3)

	found := s.findXYWings()
	if !found {
		t.Errorf("XY-Wing technique should have been found")
	}

	if p.Get(0, 0).HasCandidate(3) {
		t.Errorf("Target (0,0) should have eliminated candidate 3")
	}
}

func TestFindXYZWing(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)

	setCandidate := func(r, c, v int) {
		p.Get(r, c).Candidates.Add(v)
		s.rows[r].Unsolved[v].Add(c)
		s.columns[c].Unsolved[v].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[v].Add(boxLoc)
	}

	// Clear all candidates
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			p.Get(r, c).Candidates.Clear()
			for v := 1; v <= 9; v++ {
				s.rows[r].Unsolved[v].Clear()
				s.columns[c].Unsolved[v].Clear()
				s.boxes[p.Get(r, c).Box()].Unsolved[v].Clear()
			}
		}
	}

	// XYZ-Wing Setup
	// Pivot: (4,4) with candidates {1, 2, 3}
	setCandidate(4, 4, 1)
	setCandidate(4, 4, 2)
	setCandidate(4, 4, 3)

	// Pincer 1: (4,5) with candidates {1, 3} (Same Row & Box)
	setCandidate(4, 5, 1)
	setCandidate(4, 5, 3)

	// Pincer 2: (2,4) with candidates {2, 3} (Same Col, DIFFERENT Box)
	// (2,4) is in Box 1 (Row 2, Col 4)
	setCandidate(2, 4, 2)
	setCandidate(2, 4, 3)

	// Target: (5,4) with candidate {3} (Sees all three)
	// Sees (4,4) via Col 4 / Box 4
	// Sees (4,5) via Box 4
	// Sees (2,4) via Col 4
	setCandidate(5, 4, 3)

	found := s.findXYZWings()
	if !found {
		t.Errorf("XYZ-Wing technique should have been found")
	}

	if p.Get(5, 4).HasCandidate(3) {
		t.Errorf("Target (5,4) should have eliminated candidate 3")
	}
}
