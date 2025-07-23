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
	for val, locs := range g.Unsolved {
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
	printChecking("Naked Pair")
	found := false
	for i := range 9 {
		found = found ||
			s.checkNakedPairsForGroup(s.rowGroups[i]) ||
			s.checkNakedPairsForGroup(s.colGroups[i]) ||
			s.checkNakedPairsForGroup(s.houseGroups[i])
	}
	return found
}

func (s *Solver) checkNakedPairsForGroup(g *Group) bool {
	found := false
	candidates := make(map[int]*set.Set[int8])
	for i, c := range g.Cells {
		// Collect a map of all locations with exactly 2 candidate values.
		if c.NumCandidates() == 2 {
			candidates[i] = c.Candidates
		}
	}
	if len(candidates) < 2 {
		// We need at least 2 candidate values to have a pair.
		return false
	}

	locs := mapKeys(candidates)
	for i := 0; i < len(locs)-1; i++ {
		for j := i + 1; j < len(locs); j++ {
			a, b := locs[i], locs[j]
			valueSet := set.Union(candidates[a], candidates[b])
			if valueSet.Size() != 2 {
				// If the union of the location sets does not have exactly 2 elements, then
				// this is not a naked pair.
				continue
			}

			locSet := set.NewSet(a, b)
			found = found || s.eliminateFromOtherLocs(
				g, valueSet, locSet, "Naked Pair")
		}
	}

	return found
}

// eliminateFromOtherLocs removes the candidates listed in values from all
// cells that are not listed in locs.
func (s *Solver) eliminateFromOtherLocs(
	g *Group, values *set.Set[int8], locs LocSet, basePattern string,
) bool {
	pattern := fmt.Sprintf("%s (%s)", basePattern, g.GroupType)
	found := false
	for l := range 9 {
		if locs.Contains(l) {
			continue
		}
		c := g.Cells[l]
		for _, v := range values.Values() {
			if c.HasCandidate(v) {
				printEliminate(pattern, c.Row, c.Col, v)
				s.removeCellCandidate(c.Row, c.Col, v)
				found = true
			}
		}
	}
	return found
}

func (s *Solver) findLockedCandidates() bool {
	found := false
	return found
}

func (s *Solver) findPointingTuples() bool {
	printChecking("Pointing Tuple")
	found := false
	for i := range 9 {
		found = found || s.checkPointingTuplesForHouse(s.houseGroups[i])
	}
	return found
}

func (s *Solver) checkPointingTuplesForHouse(g *Group) bool {
	candidates := filterMap(g.Unsolved, func(_ int8, l LocSet) bool {
		return l.Size() == 2 || l.Size() == 3
	})

	found := false
	for val, locs := range candidates {
		valueSet := set.NewSet(val)
		cells := g.cellsFromLocs(locs.Values())
		if row, ok := g.sharedRow(locs); ok {
			cols := transformSlice(cells, func(c *board.Cell) int {
				return c.Col
			})
			locSet := set.NewSet(cols...)
			found = found || s.eliminateFromOtherLocs(
				s.rowGroups[row], valueSet, locSet, "Pointing Tuple")
		}
		if col, ok := g.sharedCol(locs); ok {
			rows := transformSlice(cells, func(c *board.Cell) int {
				return c.Row
			})
			locSet := set.NewSet(rows...)
			found = found || s.eliminateFromOtherLocs(
				s.colGroups[col], valueSet, locSet, "Pointing Tuple")
		}
	}
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
	candidates := filterMap(g.Unsolved, func(_ int8, l LocSet) bool {
		return l.Size() == 2
	})
	if len(candidates) < 2 {
		// We need at least 2 candidate values to have a pair.
		return false
	}

	found := false
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
			found = found || s.eliminateOtherValues(
				g, valueSet, locSet, "Hidden Pair")
		}
	}
	return found
}

// eliminateOtherValues removes candidates that are not listed in values
// from the cells in locs.
func (s *Solver) eliminateOtherValues(
	g *Group, values *set.Set[int8], locs LocSet, basePattern string,
) bool {
	pattern := fmt.Sprintf("%s (%s)", basePattern, g.GroupType)
	found := false
	for _, l := range locs.Values() {
		c := g.Cells[l]
		for _, v := range c.CandidateValues() {
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
	candidates := filterMap(g.Unsolved, func(_ int8, l LocSet) bool {
		return l.Size() == 2 || l.Size() == 3
	})
	if len(candidates) < 3 {
		// We need at least 3 candidate values to have a triple.
		return false
	}

	found := false
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
				found = found || s.eliminateOtherValues(
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
	printChecking("Hidden Quadruple")
	found := false
	for i := range 9 {
		found = found ||
			s.checkHiddenQuadruplesForGroup(s.rowGroups[i]) ||
			s.checkHiddenQuadruplesForGroup(s.colGroups[i]) ||
			s.checkHiddenQuadruplesForGroup(s.houseGroups[i])
	}
	return found
}

func (s *Solver) checkHiddenQuadruplesForGroup(g *Group) bool {
	found := false
	candidates := filterMap(g.Unsolved, func(_ int8, l LocSet) bool {
		return l.Size() == 2 || l.Size() == 3 || l.Size() == 4
	})
	if len(candidates) < 4 {
		// We need at least 4 candidate values to have a quadruple.
		return false
	}

	values := mapKeys(candidates)
	for i := 0; i < len(values)-3; i++ {
		for j := i + 1; j < len(values)-2; j++ {
			for k := j + 1; k < len(values)-1; k++ {
				for n := k + 1; n < len(values); n++ {
					w, x, y, z := values[i], values[j], values[k], values[n]
					locSet := set.Union(
						candidates[w],
						candidates[x],
						candidates[y],
						candidates[z],
					)
					if locSet.Size() != 4 {
						// If the union of the location sets does not have exactly 4 elements, then
						// this is not a hidden quadruple.
						continue
					}

					valueSet := set.NewSet(w, x, y, z)
					found = found || s.eliminateOtherValues(
						g, valueSet, locSet, "Hidden Quadruple")
				}
			}
		}
	}

	return found
}
