package solver

import (
	"fmt"
	"time"

	"github.com/kpitt/sudoku/internal/bitset"
	"github.com/kpitt/sudoku/internal/puzzle"
)

type (
	Solver struct {
		puzzle *puzzle.Puzzle
		*Options

		techniques []Technique

		rows    []*House
		columns []*House
		boxes   []*House

		// Many techniques need to be applied to all lines (rows and columns) or
		// all houses.  We can simplify those checks by precalulating a list for
		// each of those sets.
		lines  []*House // all rows and columns
		houses []*House // all rows, columns, and boxes

		solution []*SolutionStep

		// stats
		SolveTime time.Duration
	}

	// Options that control the behavior of the solver.
	// The options should be defined such that the zero-value represents the
	// default behavior.
	Options struct {
		LiveLog          bool // print step descriptions as they are found
		EnableDebug      bool // print additional progress info for debugging
		EnableBruteForce bool // enable brute-force search as a last resort
	}
)

// Convenient type aliases that give semantic meaning to commonly used maps
// and sets.
type (
	LocSet = bitset.BitSet16
	ValSet = bitset.BitSet16
	// ValLocMap maps a value to the set of locations where it is a candidate.
	// Index 0 is unused, indices 1-9 correspond to values 1-9.
	ValLocMap = [10]LocSet
)

const (
	allLocBits = 0b0111111111
)

func NewSolver(p *puzzle.Puzzle, opts *Options) *Solver {
	if opts == nil {
		opts = &Options{}
	}

	s := &Solver{puzzle: p, Options: opts}
	s.initTechniques()
	s.solution = make([]*SolutionStep, 0, 81)

	for i := range 9 {
		row := NewHouse(kindRow, i)
		s.rows = append(s.rows, row)
		s.lines = append(s.lines, row)
		s.houses = append(s.houses, row)
		col := NewHouse(kindColumn, i)
		s.columns = append(s.columns, col)
		s.lines = append(s.lines, col)
		s.houses = append(s.houses, col)
		box := NewHouse(kindBox, i)
		s.boxes = append(s.boxes, box)
		s.houses = append(s.houses, box)
	}

	// Collect the cells that belong to each house.
	for r := range 9 {
		for c := range 9 {
			cell := p.Grid[r][c]
			s.rows[r].Cells[c] = cell
			s.columns[c].Cells[r] = cell
			box, loc := getBoxLoc(r, c)
			s.boxes[box].Cells[loc] = cell
		}
	}

	return s
}

// processInitialValues applies the initial givens, and any other values that
// have already been placed, to the cached Solver candidates and processes any
// Naked Singles in the initial puzzle.
func (s *Solver) processInitialValues() {
	s.printProgress("Processing initial values")
	b := s.puzzle
	for r := range 9 {
		for c := range 9 {
			cell := b.Grid[r][c]
			if cell.IsSolved() {
				s.eliminateCandidates(r, c, cell.Value())
			}
		}
	}
}

// Solve attempts to solve a Sudoku puzzle by first applying known deductive
// solving techniques. If those don't solve the puzzle completely, it falls back
// to using the Dancing Links algorithm as a last resort.
func (s *Solver) Solve() {
	defer s.solveTimer(time.Now())

	s.processInitialValues()

SolverLoop:
	for !s.puzzle.IsSolved() {
		// Check techniques in roughly the order that a human solver would apply
		// them, starting with the simplest techniques and moving to more complex
		// ones. If a check results in any change to the puzzle state, then
		// restart the solver loop. Otherwise, move on to the next technique.

		for _, t := range s.techniques {
			if t.Check != nil {
				s.printChecking(t.Name)
				if t.Check() {
					continue SolverLoop
				}
			}
		}

		// If we get here, then the known techniques were not sufficient to find
		// a solution, so just exit with a partial solution.
		break
	}

	// If we didn't get a complete solution, use a "Dancing Links" brute-force
	// search as a last resort to solve the remaining cells.
	// Note that we could end up here because the puzzle has multiple solutions,
	// but it would be expensive to test every possible path to ensure that the
	// solution is unique. We just assume that the puzzle has only one solution,
	// and always use the first solution found.
	if !s.puzzle.IsSolved() && s.EnableBruteForce {
		s.printChecking("Brute Force")
		s.findBruteForce()
	}
}

func (s *Solver) solveTimer(start time.Time) {
	s.SolveTime = time.Since(start)
}

func (s *Solver) PlaceValue(r, c int, val int) {
	if s.puzzle.PlaceValue(r, c, val) {
		s.eliminateCandidates(r, c, val)
	}
}

// eliminateCandidates removes val from all cached candidates for the row,
// column, and box containing cell (r,c).
func (s *Solver) eliminateCandidates(r, c int, val int) {
	// Get the peer locations in the row, column, and box of cell (r,c) that
	// contain val as a candidate.
	row := s.rows[r]
	peerCols := *row.Locations(val)
	peerCols.Remove(c)
	col := s.columns[c]
	peerRows := *col.Locations(val)
	peerRows.Remove(r)
	boxNum, boxLoc := getBoxLoc(r, c)
	box := s.boxes[boxNum]
	peerBoxLocs := *box.Locations(val)
	peerBoxLocs.Remove(boxLoc)

	// Remove value from the cached candidates for the row, column, and box of
	// cell (r,c).
	row.RemoveCandidateValue(val, c)
	col.RemoveCandidateValue(val, r)
	box.RemoveCandidateValue(val, boxLoc)

	// Remove (r, c) as a candidate location for val in all peer cells.
	for pc := range peerCols.All() {
		s.removeCellCandidate(r, pc, val)
	}
	for pr := range peerRows.All() {
		s.removeCellCandidate(pr, c, val)
	}
	br, bc := getBoxBase(r, c)
	for pbl := range peerBoxLocs.All() {
		rb, cb := br+pbl/3, bc+pbl%3
		// Don't reprocess cells which are in the same row or column that we've already processed.
		if rb != r && cb != c {
			s.removeCellCandidate(rb, cb, val)
		}
	}
}

func (s *Solver) removeCellCandidate(r, c int, val int) {
	cell := s.puzzle.Grid[r][c]

	// Make sure val is removed from the candidates for this cell.
	cell.RemoveCandidate(val)

	// Also remove this cell from the cached locations for value.
	s.rows[r].RemoveCandidateLoc(val, c)
	s.columns[c].RemoveCandidateLoc(val, r)
	box, boxLoc := getBoxLoc(r, c)
	s.boxes[box].RemoveCandidateLoc(val, boxLoc)

	// A "Naked Single" is a cell that has only one possible value.
	// Checking for a "Naked Single" each time a candidate is removed narrows
	// down the possible options more quickly, and doesn't require iterating
	// over the entire puzzle grid at the start of each solver pass.
	if cell.NumCandidates() == 1 {
		step := NewStep(kindNakedSingle).
			WithPlacedValue(r, c, cell.CandidateValues()[0])
		s.applyStep(step)
	}
}

func (s *Solver) appendNextStep(step *SolutionStep) {
	s.solution = append(s.solution, step)
	if s.LiveLog {
		fmt.Printf("%2d. %s\n", len(s.solution), s.FormatStep(step))
	}
}

func (s *Solver) applyStep(step *SolutionStep) {
	s.appendNextStep(step)
	if step.IsSingle() {
		// Place the value for this step in the puzzle grid.
		index := step.indices[0]
		r, c := rowColFromIndex(index)
		s.PlaceValue(r, c, step.values[0])
	} else {
		// Apply the candidates eliminated by this step.
		for _, dc := range step.deletedCandidates {
			r, c := rowColFromIndex(dc.Index)
			s.removeCellCandidate(r, c, dc.Value)
		}
	}
}

// SolveBruteForce uses only the "Dancing Links" brute force search to solve
// the puzzle.  This is provided primarily for comparing the performance of
// the deductive solver against a pure brute-force approach.
func (s *Solver) SolveBruteForce() {
	defer s.solveTimer(time.Now())

	dl := NewDancingLinks(s.puzzle)
	dlOptions := &DancingLinksOptions{
		EnableDebug: s.EnableDebug,
		TimeLimit:   5 * time.Second,
	}

	solved, _ := dl.SolveWithStats(dlOptions)
	if solved {
		s.applyBruteForceSteps(dl)
	}
}
