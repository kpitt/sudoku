package solver

import (
	"fmt"
	"slices"

	"github.com/kpitt/sudoku/internal/puzzle"
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
		bases             []*House
		covers            []*House
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

func (step *SolutionStep) WithValues(values ...int) *SolutionStep {
	step.values = append(step.values, values...)
	return step
}

func (step *SolutionStep) WithIndices(indices ...int) *SolutionStep {
	step.indices = append(step.indices, indices...)
	return step
}

func (step *SolutionStep) WithCells(cells ...*puzzle.Cell) *SolutionStep {
	for _, cell := range cells {
		step.indices = append(step.indices, cell.Row*9+cell.Col)
	}
	return step
}

func (step *SolutionStep) WithPlacedValue(r, c int, v int) *SolutionStep {
	step.indices = []int{r*9 + c}
	step.values = []int{v}
	return step
}

func (step *SolutionStep) WithBases(bases ...*House) *SolutionStep {
	step.bases = append(step.bases, bases...)
	return step
}

func (step *SolutionStep) WithCovers(covers ...*House) *SolutionStep {
	step.covers = append(step.covers, covers...)
	return step
}

func (step *SolutionStep) IsSingle() bool {
	return step.technique == kindNakedSingle || step.technique == kindHiddenSingle
}

func (step *SolutionStep) DeleteCandidate(row, col, value int) {
	step.deletedCandidates = append(step.deletedCandidates, Candidate{Index: row*9 + col, Value: value})
}

func (step *SolutionStep) Format() string {
	switch step.technique {
	case kindNakedSingle:
		return step.formatNakedSingle()

	case kindHiddenSingle:
		return step.formatHiddenSingle()

	case kindNakedPair,
		kindNakedTriple,
		kindNakedQuadruple,
		kindHiddenPair,
		kindHiddenTriple,
		kindHiddenQuadruple:

		return step.formatNakedOrHiddenSubset()

	case kindLockedCandidate,
		kindPointingTuple:

		return step.formatLockedCandidate()

	case kindXWing,
		kindSwordfish,
		kindJellyfish,
		kindFinnedXWing,
		kindFinnedSwordfish,
		kindFinnedJellyfish:

		return step.formatFish()

	case kindXYWing,
		kindXYZWing:

		return step.formatXYZWing()

	case kindAvoidableRectangle,
		kindUniqueRectangle,
		kindHiddenRectangle,
		kindPointingRectangle:

		return step.formatUniqueRectangle()

	case kindEmptyRectangle:

		return step.formatGeneric()

	case kindSkyscraper,
		kindTwoStringKite,
		kindColorChain:

		return step.formatGeneric()
	}

	return step.stepName()
}

func (step *SolutionStep) stepName() string {
	return techniqueName(step.technique)
}

func (step *SolutionStep) formatNamedStep(desc string) string {
	if desc == "" {
		return step.stepName()
	}
	return fmt.Sprintf("%s: %s", step.stepName(), desc)
}

func (step *SolutionStep) formatNakedSingle() string {
	return step.formatNamedStep(step.formatPlacedValue())
}

func (step *SolutionStep) formatHiddenSingle() string {
	desc := fmt.Sprintf("%d in %s => %s",
		step.values[0], formatHouse(step.house), step.formatPlacedValue())
	return step.formatNamedStep(desc)
}

func (step *SolutionStep) formatPlacedValue() string {
	if len(step.values) == 0 || len(step.indices) == 0 {
		return ""
	}

	// index references a cell as a single value in the range 0-80, where
	// index = r*9 + c, so we need to convert it back to a row and column.
	return fmt.Sprintf("%s=%d", formatCell(step.indices[0]), step.values[0])
}

func (step *SolutionStep) formatElimination(format string, a ...any) string {
	var desc string
	if format == "" {
		desc = step.formatDeletedCandidates()
	} else {
		desc = fmt.Sprintf("%s => %s",
			fmt.Sprintf(format, a...), step.formatDeletedCandidates())
	}
	return step.formatNamedStep(desc)
}

func (step *SolutionStep) formatGeneric() string {
	return step.formatElimination("")
}

func (step *SolutionStep) formatNakedOrHiddenSubset() string {
	return step.formatElimination("%s in %s",
		step.formatValuesList(), step.formatIndices())
}

func (step *SolutionStep) formatLockedCandidate() string {
	return step.formatElimination("%d in %s",
		step.values[0], formatHouse(step.house))
}

func (step *SolutionStep) formatFish() string {
	return step.formatElimination("%d %s %s", step.values[0],
		formatHouses(step.bases), formatHouses(step.covers))
}

func (step *SolutionStep) formatXYZWing() string {
	// For an XY-Wing or XYZ-Wing, the eliminated z value should be the last
	// digit in the value sequence, and the x,y values should be in sorted
	// order.  To accomplish this, we first get the z value from the deleted
	// candidates.  We then swap the z value to the last position in the values
	// slice, and sort the rest of the values.
	z := step.deletedCandidates[0].Value
	for i := range 2 {
		if step.values[i] == z {
			// Swap the z value to the end of the values slice.
			step.values[i], step.values[2] = step.values[2], step.values[i]
			break
		}
	}
	// Sort the x and y values in ascending order.
	slices.Sort(step.values[:2])
	return step.formatElimination("%s in %s",
		step.formatValuesWing(), step.formatIndices())
}

func (step *SolutionStep) formatUniqueRectangle() string {
	return step.formatElimination("%s in %s",
		step.formatValuesWing(), step.formatRectIndices())
}

func (step *SolutionStep) formatDeletedCandidates() string {
	// First, organize the candidates into a map by value.
	locsByValue := make(map[int][]int)
	for _, c := range step.deletedCandidates {
		locsByValue[c.Value] = append(locsByValue[c.Value], c.Index)
	}
	// Then, process the values in order and format each list of candidates.
	orderedValues := mapKeys(locsByValue)
	slices.Sort(orderedValues)
	var result string
	for i, v := range orderedValues {
		if i > 0 {
			result += ", "
		}
		result += formatCellsCompact(locsByValue[v])
		result += fmt.Sprintf("<>%d", v)
	}
	return result
}

// formatValuesList formats all values of the step as a comma-separated list in
// sorted order.
func (step *SolutionStep) formatValuesList() string {
	slices.Sort(step.values)
	return formatDigitsSeparated(step.values, ',')
}

// formatValuesWing formats all values of the step as a slash-separated wing
// sequence.  Values are assumed to already be in the desired order.
func (step *SolutionStep) formatValuesWing() string {
	return formatDigitsSeparated(step.values, '/')
}

// formatIndices formats all indices of the step as a compact cell-reference
// string, e.g. "r1c12,r3c1".
func (step *SolutionStep) formatIndices() string {
	return formatCellsCompact(step.indices)
}

// formatRectIndices formats the indices of the step as a condensed rectangle
// string, e.g. "r13c12" (equivalent to "r1c12,r3c12").
func (step *SolutionStep) formatRectIndices() string {
	return formatRectCompact(step.indices)
}

func formatCell(index int) string {
	r, c := rowColFromIndex(index)
	return fmt.Sprintf("r%dc%d", r+1, c+1)
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
			result += formatCell(cells[0])
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

func formatRectCompact(cells []int) string {
	// A rectangle is defined by either 3 or 4 cells.
	// For anything else, fallback to the normal notation.
	if len(cells) != 3 && len(cells) != 4 {
		return formatCellsCompact(cells)
	}

	// This function assumes the cells form a valid rectangle, so any three
	// cells are sufficient to define the rectangle. If there is a 4th cell,
	// it is ignored.

	r1, c1 := rowColFromIndex(cells[0])
	r2, c2 := rowColFromIndex(cells[1])
	r3, c3 := rowColFromIndex(cells[2])
	// Make sure r1 < r2 and c1 < c2.
	if r1 == r2 {
		if r2 < r3 {
			r2 = r3
		} else {
			r1 = r3
		}
	}
	if c1 == c2 {
		if c2 < c3 {
			c2 = c3
		} else {
			c1 = c3
		}
	}

	return fmt.Sprintf("r%d%dc%d%d", r1+1, r2+1, c1+1, c2+1)
}

func formatHouse(h *House) string {
	return fmt.Sprintf("%s%d", houseKindShortNames[h.Kind], h.Index+1)
}

// formatHouses formats a compact representation of a list of houses in the
// form "h123", where "h" is the house kind ("r" for row, "c" for column,
// "b" for box), and "123" is the concatenation of the indices of the houses.
// This assumes that all houses are of the same kind, which is typical for the
// base and cover sets used in the various "Fish" techniques.  If houses are of
// different kinds, then only the first kind will be formatted, and the rest
// will be ignored.
func formatHouses(houses []*House) string {
	// Sort the houses by their kind and index for consistent formatting.
	slices.SortFunc(houses, func(a, b *House) int {
		if a.Kind != b.Kind {
			return int(a.Kind) - int(b.Kind)
		}
		return a.Index - b.Index
	})

	kind := houses[0].Kind
	digits := make([]int, 0, len(houses))
	for _, h := range houses {
		if h.Kind != kind {
			break // We only format houses of the same kind.
		}
		digits = append(digits, h.Index+1) // Store indices as 1-based.
	}
	return fmt.Sprintf("%s%s", houseKindShortNames[kind], formatDigitsCompact(digits))
}

func formatDigitsCompact(digits []int) string {
	// Sort the digits for consistent formatting.
	slices.Sort(digits)

	result := make([]rune, 0, len(digits))
	for _, d := range digits {
		result = append(result, rune('0'+d))
	}
	return string(result)
}

// formatDigitsSeparated formats a list of digits separated by the specified
// separator rune.  Digits are assumed to already be in the desired order.
func formatDigitsSeparated(digits []int, sep rune) string {
	result := make([]rune, 0, 2*len(digits)-1)
	for i, d := range digits {
		if i > 0 {
			result = append(result, sep)
		}
		result = append(result, rune('0'+d))
	}
	return string(result)
}
