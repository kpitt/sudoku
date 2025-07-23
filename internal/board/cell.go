package board

import "github.com/kpitt/sudoku/internal/set"

type Cell struct {
	Row, Col int
	IsFixed  bool

	value      int8
	Candidates *set.Set[int8]
}

func NewCell(r, c int) *Cell {
	return &Cell{
		Row: r, Col: c,
		Candidates: set.NewSet[int8](1, 2, 3, 4, 5, 6, 7, 8, 9),
	}
}

func (c *Cell) IsLocked() bool {
	return c.value > 0
}

func (c *Cell) LockedValue() int8 {
	return c.value
}

func (c *Cell) LockValue(val int8) {
	c.value = val
	c.Candidates.Clear()
}

func (c *Cell) NumCandidates() int {
	return c.Candidates.Size()
}

func (c *Cell) CandidateValues() []int8 {
	return c.Candidates.Values()
}

func (c *Cell) HasCandidate(val int8) bool {
	return c.Candidates.Contains(val)
}

func (c *Cell) RemoveCandidate(val int8) {
	c.Candidates.Remove(val)
}

// House returns the index of the house that contains this cell.  Houses are
// numbered from left-to-right, top-to-bottom across the 3x3 grid of houses,
// with house 0 at the top-left and house 8 at the bottom-right.
func (c *Cell) House() int {
	return (c.Row - c.Row%3) + c.Col/3
}

// HouseCoordiinates returns the house coordinates of this cell.  The house
// coordinates consist of the house number, and the row and column in the 3x3
// grid of the house.
func (c *Cell) HouseCoordinates() (house, row, col int) {
	houseRow, houseCol := c.Row/3, c.Col/3
	house = houseRow*3 + houseCol
	baseRow, baseCol := houseRow*3, houseCol*3
	row, col = c.Row-baseRow, c.Col-baseCol
	return house, row, col
}

func (c *Cell) setFixedValue(val int8) {
	c.IsFixed = true
	c.LockValue(val)
}
