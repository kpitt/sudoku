package solver

import (
	"github.com/fatih/color"
	"github.com/kpitt/sudoku/internal/board"
)

type Solver struct {
	board *board.Board

	rowGroups   [9]*Group
	colGroups   [9]*Group
	houseGroups [9]*Group
}

func NewSolver(b *board.Board) *Solver {
	s := &Solver{board: b}

	for i := range 9 {
		s.rowGroups[i] = NewGroup()
		s.colGroups[i] = NewGroup()
		s.houseGroups[i] = NewGroup()
	}

	s.initializeCandidates()

	return s
}

func (s *Solver) initializeCandidates() {
	printProgress("Processing initial board state")
	b := s.board
	for r := range 9 {
		for c := range 9 {
			cell := b.Cells[r][c]
			if cell.IsFixed {
				s.eliminateCandidates(r, c, cell.LockedValue())
			}
		}
	}
}

// Solve attempts to solve a Sudoku puzzle by repeatedly applying known solving
// patterns to find solved cells and eliminate candidate values until the puzzle
// is completely solved, or until no more candidates can be eliminated (partial
// solution).
func (s *Solver) Solve() {
	b := s.board

	var pass int
	for !b.IsSolved() {
		pass = pass + 1
		color.HiYellow("Solver Pass %d:", pass)

		// "Naked Single" and "Hidden Single" are the only patterns that detect
		// an exact solution for a given cell.  The "Naked Single" pattern is
		// checked each time a candidate is removed from a cell, so we only
		// need to look for the "Hidden Single" pattern here.

		if s.findHiddenSingles() {
			continue
		}

		// The remaining patterns are used to eliminate candidate values.
		// Patterns are processed in increasing order of how complex the pattern
		// is to detect.  If a pattern eliminates at least one candidate, then
		// we go back check the simplest patterns again.  Otherwise, we move on
		// to the next pattern.

		if s.findNakedOrHiddenPairs() {
			continue
		}
		if s.findLockedCandidates() {
			continue
		}
		if s.findPointingTuples() {
			continue
		}
		if s.findNakedOrHiddenTriples() {
			continue
		}
		if s.findXWings() {
			continue
		}

		// If we were unable to find any of the known patterns, then we've
		// eliminated as many candidates as we can.  All we can do now is break
		// out of the solver loop with a partial solution.
		break
	}
	color.HiYellow("Total Solver Passes: %d", pass)
}

func (s *Solver) LockValue(r, c int, val int8, pattern string) {
	if s.board.LockValue(r, c, val) {
		printFound(pattern, r, c, val)
		s.eliminateCandidates(r, c, val)
	}
}

// eliminateCandidates removes val as a candidate value for row r, column c, and
// the house containing cell (r,c), and also removes cell (r,c) as a possible
// location for any other values in that row, column, and house.
func (s *Solver) eliminateCandidates(r, c int, val int8) {
	// Remove value from the cached candidates for the row, column, and house
	// of cell (r,c).
	s.rowGroups[r].RemoveCandidateValue(val, c)
	s.colGroups[c].RemoveCandidateValue(val, r)
	house, houseCell, rowBase, colBase := getHouseInfo(r, c)
	s.houseGroups[house].RemoveCandidateValue(val, houseCell)

	for i := range 9 {
		s.removeCellCandidate(r, i, val) // remove candidate from row r
		s.removeCellCandidate(i, c, val) // remove candidate from column c
		// remove candidate from the house of (r,c)
		s.removeCellCandidate(rowBase+i/3, colBase+i%3, val)
	}
}

func (s *Solver) removeCellCandidate(r, c int, val int8) {
	b := s.board
	cell := b.Cells[r][c]
	if cell.IsLocked() || !cell.IsCandidate(val) {
		return
	}

	// Remove val from the candidates for this cell.
	cell.RemoveCandidate(val)

	// Also remove this cell from the cached locations for value.
	s.rowGroups[r].RemoveCandidateCell(val, c)
	s.colGroups[c].RemoveCandidateCell(val, r)
	house, houseCell, _, _ := getHouseInfo(r, c)
	s.houseGroups[house].RemoveCandidateCell(val, houseCell)

	// A "Naked Single" is a cell that has only one possible value.
	// Checking for a "Naked Single" each time a candidate is removed narrows
	// down the possible options more quickly, and doesn't require iterating
	// over the entire board at the start of each solver pass.
	if cell.NumCandidates() == 1 {
		s.LockValue(r, c, cell.Candidates()[0], "Naked Single")
	}
}

func getHouseInfo(row, col int) (houseIndex, cellIndex, baseRow, baseCol int) {
	houseRow, houseCol := row/3, col/3
	houseIndex = houseRow*3 + houseCol
	baseRow, baseCol = houseRow*3, houseCol*3
	cellIndex = (row-baseRow)*3 + (col - baseCol)
	return houseIndex, cellIndex, baseRow, baseCol
}

func getHouseCellLoc(houseIndex, cellIndex int) (row, col int) {
	houseRow, houseCol := houseIndex/3, houseIndex%3
	cellRow, cellCol := cellIndex/3, cellIndex%3
	return houseRow*3 + cellRow, houseCol*3 + cellCol
}
