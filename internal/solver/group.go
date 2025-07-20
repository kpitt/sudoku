package solver

import "github.com/kpitt/sudoku/internal/set"

// A Group represents any row, column, or house that must contain each of the
// digits from 1 to 9.  Each Group caches information about the remaining cells
// that are possible locations for each digit, which makes it easier to check
// for certain patterns.
type Group struct {
	unsolved map[int8]LocSet
}

type LocSet = *set.Set[int]

var emptyLocations = set.NewSet[int]()

func NewGroup() *Group {
	g := &Group{
		unsolved: make(map[int8]LocSet),
	}
	for v := int8(1); v <= 9; v = v + 1 {
		g.unsolved[v] = set.NewSet(0, 1, 2, 3, 4, 5, 6, 7, 8)
	}
	return g
}

// RemoveCandidateCell removes cell from the candidate locations for value val.
func (g *Group) RemoveCandidateCell(val int8, cell int) {
	if cells := g.unsolved[val]; cells != nil {
		cells.Remove(cell)
		if cells.Size() == 0 {
			delete(g.unsolved, val)
		}
	}
}

// RemoveCandidateValue removes all candidate locations that conflict with a
// locked value of val in cell.
func (g *Group) RemoveCandidateValue(val int8, cell int) {
	// val is no longer an unsolved candidate for any cell in this group.
	delete(g.unsolved, val)
	// If cell is locked, then no other value can appear in that location.
	for _, locs := range g.unsolved {
		locs.Remove(cell)
	}
}

func (g *Group) Unsolved() map[int8]LocSet {
	return g.unsolved
}

func (g *Group) NumUnsolved() int {
	return len(g.unsolved)
}

func (g *Group) UnsolvedDigits() []int8 {
	digits := make([]int8, 0, len(g.unsolved))
	for k := range g.unsolved {
		digits = append(digits, k)
	}
	return digits
}

func (g *Group) NumLocations(val int8) int {
	if loc, ok := g.unsolved[val]; ok {
		return loc.Size()
	}
	return 0
}

func (g *Group) Locations(val int8) LocSet {
	if loc, ok := g.unsolved[val]; ok {
		return loc
	}
	return emptyLocations
}
