package solver

import "fmt"

// findNakedSingles locks the value of any cells that match the "Naked Single"
// pattern.  A "Naked Single" is a cell that has only one possible value.
func (s *Solver) findNakedSingles() bool {
	printChecking("Naked Single")
	b := s.board
	found := false
	for r := range 9 {
		for c := range 9 {
			cell := b.Cells[r][c]
			if !cell.IsLocked() && cell.NumCandidates() == 1 {
				val := cell.Candidates()[0]
				s.LockValue(r, c, val)
				found = true
				printFound("Naked Single", r, c, val)
			}
		}
	}
	return found
}

// findHiddenSingles locks the value of any cells that match the "Hidden Single"
// pattern.  A "Hidden Single" is the only cell that contains a particular
// candidate in its row, column, or house.
func (s *Solver) findHiddenSingles() bool {
	printChecking("Hidden Single")
	found := false
	lockValue := func(r, c int, val int8, groupType string) {
		s.LockValue(r, c, val)
		found = true
		pattern := fmt.Sprintf("Hidden Single (%s)", groupType)
		printFound(pattern, r, c, val)
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
