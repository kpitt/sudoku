package board

func (b *Board) Solve() {
	for !b.IsSolved() {
		if b.findNakedSingles() {
			continue
		}
		if b.findHiddenSingles() {
			continue
		}

		// If none of our known techniques was able to reduce the possible
		// candidates, then we've solved as much as we can, so all we can do
		// is break out of the solver loop.
		break
	}
}

func (b *Board) findNakedSingles() bool {
	found := false
	for r := range 9 {
		for c := range 9 {
			cell := b.cells[r][c]
			if !cell.IsLocked() && cell.NumCandidates() == 1 {
				val := cell.Candidates()[0]
				b.LockValue(r, c, val)
				found = true
			}
		}
	}
	return found
}

func (b *Board) findHiddenSingles() bool {
	found := false
	return found
}
