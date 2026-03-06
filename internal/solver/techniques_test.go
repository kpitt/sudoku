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

func TestFindXWing(t *testing.T) {
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

	// X-Wing Setup (Rows as Base)
	// R1C1, R1C8
	// R4C1, R4C8
	setCandidate(1, 1)
	setCandidate(1, 8)
	setCandidate(4, 1)
	setCandidate(4, 8)

	// Target to eliminate (Col 1, not in base rows)
	setCandidate(0, 1)
	// Add noise to Row 0 so it has > 2 candidates, preventing it from being
	// selected as a base line for X-Wing (Size 2).
	setCandidate(0, 2)
	setCandidate(0, 3)

	found := s.findXWings()
	if !found {
		t.Errorf("X-Wing technique should have been found")
	}

	if p.Get(0, 1).HasCandidate(val) {
		t.Errorf("Target (0,1) should have eliminated candidate %d", val)
	}
}

func TestFindSwordfish(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)
	val := 2

	setCandidate := func(r, c int) {
		p.Get(r, c).Candidates.Add(val)
		s.rows[r].Unsolved[val].Add(c)
		s.columns[c].Unsolved[val].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[val].Add(boxLoc)
	}

	// Clear candidates for val 2
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

	// Swordfish Setup (Rows as Base)
	// R1: C1, C4
	// R4: C4, C7
	// R7: C1, C7
	setCandidate(1, 1)
	setCandidate(1, 4)
	setCandidate(4, 4)
	setCandidate(4, 7)
	setCandidate(7, 1)
	setCandidate(7, 7)

	// Target to eliminate (Col 1, Row 0)
	setCandidate(0, 1)
	// Add noise to Row 0 so it has > 3 candidates, preventing it from being
	// selected as a base line for Swordfish (Size 3).
	setCandidate(0, 2)
	setCandidate(0, 3)
	setCandidate(0, 5)

	found := s.findSwordfish()
	if !found {
		t.Errorf("Swordfish technique should have been found")
	}

	if p.Get(0, 1).HasCandidate(val) {
		t.Errorf("Target (0,1) should have eliminated candidate %d", val)
	}
}

func TestFindJellyfish(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)
	val := 3

	setCandidate := func(r, c int) {
		p.Get(r, c).Candidates.Add(val)
		s.rows[r].Unsolved[val].Add(c)
		s.columns[c].Unsolved[val].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[val].Add(boxLoc)
	}

	// Clear candidates for val 3
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

	// Jellyfish Setup (Rows as Base)
	// R1: C1, C2, C4, C5
	// R2: C1, C2, C4, C5
	// R4: C1, C2, C4, C5
	// R5: C1, C2, C4, C5
	rows := []int{1, 2, 4, 5}
	cols := []int{1, 2, 4, 5}

	for _, r := range rows {
		for _, c := range cols {
			setCandidate(r, c)
		}
	}

	// Target to eliminate (Col 1, Row 0)
	setCandidate(0, 1)
	// Add noise to Row 0 so it has > 4 candidates, preventing it from being
	// selected as a base line for Jellyfish (Size 4).
	setCandidate(0, 2)
	setCandidate(0, 3)
	setCandidate(0, 6)
	setCandidate(0, 7)
	setCandidate(0, 8)

	found := s.findJellyfish()
	if !found {
		t.Errorf("Jellyfish technique should have been found")
	}

	if p.Get(0, 1).HasCandidate(val) {
		t.Errorf("Target (0,1) should have eliminated candidate %d", val)
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

func TestFindNakedPairs(t *testing.T) {
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

	// Naked Pair in Row 0: Cells (0,0) and (0,1) with candidates {1, 2}
	setCandidate(0, 0, 1)
	setCandidate(0, 0, 2)
	setCandidate(0, 1, 1)
	setCandidate(0, 1, 2)

	// Other cells in Row 0 that have candidates 1 and 2
	setCandidate(0, 2, 1) // Target to eliminate
	setCandidate(0, 3, 2) // Target to eliminate

	// Add other candidates so these cells aren't empty
	setCandidate(0, 2, 3)
	setCandidate(0, 3, 4)

	found := s.findNakedPairs()
	if !found {
		t.Errorf("Naked Pair technique should have been found")
	}

	if p.Get(0, 2).HasCandidate(1) {
		t.Errorf("Target (0,2) should have eliminated candidate 1")
	}
	if p.Get(0, 3).HasCandidate(2) {
		t.Errorf("Target (0,3) should have eliminated candidate 2")
	}
}

func TestFindNakedTriples(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)

	setCandidate := func(r, c, v int) {
		p.Get(r, c).Candidates.Add(v)
		s.rows[r].Unsolved[v].Add(c)
		s.columns[c].Unsolved[v].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[v].Add(boxLoc)
	}

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

	// Naked Triple: {1,2}, {2,3}, {1,3}
	setCandidate(0, 0, 1)
	setCandidate(0, 0, 2)
	setCandidate(0, 1, 2)
	setCandidate(0, 1, 3)
	setCandidate(0, 2, 1)
	setCandidate(0, 2, 3)

	setCandidate(0, 4, 1) // Target
	setCandidate(0, 4, 2) // Target
	setCandidate(0, 4, 4) // Noise

	found := s.findNakedTriples()
	if !found {
		t.Errorf("Naked Triple technique should have been found")
	}

	if p.Get(0, 4).HasCandidate(1) || p.Get(0, 4).HasCandidate(2) {
		t.Errorf("Target (0,4) should have eliminated candidates 1 and 2")
	}
}

func TestFindNakedQuadruples(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)

	setCandidate := func(r, c, v int) {
		p.Get(r, c).Candidates.Add(v)
		s.rows[r].Unsolved[v].Add(c)
		s.columns[c].Unsolved[v].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[v].Add(boxLoc)
	}

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

	setCandidate(0, 0, 1)
	setCandidate(0, 0, 2)
	setCandidate(0, 1, 2)
	setCandidate(0, 1, 3)
	setCandidate(0, 2, 3)
	setCandidate(0, 2, 4)
	setCandidate(0, 3, 1)
	setCandidate(0, 3, 4)

	setCandidate(0, 5, 1) // Target
	setCandidate(0, 5, 5) // Noise

	found := s.findNakedQuadruples()
	if !found {
		t.Errorf("Naked Quadruple technique should have been found")
	}

	if p.Get(0, 5).HasCandidate(1) {
		t.Errorf("Target (0,5) should have eliminated candidate 1")
	}
}

func TestFindHiddenPairs(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)

	setCandidate := func(r, c, v int) {
		p.Get(r, c).Candidates.Add(v)
		s.rows[r].Unsolved[v].Add(c)
		s.columns[c].Unsolved[v].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[v].Add(boxLoc)
	}

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

	setCandidate(0, 0, 1)
	setCandidate(0, 0, 2)
	setCandidate(0, 1, 1)
	setCandidate(0, 1, 2)

	setCandidate(0, 0, 3) // Target
	setCandidate(0, 1, 4) // Target

	setCandidate(0, 2, 3)
	setCandidate(0, 3, 4)
	setCandidate(0, 4, 5)

	found := s.findHiddenPairs()
	if !found {
		t.Errorf("Hidden Pair technique should have been found")
	}

	if p.Get(0, 0).HasCandidate(3) {
		t.Errorf("Target (0,0) should have eliminated candidate 3")
	}
	if p.Get(0, 1).HasCandidate(4) {
		t.Errorf("Target (0,1) should have eliminated candidate 4")
	}
}

func TestFindHiddenTriples(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)

	setCandidate := func(r, c, v int) {
		p.Get(r, c).Candidates.Add(v)
		s.rows[r].Unsolved[v].Add(c)
		s.columns[c].Unsolved[v].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[v].Add(boxLoc)
	}

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

	setCandidate(0, 0, 1)
	setCandidate(0, 0, 2)
	setCandidate(0, 1, 2)
	setCandidate(0, 1, 3)
	setCandidate(0, 2, 1)
	setCandidate(0, 2, 3)

	setCandidate(0, 0, 4) // Target
	setCandidate(0, 1, 5) // Target
	setCandidate(0, 2, 6) // Target

	setCandidate(0, 3, 4)
	setCandidate(0, 3, 5)
	setCandidate(0, 4, 5)
	setCandidate(0, 4, 6)

	found := s.findHiddenTriples()
	if !found {
		t.Errorf("Hidden Triple technique should have been found")
	}

	if p.Get(0, 0).HasCandidate(4) {
		t.Errorf("Target (0,0) should have eliminated candidate 4")
	}
	if p.Get(0, 1).HasCandidate(5) {
		t.Errorf("Target (0,1) should have eliminated candidate 5")
	}
	if p.Get(0, 2).HasCandidate(6) {
		t.Errorf("Target (0,2) should have eliminated candidate 6")
	}
}

func TestFindHiddenQuadruples(t *testing.T) {
	p := puzzle.NewPuzzle()
	s := NewSolver(p, nil)

	setCandidate := func(r, c, v int) {
		p.Get(r, c).Candidates.Add(v)
		s.rows[r].Unsolved[v].Add(c)
		s.columns[c].Unsolved[v].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Get(r, c).Box()].Unsolved[v].Add(boxLoc)
	}

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

	setCandidate(0, 0, 1)
	setCandidate(0, 0, 2)
	setCandidate(0, 1, 2)
	setCandidate(0, 1, 3)
	setCandidate(0, 2, 3)
	setCandidate(0, 2, 4)
	setCandidate(0, 3, 1)
	setCandidate(0, 3, 4)

	setCandidate(0, 0, 5) // Target
	setCandidate(0, 1, 6) // Target

	setCandidate(0, 4, 5)
	setCandidate(0, 4, 6)
	setCandidate(0, 5, 7)
	setCandidate(0, 5, 8)

	found := s.findHiddenQuadruples()
	if !found {
		t.Errorf("Hidden Quadruple technique should have been found")
	}

	if p.Get(0, 0).HasCandidate(5) {
		t.Errorf("Target (0,0) should have eliminated candidate 5")
	}
	if p.Get(0, 1).HasCandidate(6) {
		t.Errorf("Target (0,1) should have eliminated candidate 6")
	}
}
