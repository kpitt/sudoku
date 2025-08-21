package puzzle

import "github.com/kpitt/sudoku/internal/set"

type Cell struct {
	Row, Col int
	IsGiven  bool

	value      int8
	Candidates *set.Set[int8]
}

func NewCell(r, c int) *Cell {
	return &Cell{
		Row: r, Col: c,
		Candidates: set.NewSet[int8](1, 2, 3, 4, 5, 6, 7, 8, 9),
	}
}

// IsSolved returns true if a solved value has been placed in this cell.
func (c *Cell) IsSolved() bool {
	return c.value > 0
}

func (c *Cell) Value() int8 {
	return c.value
}

// PlaceValue places a solved value into the cell, clearing any remaining
// candidates.
func (c *Cell) PlaceValue(val int8) {
	c.value = val
	c.Candidates.Clear()
}

// GivenValue places an initial value into the cell, marking it as a given
// value that cannot be changed.  This is used for the initial puzzle setup.
func (c *Cell) GivenValue(val int8) {
	c.IsGiven = true
	c.PlaceValue(val)
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

// Box returns the index of the 3x3 box that contains this cell.  Boxes are
// numbered from left-to-right and top-to-bottom, with box 0 at the top-left
// and box 8 at the bottom-right.
func (c *Cell) Box() int {
	return (c.Row - c.Row%3) + c.Col/3
}

// BoxCoordinates returns the box coordinates of this cell, which consist of
// the index of the box that contains the cell and the index of the cell
// within the box.
func (c *Cell) BoxCoordinates() (box, index int) {
	box = c.Row/3*3 + c.Col/3
	row, col := c.Row%3, c.Col%3
	return box, row*3 + col
}

// SameCell returns true if the other Cell refers to the same cell location
// as this Cell (i.e. both cells have the same row and column).
func (c *Cell) SameCell(other *Cell) bool {
	return c.Row == other.Row && c.Col == other.Col
}
