package solver

import (
	"fmt"

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
	found := false
	candidates := g.UnsolvedWhere(func(_ int8, l LocSet) bool {
		return l.Size() == 2
	})
	if len(candidates) < 2 {
		// We need at least 2 candidate values to have a pair.
		return false
	}

	values := mapKeys(candidates)
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			x, y := values[i], values[j]
			locSet := set.Union(candidates[x], candidates[y])
			if locSet.Size() != 2 {
				// If the union of the location sets does not have exactly 2 elements, then
				// this is not a hidden pair.
				continue
			}

			valueSet := set.NewSet(x, y)
			found = found || s.eliminateHiddenTupleCandidates(
				g, valueSet, locSet, "Hidden Pair")
		}
	}

	return found
}

func (s *Solver) eliminateHiddenTupleCandidates(
	g *Group, values *set.Set[int8], locs *set.Set[int], basePattern string,
) bool {
	pattern := fmt.Sprintf("%s (%s)", basePattern, g.GroupType)
	found := false
	for _, l := range locs.Values() {
		c := g.Cells[l]
		for _, v := range c.Candidates() {
			if !values.Contains(v) {
				printEliminate(pattern, c.Row, c.Col, v)
				s.removeCellCandidate(c.Row, c.Col, v)
				found = true
			}
		}
	}
	return found
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
	printChecking("Hidden Triple")
	found := false
	for i := range 9 {
		found = found ||
			s.checkHiddenTriplesForGroup(s.rowGroups[i]) ||
			s.checkHiddenTriplesForGroup(s.colGroups[i]) ||
			s.checkHiddenTriplesForGroup(s.houseGroups[i])
	}
	return found
}

func (s *Solver) checkHiddenTriplesForGroup(g *Group) bool {
	found := false
	candidates := g.UnsolvedWhere(func(_ int8, l LocSet) bool {
		return l.Size() == 2 || l.Size() == 3
	})
	if len(candidates) < 3 {
		// We need at least 3 candidate values to have a triple.
		return false
	}

	values := mapKeys(candidates)
	for i := 0; i < len(values)-2; i++ {
		for j := i + 1; j < len(values)-1; j++ {
			for k := j + 1; k < len(values); k++ {
				x, y, z := values[i], values[j], values[k]
				locSet := set.Union(candidates[x], candidates[y], candidates[z])
				if locSet.Size() != 3 {
					// If the union of the location sets does not have exactly 3 elements, then
					// this is not a hidden triple.
					continue
				}

				valueSet := set.NewSet(x, y, z)
				found = found || s.eliminateHiddenTupleCandidates(
					g, valueSet, locSet, "Hidden Triple")
			}
		}
	}

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
