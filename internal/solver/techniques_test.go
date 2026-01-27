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
		p.Grid[r][c].Candidates.Add(val)
		s.rows[r].Unsolved[val].Add(c)
		s.columns[c].Unsolved[val].Add(r)
		_, boxLoc := getBoxLoc(r, c)
		s.boxes[p.Grid[r][c].Box()].Unsolved[val].Add(boxLoc)
	}

	// Prepare state:
	// Clear all candidates for val 1 first
	for r := 0; r < 9; r++ {
		s.rows[r].Unsolved[val].Clear()
		for c := 0; c < 9; c++ {
			p.Grid[r][c].RemoveCandidate(val)
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
	if p.Grid[3][7].HasCandidate(val) {
		t.Errorf("Target (3,7) should have eliminated candidate %d", val)
	}
}
