package solver

import (
	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/set"
)

// A House represents any row, column, or box that must contain each of the
// digits from 1 to 9.  The House maps each unsolved digit to the possible
// locations for that value, which makes it easier to check for certain
// patterns.
type House struct {
	Unsolved ValLocMap
	Cells    [9]*puzzle.Cell
	Type     string
	Index    int
}

type UnsolvedFilter = func(int, LocSet) bool

var emptyLocations = set.NewSet[int]()

func NewHouse(houseType string, index int) *House {
	h := &House{
		Unsolved: make(ValLocMap),
		Type:     houseType,
		Index:    index,
	}
	for i := range 9 {
		h.Unsolved[i+1] = set.NewSet(0, 1, 2, 3, 4, 5, 6, 7, 8)
	}
	return h
}

// RemoveCandidateCell removes cell from the candidate locations for value val.
func (h *House) RemoveCandidateCell(val int, cell int) {
	if cells := h.Unsolved[val]; cells != nil {
		cells.Remove(cell)
		if cells.Size() == 0 {
			delete(h.Unsolved, val)
		}
	}
}

// RemoveCandidateValue removes all candidate locations that conflict with a
// locked value of val in cell.
func (h *House) RemoveCandidateValue(val int, cell int) {
	// val is no longer an unsolved candidate for any cell in this house.
	delete(h.Unsolved, val)
	// If cell is locked, then no other value can appear in that location.
	for _, locs := range h.Unsolved {
		locs.Remove(cell)
	}
}

func (h *House) NumUnsolved() int {
	return len(h.Unsolved)
}

func (h *House) UnsolvedDigits() []int {
	digits := make([]int, 0, len(h.Unsolved))
	for k := range h.Unsolved {
		digits = append(digits, k)
	}
	return digits
}

func (h *House) NumLocations(val int) int {
	if loc, ok := h.Unsolved[val]; ok {
		return loc.Size()
	}
	return 0
}

func (h *House) Locations(val int) LocSet {
	if loc, ok := h.Unsolved[val]; ok {
		return loc
	}
	return emptyLocations
}

// sharedRow returns the row and true if all cells for the locations in locs
// are in the same row.  Otherwise, returns 0 and false.
func (h *House) sharedRow(locs LocSet) (row int, ok bool) {
	cells := h.cellsFromLocs(locs.Values())
	row = cells[0].Row
	for _, c := range cells[1:] {
		if c.Row != row {
			return 0, false
		}
	}
	return row, true
}

// sharedCol returns the column and true if all cells for the locations in locs
// are in the same column.  Otherwise, returns 0 and false.
func (h *House) sharedCol(locs LocSet) (col int, ok bool) {
	cells := h.cellsFromLocs(locs.Values())
	col = cells[0].Col
	for _, c := range cells[1:] {
		if c.Col != col {
			return 0, false
		}
	}
	return col, true
}

// sharedBox returns the box and true if all cells for the locations in locs
// are in the same box.  Otherwise, returns 0 and false.
func (h *House) sharedBox(locs LocSet) (box int, ok bool) {
	cells := h.cellsFromLocs(locs.Values())
	box = cells[0].Box()
	for _, c := range cells[1:] {
		if c.Box() != box {
			return 0, false
		}
	}
	return box, true
}

func (h *House) cellsFromLocs(locs []int) []*puzzle.Cell {
	return transformSlice(locs, func(l int) *puzzle.Cell {
		return h.Cells[l]
	})
}
