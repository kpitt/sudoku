package solver

import (
	"github.com/kpitt/sudoku/internal/bitset"
	"github.com/kpitt/sudoku/internal/puzzle"
)

// A House represents any row, column, or box that must contain each of the
// digits from 1 to 9.  The House maps each unsolved digit to the possible
// locations for that value, which makes it easier to check for certain
// patterns.
type House struct {
	Unsolved ValLocMap
	Cells    [9]*puzzle.Cell
	Kind     houseKind
	Index    int
}

type houseKind int

const (
	kindRow houseKind = iota
	kindColumn
	kindBox
)

// houseKindNames is a list of display names for each house kind.
// The order of the names must match the order of the houseKind constants.
var houseKindNames = []string{
	"Row",
	"Column",
	"Box",
}

// houseKindShortNames is a list of single-letter short names for each house
// kind.  These are used for formatting house references for solution steps.
// The order of the names must match the order of the houseKind constants.
var houseKindShortNames = []string{
	"r",
	"c",
	"b",
}

type UnsolvedFilter = func(int, LocSet) bool

func NewHouse(kind houseKind, index int) *House {
	h := &House{
		Kind:  kind,
		Index: index,
	}
	for i := range 9 {
		h.Unsolved[i+1] = bitset.BitSet16(allLocBits)
	}
	return h
}

func (h *House) Name() string {
	return houseKindNames[h.Kind]
}

// RemoveCandidateLoc removes loc from the candidate locations for value val.
func (h *House) RemoveCandidateLoc(val int, loc int) {
	h.Unsolved[val].Remove(loc)
}

// RemoveCandidateValue removes all candidate locations that conflict with a
// solved value of val in loc.
func (h *House) RemoveCandidateValue(val int, loc int) {
	// val is no longer an unsolved candidate for any cell in this house.
	h.Unsolved[val].Clear()
	// Location loc is solved, so no other value can be placed there.
	for i := 1; i <= 9; i++ {
		h.Unsolved[i].Remove(loc)
	}
}

func (h *House) NumUnsolved() int {
	count := 0
	for i := 1; i <= 9; i++ {
		if !h.Unsolved[i].Empty() {
			count++
		}
	}
	return count
}

func (h *House) UnsolvedDigits() []int {
	digits := make([]int, 0, 9)
	for i := 1; i <= 9; i++ {
		if !h.Unsolved[i].Empty() {
			digits = append(digits, i)
		}
	}
	return digits
}

func (h *House) NumLocations(val int) int {
	return h.Unsolved[val].Size()
}

func (h *House) Locations(val int) *LocSet {
	return &h.Unsolved[val]
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

func (h *House) indexesFromLocs(locs []int) []int {
	return transformSlice(locs, func(l int) int {
		return h.Cells[l].Index()
	})
}
