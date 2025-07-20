package board

import "github.com/kpitt/sudoku/internal/set"

type Cell struct {
	IsFixed bool

	value      int8
	candidates *set.Set[int8]
}

func NewCell() *Cell {
	return &Cell{
		candidates: set.NewSet[int8](1, 2, 3, 4, 5, 6, 7, 8, 9),
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
	c.candidates.Clear()
}

func (c *Cell) NumCandidates() int {
	return c.candidates.Size()
}

func (c *Cell) Candidates() []int8 {
	return c.candidates.Values()
}

func (c *Cell) IsCandidate(val int8) bool {
	return c.candidates.Contains(val)
}

func (c *Cell) RemoveCandidate(val int8) {
	c.candidates.Remove(val)
}

func (c *Cell) setFixedValue(val int8) {
	c.IsFixed = true
	c.LockValue(val)
}
