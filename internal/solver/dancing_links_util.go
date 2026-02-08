package solver

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

// DancingLinksOptions configures the Dancing Links solver behavior
type DancingLinksOptions struct {
	EnableDebug bool
	TimeLimit   time.Duration
}

// DancingLinksStats tracks solving statistics
type DancingLinksStats struct {
	NodesVisited   int
	BacktrackCount int
	SolutionsFound int
	TimeElapsed    time.Duration
	MatrixSize     MatrixInfo
}

// MatrixInfo provides information about the constraint matrix
type MatrixInfo struct {
	Columns    int
	Rows       int
	TotalNodes int
	Density    float64 // percentage of non-zero entries
}

// SolveWithStats solves using Dancing Links and returns detailed statistics
func (dl *DancingLinks) SolveWithStats(options *DancingLinksOptions) (bool, *DancingLinksStats) {
	if options == nil {
		options = &DancingLinksOptions{TimeLimit: 10 * time.Second}
	}

	stats := &DancingLinksStats{
		MatrixSize: dl.getMatrixInfo(),
	}

	start := time.Now()
	defer func() {
		stats.TimeElapsed = time.Since(start)
	}()

	// Set up timeout if specified
	var timeout <-chan time.Time
	if options.TimeLimit > 0 {
		timeout = time.After(options.TimeLimit)
	}

	solved := dl.searchWithStats(stats, options, timeout)
	return solved, stats
}

// searchWithStats implements search with statistics tracking
func (dl *DancingLinks) searchWithStats(stats *DancingLinksStats, options *DancingLinksOptions, timeout <-chan time.Time) bool {
	// Check timeout
	select {
	case <-timeout:
		return false
	default:
	}

	stats.NodesVisited++

	if dl.Header.Right == &dl.Header.Node {
		// All columns covered - solution found
		stats.SolutionsFound++
		return true
	}

	// Choose column with minimum size (heuristic)
	col := dl.chooseColumn()
	if options.EnableDebug {
		printDebug("Choosing column %s with %d options", col.Name, col.Size)
	}

	dl.cover(col)

	// Try each row in the chosen column
	for r := col.Down; r != &col.Node; r = r.Down {
		dl.solution = append(dl.solution, r.RowID)

		// Cover all other columns in this row
		for j := r.Right; j != r; j = j.Right {
			dl.cover(j.Column)
		}

		// Recursively search
		if dl.searchWithStats(stats, options, timeout) {
			return true
		}

		// Backtrack: uncover columns in reverse order
		for j := r.Left; j != r; j = j.Left {
			dl.uncover(j.Column)
		}

		if options.EnableDebug {
			if can, ok := dl.candidates[r.RowID]; ok {
				r, c := rowColFromIndex(can.Index)
				printDebug("Backtracking: no solution for R%dC%d#%d", r+1, c+1, can.Value)
			}
		}
		dl.solution = dl.solution[:len(dl.solution)-1]
		stats.BacktrackCount++
	}

	dl.uncover(col)
	return false
}

// getMatrixInfo calculates information about the constraint matrix
func (dl *DancingLinks) getMatrixInfo() MatrixInfo {
	info := MatrixInfo{}

	// Count columns
	for col := dl.Header.Right; col != &dl.Header.Node; col = col.Right {
		info.Columns++
	}

	// Count rows and total nodes
	info.Rows = len(dl.Rows)

	totalNodes := 0
	for _, row := range dl.Rows {
		if row != nil {
			nodes := 1
			for n := row.Right; n != row; n = n.Right {
				nodes++
			}
			totalNodes += nodes
		}
	}

	info.TotalNodes = totalNodes

	// Calculate density (percentage of filled cells in the matrix)
	if info.Columns > 0 && info.Rows > 0 {
		totalCells := info.Columns * info.Rows
		info.Density = float64(totalNodes) / float64(totalCells) * 100.0
	}

	return info
}

// PrintStats displays solving statistics in a formatted way
func (stats *DancingLinksStats) PrintStats() {
	fmt.Printf("\n%s\n", color.HiCyanString("Dancing Links Statistics"))
	fmt.Printf("%s\n", color.HiCyanString("========================"))

	fmt.Printf("Matrix Info:\n")
	fmt.Printf("  Columns:     %s\n", color.HiYellowString("%d", stats.MatrixSize.Columns))
	fmt.Printf("  Rows:        %s\n", color.HiYellowString("%d", stats.MatrixSize.Rows))
	fmt.Printf("  Total Nodes: %s\n", color.HiYellowString("%d", stats.MatrixSize.TotalNodes))
	fmt.Printf("  Density:     %s\n", color.HiYellowString("%.2f%%", stats.MatrixSize.Density))

	fmt.Printf("\nSearch Statistics:\n")
	fmt.Printf("  Nodes Visited:   %s\n", color.HiGreenString("%d", stats.NodesVisited))
	fmt.Printf("  Backtracks:      %s\n", color.HiRedString("%d", stats.BacktrackCount))
	fmt.Printf("  Solutions Found: %s\n", color.HiGreenString("%d", stats.SolutionsFound))
	fmt.Printf("  Time Elapsed:    %s\n", color.HiBlueString("%v", stats.TimeElapsed))

	if stats.TimeElapsed.Nanoseconds() > 0 {
		nodesPerSec := float64(stats.NodesVisited) / stats.TimeElapsed.Seconds()
		fmt.Printf("  Nodes/Second:    %s\n", color.HiMagentaString("%.0f", nodesPerSec))
	}
}

// PrintMatrix prints a visual representation of the constraint matrix (for debugging)
func (dl *DancingLinks) PrintMatrix() {
	fmt.Printf("\n%s\n", color.HiCyanString("Constraint Matrix Structure"))
	fmt.Printf("%s\n", color.HiCyanString("==========================="))

	// Print column headers
	fmt.Print("Columns: ")
	count := 0
	for col := dl.Header.Right; col != &dl.Header.Node && count < 10; col = col.Right {
		fmt.Printf("%s ", color.HiYellowString(col.Column.Name))
		count++
	}
	if count == 10 {
		totalCols := dl.getMatrixInfo().Columns
		fmt.Printf("... (%d more)", totalCols-10)
	}
	fmt.Println()

	// Print first few rows
	fmt.Printf("Rows (%d total):\n", len(dl.Rows))
	for i, row := range dl.Rows {
		if i >= 5 { // Only show first 5 rows
			fmt.Printf("... (%d more rows)\n", len(dl.Rows)-5)
			break
		}

		if row != nil {
			r, c, val := dl.decodeRow(row.RowID)
			fmt.Printf("  Row %d: R%dC%d=%d -> ", i, r, c, val)

			// Show which columns this row covers
			nodeCount := 1
			fmt.Printf("%s ", row.Column.Name)
			for n := row.Right; n != row && nodeCount < 4; n = n.Right {
				fmt.Printf("%s ", n.Column.Name)
				nodeCount++
			}
			fmt.Println()
		}
	}
}

// CountSolutions counts the total number of solutions (useful for puzzles with multiple solutions)
func (dl *DancingLinks) CountSolutions(maxSolutions int) int {
	originalSolution := make([]int, len(dl.solution))
	copy(originalSolution, dl.solution)

	solutionCount := 0
	dl.countSolutionsRecursive(&solutionCount, maxSolutions)

	// Restore original solution state
	dl.solution = originalSolution
	return solutionCount
}

// countSolutionsRecursive implements the recursive solution counting
func (dl *DancingLinks) countSolutionsRecursive(count *int, maxSolutions int) {
	if *count >= maxSolutions {
		return
	}

	if dl.Header.Right == &dl.Header.Node {
		// Found a solution
		*count++
		return
	}

	col := dl.chooseColumn()
	dl.cover(col)

	for r := col.Down; r != &col.Node; r = r.Down {
		dl.solution = append(dl.solution, r.RowID)

		for j := r.Right; j != r; j = j.Right {
			dl.cover(j.Column)
		}

		dl.countSolutionsRecursive(count, maxSolutions)

		for j := r.Left; j != r; j = j.Left {
			dl.uncover(j.Column)
		}

		dl.solution = dl.solution[:len(dl.solution)-1]

		if *count >= maxSolutions {
			break
		}
	}

	dl.uncover(col)
}
