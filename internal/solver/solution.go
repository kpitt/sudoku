package solver

import (
	"fmt"
	"slices"
)

type (
	// A Candidate represents a single candidate identified by its row, column,
	// and value.
	Candidate struct {
		Row, Col int
		Value    int
	}

	// A SolutionStep represents a step in the solution process, which includes
	// the name of the technique and a list of candidates eliminated by this step.
	SolutionStep struct {
		name              string
		deletedCandidates []Candidate
	}
)

func NewSolutionStep(name string) *SolutionStep {
	return &SolutionStep{
		name:              name,
		deletedCandidates: make([]Candidate, 0),
	}
}

func (ss *SolutionStep) deleteCandidate(row, col, value int) {
	ss.deletedCandidates = append(ss.deletedCandidates, Candidate{Row: row, Col: col, Value: value})
}

func (ss *SolutionStep) Format() string {
	return fmt.Sprintf("%s: => %s", ss.name, ss.formatDeletedCandidates())
}

func (ss *SolutionStep) formatDeletedCandidates() string {
	if len(ss.deletedCandidates) == 0 {
		return ""
	}

	// First, organize the candidates into a map by value.
	values := make(map[int][]Candidate)
	for _, c := range ss.deletedCandidates {
		values[c.Value] = append(values[c.Value], c)
	}
	// Then, process the values in order and format each list of candidates.
	var result string
	orderedValues := mapKeys(values)
	slices.Sort(orderedValues)
	for i, v := range orderedValues {
		if i > 0 {
			result += ", "
		}
		result += formatCellsCompact(values[v])
		result += fmt.Sprintf("<>%d", v)
	}
	return result
}

func formatCellsCompact(cells []Candidate) string {
	if len(cells) == 0 {
		return ""
	}

	// Sort the cells by row and column for consistent formatting.
	if len(cells) > 1 {
		slices.SortFunc(cells, func(a, b Candidate) int {
			if a.Row != b.Row {
				return a.Row - b.Row
			}
			return a.Col - b.Col
		})
	}

	var result string
	for len(cells) > 0 {
		if len(result) > 0 {
			result += ","
		}
		if len(cells) == 1 {
			result += formatCellRef(cells[0].Row, cells[0].Col)
			break
		}

		remainingCells := make([]Candidate, 0, len(cells))
		rows, cols := []rune{}, []rune{}
		var row, col int
		appendRow := func(r int) {
			rows = append(rows, rune('1'+r))
		}
		appendCol := func(c int) {
			cols = append(cols, rune('1'+c))
		}
		for i, c := range cells {
			if i == 0 {
				// First cell
				appendRow(c.Row)
				appendCol(c.Col)
				row, col = c.Row, c.Col
			} else if c.Row == row && len(rows) == 1 {
				appendCol(c.Col)
			} else if c.Col == col && len(cols) == 1 {
				appendRow(c.Row)
			} else {
				remainingCells = append(remainingCells, c)
			}
		}
		result += "r" + string(rows) + "c" + string(cols)
		cells = remainingCells
	}
	return result
}
