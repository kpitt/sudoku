package solver

import (
	"fmt"
	"time"

	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/set"
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
		NumChecks int
		SolveTime time.Duration
	}

	Options struct {
		LiveLog     bool
		EnableDebug bool
	}
)

// Convenient type aliases that give semantic meaning to commonly used maps
// and sets.
type (
	LocSet    = *set.Set[int]
	ValSet    = *set.Set[int]
	LocValMap = map[int]ValSet
	ValLocMap = map[int]LocSet
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

func (s *Solver) initCandidates() {
	s.printProgress("Initializing solver candidates")
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
	defer s.solveTimer(time.Now())

	s.initCandidates()

	s.NumChecks = 0
SolverLoop:
	for !s.puzzle.IsSolved() {
		// Check techniques in roughly the order that a human solver would apply
		// them, starting with the simplest techniques and moving to more complex
		// ones.  If a check results in any change to the puzzle state, then
		// restart the solver loop.  Otherwise, move on to the next technique.

		for _, t := range s.techniques {
			if t.Check != nil {
				s.printChecking(t.Name)
				s.NumChecks += 1
				if t.Check() {
					continue SolverLoop
				}
			}
		}

		// If none of the known techniques allow us to eliminate any additional
		// candidates, then we've solved as much of the puzzle as we can, so
		// all we can do is exit with a partial solution.
		break
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
	peerCols := row.Locations(val)
	peerCols.Remove(c)
	col := s.columns[c]
	peerRows := col.Locations(val)
	peerRows.Remove(r)
	boxNum, boxLoc := getBoxLoc(r, c)
	box := s.boxes[boxNum]
	peerBoxLocs := box.Locations(val)
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

func (s *Solver) applyStep(step *SolutionStep) {
	s.solution = append(s.solution, step)
	if s.LiveLog {
		fmt.Printf("%2d. %s\n", len(s.solution), s.FormatStep(step))
	}
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
