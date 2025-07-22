package solver

import (
	"fmt"

	"github.com/kpitt/sudoku/internal/board"
	"github.com/kpitt/sudoku/internal/set"
)

// findHiddenSingles locks the value of any cells that match the "Hidden Single"
// pattern.  A "Hidden Single" is the only cell that contains a particular
// candidate in its row, column, or house.
func (s *Solver) findHiddenSingles() bool {
	printChecking("Hidden Single")
	found := false
	for i := range 9 {
		found = found ||
			s.checkHiddenSinglesForGroup(s.rowGroups[i]) ||
			s.checkHiddenSinglesForGroup(s.colGroups[i]) ||
			s.checkHiddenSinglesForGroup(s.houseGroups[i])
	}
	return found
}

func (s *Solver) checkHiddenSinglesForGroup(g *Group) bool {
	pattern := fmt.Sprintf("Hidden Single (%s)", g.GroupType)
	found := false
	for val, locs := range g.Unsolved() {
		if locs.Size() == 1 {
			index := locs.Values()[0]
			cell := g.Cells[index]
			s.LockValue(cell.Row, cell.Col, val, pattern)
			found = true
		}
	}
	return found
}

func (s *Solver) findNakedPairs() bool {
	found := false
	return found
}

func (s *Solver) findLockedCandidates() bool {
	found := false
	return found
}

func (s *Solver) findPointingTuples() bool {
	found := false
	return found
}

func (s *Solver) findHiddenPairs() bool {
	printChecking("Hidden Pair")
	found := false
	for i := range 9 {
		found = found ||
			s.checkHiddenPairsForGroup(s.rowGroups[i]) ||
			s.checkHiddenPairsForGroup(s.colGroups[i]) ||
			s.checkHiddenPairsForGroup(s.houseGroups[i])
	}
	return found
}

func (s *Solver) checkHiddenPairsForGroup(g *Group) bool {
	pattern := fmt.Sprintf("Hidden Pair (%s)", g.GroupType)
	found := false
	candidates := g.UnsolvedWhere(func(_ int8, l LocSet) bool {
		return l.Size() == 2
	})
	checked := set.NewSet[int8]()
	// Check all values that have exactly 2 possible locations.
	for valA, locsA := range candidates {
		checked.Add(valA)
		// Find a different value that has not been checked, and also has exactly 2
		// possible locations.
		for valB, locsB := range candidates {
			// Skip any candidate that has already been checked (including the current
			// value of A).
			if checked.Contains(valB) {
				continue
			}

			locsAB := set.Union(locsA, locsB)
			if locsAB.Size() != 2 {
				// If the union of the location sets does not have exactly 2 elements, then
				// this is not a hidden pair.
				continue
			}

			locs := locsAB.Values()
			cellA, cellB := g.Cells[locs[0]], g.Cells[locs[1]]
			if cellA.NumCandidates() == 2 && cellB.NumCandidates() == 2 {
				// Neither cell has any additional candidates, so this is not a hidden pair.
				continue
			}

			s.eliminateHiddenPairCandidates(cellA, valA, valB, pattern)
			s.eliminateHiddenPairCandidates(cellB, valA, valB, pattern)
			found = true
		}
	}
	return found
}

func (s *Solver) eliminateHiddenPairCandidates(c *board.Cell, x, y int8, pattern string) {
	for _, v := range c.Candidates() {
		if v != x && v != y {
			printEliminate(pattern, c.Row, c.Col, v)
			s.removeCellCandidate(c.Row, c.Col, v)
		}
	}
}

func (s *Solver) findNakedTriples() bool {
	found := false
	return found
}

func (s *Solver) findXWings() bool {
	found := false
	return found
}

func (s *Solver) findHiddenTriples() bool {
	found := false
	return found
}

func (s *Solver) findNakedQuadruples() bool {
	found := false
	return found
}

func (s *Solver) findYWings() bool {
	found := false
	return found
}

func (s *Solver) findAvoidableRectangles() bool {
	found := false
	return found
}

func (s *Solver) findXYZWings() bool {
	found := false
	return found
}

func (s *Solver) findHiddenQuadruples() bool {
	found := false
	return found
}
