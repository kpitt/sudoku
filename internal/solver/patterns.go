package solver

import "fmt"

// findHiddenSingles locks the value of any cells that match the "Hidden Single"
// pattern.  A "Hidden Single" is the only cell that contains a particular
// candidate in its row, column, or house.
func (s *Solver) findHiddenSingles() bool {
	printChecking("Hidden Single")
	found := false
	lockValue := func(r, c int, val int8, groupType string) {
		pattern := fmt.Sprintf("Hidden Single (%s)", groupType)
		s.LockValue(r, c, val, pattern)
		found = true
	}
	for i := range 9 {
		row := s.rowGroups[i]
		for val, locs := range row.Unsolved() {
			if locs.Size() == 1 {
				cols := locs.Values()
				lockValue(i, cols[0], val, "Row")
			}
		}
		col := s.colGroups[i]
		for val, locs := range col.Unsolved() {
			if locs.Size() == 1 {
				rows := locs.Values()
				lockValue(rows[0], i, val, "Column")
			}
		}
		house := s.houseGroups[i]
		for val, locs := range house.Unsolved() {
			if locs.Size() == 1 {
				cells := locs.Values()
				r, c := getHouseCellLoc(i, cells[0])
				lockValue(r, c, val, "House")
			}
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
