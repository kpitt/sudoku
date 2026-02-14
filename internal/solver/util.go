package solver

import (
	"github.com/kpitt/sudoku/internal/puzzle"
)

func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func filterSlice[T any](s []T, filter func(T) bool) []T {
	var filtered []T
	for _, v := range s {
		if filter(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func transformSlice[TSource, TTarget any](
	source []TSource, transform func(TSource) TTarget,
) []TTarget {
	target := make([]TTarget, 0, len(source))
	for _, s := range source {
		target = append(target, transform(s))
	}
	return target
}

// getBoxLoc returns the box number and location within the box for the cell at
// row r and column c.
func getBoxLoc(r, c int) (box, loc int) {
	return r/3*3 + c/3, (r%3)*3 + c%3
}

// getBoxBase returns the row and column of the top-left cell in the box that
// contains the cell at (r, c).
func getBoxBase(r, c int) (rb, cb int) {
	return r / 3 * 3, c / 3 * 3
}

// sameCandidates returns true if 2 cells a and b have exactly the same set of
// candidate values.
func sameCandidates(a, b *puzzle.Cell) bool {
	return a.Candidates.Equal(b.Candidates)
}

func rowColFromIndex(index int) (row, col int) {
	return index / 9, index % 9
}
