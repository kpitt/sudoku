package board

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Solve attempts to solve a Sudoku puzzle by repeatedly applying known solving
// patterns to find solved cells and eliminate candidate values until the puzzle
// is completely solved, or until no more candidates can be eliminated (partial
// solution).
func (b *Board) Solve() {
	var pass int
	for !b.IsSolved() {
		pass = pass + 1
		color.HiYellow("Solver Pass %d:", pass)

		// "Naked Single" and "Hidden Single" are the only patterns that detect
		// an exact solution for a given cell.

		if b.findNakedSingles() {
			continue
		}
		if b.findHiddenSingles() {
			continue
		}

		// The remaining patterns are used to eliminate candidate values.
		// Patterns are processed in increasing order of how complex the pattern
		// is to detect.  If a pattern eliminates at least one candidate, then
		// we go back check the simplest patterns again.  Otherwise, we move on
		// to the next pattern.

		if b.findNakedOrHiddenPairs() {
			continue
		}
		if b.findLockedCandidates() {
			continue
		}
		if b.findPointingTuples() {
			continue
		}
		if b.findNakedOrHiddenTriples() {
			continue
		}
		if b.findXWings() {
			continue
		}

		// If we were unable to find any of the known patterns, then we've
		// eliminated as many candidates as we can.  All we can do now is break
		// out of the solver loop with a partial solution.
		break
	}
	color.HiYellow("Total Solver Passes: %d", pass)
}

// findNakedSingles locks the value of any cells that match the "Naked Single"
// pattern.  A "Naked Single" is a cell that has only one possible value.
func (b *Board) findNakedSingles() bool {
	found := false
	for r := range 9 {
		for c := range 9 {
			cell := b.cells[r][c]
			if !cell.IsLocked() && cell.NumCandidates() == 1 {
				val := cell.Candidates()[0]
				b.LockValue(r, c, val)
				found = true
				printFound("Naked Single", r, c, val)
			}
		}
	}
	return found
}

// findHiddenSingles locks the value of any cells that match the "Hidden Single"
// pattern.  A "Hidden Single" is the only cell that contains a particular
// candidate in its row, column, or house.
func (b *Board) findHiddenSingles() bool {
	found := false
	lockValue := func(r, c int, val int8, groupType string) {
		b.LockValue(r, c, val)
		found = true
		pattern := fmt.Sprintf("Hidden Single (%s)", groupType)
		printFound(pattern, r, c, val)
	}
	for i := range 9 {
		row := b.rowGroups[i]
		for val, locs := range row.Unsolved() {
			if locs.Size() == 1 {
				cols := locs.Values()
				lockValue(i, cols[0], val, "Row")
			}
		}
		col := b.colGroups[i]
		for val, locs := range col.Unsolved() {
			if locs.Size() == 1 {
				rows := locs.Values()
				lockValue(rows[0], i, val, "Column")
			}
		}
		house := b.houseGroups[i]
		for val, locs := range house.Unsolved() {
			if locs.Size() == 1 {
				cells := locs.Values()
				r, c := getHouseCellLoc(i, cells[0])
				lockValue(r, c, val, "House")
			}
		}
	}
	return found
}

func (b *Board) findNakedOrHiddenPairs() bool {
	found := false
	return found
}

func (b *Board) findLockedCandidates() bool {
	found := false
	return found
}

func (b *Board) findPointingTuples() bool {
	found := false
	return found
}

func (b *Board) findNakedOrHiddenTriples() bool {
	found := false
	return found
}

func (b *Board) findXWings() bool {
	found := false
	return found
}

func printFound(pattern string, r, c int, val int8) {
	fmt.Fprintf(os.Stderr, "%s: (%d,%d) = %d\n", pattern, r, c, val)
}
