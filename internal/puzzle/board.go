package puzzle

import (
	"fmt"
)

type Board struct {
	Cells [9][9]*Cell

	// Holds counts of how many of each digit still needs to be placed.  If the
	// count for a digit reaches 0, then that digit is completely solved.
	// Index 0 holds the total count of unsolved cells on the board.  When this
	// value reaches 0, the puzzle is completely solved.
	unsolvedCounts [10]int
}

func NewBoard() *Board {
	b := &Board{}
	for r := range 9 {
		for c := range 9 {
			b.Cells[r][c] = NewCell(r, c)
		}
	}

	for digit := range 10 {
		if digit == 0 {
			// Digit 0 represents the total count of unsolved cells.
			b.unsolvedCounts[digit] = 9 * 9
		} else {
			b.unsolvedCounts[digit] = 9
		}
	}

	return b
}

func (b *Board) IsSolved() bool {
	return b.unsolvedCounts[0] == 0
}

func (b *Board) IsDigitSolved(digit int8) bool {
	return b.unsolvedCounts[digit] == 0
}

func (b *Board) setFixedValue(r, c int, val int8) {
	b.Cells[r][c].setFixedValue(val)
	b.updateUnsolvedCounts(r, c, val)
}

func (b *Board) LockValue(r, c int, val int8) bool {
	cell := b.Cells[r][c]
	if cell.IsLocked() {
		if cell.LockedValue() != val {
			boardStateError(fmt.Sprintf("conflicting locked values at (%d,%d)", r, c))
		}
		return false
	}

	cell.LockValue(val)
	b.updateUnsolvedCounts(r, c, val)
	return true
}

func (b *Board) updateUnsolvedCounts(r, c int, val int8) {
	b.unsolvedCounts[0] = b.unsolvedCounts[0] - 1
	b.unsolvedCounts[val] = b.unsolvedCounts[val] - 1
	if b.unsolvedCounts[val] < 0 {
		boardStateError(fmt.Sprintf("too many instances of digit %d when locking cell (%d,%d)", val, r, c))
	}
}
