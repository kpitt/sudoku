package solver

import "fmt"

// findHiddenSingles locks the value of any cells that match the "Hidden Single"
// pattern.  A "Hidden Single" is the only cell that contains a particular
// candidate in its row, column, or house.
func (s *Solver) findHiddenSingles() bool {
	printChecking("Hidden Single")
	found := false
	for i := range 9 {
		found = found ||
			s.checkHiddenSinglesForGroup(s.rowGroups[i]) ||
			s.checkHiddenSinglesForGroup(s.colGroups[i]) ||
			s.checkHiddenSinglesForGroup(s.houseGroups[i])
	}
	return found
}

func (s *Solver) checkHiddenSinglesForGroup(g *Group) bool {
	pattern := fmt.Sprintf("Hidden Single (%s)", g.GroupType)
	found := false
	for val, locs := range g.Unsolved() {
		if locs.Size() == 1 {
			index := locs.Values()[0]
			cell := g.Cells[index]
			s.LockValue(cell.Row, cell.Col, val, pattern)
			found = true
		}
	}
	return found
}

func (s *Solver) findNakedOrHiddenPairs() bool {
	found := false
	return found
}

func (s *Solver) findLockedCandidates() bool {
	found := false
	return found
}

func (s *Solver) findPointingTuples() bool {
	found := false
	return found
}

func (s *Solver) findNakedOrHiddenTriples() bool {
	found := false
	return found
}

func (s *Solver) findXWings() bool {
	found := false
	return found
}
