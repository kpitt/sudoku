package solver

import (
	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/set"
)

// A Group represents any row, column, or house that must contain each of the
// digits from 1 to 9.  Each Group caches information about the remaining cells
// that are possible locations for each digit, which makes it easier to check
// for certain patterns.
type Group struct {
	Unsolved   map[int8]LocSet
	Cells      [9]*puzzle.Cell
	GroupType  string
	BoardIndex int
}

type UnsolvedFilter = func(int8, LocSet) bool

var emptyLocations = set.NewSet[int]()

func NewGroup(groupType string, index int) *Group {
	g := &Group{
		Unsolved:   make(map[int8]LocSet),
		GroupType:  groupType,
		BoardIndex: index,
	}
	for i := range 9 {
		g.Unsolved[int8(i+1)] = set.NewSet(0, 1, 2, 3, 4, 5, 6, 7, 8)
	}
	return g
}

// RemoveCandidateCell removes cell from the candidate locations for value val.
func (g *Group) RemoveCandidateCell(val int8, cell int) {
	if cells := g.Unsolved[val]; cells != nil {
		cells.Remove(cell)
		if cells.Size() == 0 {
			delete(g.Unsolved, val)
		}
	}
}

// RemoveCandidateValue removes all candidate locations that conflict with a
// locked value of val in cell.
func (g *Group) RemoveCandidateValue(val int8, cell int) {
	// val is no longer an unsolved candidate for any cell in this group.
	delete(g.Unsolved, val)
	// If cell is locked, then no other value can appear in that location.
	for _, locs := range g.Unsolved {
		locs.Remove(cell)
	}
}

func (g *Group) NumUnsolved() int {
	return len(g.Unsolved)
}

func (g *Group) UnsolvedDigits() []int8 {
	digits := make([]int8, 0, len(g.Unsolved))
	for k := range g.Unsolved {
		digits = append(digits, k)
	}
	return digits
}

func (g *Group) NumLocations(val int8) int {
	if loc, ok := g.Unsolved[val]; ok {
		return loc.Size()
	}
	return 0
}

func (g *Group) Locations(val int8) LocSet {
	if loc, ok := g.Unsolved[val]; ok {
		return loc
	}
	return emptyLocations
}

// sharedRow returns the row and true if all cells for the locations in locs
// are in the same row.  Otherwise, returns 0 and false.
func (g *Group) sharedRow(locs LocSet) (row int, ok bool) {
	cells := g.cellsFromLocs(locs.Values())
	row = cells[0].Row
	for _, c := range cells[1:] {
		if c.Row != row {
			return 0, false
		}
	}
	return row, true
}

// sharedCol returns the column and true if all cells for the locations in locs
// are in the same row.  Otherwise, returns 0 and false.
func (g *Group) sharedCol(locs LocSet) (col int, ok bool) {
	cells := g.cellsFromLocs(locs.Values())
	col = cells[0].Col
	for _, c := range cells[1:] {
		if c.Col != col {
			return 0, false
		}
	}
	return col, true
}

// sharedHouse returns the house and true if all cells for the locations in locs
// are in the same row.  Otherwise, returns 0 and false.
func (g *Group) sharedHouse(locs LocSet) (house int, ok bool) {
	cells := g.cellsFromLocs(locs.Values())
	house = cells[0].House()
	for _, c := range cells[1:] {
		if c.House() != house {
			return 0, false
		}
	}
	return house, true
}

func (g *Group) cellsFromLocs(locs []int) []*puzzle.Cell {
	return transformSlice(locs, func(l int) *puzzle.Cell {
		return g.Cells[l]
	})
}
