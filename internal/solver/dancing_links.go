package solver

import (
	"fmt"

	"github.com/kpitt/sudoku/internal/puzzle"
)

// Node represents a node in the Dancing Links data structure
type Node struct {
	Left, Right, Up, Down *Node
	Column                *ColumnNode
	RowID                 int // identifies which constraint row this node belongs to
}

// ColumnNode represents a column header in the Dancing Links matrix
type ColumnNode struct {
	Node
	Size int    // number of nodes in this column
	Name string // column identifier for debugging
}

// DancingLinks implements Knuth's Algorithm X using Dancing Links
type DancingLinks struct {
	Header   *ColumnNode
	Rows     []*Node        // first node of each row for solution reconstruction
	Puzzle   *puzzle.Puzzle // reference to the sudoku puzzle
	solution []int          // tracks selected rows in current solution

	candidates map[int]Candidate // maps each row ID to the candidate it represents
}

// NewDancingLinks creates a new Dancing Links solver for the given puzzle
func NewDancingLinks(p *puzzle.Puzzle) *DancingLinks {
	dl := &DancingLinks{
		Puzzle:   p,
		solution: make([]int, 0, 81),
	}
	dl.buildMatrix()
	return dl
}

// buildMatrix constructs the exact cover matrix for the sudoku puzzle
func (dl *DancingLinks) buildMatrix() {
	// Create header node
	dl.Header = &ColumnNode{Name: "header"}
	dl.Header.Left = &dl.Header.Node
	dl.Header.Right = &dl.Header.Node

	// For 9x9 Sudoku, we have 4 types of constraints:
	// 1. Cell constraints: each cell must have exactly one value (81 constraints)
	// 2. Row constraints: each row must have each digit 1-9 exactly once (81 constraints)
	// 3. Column constraints: each column must have each digit 1-9 exactly once (81 constraints)
	// 4. Box constraints: each 3x3 box must have each digit 1-9 exactly once (81 constraints)
	// Total: 324 constraints

	columns := make([]*ColumnNode, 324)

	// Create column headers
	for i := range 324 {
		col := &ColumnNode{Name: dl.getColumnName(i)}
		col.Up = &col.Node
		col.Down = &col.Node
		col.Column = col
		columns[i] = col

		// Link column to header
		col.Left = dl.Header.Left
		col.Right = &dl.Header.Node
		dl.Header.Left.Right = &col.Node
		dl.Header.Left = &col.Node
	}

	// Create rows for each possible (row, col, value) combination
	dl.Rows = make([]*Node, 0, 729) // 9*9*9 = 729 possible combinations
	// Also create map for recording the candidate represented by each row.
	dl.candidates = make(map[int]Candidate)

	for r := range 9 {
		for c := range 9 {
			cell := dl.Puzzle.Grid[r][c]

			// If cell is already filled, only create row for that value
			if cell.IsSolved() {
				val := cell.Value()
				rowNodes := dl.createRowNodes(r, c, int(val), columns)
				dl.Rows = append(dl.Rows, rowNodes[0])
			} else {
				// Create rows for all possible values this cell can have
				for val := 1; val <= 9; val++ {
					if cell.HasCandidate(val) {
						rowNodes := dl.createRowNodes(r, c, val, columns)
						dl.Rows = append(dl.Rows, rowNodes[0])
					}
				}
			}
		}
	}
}

// createRowNodes creates the four nodes for a (row, col, value) combination
func (dl *DancingLinks) createRowNodes(r, c, val int, columns []*ColumnNode) []*Node {
	// Calculate column indices for the four constraints
	cellConstraint := r*9 + c
	rowConstraint := 81 + r*9 + (val - 1)
	colConstraint := 162 + c*9 + (val - 1)
	boxConstraint := 243 + (r/3*3+c/3)*9 + (val - 1)

	constraintCols := []int{cellConstraint, rowConstraint, colConstraint, boxConstraint}
	nodes := make([]*Node, 4)
	rowID := len(dl.Rows)
	// Record the candidate for this row ID
	dl.candidates[rowID] = Candidate{Index: cellConstraint, Value: val}

	// Create nodes for each constraint
	for i, colIdx := range constraintCols {
		node := &Node{
			Column: columns[colIdx],
			RowID:  rowID,
		}
		nodes[i] = node

		// Link node into column
		node.Down = columns[colIdx].Down
		node.Up = &columns[colIdx].Node
		columns[colIdx].Down.Up = node
		columns[colIdx].Down = node
		columns[colIdx].Size++
	}

	// Link nodes horizontally in circular fashion
	for i := range 4 {
		nodes[i].Left = nodes[(i+3)%4]
		nodes[i].Right = nodes[(i+1)%4]
	}

	return nodes
}

// getColumnName returns a descriptive name for the column at the given index
func (dl *DancingLinks) getColumnName(index int) string {
	if index < 81 {
		r, c := index/9, index%9
		return fmt.Sprintf("R%dC%d", r, c)
	} else if index < 162 {
		idx := index - 81
		r, val := idx/9, idx%9+1
		return fmt.Sprintf("R%d#%d", r, val)
	} else if index < 243 {
		idx := index - 162
		c, val := idx/9, idx%9+1
		return fmt.Sprintf("C%d#%d", c, val)
	} else {
		idx := index - 243
		box, val := idx/9, idx%9+1
		return fmt.Sprintf("B%d#%d", box, val)
	}
}

// Solve attempts to solve the sudoku using Dancing Links Algorithm X
func (dl *DancingLinks) Solve() bool {
	return dl.search()
}

// search implements the recursive search algorithm
func (dl *DancingLinks) search() bool {
	if dl.Header.Right == &dl.Header.Node {
		// All columns covered - solution found
		return dl.applySolution()
	}

	// Choose column with minimum size (heuristic)
	col := dl.chooseColumn()
	dl.cover(col)

	// Try each row in the chosen column
	for r := col.Down; r != &col.Node; r = r.Down {
		dl.solution = append(dl.solution, r.RowID)

		// Cover all other columns in this row
		for j := r.Right; j != r; j = j.Right {
			dl.cover(j.Column)
		}

		// Recursively search
		if dl.search() {
			return true
		}

		// Backtrack: uncover columns in reverse order
		for j := r.Left; j != r; j = j.Left {
			dl.uncover(j.Column)
		}

		dl.solution = dl.solution[:len(dl.solution)-1]
	}

	dl.uncover(col)
	return false
}

// chooseColumn selects the column with the fewest nodes (MRV heuristic)
func (dl *DancingLinks) chooseColumn() *ColumnNode {
	var chosen *ColumnNode
	minSize := int(^uint(0) >> 1) // max int

	for col := dl.Header.Right; col != &dl.Header.Node; col = col.Right {
		columnNode := col.Column
		if columnNode.Size < minSize {
			chosen = columnNode
			minSize = columnNode.Size
		}
	}

	return chosen
}

// cover removes a column and all rows that intersect with it
func (dl *DancingLinks) cover(col *ColumnNode) {
	// Remove column header from list
	col.Right.Left = col.Left
	col.Left.Right = col.Right

	// Remove all rows that have a node in this column
	for i := col.Down; i != &col.Node; i = i.Down {
		for j := i.Right; j != i; j = j.Right {
			// Remove node from its column
			j.Down.Up = j.Up
			j.Up.Down = j.Down
			j.Column.Size--
		}
	}
}

// uncover restores a column and all rows that intersect with it
func (dl *DancingLinks) uncover(col *ColumnNode) {
	// Restore all rows that have a node in this column
	for i := col.Up; i != &col.Node; i = i.Up {
		for j := i.Left; j != i; j = j.Left {
			// Restore node to its column
			j.Column.Size++
			j.Down.Up = j
			j.Up.Down = j
		}
	}

	// Restore column header to list
	col.Right.Left = &col.Node
	col.Left.Right = &col.Node
}

// applySolution applies the found solution to the puzzle
func (dl *DancingLinks) applySolution() bool {
	for _, rowID := range dl.solution {
		if rowID >= len(dl.Rows) {
			continue
		}

		// Decode the row to get (r, c, val)
		r, c, val := dl.decodeRow(rowID)

		// Apply to puzzle if not already set
		cell := dl.Puzzle.Grid[r][c]
		if !cell.IsSolved() {
			if !dl.Puzzle.PlaceValue(r, c, val) {
				return false
			}
		}
	}
	return true
}

// decodeRow extracts the row, column, and value from a row ID
func (dl *DancingLinks) decodeRow(rowID int) (int, int, int) {
	if can, ok := dl.candidates[rowID]; ok {
		r, c := rowColFromIndex(can.Index)
		return r, c, can.Value
	}

	return 0, 0, 1 // fallback
}

// SolveDancingLinks solves the sudoku using Dancing Links algorithm
func (s *Solver) SolveDancingLinks() bool {
	dl := NewDancingLinks(s.puzzle)
	return dl.Solve()
}
