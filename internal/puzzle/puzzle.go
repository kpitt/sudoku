package puzzle

import (
	"fmt"
)

type Puzzle struct {
	Grid [9][9]*Cell

	// Holds counts of how many of each digit still needs to be placed.  If the
	// count for a digit reaches 0, then that digit is completely solved.
	// Index 0 holds the total count of unsolved grid cells.  When this value
	// reaches 0, the puzzle is completely solved.
	unsolvedCounts [10]int
}

func NewPuzzle() *Puzzle {
	p := &Puzzle{}
	for r := range 9 {
		for c := range 9 {
			p.Grid[r][c] = NewCell(r, c)
		}
	}

	for digit := range 10 {
		if digit == 0 {
			// Digit 0 represents the total count of unsolved cells.
			p.unsolvedCounts[digit] = 9 * 9
		} else {
			p.unsolvedCounts[digit] = 9
		}
	}

	return p
}

func (p *Puzzle) IsSolved() bool {
	return p.unsolvedCounts[0] == 0
}

func (p *Puzzle) IsDigitSolved(digit int) bool {
	return p.unsolvedCounts[digit] == 0
}

func (p *Puzzle) GivenValue(r, c int, val int) {
	p.Grid[r][c].GivenValue(val)
	p.updateUnsolvedCounts(r, c, val)
}

func (p *Puzzle) PlaceValue(r, c int, val int) bool {
	cell := p.Grid[r][c]
	if cell.IsSolved() {
		if cell.Value() != val {
			puzzleStateError(fmt.Sprintf("conflicting cell values %d and %d at (%d,%d)",
				cell.Value(), val, r+1, c+1))
		}
		return false
	}

	cell.PlaceValue(val)
	p.updateUnsolvedCounts(r, c, val)
	return true
}

func (p *Puzzle) updateUnsolvedCounts(r, c int, val int) {
	p.unsolvedCounts[0] = p.unsolvedCounts[0] - 1
	p.unsolvedCounts[val] = p.unsolvedCounts[val] - 1
	if p.unsolvedCounts[val] < 0 {
		puzzleStateError(fmt.Sprintf("too many instances of digit %d when placing cell (%d,%d)", val, r, c))
	}
}
