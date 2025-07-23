package solver

import (
	"fmt"

	"github.com/kpitt/sudoku/internal/board"
	"github.com/kpitt/sudoku/internal/set"
)

// ***** IMPORTANT NOTE *****
// If you are combining multiple boolean checks and you want them all to run
// regardless of whether other checks succeed or fail, then each check **MUST**
// come before the first `||` operator in the combining expression.  This
// implies that a separate combining statement is needed for each check, and
// the accumulator variable must come **AFTER** the check expression.  The OR
// operator (`||`) will short-circuit after the first `true` expression, so any
// check that follows an OR operator might not run.

// findHiddenSingles locks the value of any cells that match the "Hidden Single"
// pattern.  A "Hidden Single" is the only cell that contains a particular
// candidate in its row, column, or house.
func (s *Solver) findHiddenSingles() bool {
	printChecking("Hidden Single")
	found := false
	for i := range 9 {
		found = s.checkHiddenSinglesForGroup(s.rowGroups[i]) || found
		found = s.checkHiddenSinglesForGroup(s.colGroups[i]) || found
		found = s.checkHiddenSinglesForGroup(s.houseGroups[i]) || found
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
		found = s.checkNakedPairsForGroup(s.rowGroups[i]) || found
		found = s.checkNakedPairsForGroup(s.colGroups[i]) || found
		found = s.checkNakedPairsForGroup(s.houseGroups[i]) || found
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
			found = s.eliminateFromOtherLocs(g, valueSet, locSet, "Naked Pair") || found
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
	printChecking("Locked Candidate")
	found := false
	for i := range 9 {
		// We only need to check rows and columns for Locked Candidates.
		found = s.checkLockedCandidatesForRowCol(s.rowGroups[i]) || found
		found = s.checkLockedCandidatesForRowCol(s.colGroups[i]) || found
	}
	return found
}

func (s *Solver) checkLockedCandidatesForRowCol(g *Group) bool {
	candidates := filterMap(g.Unsolved, func(_ int8, l LocSet) bool {
		// If we have more than 3 candidates in a row or column, then they can't all
		// be in the same house.
		return l.Size() <= 3
	})

	found := false
	for val, locs := range candidates {
		valueSet := set.NewSet(val)
		cells := g.cellsFromLocs(locs.Values())
		if house, ok := g.sharedHouse(locs); ok {
			houseCells := transformSlice(cells, func(c *board.Cell) int {
				_, hr, hc := c.HouseCoordinates()
				return hr*3 + hc
			})
			locSet := set.NewSet(houseCells...)
			found = s.eliminateFromOtherLocs(
				s.houseGroups[house], valueSet, locSet, "Locked Candidate") || found
		}
	}
	return found
}

func (s *Solver) findPointingTuples() bool {
	printChecking("Pointing Tuple")
	found := false
	for i := range 9 {
		// We only need to check houses for Pointing Tuples.
		found = s.checkPointingTuplesForHouse(s.houseGroups[i]) || found
	}
	return found
}

func (s *Solver) checkPointingTuplesForHouse(g *Group) bool {
	candidates := filterMap(g.Unsolved, func(_ int8, l LocSet) bool {
		// If we have more than 3 candidates in a house, then they can't all be in the
		// same row or column.
		return l.Size() <= 3
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
			found = s.eliminateFromOtherLocs(
				s.rowGroups[row], valueSet, locSet, "Pointing Tuple") || found
		}
		if col, ok := g.sharedCol(locs); ok {
			rows := transformSlice(cells, func(c *board.Cell) int {
				return c.Row
			})
			locSet := set.NewSet(rows...)
			found = s.eliminateFromOtherLocs(
				s.colGroups[col], valueSet, locSet, "Pointing Tuple") || found
		}
	}
	return found
}

func (s *Solver) findHiddenPairs() bool {
	printChecking("Hidden Pair")
	found := false
	for i := range 9 {
		found = s.checkHiddenPairsForGroup(s.rowGroups[i]) || found
		found = s.checkHiddenPairsForGroup(s.colGroups[i]) || found
		found = s.checkHiddenPairsForGroup(s.houseGroups[i]) || found
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
			found = s.eliminateOtherValues(g, valueSet, locSet, "Hidden Pair") || found
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
	printChecking("Naked Triple")
	found := false
	for i := range 9 {
		found = s.checkNakedTriplesForGroup(s.rowGroups[i]) || found
		found = s.checkNakedTriplesForGroup(s.colGroups[i]) || found
		found = s.checkNakedTriplesForGroup(s.houseGroups[i]) || found
	}
	return found
}

func (s *Solver) checkNakedTriplesForGroup(g *Group) bool {
	found := false
	candidates := make(map[int]*set.Set[int8])
	for i, c := range g.Cells {
		// Collect a map of all locations with either 2 or 3 candidate values.
		if c.NumCandidates() == 2 || c.NumCandidates() == 3 {
			candidates[i] = c.Candidates
		}
	}
	if len(candidates) < 3 {
		// We need at least 3 candidate values to have a triple.
		return false
	}

	locs := mapKeys(candidates)
	for i := 0; i < len(locs)-2; i++ {
		for j := i + 1; j < len(locs)-1; j++ {
			for k := j + 1; k < len(locs); k++ {
				a, b, c := locs[i], locs[j], locs[k]
				valueSet := set.Union(candidates[a], candidates[b], candidates[c])
				if valueSet.Size() != 3 {
					// If the union of the location sets does not have exactly 3 elements, then
					// this is not a naked triple.
					continue
				}

				locSet := set.NewSet(a, b, c)
				found = s.eliminateFromOtherLocs(
					g, valueSet, locSet, "Naked Triple") || found
			}
		}
	}

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
		found = s.checkHiddenTriplesForGroup(s.rowGroups[i]) || found
		found = s.checkHiddenTriplesForGroup(s.colGroups[i]) || found
		found = s.checkHiddenTriplesForGroup(s.houseGroups[i]) || found
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
				found = s.eliminateOtherValues(
					g, valueSet, locSet, "Hidden Triple") || found
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
	printChecking("Y Wing")
	// Collect a list of all cells with exactly 2 candidates.  Note that we'll
	// still need to re-check each cell later because additional candidates may
	// get removed along the way.
	b := s.board
	var candidates []*board.Cell
	for r := range 9 {
		for c := range 9 {
			if b.Cells[r][c].NumCandidates() == 2 {
				candidates = append(candidates, b.Cells[r][c])
			}
		}
	}
	if len(candidates) < 3 {
		// A Y Wing requires a base cell and 2 wing cells, so we need at least
		// 3 candidates.
		return false
	}

	found := false
	// Try each candidate as the base cell, checking it against all of the other
	// candidates.
	for _, base := range candidates {
		if base.NumCandidates() == 2 {
			found = s.checkYWingsForCell(base, candidates) || found
		}
	}
	return found
}

func (s *Solver) checkYWingsForCell(
	base *board.Cell, candidates []*board.Cell,
) bool {
	// Get the base x and y values.
	values := base.CandidateValues()
	x, y := values[0], values[1]

	// Find the candidate cells that can be seen by the base cell and have either
	// x or y as a candidate, but not both.  Collect the cells into separate lists
	// for cells that have x but not y and cells that have y but not x.
	var xCells, yCells []*board.Cell
	for _, cell := range candidates {
		if cell.SameCell(base) || cell.NumCandidates() != 2 || !seesCell(cell, base) {
			continue
		}
		if cell.HasCandidate(x) && !cell.HasCandidate(y) {
			xCells = append(xCells, cell)
		} else if !cell.HasCandidate(x) && cell.HasCandidate(y) {
			yCells = append(yCells, cell)
		}
	}
	if len(xCells) == 0 || len(yCells) == 0 {
		// We need at least one candidate cell for each value to have a Y Wing.
		return false
	}

	found := false
	// Check each of the x-cells against each of the y-cells to see if they share
	// a common 3rd value z.
	for _, xc := range xCells {
		if xc.NumCandidates() != 2 {
			continue
		}
		cellVals := xc.CandidateValues()
		// Set z to the non-x value in the candidate cell.
		z := cellVals[0]
		if z == x {
			z = cellVals[1]
		}
		// Look for a y-cell that also contains z and is not visible from the x-cell.
		for _, yc := range yCells {
			if !yc.HasCandidate(z) || seesCell(xc, yc) {
				continue
			}
			found = s.eliminateYWingCells(z, xc, yc) || found
		}
	}
	return found
}

// eliminateYWingCells removes candidate value z from all cells that see both
// xCell and yCell.  This assumes that xCell and yCell cannot see each other.
func (s *Solver) eliminateYWingCells(z int8, xCell, yCell *board.Cell) bool {
	seesYCell := func(cell *board.Cell) bool {
		return seesCell(cell, yCell)
	}
	removeZs := func(g *Group) bool {
		// Find candidate locations for value z in group g, which is assumed to be a
		// group that contains xCell.
		if locs, ok := g.Unsolved[z]; ok {
			// Select only the cells that also see yCell.
			cells := g.cellsFromLocs(locs.Values())
			cells = filterSlice(cells, seesYCell)
			for _, zCell := range cells {
				printEliminate("Y Wing", zCell.Row, zCell.Col, z)
				s.removeCellCandidate(zCell.Row, zCell.Col, z)
			}
			// Return true if we found any candidates to remove.
			return len(cells) != 0
		}
		return false
	}

	found := removeZs(s.rowGroups[xCell.Row])
	found = removeZs(s.colGroups[xCell.Col]) || found
	found = removeZs(s.houseGroups[xCell.House()]) || found
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
		found = s.checkHiddenQuadruplesForGroup(s.rowGroups[i]) || found
		found = s.checkHiddenQuadruplesForGroup(s.colGroups[i]) || found
		found = s.checkHiddenQuadruplesForGroup(s.houseGroups[i]) || found
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
					found = s.eliminateOtherValues(
						g, valueSet, locSet, "Hidden Quadruple") || found
				}
			}
		}
	}

	return found
}
