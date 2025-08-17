package solver

import (
	"slices"

	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/set"
)

type techniqueKind int

const (
	kindNakedSingle techniqueKind = iota
	kindHiddenSingle
	kindNakedPair
	kindLockedCandidate
	kindPointingTuple
	kindHiddenPair
	kindNakedTriple
	kindXWing
	kindHiddenTriple
	kindNakedQuadruple
	kindXYWing
	kindAvoidableRectangle
	kindXYZWing
	kindHiddenQuadruple
	kindUniqueRectangle
	kindHiddenRectangle
	kindPointingRectangle
	kindSwordfish
	kindJellyfish
	kindSkyscraper
	kindTwoStringKite
	kindEmptyRectangle
	kindColorChain
	kindFinnedXWing
	kindFinnedSwordfish
	kindFinnedJellyfish
)

// techniqueNames is a list of display names for each technique kind.
// The order of the names must match the order of the techniqueKind constants.
var techniqueNames = []string{
	"Naked Single",
	"Hidden Single",
	"Naked Pair",
	"Locked Candidate",
	"Pointing Tuple",
	"Hidden Pair",
	"Naked Triple",
	"X-Wing",
	"Hidden Triple",
	"Naked Quadruple",
	"XY-Wing",
	"Avoidable Rectangle",
	"XYZ-Wing",
	"Hidden Quadruple",
	"Unique Rectangle",
	"Hidden Rectangle",
	"Pointing Rectangle",
	"Swordfish",
	"Jellyfish",
	"Skyscraper",
	"2-String Kite",
	"Empty Rectangle",
	"Color Chain",
	"Finned X-Wing",
	"Finned Swordfish",
	"Finned Jellyfish",
}

func techniqueName(tk techniqueKind) string {
	return techniqueNames[tk]
}

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
	printChecking(kindHiddenSingle)
	found := false
	for _, h := range s.houses {
		found = s.checkHiddenSinglesForHouse(h) || found
	}
	return found
}

func (s *Solver) checkHiddenSinglesForHouse(h *House) bool {
	for val, locs := range h.Unsolved {
		if locs.Size() == 1 {
			index := locs.Values()[0]
			cell := h.Cells[index]
			step := NewSolutionStep(kindHiddenSingle).
				WithHouse(h).
				WithPlacedValue(cell.Row, cell.Col, val)
			s.applyStep(step)
			return true
		}
	}
	return false
}

func (s *Solver) findNakedPairs() (step *SolutionStep, found bool) {
	printChecking(kindNakedPair)
	for _, h := range s.houses {
		if step, found = s.checkNakedPairsForHouse(h); found {
			return step, true
		}
	}

	return nil, false
}

func (s *Solver) checkNakedPairsForHouse(h *House) (step *SolutionStep, found bool) {
	values := make(LocValMap)
	for i, c := range h.Cells {
		// Collect a map of all locations with exactly 2 candidate values.
		if c.NumCandidates() == 2 {
			values[i] = c.Candidates
		}
	}
	if len(values) < 2 {
		// We need at least 2 candidate values to have a pair.
		return nil, false
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
			step = NewSolutionStep(kindNakedPair).
				WithCells(h.cellsFromLocs(locSet.Values())...).
				WithValues(valueSet.Values()...)
			if s.eliminateFromOtherLocs(h, valueSet, locSet, step) {
				return step.WithHouse(h), true
			}
		}
	}

	return nil, false
}

// eliminateFromOtherLocs removes the candidates listed in values from all
// cells that are not listed in locs.
func (s *Solver) eliminateFromOtherLocs(
	h *House, values ValSet, locs LocSet, step *SolutionStep,
) bool {
	found := false
	for l := range 9 {
		if locs.Contains(l) {
			continue
		}
		c := h.Cells[l]
		for _, v := range values.Values() {
			if c.HasCandidate(v) {
				step.DeleteCandidate(c.Row, c.Col, v)
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
	houses []*House, values ValSet, locs LocSet, step *SolutionStep,
) bool {
	updated := false
	for _, g := range houses {
		updated = s.eliminateFromOtherLocs(g, values, locs, step) || updated
	}

	return updated
}

func (s *Solver) findLockedCandidates() (step *SolutionStep, found bool) {
	printChecking(kindLockedCandidate)
	for i := range 9 {
		// We only need to check rows and columns for Locked Candidates.
		if step, found = s.checkLockedCandidatesForLine(s.rows[i]); found {
			return step, true
		}
		if step, found = s.checkLockedCandidatesForLine(s.columns[i]); found {
			return step, true
		}
	}
	return nil, false
}

func (s *Solver) checkLockedCandidatesForLine(line *House) (step *SolutionStep, found bool) {
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
			step = NewSolutionStep(kindLockedCandidate).
				WithValues(val).
				WithHouse(line)
			if s.eliminateFromOtherLocs(s.boxes[box], valueSet, locSet, step) {
				return step, true
			}
		}
	}

	return nil, false
}

func (s *Solver) findPointingTuples() (step *SolutionStep, found bool) {
	printChecking(kindPointingTuple)
	for i := range 9 {
		// We only need to check boxes for Pointing Tuples.
		if step, found = s.checkPointingTuplesForBox(s.boxes[i]); found {
			return step, true
		}
	}
	return nil, false
}

func (s *Solver) checkPointingTuplesForBox(box *House) (step *SolutionStep, found bool) {
	candidates := filterMap(box.Unsolved, func(_ int, l LocSet) bool {
		// If we have more than 3 candidates in a single box, then they can't all
		// be in the same line.
		return l.Size() <= 3
	})

	for val, locs := range candidates {
		step = NewSolutionStep(kindPointingTuple).
			WithValues(val).
			WithHouse(box)
		valueSet := set.NewSet(val)
		cells := box.cellsFromLocs(locs.Values())
		if row, ok := box.sharedRow(locs); ok {
			cols := transformSlice(cells, func(c *puzzle.Cell) int {
				return c.Col
			})
			locSet := set.NewSet(cols...)
			if s.eliminateFromOtherLocs(s.rows[row], valueSet, locSet, step) {
				return step, true
			}
		}
		if col, ok := box.sharedCol(locs); ok {
			rows := transformSlice(cells, func(c *puzzle.Cell) int {
				return c.Row
			})
			locSet := set.NewSet(rows...)
			if s.eliminateFromOtherLocs(s.columns[col], valueSet, locSet, step) {
				return step, true
			}
		}
	}

	return nil, false
}

func (s *Solver) findHiddenPairs() (step *SolutionStep, found bool) {
	printChecking(kindHiddenPair)
	for _, h := range s.houses {
		if step, found = s.checkHiddenPairsForHouse(h); found {
			return step, true
		}
	}

	return nil, false
}

func (s *Solver) checkHiddenPairsForHouse(h *House) (step *SolutionStep, found bool) {
	locs := filterMap(h.Unsolved, func(_ int, l LocSet) bool {
		return l.Size() == 2
	})
	if len(locs) < 2 {
		// We need at least 2 candidate values to have a pair.
		return nil, false
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
			step = NewSolutionStep(kindHiddenPair).
				WithCells(h.cellsFromLocs(locSet.Values())...).
				WithValues(valueSet.Values()...)
			if s.eliminateOtherValues(h, valueSet, locSet, step) {
				return step.WithHouse(h), true
			}
		}
	}

	return nil, false
}

// eliminateOtherValues removes candidates that are not listed in values from
// the cells in locs.
func (s *Solver) eliminateOtherValues(
	h *House, values ValSet, locs LocSet, step *SolutionStep,
) bool {
	found := false
	for _, l := range locs.Values() {
		c := h.Cells[l]
		for _, v := range c.CandidateValues() {
			if !values.Contains(v) {
				step.DeleteCandidate(c.Row, c.Col, v)
				found = true
			}
		}
	}

	return found
}

func (s *Solver) findNakedTriples() (step *SolutionStep, found bool) {
	printChecking(kindNakedTriple)
	for _, h := range s.houses {
		if step, found = s.checkNakedTriplesForHouse(h); found {
			return step, true
		}
	}

	return nil, false
}

func (s *Solver) checkNakedTriplesForHouse(h *House) (step *SolutionStep, found bool) {
	values := make(LocValMap)
	for i, c := range h.Cells {
		// Collect a map of all locations with either 2 or 3 candidate values.
		if c.NumCandidates() == 2 || c.NumCandidates() == 3 {
			values[i] = c.Candidates
		}
	}
	if len(values) < 3 {
		// We need at least 3 candidate values to have a triple.
		return nil, false
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
				step = NewSolutionStep(kindNakedTriple).
					WithCells(h.cellsFromLocs(locSet.Values())...).
					WithValues(valueSet.Values()...)
				if s.eliminateFromOtherLocs(h, valueSet, locSet, step) {
					return step.WithHouse(h), true
				}
			}
		}
	}

	return nil, false
}

func (s *Solver) findHiddenTriples() (step *SolutionStep, found bool) {
	printChecking(kindHiddenTriple)
	for _, h := range s.houses {
		if step, found = s.checkHiddenTriplesForHouse(h); found {
			return step, true
		}
	}

	return nil, false
}

func (s *Solver) checkHiddenTriplesForHouse(h *House) (step *SolutionStep, found bool) {
	locs := filterMap(h.Unsolved, func(_ int, l LocSet) bool {
		return l.Size() == 2 || l.Size() == 3
	})
	if len(locs) < 3 {
		// We need at least 3 candidate values to have a triple.
		return nil, false
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
				step = NewSolutionStep(kindHiddenTriple).
					WithCells(h.cellsFromLocs(locSet.Values())...).
					WithValues(valueSet.Values()...)
				if s.eliminateOtherValues(h, valueSet, locSet, step) {
					return step.WithHouse(h), true
				}
			}
		}
	}

	return nil, false
}

func (s *Solver) findNakedQuadruples() (step *SolutionStep, found bool) {
	printChecking(kindNakedQuadruple)
	for _, h := range s.houses {
		if step, found = s.checkNakedQuadruplesForHouse(h); found {
			return step, true
		}
	}

	return nil, false
}

func (s *Solver) checkNakedQuadruplesForHouse(h *House) (step *SolutionStep, found bool) {
	values := make(LocValMap)
	for i, c := range h.Cells {
		// Collect a map of all locations with either 2, 3 or 4 candidate values.
		if c.NumCandidates() == 2 || c.NumCandidates() == 3 || c.NumCandidates() == 4 {
			values[i] = c.Candidates
		}
	}
	if len(values) < 4 {
		// We need at least 4 candidate values to have a quadruple.
		return nil, false
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
					step = NewSolutionStep(kindNakedQuadruple).
						WithCells(h.cellsFromLocs(locSet.Values())...).
						WithValues(valueSet.Values()...)
					if s.eliminateFromOtherLocs(h, valueSet, locSet, step) {
						return step.WithHouse(h), true
					}
				}
			}
		}
	}

	return nil, false
}

func (s *Solver) findXYWings() (step *SolutionStep, found bool) {
	printChecking(kindXYWing)
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
		return nil, false
	}

	// Try each candidate as the pivot cell, checking it against all of the other
	// candidates.
	for _, pivot := range candidates {
		if step, found = s.checkXYWingsForPivot(pivot, candidates); found {
			return step, true
		}
	}
	return nil, false
}

func (s *Solver) checkXYWingsForPivot(
	pivot *puzzle.Cell, candidates []*puzzle.Cell,
) (step *SolutionStep, found bool) {
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
		return nil, false
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
			step = NewSolutionStep(kindXYWing).
				WithCells(pivot, xc, yc).
				WithValues(x, y, z)
			if s.eliminateXYWingCells(z, xc, yc, step) {
				return step, true
			}
		}
	}

	return nil, false
}

// eliminateXYWingCells removes candidate value z from all cells that see both
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
				ss.DeleteCandidate(zCell.Row, zCell.Col, z)
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

func (s *Solver) findAvoidableRectangles() (step *SolutionStep, found bool) {
	return nil, false
}

// findXYZ searches for 3 cells that fit the "XYZ-Wing" pattern.  An XYZ-Wing
// consists of a pivot cell with 3 candidate values x,y,z, and two pincer cells
// that each see the pivot cell but don't see each other.  One pincer must have
// candidate values x,z and the other must have candidate values y,z.  One of
// these cells must have the value z, so z can be eliminated as a candidate for
// any cell that sees all three.  Note that one pincer *MUST* be in the same
// box as the pivot cell in order for it to be possible for any cell to see the
// pivot and both pincers.
func (s *Solver) findXYZWings() (step *SolutionStep, found bool) {
	printChecking(kindXYZWing)
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
		if step, found = s.checkXYZWingsForPivot(pivot); found {
			return step, true
		}
	}
	return nil, false
}

func (s *Solver) checkXYZWingsForPivot(pivot *puzzle.Cell) (step *SolutionStep, found bool) {
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
		return nil, false
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

		yzCells := slices.Concat(
			s.rows[pivot.Row].Cells[:],
			s.columns[pivot.Col].Cells[:],
		)
		for _, yzCell := range yzCells {
			step = NewSolutionStep(kindXYZWing)
			if isYZCandidate(yzCell) &&
				s.eliminateXYZWingCells(pivot, xzCell, yzCell, step) {

				step.WithCells(pivot, xzCell, yzCell).
					WithValues(pivot.CandidateValues()...)
				return step, true
			}
		}
	}

	return nil, false
}

// eliminateXYZWingCells removes candidate value z from any cells that see all
// three of xyzCell, xzCell, and yzCell.  The value x is the one candidate value
// that appears as a candidate in all 3 cells.  This assumes that xzCell and
// yzCell cannot see each other, and that xzCell is in the same box as xyzCell.
func (s *Solver) eliminateXYZWingCells(xyzCell, xzCell, yzCell *puzzle.Cell, step *SolutionStep) bool {
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
		step.DeleteCandidate(xCell.Row, xCell.Col, z)
	}
	return true
}

func (s *Solver) findHiddenQuadruples() (step *SolutionStep, found bool) {
	printChecking(kindHiddenQuadruple)
	for _, h := range s.houses {
		if step, found = s.checkHiddenQuadruplesForHouse(h); found {
			return step, true
		}
	}

	return nil, false
}

func (s *Solver) checkHiddenQuadruplesForHouse(h *House) (step *SolutionStep, found bool) {
	locs := filterMap(h.Unsolved, func(_ int, l LocSet) bool {
		return l.Size() == 2 || l.Size() == 3 || l.Size() == 4
	})
	if len(locs) < 4 {
		// We need at least 4 candidate values to have a quadruple.
		return nil, false
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
					step = NewSolutionStep(kindHiddenQuadruple).
						WithCells(h.cellsFromLocs(locSet.Values())...).
						WithValues(valueSet.Values()...)
					if s.eliminateOtherValues(h, valueSet, locSet, step) {
						return step.WithHouse(h), true
					}
				}
			}
		}
	}

	return nil, false
}

// FISH TECHNIQUES

func (s *Solver) findXWings() (step *SolutionStep, found bool) {
	return s.findFishOfSize(2, kindXWing)
}

func (s *Solver) findSwordfish() (step *SolutionStep, found bool) {
	return s.findFishOfSize(3, kindSwordfish)
}

func (s *Solver) findJellyfish() (step *SolutionStep, found bool) {
	return s.findFishOfSize(4, kindJellyfish)
}

func (s *Solver) findFishOfSize(
	fishSize int, fishKind techniqueKind,
) (step *SolutionStep, found bool) {
	printChecking(fishKind)
	find := func(baseLines, coverLines []*House) bool {
		step, found = s.findFishInLines(fishSize, fishKind, baseLines, coverLines)
		return found
	}
	if find(s.rows, s.columns) {
		return step, true
	}
	if find(s.columns, s.rows) {
		return step, true
	}
	return nil, false
}

func (s *Solver) findFishInLines(
	fishSize int,
	fishKind techniqueKind,
	baseLines, coverLines []*House,
) (step *SolutionStep, found bool) {
	for _, base := range baseLines {
		for val, locs := range base.Unsolved {
			// A fish line must have no more than fishSize candidate locations
			// for a value. We assume that all singles and smaller fish have
			// already been found.
			if locs.Size() > fishSize {
				continue
			}

			if step, found = s.checkFishForValue(fishSize, fishKind, val, base, baseLines, coverLines); found {
				return step, true
			}
		}
	}

	return nil, false
}

func (s *Solver) checkFishForValue(
	fishSize int,
	fishKind techniqueKind,
	val int,
	base1 *House,
	baseLines, coverLines []*House,
) (step *SolutionStep, found bool) {
	// Find all base lines other than base1 that have either 2 or 3 candidate
	// locations for val.
	candidates := filterSlice(baseLines, func(b2 *House) bool {
		numLocs := b2.NumLocations(val)
		return b2.Index != base1.Index && numLocs >= 2 && numLocs <= fishSize
	})

	step = NewSolutionStep(fishKind).WithValues(val)
	valueSet := set.NewSet(val)
	// Variables for storing search results.
	fishLines := []int{base1.Index}
	var coverLocs LocSet

	// Must forward-declare func so we can call it self-recursively.
	var checkLines func(lines []*House, fishLocs LocSet) bool
	checkLines = func(lines []*House, fishLocs LocSet) bool {
		for i, line := range lines {
			locs := set.Union(fishLocs, line.Unsolved[val])
			if locs.Size() > fishSize {
				// Too many locations, so this line can't be part of the fish.
				continue
			}
			fishLines = append(fishLines, line.Index)
			if len(fishLines) == fishSize {
				// We found enough lines, so we have a fish.
				// fishLines contains the base lines that make up the fish,
				// but we need to save the set containing the cover lines.
				coverLocs = locs
				return true
			}
			// Check recursively until we have enough lines.
			if checkLines(lines[i+1:], locs) {
				return true
			}
			// No fish found, so backtrack the last line and keep trying.
			fishLines = fishLines[:len(fishLines)-1]
		}
		// No more candidate lines, so we don't have a fish.
		return false
	}

	if checkLines(candidates, base1.Unsolved[val]) {
		// We found a fish.
		bases := transformSlice(fishLines, func(x int) *House {
			return baseLines[x]
		})
		step.WithBases(bases...)
		covers := transformSlice(coverLocs.Values(), func(y int) *House {
			return coverLines[y]
		})
		step.WithCovers(covers...)
		locSet := set.NewSet(fishLines...)
		if s.eliminateFromOtherLocsMulti(covers, valueSet, locSet, step) {
			return step, true
		}
	}

	return nil, false
}

func (s *Solver) findUniqueRectangles() (step *SolutionStep, found bool) {
	printChecking(kindUniqueRectangle)
	b := s.puzzle
	// Check each cell with exactly 2 candidate values to see if it is the base
	// corner of a unique rectangle.
	for r := range 9 {
		for c := range 9 {
			cell := b.Grid[r][c]
			if cell.NumCandidates() != 2 {
				continue
			}
			if step, found = s.checkUniqueRectangleForCell(cell); found {
				return step, true
			}
		}
	}

	return nil, false
}

func (s *Solver) checkUniqueRectangleForCell(base *puzzle.Cell) (step *SolutionStep, found bool) {
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
		return nil, false
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
		return nil, false
	}

	// The 2 wing cells must be in different boxes, but one of them must be in
	// the same box as the base.
	if rowWing.Box() != colWing.Box() &&
		(rowWing.Box() == base.Box() || colWing.Box() == base.Box()) {

		// These cells form a unique rectangle, so we can eliminate their candidates
		// from the cell at the 4th corner of the rectangle, which will have the
		// same row as the column-wing and the same column as the row-wing.
		step = NewSolutionStep(kindUniqueRectangle).
			WithValues(base.Candidates.Values()...).
			WithCells(base, rowWing, colWing)
		if s.eliminateValuesFromCell(colWing.Row, rowWing.Col, base.Candidates, step) {
			return step, true
		}
	}

	return nil, false
}

// eliminateValuesFromCell removes all candidates listed in values from the cell
// at (r,c).
func (s *Solver) eliminateValuesFromCell(
	r, c int, values ValSet, step *SolutionStep,
) bool {
	cell := s.puzzle.Grid[r][c]
	found := false
	for _, v := range values.Values() {
		if cell.HasCandidate(v) {
			step.DeleteCandidate(r, c, v)
			found = true
		}
	}
	return found
}
