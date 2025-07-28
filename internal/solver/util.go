package solver

import (
	"github.com/kpitt/sudoku/internal/board"
	"github.com/kpitt/sudoku/internal/set"
)

func mapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func filterMap[K comparable, V any](
	m map[K]V, filter func(K, V) bool,
) map[K]V {
	filtered := make(map[K]V)
	for k, v := range m {
		if filter(k, v) {
			filtered[k] = v
		}
	}
	return filtered
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

// seesCell returns true if cell a sees cell b (i.e. they share a row, column,
// or house).
func seesCell(a, b *board.Cell) bool {
	// Two cells can see each other if they have the same row, the same column,
	// or the same house and they are not the same cell.
	return !a.SameCell(b) &&
		(a.Row == b.Row || a.Col == b.Col || a.House() == b.House())
}

// sameCandidates returns true if 2 cells a and b have exactly the same set of
// candidate values.
func sameCandidates(a, b *board.Cell) bool {
	sizeA := a.NumCandidates()
	if b.NumCandidates() != sizeA {
		return false
	}
	values := set.Union(a.Candidates, b.Candidates)
	return values.Size() == sizeA
}
