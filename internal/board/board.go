package board

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

type Board struct {
	cells         [9][9]*Cell
	unlockedCount int

	rowGroups   [9]*Group
	colGroups   [9]*Group
	houseGroups [9]*Group
}

func NewBoard() *Board {
	b := &Board{unlockedCount: 9 * 9}
	for r := range 9 {
		for c := range 9 {
			b.cells[r][c] = NewCell()
		}
	}

	for i := range 9 {
		b.rowGroups[i] = NewGroup()
		b.colGroups[i] = NewGroup()
		b.houseGroups[i] = NewGroup()
	}

	return b
}

func (b *Board) IsSolved() bool {
	return b.unlockedCount == 0
}

func (b *Board) FixValue(r, c int, val int8) {
	printFixed(r, c, val)
	b.LockValue(r, c, val)
	b.cells[r][c].IsFixed = true
}

func (b *Board) LockValue(r, c int, val int8) {
	cell := b.cells[r][c]
	if cell.IsLocked() {
		if cell.LockedValue() != val {
			boardStateError(fmt.Sprintf("conflicting locked values at (%d,%d)", r, c))
		}
		return
	}

	cell.LockValue(val)
	b.unlockedCount = b.unlockedCount - 1
	b.eliminateCandidates(r, c, val)
}

// eliminateCandidates removes val as a candidate value for row r, column c, and
// the house containing cell (r,c), and also removes cell (r,c) as a possible
// location for any other values in that row, column, and house.
func (b *Board) eliminateCandidates(r, c int, val int8) {
	// Remove value from the cached candidates for the row, column, and house
	// of cell (r,c).
	b.rowGroups[r].RemoveCandidateValue(val, c)
	b.colGroups[c].RemoveCandidateValue(val, r)
	house, houseCell, rowBase, colBase := getHouseInfo(r, c)
	b.houseGroups[house].RemoveCandidateValue(val, houseCell)

	for i := range 9 {
		b.removeCellCandidate(r, i, val) // remove candidate from row r
		b.removeCellCandidate(i, c, val) // remove candidate from column c
		// remove candidate from the house of (r,c)
		b.removeCellCandidate(rowBase+i/3, colBase+i%3, val)
	}
}

func (b *Board) removeCellCandidate(r, c int, val int8) {
	cell := b.cells[r][c]
	if cell.IsLocked() || !cell.IsCandidate(val) {
		return
	}

	// Remove val from the candidates for this cell.
	cell.RemoveCandidate(val)

	// Also remove this cell from the cached locations for value.
	b.rowGroups[r].RemoveCandidateCell(val, c)
	b.colGroups[c].RemoveCandidateCell(val, r)
	house, houseCell, _, _ := getHouseInfo(r, c)
	b.houseGroups[house].RemoveCandidateCell(val, houseCell)
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

func printFixed(r, c int, val int8) {
	msg := color.YellowString("Fixed Value: (%d,%d) = %d", r, c, val)
	fmt.Fprintln(os.Stderr, msg)
}
