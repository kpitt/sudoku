package board

type Cell struct {
	value      int8
	candidates *Set[int8]
	isFixed    bool
}

func NewCell() *Cell {
	return &Cell{
		candidates: NewSet[int8](1, 2, 3, 4, 5, 6, 7, 8, 9),
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
}

func (c *Cell) FixValue(val int8) {
	c.value = val
	c.isFixed = true
}

func (c *Cell) NumCandidates() int {
	return c.candidates.Size()
}

func (c *Cell) Candidates() []int8 {
	return c.candidates.Elements()
}

func (c *Cell) IsCandidate(val int8) bool {
	return c.candidates.Contains(val)
}

func (c *Cell) RemoveCandidate(val int8) {
	c.candidates.Remove(val)
}
