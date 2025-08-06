package solver

import (
	"fmt"
	"slices"
)

type (
	// A Candidate represents a single candidate identified by its index and
	// value.  Index encodes the row and col of the candidate cell as a single
	// value in the range 0-80, where Index = row*9 + col.
	Candidate struct {
		Index int
		Value int
	}

	// A SolutionStep represents a step in the solution process, which includes
	// the name of the technique and a list of candidates eliminated by this step.
	SolutionStep struct {
		technique         techniqueKind
		house             *House
		values            []int
		indices           []int
		deletedCandidates []Candidate
	}
)

func NewSolutionStep(tk techniqueKind) *SolutionStep {
	return &SolutionStep{
		technique:         tk,
		deletedCandidates: make([]Candidate, 0),
	}
}

func (step *SolutionStep) WithHouse(h *House) *SolutionStep {
	step.house = h
	return step
}

func (step *SolutionStep) WithValue(v int) *SolutionStep {
	step.values = append(step.values, v)
	return step
}

func (step *SolutionStep) WithIndex(i int) *SolutionStep {
	step.indices = append(step.indices, i)
	return step
}

func (step *SolutionStep) WithPlacedValue(r, c int, v int) *SolutionStep {
	step.indices = []int{r*9 + c}
	step.values = []int{v}
	return step
}

func (step *SolutionStep) IsSingle() bool {
	return step.technique == kindNakedSingle || step.technique == kindHiddenSingle
}

func (step *SolutionStep) DeleteCandidate(row, col, value int) {
	step.deletedCandidates = append(step.deletedCandidates, Candidate{Index: row*9 + col, Value: value})
}

func (step *SolutionStep) Format() string {
	var name string
	if step.house != nil {
		name = houseCheckName(step.technique, step.house)
	} else {
		name = techniqueName(step.technique)
	}

	if step.IsSingle() {
		return fmt.Sprintf("%s: %s", name, step.formatPlacedValue())
	}
	return fmt.Sprintf("%s:%s", name, step.formatDeletedCandidates())
}

func (step *SolutionStep) formatPlacedValue() string {
	if len(step.values) == 0 || len(step.indices) == 0 {
		return ""
	}

	// index references a cell as a single value in the range 0-80, where
	// index = r*9 + c, so we need to convert it back to a row and column.
	return fmt.Sprintf("%s=%d", formatCellRef(step.indices[0]), step.values[0])
}

func (step *SolutionStep) formatDeletedCandidates() string {
	if len(step.deletedCandidates) == 0 {
		return ""
	}

	// First, organize the candidates into a map by value.
	values := make(map[int][]int)
	for _, c := range step.deletedCandidates {
		values[c.Value] = append(values[c.Value], c.Index)
	}
	// Then, process the values in order and format each list of candidates.
	orderedValues := mapKeys(values)
	slices.Sort(orderedValues)
	result := " => "
	for i, v := range orderedValues {
		if i > 0 {
			result += ", "
		}
		result += formatCellsCompact(values[v])
		result += fmt.Sprintf("<>%d", v)
	}
	return result
}

func formatCellsCompact(cells []int) string {
	if len(cells) == 0 {
		return ""
	}

	// Sort the cell indexes for consistent formatting.
	if len(cells) > 1 {
		slices.Sort(cells)
	}

	var result string
	for len(cells) > 0 {
		if len(result) > 0 {
			result += ","
		}

		// Short-circuit path: If there's only one cell, just format it directly.
		if len(cells) == 1 {
			result += formatCellRef(cells[0])
			break
		}

		remainingCells := make([]int, 0, len(cells))
		rows, cols := []rune{}, []rune{}
		appendRow := func(r int) {
			rows = append(rows, rune('1'+r))
		}
		appendCol := func(c int) {
			cols = append(cols, rune('1'+c))
		}
		var row, col int
		for i, cell := range cells {
			r, c := rowColFromIndex(cell)
			if i == 0 {
				// First cell
				row, col = r, c
				appendRow(row)
				appendCol(col)
			} else if r == row && len(rows) == 1 {
				appendCol(c)
			} else if c == col && len(cols) == 1 {
				appendRow(r)
			} else {
				// Cell is not in the same line as the first cell, so save it
				// for processing in the next pass.
				remainingCells = append(remainingCells, cell)
			}
		}
		result += "r" + string(rows) + "c" + string(cols)
		cells = remainingCells
	}

	return result
}

func techniqueName(tk techniqueKind) string {
	return techniqueNames[tk]
}

func houseCheckName(tk techniqueKind, h *House) string {
	return fmt.Sprintf("%s (%s)", techniqueNames[tk], h.Name())
}
