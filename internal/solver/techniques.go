package solver

import (
	"fmt"

	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/set"
)

// ***** IMPORTANT NOTE *****
//
// When processing a check against a set of viable candidates, _always_
// short-circuit the remaining checks after making a change that could
// invalidate the remaining candidates.  Checks should not be combined in a
// single pass unless the checks are completely independent.  If it isn't
// clear whether or not the checks are independent, go ahead and short-ciruit,
// and any additional candidates will get checked in the next solver pass.
//
// For checks that should _not_ short-circuit, be careful when using OR
// expressions to combine the results.  The safest approach is to use an
// accumulator variable and the pattern `found = check(...) || found` for each
// check.  The `||` operator will short-circuit after the first term that
// evalues to `true`, so only the first term is guaranteed to be evaluated.

// findHiddenSingles places the value of any cells that match the "Hidden
// Single" pattern.  A "Hidden Single" is the only cell that contains a
// particular candidate in its house.
func (s *Solver) findHiddenSingles() bool {
	printChecking("Hidden Single")
	found := false
	for i := range 9 {
		found = s.checkHiddenSinglesForHouse(s.rows[i]) || found
		found = s.checkHiddenSinglesForHouse(s.columns[i]) || found
		found = s.checkHiddenSinglesForHouse(s.boxes[i]) || found
	}
	return found
}

func (s *Solver) checkHiddenSinglesForHouse(h *House) bool {
	name := fmt.Sprintf("Hidden Single (%s)", h.Type)
	for val, locs := range h.Unsolved {
		if locs.Size() == 1 {
			index := locs.Values()[0]
			cell := h.Cells[index]
			s.PlaceValue(cell.Row, cell.Col, val, name)
			return true
		}
	}
	return false
}

func (s *Solver) findNakedPairs() bool {
	printChecking("Naked Pair")
	found := false
	for i := range 9 {
		found = s.checkNakedPairsForHouse(s.rows[i]) || found
		found = s.checkNakedPairsForHouse(s.columns[i]) || found
		found = s.checkNakedPairsForHouse(s.boxes[i]) || found
	}
	return found
}

func (s *Solver) checkNakedPairsForHouse(h *House) bool {
	values := make(LocValMap)
	for i, c := range h.Cells {
		// Collect a map of all locations with exactly 2 candidate values.
		if c.NumCandidates() == 2 {
			values[i] = c.Candidates
		}
	}
	if len(values) < 2 {
		// We need at least 2 candidate values to have a pair.
		return false
	}

	locs := mapKeys(values)
	for i := 0; i < len(locs)-1; i++ {
		for j := i + 1; j < len(locs); j++ {
			a, b := locs[i], locs[j]
			valueSet := set.Union(values[a], values[b])
			if valueSet.Size() != 2 {
				// If the union of the location sets does not have exactly 2 elements, then
				// this is not a naked pair.
				continue
			}

			locSet := set.NewSet(a, b)
			ss := NewSolutionStep(fmt.Sprintf("Naked Pair (%s)", h.Type))
			if s.eliminateFromOtherLocs(h, valueSet, locSet, ss) {
				printStep(ss)
				s.applyStepCandidates(ss)
				return true
			}
		}
	}

	return false
}

// eliminateFromOtherLocs removes the candidates listed in values from all
// cells that are not listed in locs.
func (s *Solver) eliminateFromOtherLocs(
	h *House, values ValSet, locs LocSet, ss *SolutionStep,
) bool {
	found := false
	for l := range 9 {
		if locs.Contains(l) {
			continue
		}
		c := h.Cells[l]
		for _, v := range values.Values() {
			if c.HasCandidate(v) {
				ss.deleteCandidate(c.Row, c.Col, v)
				found = true
			}
		}
	}

	return found
}

// eliminateFromOtherLocsMulti removes the candidates listed in values from all
// cells from each house in houses whose index is not listed in locs.  Returns
// true if at least one candidate was eliminated.
func (s *Solver) eliminateFromOtherLocsMulti(
	houses []*House, values ValSet, locs LocSet, ss *SolutionStep,
) bool {
	updated := false
	for _, g := range houses {
		updated = s.eliminateFromOtherLocs(g, values, locs, ss) || updated
	}

	return updated
}

func (s *Solver) findLockedCandidates() bool {
	printChecking("Locked Candidate")
	found := false
	for i := range 9 {
		// We only need to check rows and columns for Locked Candidates.
		found = s.checkLockedCandidatesForLine(s.rows[i]) || found
		found = s.checkLockedCandidatesForLine(s.columns[i]) || found
	}
	return found
}

func (s *Solver) checkLockedCandidatesForLine(line *House) bool {
	candidates := filterMap(line.Unsolved, func(_ int, l LocSet) bool {
		// If we have more than 3 candidates in a line, then they can't all be
		// in the same box.
		return l.Size() <= 3
	})

	for val, locs := range candidates {
		valueSet := set.NewSet(val)
		cells := line.cellsFromLocs(locs.Values())
		if box, ok := line.sharedBox(locs); ok {
			boxCells := transformSlice(cells, func(c *puzzle.Cell) int {
				_, index := c.BoxCoordinates()
				return index
			})
			locSet := set.NewSet(boxCells...)
			ss := NewSolutionStep(fmt.Sprintf("Locked Candidate (%s)", line.Type))
			if s.eliminateFromOtherLocs(s.boxes[box], valueSet, locSet, ss) {
				printStep(ss)
				s.applyStepCandidates(ss)
				return true
			}
		}
	}

	return false
}

func (s *Solver) findPointingTuples() bool {
	printChecking("Pointing Tuple")
	found := false
	for i := range 9 {
		// We only need to check boxes for Pointing Tuples.
		found = s.checkPointingTuplesForBox(s.boxes[i]) || found
	}
	return found
}

func (s *Solver) checkPointingTuplesForBox(box *House) bool {
	candidates := filterMap(box.Unsolved, func(_ int, l LocSet) bool {
		// If we have more than 3 candidates in a single box, then they can't all
		// be in the same line.
		return l.Size() <= 3
	})

	for val, locs := range candidates {
		valueSet := set.NewSet(val)
		cells := box.cellsFromLocs(locs.Values())
		if row, ok := box.sharedRow(locs); ok {
			cols := transformSlice(cells, func(c *puzzle.Cell) int {
				return c.Col
			})
			locSet := set.NewSet(cols...)
			ss := NewSolutionStep("Pointing Tuple (Row)")
			if s.eliminateFromOtherLocs(s.rows[row], valueSet, locSet, ss) {
				printStep(ss)
				s.applyStepCandidates(ss)
				return true
			}
		}
		if col, ok := box.sharedCol(locs); ok {
			rows := transformSlice(cells, func(c *puzzle.Cell) int {
				return c.Row
			})
			locSet := set.NewSet(rows...)
			ss := NewSolutionStep("Pointing Tuple (Column)")
			if s.eliminateFromOtherLocs(s.columns[col], valueSet, locSet, ss) {
				printStep(ss)
				s.applyStepCandidates(ss)
				return true
			}
		}
	}

	return false
}

func (s *Solver) findHiddenPairs() bool {
	printChecking("Hidden Pair")
	found := false
	for i := range 9 {
		found = s.checkHiddenPairsForHouse(s.rows[i]) || found
		found = s.checkHiddenPairsForHouse(s.columns[i]) || found
		found = s.checkHiddenPairsForHouse(s.boxes[i]) || found
	}
	return found
}

func (s *Solver) checkHiddenPairsForHouse(h *House) bool {
	locs := filterMap(h.Unsolved, func(_ int, l LocSet) bool {
		return l.Size() == 2
	})
	if len(locs) < 2 {
		// We need at least 2 candidate values to have a pair.
		return false
	}

	values := mapKeys(locs)
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			x, y := values[i], values[j]
			locSet := set.Union(locs[x], locs[y])
			if locSet.Size() != 2 {
				// If the union of the location sets does not have exactly 2 elements, then
				// this is not a hidden pair.
				continue
			}

			valueSet := set.NewSet(x, y)
			ss := NewSolutionStep(fmt.Sprintf("Hidden Pair (%s)", h.Type))
			if s.eliminateOtherValues(h, valueSet, locSet, ss) {
				printStep(ss)
				s.applyStepCandidates(ss)
				return true
			}
		}
	}

	return false
}

// eliminateOtherValues removes candidates that are not listed in values from
// the cells in locs.
func (s *Solver) eliminateOtherValues(
	h *House, values ValSet, locs LocSet, ss *SolutionStep,
) bool {
	found := false
	for _, l := range locs.Values() {
		c := h.Cells[l]
		for _, v := range c.CandidateValues() {
			if !values.Contains(v) {
				ss.deleteCandidate(c.Row, c.Col, v)
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
		found = s.checkNakedTriplesForHouse(s.rows[i]) || found
		found = s.checkNakedTriplesForHouse(s.columns[i]) || found
		found = s.checkNakedTriplesForHouse(s.boxes[i]) || found
	}
	return found
}

func (s *Solver) checkNakedTriplesForHouse(h *House) bool {
	values := make(LocValMap)
	for i, c := range h.Cells {
		// Collect a map of all locations with either 2 or 3 candidate values.
		if c.NumCandidates() == 2 || c.NumCandidates() == 3 {
			values[i] = c.Candidates
		}
	}
	if len(values) < 3 {
		// We need at least 3 candidate values to have a triple.
		return false
	}

	locs := mapKeys(values)
	for i := 0; i < len(locs)-2; i++ {
		for j := i + 1; j < len(locs)-1; j++ {
			for k := j + 1; k < len(locs); k++ {
				a, b, c := locs[i], locs[j], locs[k]
				valueSet := set.Union(values[a], values[b], values[c])
				if valueSet.Size() != 3 {
					// If the union of the location sets does not have exactly 3 elements, then
					// this is not a naked triple.
					continue
				}

				locSet := set.NewSet(a, b, c)
				ss := NewSolutionStep(fmt.Sprintf("Naked Triple (%s)", h.Type))
				if s.eliminateFromOtherLocs(h, valueSet, locSet, ss) {
					printStep(ss)
					s.applyStepCandidates(ss)
					return true
				}
			}
		}
	}

	return false
}

func (s *Solver) findXWings() bool {
	printChecking("X-Wing")
	found := s.findXWingsInLines(s.rows, s.columns)
	found = s.findXWingsInLines(s.columns, s.rows) || found
	return found
}

func (s *Solver) findXWingsInLines(baseLines, coverLines []*House) bool {
	for i, b1 := range baseLines[:8] {
		for x, b1Locs := range b1.Unsolved {
			if b1Locs.Size() != 2 {
				// We need exactly 2 candidates to form an X-Wing.
				continue
			}

			valueSet := set.NewSet(x)
			// Check the remaining base lines for a line that also has exactly
			// 2 candidates for the same value.
			for _, b2 := range baseLines[i+1:] {
				b2Locs := b2.Unsolved[x]
				if b2Locs == nil || b2Locs.Size() != 2 {
					continue
				}
				baseLocs := set.Union(b1Locs, b2Locs)
				// If b1 and b2 form an X-Wing, then the x values must be covered
				// by the same 2 lines in the cover set.  If this is the case,
				// the union of the x locations from b1 and b2 will have exactly
				// 2 entries, and the entries will be the indexes of the cover
				// lines.
				if baseLocs.Size() != 2 {
					continue
				}

				locSet := set.NewSet(b1.Index, b2.Index)
				covers := transformSlice(baseLocs.Values(), func(y int) *House {
					return coverLines[y]
				})
				ss := NewSolutionStep("X-Wing")
				if s.eliminateFromOtherLocsMulti(covers, valueSet, locSet, ss) {
					printStep(ss)
					s.applyStepCandidates(ss)
					return true
				}
			}
		}
	}

	return false
}

func (s *Solver) findHiddenTriples() bool {
	printChecking("Hidden Triple")
	for i := range 9 {
		if s.checkHiddenTriplesForHouse(s.rows[i]) ||
			s.checkHiddenTriplesForHouse(s.columns[i]) ||
			s.checkHiddenTriplesForHouse(s.boxes[i]) {

			return true
		}
	}
	return false
}

func (s *Solver) checkHiddenTriplesForHouse(h *House) bool {
	locs := filterMap(h.Unsolved, func(_ int, l LocSet) bool {
		return l.Size() == 2 || l.Size() == 3
	})
	if len(locs) < 3 {
		// We need at least 3 candidate values to have a triple.
		return false
	}

	values := mapKeys(locs)
	for i := 0; i < len(values)-2; i++ {
		for j := i + 1; j < len(values)-1; j++ {
			for k := j + 1; k < len(values); k++ {
				x, y, z := values[i], values[j], values[k]
				locSet := set.Union(locs[x], locs[y], locs[z])
				if locSet.Size() != 3 {
					// If the union of the location sets does not have exactly 3 elements, then
					// this is not a hidden triple.
					continue
				}

				valueSet := set.NewSet(x, y, z)
				ss := NewSolutionStep(fmt.Sprintf("Hidden Triple (%s)", h.Type))
				if s.eliminateOtherValues(h, valueSet, locSet, ss) {
					printStep(ss)
					s.applyStepCandidates(ss)
					return true
				}
			}
		}
	}

	return false
}

func (s *Solver) findNakedQuadruples() bool {
	printChecking("Naked Quadruple")
	found := false
	for i := range 9 {
		found = s.checkNakedQuadruplesForHouse(s.rows[i]) || found
		found = s.checkNakedQuadruplesForHouse(s.columns[i]) || found
		found = s.checkNakedQuadruplesForHouse(s.boxes[i]) || found
	}
	return found
}

func (s *Solver) checkNakedQuadruplesForHouse(h *House) bool {
	values := make(LocValMap)
	for i, c := range h.Cells {
		// Collect a map of all locations with either 2, 3 or 4 candidate values.
		if c.NumCandidates() == 2 || c.NumCandidates() == 3 || c.NumCandidates() == 4 {
			values[i] = c.Candidates
		}
	}
	if len(values) < 4 {
		// We need at least 4 candidate values to have a quadruple.
		return false
	}

	locs := mapKeys(values)
	for i := 0; i < len(locs)-3; i++ {
		for j := i + 1; j < len(locs)-2; j++ {
			for k := j + 1; k < len(locs)-1; k++ {
				for n := k + 1; n < len(locs); n++ {
					a, b, c, d := locs[i], locs[j], locs[k], locs[n]
					valueSet := set.Union(values[a], values[b], values[c], values[d])
					if valueSet.Size() != 4 {
						// If the union of the location sets does not have exactly 4 elements, then
						// this is not a naked quadruple.
						continue
					}

					locSet := set.NewSet(a, b, c, d)
					ss := NewSolutionStep(fmt.Sprintf("Naked Quadruple (%s)", h.Type))
					if s.eliminateFromOtherLocs(h, valueSet, locSet, ss) {
						printStep(ss)
						s.applyStepCandidates(ss)
						return true
					}
				}
			}
		}
	}

	return false
}

func (s *Solver) findXYWings() bool {
	printChecking("XY-Wing")
	// Collect a list of all cells with exactly 2 candidates.
	p := s.puzzle
	var candidates []*puzzle.Cell
	for r := range 9 {
		for c := range 9 {
			if p.Grid[r][c].NumCandidates() == 2 {
				candidates = append(candidates, p.Grid[r][c])
			}
		}
	}
	if len(candidates) < 3 {
		// An XY-Wing requires a pivot cell and 2 pincer cells, so we need at
		// least 3 candidates.
		return false
	}

	// Try each candidate as the pivot cell, checking it against all of the other
	// candidates.
	for _, pivot := range candidates {
		if s.checkXYWingsForPivot(pivot, candidates) {
			return true
		}
	}
	return false
}

func (s *Solver) checkXYWingsForPivot(
	pivot *puzzle.Cell, candidates []*puzzle.Cell,
) bool {
	// Get the x and y values.
	values := pivot.CandidateValues()
	x, y := values[0], values[1]

	// Find the candidate cells that can be seen by the pivot cell and have either
	// x or y as a candidate, but not both.  Collect the cells into separate lists
	// for cells that have x but not y and cells that have y but not x.
	var xCells, yCells []*puzzle.Cell
	for _, cell := range candidates {
		if cell.SameCell(pivot) || cell.NumCandidates() != 2 || !seesCell(cell, pivot) {
			continue
		}
		if cell.HasCandidate(x) && !cell.HasCandidate(y) {
			xCells = append(xCells, cell)
		} else if !cell.HasCandidate(x) && cell.HasCandidate(y) {
			yCells = append(yCells, cell)
		}
	}
	if len(xCells) == 0 || len(yCells) == 0 {
		// We need at least one candidate cell for each value to have an XY-Wing.
		return false
	}

	// Check each of the x-cells against each of the y-cells to see if they share
	// a common 3rd value z.
	for _, xc := range xCells {
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
			ss := NewSolutionStep("XY-Wing")
			if s.eliminateXYWingCells(z, xc, yc, ss) {
				printStep(ss)
				s.applyStepCandidates(ss)
				return true
			}
		}
	}

	return false
}

// eliminateYWingCells removes candidate value z from all cells that see both
// xCell and yCell.  This assumes that xCell and yCell cannot see each other.
func (s *Solver) eliminateXYWingCells(z int, xCell, yCell *puzzle.Cell, ss *SolutionStep) bool {
	seesYCell := func(cell *puzzle.Cell) bool {
		return seesCell(cell, yCell)
	}
	removeZs := func(h *House) bool {
		// Find candidate locations for value z in house h, which is assumed to
		// be a house that contains xCell.
		if locs, ok := h.Unsolved[z]; ok {
			// Select only the cells that also see yCell.
			cells := h.cellsFromLocs(locs.Values())
			cells = filterSlice(cells, seesYCell)
			for _, zCell := range cells {
				ss.deleteCandidate(zCell.Row, zCell.Col, z)
			}
			// Return true if we found any candidates to remove.
			return len(cells) != 0
		}
		return false
	}

	found := removeZs(s.rows[xCell.Row])
	found = removeZs(s.columns[xCell.Col]) || found
	found = removeZs(s.boxes[xCell.Box()]) || found
	return found
}

func (s *Solver) findAvoidableRectangles() bool {
	found := false
	return found
}

// findXYZ searches for 3 cells that fit the "XYZ-Wing" pattern.  An XYZ-Wing
// consists of a pivot cell with 3 candidate values x,y,z, and two pincer cells
// that each see the pivot cell but don't see each other.  One pincer must have
// candidate values x,z and the other must have candidate values y,z.  One of
// these cells must have the value z, so z can be eliminated as a candidate for
// any cell that sees all three.  Note that one pincer *MUST* be in the same
// box as the pivot cell in order for it to be possible for any cell to see the
// pivot and both pincers.
func (s *Solver) findXYZWings() bool {
	printChecking("XYZ-Wing")
	// Collect a list of all cells with exactly 3 candidates.
	p := s.puzzle
	var candidates []*puzzle.Cell
	for r := range 9 {
		for c := range 9 {
			if p.Grid[r][c].NumCandidates() == 3 {
				candidates = append(candidates, p.Grid[r][c])
			}
		}
	}

	// Check each candidate as a possible pivot cell for an XYZ-Wing.
	for _, pivot := range candidates {
		if s.checkXYZWingsForPivot(pivot) {
			return true
		}
	}
	return false
}

func (s *Solver) checkXYZWingsForPivot(pivot *puzzle.Cell) bool {
	// Find cells in the same box as the pivot cell which have exactly 2
	// candidates that both appear in the pivot cell.
	box := s.boxes[pivot.Box()]
	var xzCells []*puzzle.Cell
	for _, cell := range box.Cells {
		if cell.NumCandidates() == 2 {
			// The pivot cell can't match here because it has 3 candidates.
			values := cell.CandidateValues()
			if pivot.HasCandidate(values[0]) && pivot.HasCandidate(values[1]) {
				xzCells = append(xzCells, cell)
			}
		}
	}
	if len(xzCells) == 0 {
		// No valid candidates found.
		return false
	}

	for _, xzCell := range xzCells {
		// Find the y value that does not appear in the xz-cell candidate.
		var y int
		for _, val := range pivot.CandidateValues() {
			if !xzCell.HasCandidate(val) {
				y = val
				break
			}
		}

		// Now find a cell in the same row or column as the pivot cell that has
		// exactly 2 candidate values, where one candidate is y and the other
		// is one of the candidates in xzCell.
		isYZCandidate := func(cell *puzzle.Cell) bool {
			if cell.Box() == pivot.Box() ||
				cell.NumCandidates() != 2 ||
				!cell.HasCandidate(y) {

				return false
			}
			for _, val := range cell.CandidateValues() {
				if val != y && !xzCell.HasCandidate(val) {
					return false
				}
			}
			return true
		}

		ss := NewSolutionStep("XYZ-Wing")
		r, c := pivot.Row, pivot.Col
		for _, rowCell := range s.rows[r].Cells {
			if isYZCandidate(rowCell) &&
				s.eliminateXYZWingCells(pivot, xzCell, rowCell, ss) {

				printStep(ss)
				s.applyStepCandidates(ss)
				return true
			}
		}
		for _, colCell := range s.columns[c].Cells {
			if isYZCandidate(colCell) &&
				s.eliminateXYZWingCells(pivot, xzCell, colCell, ss) {

				printStep(ss)
				s.applyStepCandidates(ss)
				return true
			}
		}
	}

	return false
}

// eliminateXYZWingCells removes candidate value z from any cells that see all
// three of xyzCell, xzCell, and yzCell.  The value x is the one candidate value
// that appears as a candidate in all 3 cells.  This assumes that xzCell and
// yzCell cannot see each other, and that xzCell is in the same box as xyzCell.
func (s *Solver) eliminateXYZWingCells(xyzCell, xzCell, yzCell *puzzle.Cell, ss *SolutionStep) bool {
	// The z value is the only common candidate between xzCell and yzCell.
	var z int
	for _, val := range xzCell.CandidateValues() {
		if yzCell.HasCandidate(val) {
			z = val
			break
		}
	}

	// The only cells that could possibly see all three XYZ-Wing cells are the
	// other cells in the same box as xyzCell and xzCell, so we just need to
	// check the candidate locations for value z in that box and select the
	// ones that can see yzCell.
	box := s.boxes[xyzCell.Box()]
	locs := box.Unsolved[z]
	cells := box.cellsFromLocs(locs.Values())
	cells = filterSlice(cells, func(cell *puzzle.Cell) bool {
		return !cell.SameCell(xyzCell) &&
			!cell.SameCell(xzCell) &&
			seesCell(cell, yzCell)
	})
	if len(cells) == 0 {
		// No candidates found to eliminate.
		return false
	}

	for _, xCell := range cells {
		ss.deleteCandidate(xCell.Row, xCell.Col, z)
	}
	return true
}

func (s *Solver) findHiddenQuadruples() bool {
	printChecking("Hidden Quadruple")
	found := false
	for i := range 9 {
		found = s.checkHiddenQuadruplesForHouse(s.rows[i]) || found
		found = s.checkHiddenQuadruplesForHouse(s.columns[i]) || found
		found = s.checkHiddenQuadruplesForHouse(s.boxes[i]) || found
	}
	return found
}

func (s *Solver) checkHiddenQuadruplesForHouse(h *House) bool {
	locs := filterMap(h.Unsolved, func(_ int, l LocSet) bool {
		return l.Size() == 2 || l.Size() == 3 || l.Size() == 4
	})
	if len(locs) < 4 {
		// We need at least 4 candidate values to have a quadruple.
		return false
	}

	values := mapKeys(locs)
	for i := 0; i < len(values)-3; i++ {
		for j := i + 1; j < len(values)-2; j++ {
			for k := j + 1; k < len(values)-1; k++ {
				for n := k + 1; n < len(values); n++ {
					w, x, y, z := values[i], values[j], values[k], values[n]
					locSet := set.Union(locs[w], locs[x], locs[y], locs[z])
					if locSet.Size() != 4 {
						// If the union of the location sets does not have exactly 4 elements, then
						// this is not a hidden quadruple.
						continue
					}

					valueSet := set.NewSet(w, x, y, z)
					ss := NewSolutionStep(fmt.Sprintf("Hidden Quadruple (%s)", h.Type))
					if s.eliminateOtherValues(h, valueSet, locSet, ss) {
						printStep(ss)
						s.applyStepCandidates(ss)
						return true
					}
				}
			}
		}
	}

	return false
}

func (s *Solver) findSwordfish() bool {
	printChecking("Swordfish")
	found := s.findSwordfishInLines(s.rows, s.columns)
	found = s.findSwordfishInLines(s.columns, s.rows) || found
	return found
}

func (s *Solver) findSwordfishInLines(baseLines, coverLines []*House) bool {
	for _, base := range baseLines {
		for val, locs := range base.Unsolved {
			// A Swordfish needs at least one line with exactly 3 candidate
			// locations for a value.
			if locs.Size() == 3 && s.checkSwordfishForValue(val, base, baseLines, coverLines) {
				return true
			}
		}
	}

	return false
}

func (s *Solver) checkSwordfishForValue(val int, base1 *House, baseLines, coverLines []*House) bool {
	// Find all base lines other than base1 that have either 2 or 3 candidate
	// locations for val.
	candidates := filterSlice(baseLines, func(b2 *House) bool {
		numLocs := b2.NumLocations(val)
		return b2.Index != base1.Index && numLocs >= 2 && numLocs <= 3
	})

	valueSet := set.NewSet(val)
	locSet := set.NewSet(base1.Index)
	base1Locs := base1.Unsolved[val]
	for _, line := range candidates {
		locs := line.Unsolved[val]
		if set.Union(base1Locs, locs).Size() != 3 {
			// The locations in this line don't match the locations in base1, so
			// it can't be part of the Swordfish.
			continue
		}

		// Add the line to the set of matched base indexes.  If we've found
		// 3 matching lines, then we have a Swordfish and we can check for
		// eliminated candidates.
		locSet.Add(line.Index)
		if locSet.Size() < 3 {
			continue
		}

		covers := transformSlice(base1Locs.Values(), func(y int) *House {
			return coverLines[y]
		})
		ss := NewSolutionStep("Swordfish")
		if s.eliminateFromOtherLocsMulti(covers, valueSet, locSet, ss) {
			printStep(ss)
			s.applyStepCandidates(ss)
			return true
		}
		break
	}

	return false
}

func (s *Solver) findJellyfish() bool {
	printChecking("Jellyfish")
	found := s.findJellyfishInLines(s.rows, s.columns)
	found = s.findJellyfishInLines(s.columns, s.rows) || found
	return found
}

func (s *Solver) findJellyfishInLines(baseLines, coverLines []*House) bool {
	for _, base := range baseLines {
		for val, locs := range base.Unsolved {
			// A Swordfish needs at least one line with exactly 4 candidate
			// locations for a value.
			if locs.Size() == 4 && s.checkJellyfishForValue(val, base, baseLines, coverLines) {
				return true
			}
		}
	}

	return false
}

func (s *Solver) checkJellyfishForValue(val int, base1 *House, baseLines, coverLines []*House) bool {
	// Find all base lines other than base1 that have at least 2 but not more
	// than 4 candidate locations for val.
	candidates := filterSlice(baseLines, func(b2 *House) bool {
		numLocs := b2.NumLocations(val)
		return b2.Index != base1.Index && numLocs >= 2 && numLocs <= 4
	})

	valueSet := set.NewSet(val)
	locSet := set.NewSet(base1.Index)
	baseLocs := base1.Unsolved[val]
	for _, g := range candidates {
		locs := g.Unsolved[val]
		if set.Union(baseLocs, locs).Size() != 4 {
			// The locations in this line don't match the locations in base1, so
			// it can't be part of the Swordfish.
			continue
		}

		// Add the line to the set of matched base indexes.  If we've found
		// 4 matching lines, then we have a Jellyfish and we can check for
		// eliminated candidates.
		locSet.Add(g.Index)
		if locSet.Size() < 4 {
			continue
		}

		covers := transformSlice(baseLocs.Values(), func(y int) *House {
			return coverLines[y]
		})
		ss := NewSolutionStep("Jellyfish")
		if s.eliminateFromOtherLocsMulti(covers, valueSet, locSet, ss) {
			printStep(ss)
			s.applyStepCandidates(ss)
			return true
		}
		break
	}

	return false
}

func (s *Solver) findUniqueRectangles() bool {
	printChecking("Unique Rectangle")
	found := false
	b := s.puzzle
	// Check each cell with exactly 2 candidate values to see if it is the base
	// corner of a unique rectangle.
	for r := range 9 {
		for c := range 9 {
			cell := b.Grid[r][c]
			if cell.NumCandidates() == 2 && s.checkUniqueRectangleForCell(cell) {
				return true
			}
		}
	}
	return found
}

func (s *Solver) checkUniqueRectangleForCell(base *puzzle.Cell) bool {
	b := s.puzzle

	// Look for a cell in the same row as base with the same pair of candidates.
	var rowWing *puzzle.Cell
	for c := range 9 {
		if c != base.Col {
			cell := b.Grid[base.Row][c]
			if sameCandidates(base, cell) {
				rowWing = cell
				break
			}
		}
	}
	if rowWing == nil {
		return false
	}

	// Look for a cell in the same column as base with the same pair of candidates.
	var colWing *puzzle.Cell
	for r := range 9 {
		if r != base.Row {
			cell := b.Grid[r][base.Col]
			if sameCandidates(base, cell) {
				colWing = cell
				break
			}
		}
	}
	if colWing == nil {
		return false
	}

	// The 2 wing cells must be in different boxes, but one of them must be in
	// the same box as the base.
	if rowWing.Box() != colWing.Box() &&
		(rowWing.Box() == base.Box() || colWing.Box() == base.Box()) {

		// These cells form a unique rectangle, so we can eliminate their candidates
		// from the cell at the 4th corner of the rectangle, which will have the
		// same row as the column-wing and the same column as the row-wing.
		ss := NewSolutionStep("Unique Rectangle")
		if s.eliminateValuesFromCell(colWing.Row, rowWing.Col, base.Candidates, ss) {
			printStep(ss)
			s.applyStepCandidates(ss)
			return true
		}
	}

	return false
}

// eliminateValuesFromCell removes all candidates listed in values from the cell
// at (r,c).
func (s *Solver) eliminateValuesFromCell(
	r, c int, values ValSet, ss *SolutionStep,
) bool {
	cell := s.puzzle.Grid[r][c]
	found := false
	for _, v := range values.Values() {
		if cell.HasCandidate(v) {
			ss.deleteCandidate(r, c, v)
			found = true
		}
	}
	return found
}
