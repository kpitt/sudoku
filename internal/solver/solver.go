package solver

import (
	"time"

	"github.com/fatih/color"
	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/set"
)

type Solver struct {
	puzzle *puzzle.Puzzle

	rows    []*House
	columns []*House
	boxes   []*House

	// Many techniques need to be applied to all lines (rows and columns) or
	// all houses.  We can simplify those checks by precalulating a list for
	// each of those sets.
	lines  []*House // all rows and columns
	houses []*House // all rows, columns, and boxes
}

// Convenient type aliases that give semantic meaning to commonly used maps
// and sets.
type (
	LocSet    = *set.Set[int]
	ValSet    = *set.Set[int]
	LocValMap = map[int]ValSet
	ValLocMap = map[int]LocSet
)

// Convenience alias for the signature of functions that check a technique.
type CheckFunc = func() (step *SolutionStep, found bool)

func NewSolver(p *puzzle.Puzzle) *Solver {
	s := &Solver{puzzle: p}

	for i := range 9 {
		row := NewHouse("Row", i)
		s.rows = append(s.rows, row)
		s.lines = append(s.lines, row)
		s.houses = append(s.houses, row)
		col := NewHouse("Column", i)
		s.columns = append(s.columns, col)
		s.lines = append(s.lines, col)
		s.houses = append(s.houses, col)
		box := NewHouse("Box", i)
		s.boxes = append(s.boxes, box)
		s.houses = append(s.houses, box)
	}

	// Collect the cells that belong to each house.
	for r := range 9 {
		for c := range 9 {
			cell := p.Grid[r][c]
			s.rows[r].Cells[c] = cell
			s.columns[c].Cells[r] = cell
			box, index := cell.BoxCoordinates()
			s.boxes[box].Cells[index] = cell
		}
	}

	s.initializeCandidates()

	return s
}

func (s *Solver) initializeCandidates() {
	printProgress("Processing given values")
	b := s.puzzle
	for r := range 9 {
		for c := range 9 {
			cell := b.Grid[r][c]
			if cell.IsGiven {
				s.eliminateCandidates(r, c, cell.Value())
			}
		}
	}
}

// Solve attempts to solve a Sudoku puzzle by repeatedly applying known deductive
// solving techniques to find solved cells and eliminate candidate values until
// the puzzle is completely solved, or until no more candidates can be eliminated
// (partial solution).
func (s *Solver) Solve() {
	defer solveTimer(time.Now())

	b := s.puzzle

	// List of known technique checks in the order they should be applied.
	checks := []CheckFunc{
		s.findNakedPairs,
		s.findLockedCandidates,
		s.findPointingTuples,
		s.findHiddenPairs,
		s.findNakedTriples,
		s.findXWings,
		s.findHiddenTriples,
		s.findNakedQuadruples,
		s.findXYWings,
		s.findAvoidableRectangles,
		s.findXYZWings,
		s.findHiddenQuadruples,
		s.findUniqueRectangles,
		s.findSwordfish,
		s.findJellyfish,
	}

	var pass int
SolverLoop:
	for !b.IsSolved() {
		pass = pass + 1
		color.HiYellow("Solver Pass %d:", pass)

		// "Naked Single" and "Hidden Single" are the only techniques that
		// provide an exact solution for a given cell.  The Naked Single
		// technique is applied each time a candidate is removed from a cell,
		// so we only need to look for Hidden Singles here.

		if s.findHiddenSingles() {
			continue
		}

		// The remaining techniques are used to eliminate candidate values until
		// a Naked or Hidden Single is reached.  Techniques are checked in
		// roughly the order that a human solver would apply them, starting
		// with the simplest techniques and moving to more complex ones.  If a
		// technique eliminates at least one candidate, then we start again with
		// the simplest checks.  Otherwise, we move on to try the next
		// technique.

		for _, check := range checks {
			if step, found := check(); found {
				printStep(step)
				s.applyStepCandidates(step)
				continue SolverLoop
			}
		}

		// If none of the known techniques allow us to eliminate any additional
		// candidates, then we've solved as much of the puzzle as we can, so
		// all we can do is exit with a partial solution.
		break
	}

	color.HiYellow("Total Solver Passes: %d", pass)
}

func solveTimer(start time.Time) {
	elapsed := time.Since(start)
	color.HiYellow("Total Solver Time:   %v", elapsed)
}

func (s *Solver) PlaceValue(r, c int, val int, technique string) {
	if s.puzzle.PlaceValue(r, c, val) {
		printFound(technique, r, c, val)
		s.eliminateCandidates(r, c, val)
	}
}

// eliminateCandidates removes val as a candidate value for row r, column c, and
// the box containing cell (r,c).  It also removes cell (r,c) as a possible
// location for any other values in the same row, column, or box.
func (s *Solver) eliminateCandidates(r, c int, val int) {
	// Remove value from the cached candidates for the row, column, and box of
	// cell (r,c).
	s.rows[r].RemoveCandidateValue(val, c)
	s.columns[c].RemoveCandidateValue(val, r)
	box, boxCell, rowBase, colBase := getBoxInfo(r, c)
	s.boxes[box].RemoveCandidateValue(val, boxCell)

	for i := range 9 {
		s.removeCellCandidate(r, i, val) // remove candidate from row r
		s.removeCellCandidate(i, c, val) // remove candidate from column c
		// remove candidate from the box that contains (r,c)
		s.removeCellCandidate(rowBase+i/3, colBase+i%3, val)
	}
}

func (s *Solver) removeCellCandidate(r, c int, val int) {
	b := s.puzzle
	cell := b.Grid[r][c]
	if cell.IsSolved() || !cell.HasCandidate(val) {
		return
	}

	// Remove val from the candidates for this cell.
	cell.RemoveCandidate(val)

	// Also remove this cell from the cached locations for value.
	s.rows[r].RemoveCandidateCell(val, c)
	s.columns[c].RemoveCandidateCell(val, r)
	box, boxCell := cell.BoxCoordinates()
	s.boxes[box].RemoveCandidateCell(val, boxCell)

	// A "Naked Single" is a cell that has only one possible value.
	// Checking for a "Naked Single" each time a candidate is removed narrows
	// down the possible options more quickly, and doesn't require iterating
	// over the entire puzzle grid at the start of each solver pass.
	if cell.NumCandidates() == 1 {
		s.PlaceValue(r, c, cell.CandidateValues()[0], "Naked Single")
	}
}

func (s *Solver) applyStepCandidates(ss *SolutionStep) {
	// Apply the candidates eliminated by this step.
	for _, c := range ss.deletedCandidates {
		s.removeCellCandidate(c.Row, c.Col, c.Value)
	}
}

func getBoxInfo(r, c int) (box, cellIndex, baseRow, baseCol int) {
	boxRow, boxCol := r/3, c/3
	box = boxRow*3 + boxCol
	baseRow, baseCol = boxRow*3, boxCol*3
	cellIndex = (r%3)*3 + c%3
	return box, cellIndex, baseRow, baseCol
}
