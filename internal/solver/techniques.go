package solver

import (
	"slices"
	"time"

	"github.com/kpitt/sudoku/internal/bitset"
	"github.com/kpitt/sudoku/internal/puzzle"
)

type techniqueKind int

const (
	// Enum constants for each solving technique, which must be defined in the
	// order that they should be applied.
	kindNakedSingle techniqueKind = iota
	kindHiddenSingle
	kindLockedCandidatesPointing
	kindLockedCandidatesClaiming
	kindNakedPair
	kindNakedTriple
	kindHiddenPair
	kindHiddenTriple
	kindNakedQuadruple
	kindHiddenQuadruple
	kindXWing
	kindSwordfish
	kindJellyfish
	kindRemotePair
	kindBUG
	kindSkyscraper
	kindTwoStringKite
	kindEmptyRectangle
	kindWWing
	kindXYWing
	kindXYZWing
	kindAvoidableRectangle
	kindUniqueRectangle1
	kindUniqueRectangle2
	kindUniqueRectangle3
	kindUniqueRectangle4
	kindHiddenRectangle
	kindFinnedXWing
	kindFinnedSwordfish
	kindFinnedJellyfish
	kindSueDeCoq
	kindSimpleColoring
	kindMultiColoring
	kindXChain
	kindXYChain
	kindBruteForce
)

// A Technique represents a single Sudoku solving technique, represented by a
// display name and a function to check for the technique in the puzzle.
type Technique struct {
	Name  string
	Check func() bool
}

// initTechniques initializes a list of the known solving techniques and
// assigns it to the solver.
// The order must match the order of the techniqueKind constants.
func (s *Solver) initTechniques() {
	// The order of this list must match the order of the techniqueKind constants.
	s.techniques = []Technique{
		{"Naked Single", nil}, // checked during candidate removal
		{"Hidden Single", s.findHiddenSingles},
		{"Locked Candidates Type 1 (Pointing)", s.findPointingTuples},
		{"Locked Candidates Type 2 (Claiming)", s.findClaimingTuples},
		{"Naked Pair", s.findNakedPairs},
		{"Naked Triple", s.findNakedTriples},
		{"Hidden Pair", s.findHiddenPairs},
		{"Hidden Triple", s.findHiddenTriples},
		{"Naked Quadruple", s.findNakedQuadruples},
		{"Hidden Quadruple", s.findHiddenQuadruples},
		{"X-Wing", s.findXWings},
		{"Swordfish", s.findSwordfish},
		{"Jellyfish", s.findJellyfish},
		{"Remote Pair", nil},
		{"BUG+1", nil},
		{"Skyscraper", s.findSkyscraper},
		{"2-String Kite", s.findTwoStringKite},
		{"Empty Rectangle", nil},
		{"W-Wing", nil},
		{"XY-Wing", s.findXYWings},
		{"XYZ-Wing", s.findXYZWings},
		{"Avoidable Rectangle", s.findAvoidableRectangles},
		{"Unique Rectangle Type 1", s.findUniqueRectangleType1},
		{"Unique Rectangle Type 2", nil},
		{"Unique Rectangle Type 3", nil},
		{"Unique Rectangle Type 4", nil},
		{"Hidden Rectangle", nil},
		{"Finned X-Wing", nil},
		{"Finned Swordfish", nil},
		{"Finned Jellyfish", nil},
		{"Sue de Coq", nil},
		{"Simple Coloring", nil},
		{"Multi-Coloring", nil},
		{"X-Chain", nil},
		{"XY-Chain", nil},
		{"Brute Force", nil}, // custom check as last resort
	}
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
	found := false
	for _, h := range s.houses {
		found = s.checkHiddenSinglesForHouse(h) || found
	}
	return found
}

func (s *Solver) checkHiddenSinglesForHouse(h *House) bool {
	for val := 1; val <= 9; val++ {
		locs := h.Unsolved[val]
		if locs.Empty() {
			continue
		}
		if locs.Size() == 1 {
			index := locs.Value()
			cell := h.Cells[index]
			step := NewStep(kindHiddenSingle).
				WithHouse(h).
				WithPlacedValue(cell.Row*9+cell.Col, val)
			s.applyStep(step)
			return true
		}
	}
	return false
}

func (s *Solver) findNakedSubsets(size int, kind techniqueKind) (found bool) {
	return slices.ContainsFunc(s.houses, func(h *House) bool {
		return s.checkNakedSubsetsForHouse(size, kind, h)
	})
}

func (s *Solver) checkNakedSubsetsForHouse(size int, kind techniqueKind, h *House) (found bool) {
	var locsBuf [9]int
	locs := locsBuf[:0]
	// Collect a list of all locations with no more than `size` candidate values.
	for i, c := range h.Cells {
		numCand := c.NumCandidates()
		if numCand >= 2 && numCand <= size {
			locs = append(locs, i)
		}
	}
	if len(locs) < size {
		// We need at least `size` candidate values to have a subset of that size.
		return false
	}

	// Try combinations of the required size.
	var checkCombinations func(start int, currentIndices []int) bool
	checkCombinations = func(start int, currentIndices []int) bool {
		if len(currentIndices) == size {
			valueSet := bitset.BitSet16(0)
			for _, idx := range currentIndices {
				valueSet = bitset.Union(valueSet, h.Cells[locs[idx]].Candidates)
			}

			if valueSet.Size() == size {
				var locSet bitset.BitSet16
				for _, idx := range currentIndices {
					locSet.Add(locs[idx])
				}

				step := NewStep(kind)
				if s.eliminateFromOtherLocs(h, valueSet, locSet, step) {
					s.applyStep(step.
						WithIndices(h.indexesFromLocs(locSet.Values())...).
						WithValues(valueSet.Values()...).
						WithHouse(h))
					return true
				}
			}
			return false
		}

		for i := start; i < len(locs); i++ {
			if checkCombinations(i+1, append(currentIndices, i)) {
				return true
			}
		}
		return false
	}

	// Avoid allocation for small known sizes by allocating the initial buffer on the stack.
	indices := make([]int, 0, size)
	return checkCombinations(0, indices)
}

func (s *Solver) findNakedPairs() (found bool) {
	return s.findNakedSubsets(2, kindNakedPair)
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

		// Check for intersection first to avoid inner loop/allocations.
		common := c.Candidates.Intersection(values)
		if common.Empty() {
			continue
		}

		for v := range common.All() {
			step.DeleteCandidate(c.Index(), v)
			found = true
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

func (s *Solver) findClaimingTuples() (found bool) {
	for i := range 9 {
		// We only need to check rows and columns for Locked Candidates.
		if s.checkLockedCandidatesForLine(s.rows[i]) ||
			s.checkLockedCandidatesForLine(s.columns[i]) {

			return true
		}
	}
	return false
}

func (s *Solver) checkLockedCandidatesForLine(line *House) (found bool) {
	for val := 1; val <= 9; val++ {
		locs := line.Unsolved[val]
		if locs.Empty() {
			continue
		}

		// If we have more than 3 candidates in a line, then they can't all be
		// in the same box.
		if locs.Size() > 3 {
			continue
		}

		valueSet := bitset.FromValues16(val)
		if box, ok := line.sharedBox(locs); ok {
			cells := line.cellsFromLocs(locs.Values())
			boxCells := transformSlice(cells, func(c *puzzle.Cell) int {
				_, index := getBoxLoc(c.Row, c.Col)
				return index
			})
			locSet := bitset.FromValues16(boxCells...)
			step := NewStep(kindLockedCandidatesClaiming)
			if s.eliminateFromOtherLocs(s.boxes[box], valueSet, locSet, step) {
				s.applyStep(step.
					WithValues(val).
					WithHouse(line))
				return true
			}
		}
	}

	return false
}

func (s *Solver) findPointingTuples() (found bool) {
	for i := range 9 {
		// We only need to check boxes for Pointing Tuples.
		if s.checkPointingTuplesForBox(s.boxes[i]) {
			return true
		}
	}
	return false
}

func (s *Solver) checkPointingTuplesForBox(box *House) (found bool) {
	for val := 1; val <= 9; val++ {
		locs := box.Unsolved[val]
		if locs.Empty() {
			continue
		}

		// If we have more than 3 candidates in a single box, then they can't all
		// be in the same line.
		if locs.Size() > 3 {
			continue
		}

		step := NewStep(kindLockedCandidatesPointing).
			WithValues(val).
			WithHouse(box)
		valueSet := bitset.FromValues16(val)
		cells := box.cellsFromLocs(locs.Values())
		if row, ok := box.sharedRow(locs); ok {
			cols := transformSlice(cells, func(c *puzzle.Cell) int {
				return c.Col
			})
			locSet := bitset.FromValues16(cols...)
			if s.eliminateFromOtherLocs(s.rows[row], valueSet, locSet, step) {
				s.applyStep(step)
				return true
			}
		}
		if col, ok := box.sharedCol(locs); ok {
			rows := transformSlice(cells, func(c *puzzle.Cell) int {
				return c.Row
			})
			locSet := bitset.FromValues16(rows...)
			if s.eliminateFromOtherLocs(s.columns[col], valueSet, locSet, step) {
				s.applyStep(step)
				return true
			}
		}
	}

	return false
}

func (s *Solver) findHiddenSubsets(size int, kind techniqueKind) (found bool) {
	return slices.ContainsFunc(s.houses, func(h *House) bool {
		return s.checkHiddenSubsetsForHouse(size, kind, h)
	})
}

func (s *Solver) checkHiddenSubsetsForHouse(size int, kind techniqueKind, h *House) (found bool) {
	var valBuf [9]int
	values := valBuf[:0]
	for val := 1; val <= 9; val++ {
		locs := h.Unsolved[val]
		if locs.Empty() {
			continue
		}
		if locs.Size() >= 2 && locs.Size() <= size {
			values = append(values, val)
		}
	}

	if len(values) < size {
		// We need at least `size` candidate values to have a subset of that size.
		return false
	}

	var checkCombinations func(start int, currentIndices []int) bool
	checkCombinations = func(start int, currentIndices []int) bool {
		if len(currentIndices) == size {
			locSet := bitset.BitSet16(0)
			valueSet := bitset.BitSet16(0)
			for _, idx := range currentIndices {
				val := values[idx]
				locSet = bitset.Union(locSet, h.Unsolved[val])
				valueSet.Add(val)
			}

			if locSet.Size() == size {
				step := NewStep(kind)
				if s.eliminateOtherValues(h, valueSet, locSet, step) {
					s.applyStep(step.
						WithIndices(h.indexesFromLocs(locSet.Values())...).
						WithValues(valueSet.Values()...).
						WithHouse(h))
					return true
				}
			}
			return false
		}

		for i := start; i < len(values); i++ {
			if checkCombinations(i+1, append(currentIndices, i)) {
				return true
			}
		}
		return false
	}

	indices := make([]int, 0, size)
	return checkCombinations(0, indices)
}

func (s *Solver) findHiddenPairs() (found bool) {
	return s.findHiddenSubsets(2, kindHiddenPair)
}

// eliminateOtherValues removes candidates that are not listed in values from
// the cells in locs.
func (s *Solver) eliminateOtherValues(
	h *House, values ValSet, locs LocSet, step *SolutionStep,
) bool {
	found := false
	for l := range locs.All() {
		c := h.Cells[l]

		// We want to remove any candidates in `c` that are NOT in `values`.
		toRemove := c.Candidates.Difference(values)
		if toRemove.Empty() {
			continue
		}

		for v := range toRemove.All() {
			step.DeleteCandidate(c.Index(), v)
			found = true
		}
	}

	return found
}

func (s *Solver) findNakedTriples() (found bool) {
	return s.findNakedSubsets(3, kindNakedTriple)
}

func (s *Solver) findHiddenTriples() (found bool) {
	return s.findHiddenSubsets(3, kindHiddenTriple)
}

func (s *Solver) findNakedQuadruples() (found bool) {
	return s.findNakedSubsets(4, kindNakedQuadruple)
}

func (s *Solver) findXYWings() (found bool) {
	// Collect a list of all cells with exactly 2 candidates.
	var candidates []*puzzle.Cell
	for i := range 81 {
		if s.puzzle.Cell(i).NumCandidates() == 2 {
			candidates = append(candidates, s.puzzle.Cell(i))
		}
	}
	if len(candidates) < 3 {
		// An XY-Wing requires a pivot cell and 2 pincer cells, so we need at
		// least 3 candidates.
		return false
	}

	// Try each candidate as the pivot cell, checking it against its peers.
	for _, pivot := range candidates {
		if s.checkXYWingsForPivot(pivot) {
			return true
		}
	}
	return false
}

func (s *Solver) checkXYWingsForPivot(pivot *puzzle.Cell) (found bool) {
	// Get the x and y values.
	values := pivot.CandidateValues()
	x, y := values[0], values[1]

	// Find the candidate cells that can be seen by the pivot cell and have either
	// x or y as a candidate, but not both.  Collect the cells into separate lists
	// for cells that have x but not y and cells that have y but not x.
	var xCells, yCells []*puzzle.Cell

	peers := puzzle.GetPeers(pivot.Index())
	for i := range 20 {
		cell := s.puzzle.Cell(peers[i])
		if cell.Index() == pivot.Index() || cell.NumCandidates() != 2 ||
			!seesCell(cell.Index(), pivot.Index()) {

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
		// Look for a y-cell that also contains z and is not visible from the x-cell.
		for _, yc := range yCells {
			if seesCell(xc.Index(), yc.Index()) {
				continue
			}

			// The intersection of {x, z} and {y, z} must be {z}.
			// Since x != y, any intersection is the common value z.
			common := xc.Candidates.Intersection(yc.Candidates)
			if common.Empty() {
				// No common candidate, not an XY-Wing
				continue
			}

			z := common.Value()
			step := NewStep(kindXYWing)
			if s.eliminateFromIntersection(xc.Index(), yc.Index(), -1, z, step) {
				s.applyStep(step.
					WithIndices(pivot.Index(), xc.Index(), yc.Index()).
					WithValues(x, y, z))
				return true
			}
		}
	}

	return false
}

func (s *Solver) findAvoidableRectangles() (found bool) {
	// TODO: Implement "Avoidable Rectangle" technique
	return false
}

// findXYZ searches for 3 cells that fit the "XYZ-Wing" pattern.  An XYZ-Wing
// consists of a pivot cell with 3 candidate values x,y,z, and two pincer cells
// that each see the pivot cell but don't see each other.  One pincer must have
// candidate values x,z and the other must have candidate values y,z.  One of
// these cells must have the value z, so z can be eliminated as a candidate for
// any cell that sees all three.  Note that one pincer *MUST* be in the same
// box as the pivot cell in order for it to be possible for any cell to see the
// pivot and both pincers.
func (s *Solver) findXYZWings() (found bool) {
	// Collect a list of all cells with exactly 3 candidates.
	p := s.puzzle
	var candidates []*puzzle.Cell
	for i := range 81 {
		if p.Cell(i).NumCandidates() == 3 {
			candidates = append(candidates, p.Cell(i))
		}
	}

	// Check each candidate as a possible pivot cell for an XYZ-Wing.
	return slices.ContainsFunc(candidates, s.checkXYZWingsForPivot)
}

func (s *Solver) checkXYZWingsForPivot(pivot *puzzle.Cell) (found bool) {
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
		// Find the y value (pivot - xzCell).
		// Pivot::{x,y,z}, xzCell::{x,z} => Difference::{y}
		ySet := pivot.Candidates.Difference(xzCell.Candidates)
		if ySet.Empty() {
			continue
		} // Should not happen if xzCell is valid subset
		y := ySet.Value()

		// Now find a cell in the same row or column as the pivot cell.
		isYZCandidate := func(cell *puzzle.Cell) bool {
			if cell.Box() == pivot.Box() ||
				cell.NumCandidates() != 2 ||
				!cell.HasCandidate(y) {

				return false
			}

			// Verify it shares Z with xzCell
			// yzCell::{y,z}, xzCell::{x,z} => Intersection::{z}
			// (Assuming x != y, which is true)
			return cell.Candidates.Intersects(xzCell.Candidates)
		}

		yzCells := slices.Concat(
			s.rows[pivot.Row].Cells[:],
			s.columns[pivot.Col].Cells[:],
		)
		for _, yzCell := range yzCells {
			if !isYZCandidate(yzCell) {
				continue
			}
			// We now know that the 2 cells are {x,z} and {y,z}, so the intersection
			// must be {z}.
			z := xzCell.Candidates.Intersection(yzCell.Candidates).Value()
			step := NewStep(kindXYZWing)
			if s.eliminateFromIntersection(xzCell.Index(), yzCell.Index(), pivot.Index(), z, step) {
				s.applyStep(step.
					WithIndices(pivot.Index(), xzCell.Index(), yzCell.Index()).
					WithValues(pivot.CandidateValues()...))
				return true
			}
		}
	}

	return false
}

// eliminateFromIntersection removes `val` from any cell that sees ALL provided indices.
// If pivotIdx is -1, it finds intersection of (idx1, idx2).
// If pivotIdx is >= 0, it finds intersection of (idx1, idx2, pivotIdx).
func (s *Solver) eliminateFromIntersection(
	idx1, idx2, pivotIdx, val int, step *SolutionStep,
) bool {
	changed := false

	peers1 := puzzle.GetPeers(idx1)

	// Build fast lookup table for peers of idx2.
	isPeerOf2 := s.getPeerSet(idx2)

	var isPeerOfPivot [81]bool
	hasPivot := false
	if pivotIdx != -1 {
		isPeerOfPivot = s.getPeerSet(pivotIdx)
		hasPivot = true
	}

	for i := range 20 {
		candidateIdx := int(peers1[i])

		// Must be peer of idx2
		if !isPeerOf2[candidateIdx] {
			continue
		}

		// Must be peer of pivot
		if hasPivot && !isPeerOfPivot[candidateIdx] {
			continue
		}

		// Don't eliminate from the pattern cells themselves
		if candidateIdx == idx1 || candidateIdx == idx2 || candidateIdx == pivotIdx {
			continue
		}

		// Remove the candidate if it exists
		if s.puzzle.Cell(candidateIdx).HasCandidate(val) {
			changed = true
			step.DeleteCandidate(candidateIdx, val)
		}
	}
	return changed
}

// getPeerSet returns a fast lookup table for peers.
func (s *Solver) getPeerSet(idx int) [81]bool {
	peers := puzzle.GetPeers(idx)

	var lookup [81]bool
	for i := range 20 {
		peerIdx := peers[i]
		lookup[peerIdx] = true
	}
	return lookup
}

func (s *Solver) findHiddenQuadruples() (found bool) {
	return s.findHiddenSubsets(4, kindHiddenQuadruple)
}

// FISH TECHNIQUES

func (s *Solver) findXWings() (found bool) {
	return s.findFishOfSize(2, kindXWing)
}

func (s *Solver) findSwordfish() (found bool) {
	return s.findFishOfSize(3, kindSwordfish)
}

func (s *Solver) findJellyfish() (found bool) {
	return s.findFishOfSize(4, kindJellyfish)
}

func (s *Solver) findFishOfSize(fishSize int, fishKind techniqueKind) (found bool) {
	find := func(baseLines, coverLines []*House) bool {
		return s.findFishInLines(fishSize, fishKind, baseLines, coverLines)
	}
	return find(s.rows, s.columns) || find(s.columns, s.rows)
}

func (s *Solver) findFishInLines(
	fishSize int,
	fishKind techniqueKind,
	baseLines, coverLines []*House,
) (found bool) {
	for _, base := range baseLines {
		for val := 1; val <= 9; val++ {
			locs := base.Unsolved[val]
			if locs.Empty() {
				continue
			}

			// A fish line must have no more than fishSize candidate locations
			// for a value. We assume that all singles and smaller fish have
			// already been found.
			if locs.Size() > fishSize {
				continue
			}

			if s.checkFishForValue(fishSize, fishKind, val, base, baseLines, coverLines) {
				return true
			}
		}
	}

	return false
}

func (s *Solver) checkFishForValue(
	fishSize int,
	fishKind techniqueKind,
	val int,
	base1 *House,
	baseLines, coverLines []*House,
) (found bool) {
	// Find all base lines other than base1 that have either 2 or 3 candidate
	// locations for val.
	candidates := filterSlice(baseLines, func(b2 *House) bool {
		numLocs := b2.NumLocations(val)
		return b2.Index != base1.Index && numLocs >= 2 && numLocs <= fishSize
	})

	valueSet := bitset.FromValues16(val)
	// Variables for storing search results.
	fishLines := []int{base1.Index}
	var coverLocs LocSet

	// Must forward-declare func so we can call it self-recursively.
	var checkLines func(lines []*House, fishLocs LocSet) bool
	checkLines = func(lines []*House, fishLocs LocSet) bool {
		for i, line := range lines {
			locs := bitset.Union(fishLocs, line.Unsolved[val])
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

	step := NewStep(fishKind).WithValues(val)
	if checkLines(candidates, base1.Unsolved[val]) {
		// We found a fish.
		bases := transformSlice(fishLines, func(x int) *House {
			return baseLines[x]
		})
		covers := transformSlice(coverLocs.Values(), func(y int) *House {
			return coverLines[y]
		})
		locSet := bitset.FromValues16(fishLines...)
		if s.eliminateFromOtherLocsMulti(covers, valueSet, locSet, step) {
			s.applyStep(step.
				WithBases(bases...).
				WithCovers(covers...))
			return true
		}
	}

	return false
}

// SINGLE-DIGIT TECHNIQUES

func (s *Solver) findSkyscraper() (found bool) {
	check := func(baseLines []*House) bool {
		return s.checkSkyscraper(baseLines)
	}
	// We check for Skyscrapers where the base lines are rows (meaning the strong
	// links are horizontal) and then where the base lines are columns (vertical
	// strong links).
	return check(s.rows) || check(s.columns)
}

func (s *Solver) checkSkyscraper(baseLines []*House) (found bool) {
	candidates := make([]*House, 0, 9)

	for val := 1; val <= 9; val++ {
		// Find all lines that have exactly 2 candidates for this value.
		candidates = candidates[:0]
		for _, line := range baseLines {
			if line.NumLocations(val) == 2 {
				candidates = append(candidates, line)
			}
		}

		if len(candidates) < 2 {
			continue
		}

		// Check each pair of candidate lines to see if they form a Skyscraper.
		for i := 0; i < len(candidates)-1; i++ {
			base1 := candidates[i]
			for j := i + 1; j < len(candidates); j++ {
				base2 := candidates[j]

				// To form a Skyscraper, the two base lines must share exactly
				// one column (or row, if bases are columns) where the candidate
				// appears. This shared location forms the base of the Skyscraper.
				locs1 := base1.Unsolved[val]
				locs2 := base2.Unsolved[val]

				// Find intersection of locations.
				// Since we know size is 2, we can just check values.
				commonLoc := -1
				top1Loc := -1
				top2Loc := -1

				// Identify common and distinct locations
				for _, l1 := range locs1.Values() {
					if locs2.Contains(l1) {
						commonLoc = l1
					} else {
						top1Loc = l1
					}
				}

				// If we didn't find exactly one common location, or if the
				// locations are identical (2 common locations), then this isn't
				// a Skyscraper.
				// Note: If locs1 and locs2 are identical, commonLoc would be set
				// but top1Loc would remain -1.
				if commonLoc == -1 || top1Loc == -1 {
					continue
				}

				// Find top2Loc (the one in base2 that isn't common)
				for _, l2 := range locs2.Values() {
					if l2 != commonLoc {
						top2Loc = l2
						break
					}
				}

				// The "top" cells are the ends of the strong links that are NOT
				// the shared base.
				top1 := base1.Cells[top1Loc]
				top2 := base2.Cells[top2Loc]

				// The "floor" cells are the shared base cells.
				floor1 := base1.Cells[commonLoc]
				floor2 := base2.Cells[commonLoc]

				// Try to eliminate candidates from any cell that sees both tops.
				step := NewStep(kindSkyscraper).WithValues(val)
				if s.eliminateFromIntersection(top1.Index(), top2.Index(), -1, val, step) {
					s.applyStep(step.WithIndices(top1.Index(), top2.Index(), floor1.Index(), floor2.Index()))
					return true
				}
			}
		}
	}

	return false
}

func (s *Solver) findTwoStringKite() (found bool) {
	candidates := make([]*House, 0, 9)
	for val := 1; val <= 9; val++ {
		// Find rows and columns with exactly 2 candidates.
		rowCandidates := candidates[:0] // Reuse slice
		for _, row := range s.rows {
			if row.NumLocations(val) == 2 {
				rowCandidates = append(rowCandidates, row)
			}
		}
		// We can't reuse the same slice for cols since we iterate nested.
		// So let's just make a new one or manage it carefully.
		// Actually, let's just filter locally in the loop.
		// Or loop rows then loop cols.

		if len(rowCandidates) == 0 {
			continue
		}

		colCandidates := make([]*House, 0, 9)
		for _, col := range s.columns {
			if col.NumLocations(val) == 2 {
				colCandidates = append(colCandidates, col)
			}
		}

		if len(colCandidates) == 0 {
			continue
		}

		// Check each Row against each Col
		for _, row := range rowCandidates {
			for _, col := range colCandidates {
				if s.checkTwoStringKite(val, row, col) {
					return true
				}
			}
		}
	}
	return false
}

func (s *Solver) checkTwoStringKite(val int, row, col *House) (found bool) {
	// Row candidates (2 of them)
	rLocs := row.Unsolved[val].Values()
	// Col candidates (2 of them)
	cLocs := col.Unsolved[val].Values()

	// rLocs are column indices in the row.
	// cLocs are row indices in the column.

	checkPoly := func(rP, rTail, cP, cTail int) bool {
		// rP is column index of the row candidate involved in the connection
		// cP is row index of the col candidate involved in the connection

		rCell := row.Cells[rP]
		cCell := col.Cells[cP]

		// Ensure they are not the same cell (intersection)
		if rCell.Index() == cCell.Index() {
			return false
		}

		// Check if they are in the same box (connection)
		if rCell.Box() == cCell.Box() {
			// Found connection!

			// Tails are the other candidates
			tail1 := row.Cells[rTail]
			tail2 := col.Cells[cTail]

			step := NewStep(kindTwoStringKite).
				WithValues(val).
				WithBases(row, col)

			if s.eliminateFromIntersection(tail1.Index(), tail2.Index(), -1, val, step) {
				s.applyStep(step.
					WithIndices(tail1.Index(), tail2.Index(), rCell.Index(), cCell.Index()))
				// Note: visualization usually highlights tails and connection cells.
				// We pass all 4.
				return true
			}
		}
		return false
	}

	// Try all combinations
	// Pair 1: (row[0], col[0]) connect?
	if checkPoly(rLocs[0], rLocs[1], cLocs[0], cLocs[1]) {
		return true
	}
	// Pair 2: (row[0], col[1]) connect?
	if checkPoly(rLocs[0], rLocs[1], cLocs[1], cLocs[0]) {
		return true
	}
	// Pair 3: (row[1], col[0]) connect?
	if checkPoly(rLocs[1], rLocs[0], cLocs[0], cLocs[1]) {
		return true
	}
	// Pair 4: (row[1], col[1]) connect?
	if checkPoly(rLocs[1], rLocs[0], cLocs[1], cLocs[0]) {
		return true
	}

	return false
}

// UNIQUENESS TECHNIQUES

func (s *Solver) findUniqueRectangleType1() (found bool) {
	b := s.puzzle
	// Check each cell with exactly 2 candidate values to see if it is the base
	// corner of a unique rectangle.
	for r := range 9 {
		for c := range 9 {
			// TODO: r,c or index?
			cell := b.Get(r, c)
			if cell.NumCandidates() != 2 {
				continue
			}
			if s.checkUniqueRectangleForCell(cell) {
				return true
			}
		}
	}

	return false
}

func (s *Solver) checkUniqueRectangleForCell(base *puzzle.Cell) (found bool) {
	b := s.puzzle

	// Look for a cell in the same row as base with the same pair of candidates.
	var rowWing *puzzle.Cell
	for c := range 9 {
		if c != base.Col {
			cell := b.Get(base.Row, c)
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
			cell := b.Get(r, base.Col)
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
		step := NewStep(kindUniqueRectangle1)
		if s.eliminateValuesFromCell(colWing.Row, rowWing.Col, base.Candidates, step) {
			s.applyStep(step.
				WithValues(base.Candidates.Values()...).
				WithIndices(base.Index(), rowWing.Index(), colWing.Index()))
			return true
		}
	}

	return false
}

// eliminateValuesFromCell removes all candidates listed in values from the cell
// at (r,c).
func (s *Solver) eliminateValuesFromCell(
	r, c int, values ValSet, step *SolutionStep,
) bool {
	cell := s.puzzle.Get(r, c)
	found := false
	for _, v := range values.Values() {
		if cell.HasCandidate(v) {
			step.DeleteCandidate(r*9+c, v)
			found = true
		}
	}
	return found
}

// BRUTE FORCE (LAST RESORT) TECHNIQUE

// findBruteForce uses a brute force search to find a solution for any remaining
// unsolved cells.  The search uses Donald Knuth's Algorithm X with the "Dancing
// Links" technique.
func (s *Solver) findBruteForce() bool {
	dl := NewDancingLinks(s.puzzle)
	dlOptions := &DancingLinksOptions{
		EnableDebug: s.EnableDebug,
		TimeLimit:   5 * time.Second,
	}

	solved, _ := dl.SolveWithStats(dlOptions)
	if solved {
		s.applyBruteForceSteps(dl)
	}

	return solved
}

func (s *Solver) applyBruteForceSteps(dl *DancingLinks) {
	steps := dl.GetSolution()
	for _, step := range steps {
		s.appendNextStep(NewStep(kindBruteForce).WithPlacedValue(step.Index, step.Value))
		// Place the value in the puzzle.
		s.puzzle.PlaceValue(step.Index, step.Value)
	}
}
