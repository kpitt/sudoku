package board

type Board struct {
	cells         [9][9]*Cell
	unlockedCount int
}

func NewBoard() *Board {
	b := &Board{unlockedCount: 9 * 9}
	for r := range 9 {
		for c := range 9 {
			b.cells[r][c] = NewCell()
		}
	}
	return b
}

func (b *Board) IsSolved() bool {
	return b.unlockedCount == 0
}

func (b *Board) FixValue(r, c int, val int8) {
	b.cells[r][c].FixValue(val)
	b.removeCandidates(r, c, val)
}

func (b *Board) LockValue(r, c int, val int8) {
	b.cells[r][c].LockValue(val)
	b.removeCandidates(r, c, val)
}

func (b *Board) removeCandidates(r, c int, val int8) {
	b.unlockedCount = b.unlockedCount - 1
	b.removeRowCandidates(r, val)
	b.removeColCandidates(c, val)
	b.removeHouseCandidates(r, c, val)
}

func (b *Board) removeRowCandidates(row int, val int8) {
	for col := range 9 {
		b.cells[row][col].RemoveCandidate(val)
	}
}

func (b *Board) removeColCandidates(col int, val int8) {
	for row := range 9 {
		b.cells[row][col].RemoveCandidate(val)
	}
}

func (b *Board) removeHouseCandidates(row, col int, val int8) {
	houseRow, houseCol := row/3, col/3
	rowBase, colBase := houseRow*3, houseCol*3
	for r := range 3 {
		for c := range 3 {
			cell := b.cells[rowBase+r][colBase+c]
			cell.RemoveCandidate(val)
		}
	}
}

func getHouseIndex(row, col int) int {
	houseRow, houseCol := row/3, col/3
	return houseRow*3 + houseCol
}
