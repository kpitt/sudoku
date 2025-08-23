package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/solver"
)

func main() {
	fmt.Println("Dancing Links Algorithm Demonstration")
	fmt.Println("=====================================")

	// Test cases with different difficulty levels
	testCases := []struct {
		name   string
		puzzle [][]int
	}{
		{
			name: "Easy Puzzle",
			puzzle: [][]int{
				{5, 3, 0, 0, 7, 0, 0, 0, 0},
				{6, 0, 0, 1, 9, 5, 0, 0, 0},
				{0, 9, 8, 0, 0, 0, 0, 6, 0},
				{8, 0, 0, 0, 6, 0, 0, 0, 3},
				{4, 0, 0, 8, 0, 3, 0, 0, 1},
				{7, 0, 0, 0, 2, 0, 0, 0, 6},
				{0, 6, 0, 0, 0, 0, 2, 8, 0},
				{0, 0, 0, 4, 1, 9, 0, 0, 5},
				{0, 0, 0, 0, 8, 0, 0, 7, 9},
			},
		},
		{
			name: "Medium Puzzle",
			puzzle: [][]int{
				{0, 0, 0, 6, 0, 0, 4, 0, 0},
				{7, 0, 0, 0, 0, 3, 6, 0, 0},
				{0, 0, 0, 0, 9, 1, 0, 8, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 5, 0, 1, 8, 0, 0, 0, 3},
				{0, 0, 0, 3, 0, 6, 0, 4, 5},
				{0, 4, 0, 2, 0, 0, 0, 6, 0},
				{9, 0, 3, 0, 0, 0, 0, 0, 0},
				{0, 2, 0, 0, 0, 0, 1, 0, 0},
			},
		},
		{
			name: "Hard Puzzle",
			puzzle: [][]int{
				{0, 0, 0, 0, 0, 0, 0, 1, 0},
				{4, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 6, 0, 2},
				{0, 0, 0, 0, 0, 3, 0, 7, 0},
				{5, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
				{0, 0, 0, 2, 0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("\n%s %d: %s\n", color.HiBlueString("Test Case"), i+1, color.HiYellowString(testCase.name))
		fmt.Println(color.HiBlueString("Original Puzzle:"))

		// Create and setup puzzle
		p := puzzle.NewPuzzle()
		setupPuzzle(p, testCase.puzzle)
		printPuzzle(p)

		// Solve using Dancing Links
		fmt.Println(color.HiGreenString("\nSolving with Dancing Links Algorithm..."))
		start := time.Now()

		s := solver.NewSolver(p)
		solved := s.SolveDancingLinks()

		duration := time.Since(start)

		if solved {
			fmt.Printf("%s (%.3fms)\n", color.HiGreenString("✓ Solved successfully!"), float64(duration.Nanoseconds())/1e6)
			fmt.Println(color.HiBlueString("Solution:"))
			printPuzzle(p)

			// Verify solution
			if verifySolution(p) {
				fmt.Println(color.HiGreenString("✓ Solution verified as correct!"))
			} else {
				fmt.Println(color.HiRedString("✗ Solution verification failed!"))
			}
		} else {
			fmt.Printf("%s (%.3fms)\n", color.HiRedString("✗ Failed to solve"), float64(duration.Nanoseconds())/1e6)
		}

		fmt.Println(color.HiBlackString("─────────────────────────────────────"))
	}

	// Demonstrate algorithm details
	demonstrateAlgorithmDetails()
}

func setupPuzzle(p *puzzle.Puzzle, puzzle [][]int) {
	for r := range 9 {
		for c := range 9 {
			if puzzle[r][c] != 0 {
				p.GivenValue(r, c, puzzle[r][c])
			}
		}
	}
}

func printPuzzle(p *puzzle.Puzzle) {
	fmt.Println("┌───────┬───────┬───────┐")
	for r := range 9 {
		if r == 3 || r == 6 {
			fmt.Println("├───────┼───────┼───────┤")
		}
		fmt.Print("│ ")
		for c := range 9 {
			if c == 3 || c == 6 {
				fmt.Print("│ ")
			}
			cell := p.Grid[r][c]
			if cell.IsSolved() {
				if cell.IsGiven {
					fmt.Printf("%s ", color.HiBlueString("%d", cell.Value()))
				} else {
					fmt.Printf("%s ", color.HiGreenString("%d", cell.Value()))
				}
			} else {
				fmt.Print(color.HiBlackString("· "))
			}
		}
		fmt.Println("│")
	}
	fmt.Println("└───────┴───────┴───────┘")
	fmt.Printf("Legend: %s = Given, %s = Solved, %s = Empty\n",
		color.HiBlueString("Blue"), color.HiGreenString("Green"), color.HiBlackString("Gray"))
}

func verifySolution(p *puzzle.Puzzle) bool {
	// Check if puzzle is completely solved
	if !p.IsSolved() {
		return false
	}

	// Verify all constraints
	for i := range 9 {
		// Check rows
		if !verifyHouse(func(j int) int { return p.Grid[i][j].Value() }) {
			return false
		}

		// Check columns
		if !verifyHouse(func(j int) int { return p.Grid[j][i].Value() }) {
			return false
		}

		// Check 3x3 boxes
		boxRow, boxCol := i/3, i%3
		if !verifyHouse(func(j int) int {
			r, c := boxRow*3+j/3, boxCol*3+j%3
			return p.Grid[r][c].Value()
		}) {
			return false
		}
	}

	return true
}

func verifyHouse(getValue func(int) int) bool {
	seen := make(map[int]bool)
	for i := range 9 {
		val := getValue(i)
		if val < 1 || val > 9 || seen[val] {
			return false
		}
		seen[val] = true
	}
	return true
}

func demonstrateAlgorithmDetails() {
	fmt.Printf("\n%s\n", color.HiCyanString("Dancing Links Algorithm Details"))
	fmt.Println(color.HiCyanString("================================"))

	fmt.Println("\nThe Dancing Links algorithm (also known as Algorithm X) is designed to solve")
	fmt.Println("exact cover problems efficiently. For Sudoku, we model the puzzle as an exact")
	fmt.Println("cover problem with the following constraints:")

	fmt.Printf("\n%s\n", color.HiYellowString("1. Constraint Matrix Structure:"))
	fmt.Println("   • 324 columns representing all constraints")
	fmt.Println("   • 81 cell constraints: each cell must have exactly one value")
	fmt.Println("   • 81 row constraints: each row must contain digits 1-9 exactly once")
	fmt.Println("   • 81 column constraints: each column must contain digits 1-9 exactly once")
	fmt.Println("   • 81 box constraints: each 3×3 box must contain digits 1-9 exactly once")

	fmt.Printf("\n%s\n", color.HiYellowString("2. Matrix Rows:"))
	fmt.Println("   • Up to 729 rows (9×9×9) representing all possible (row, col, value) combinations")
	fmt.Println("   • Each row has exactly 4 nodes (one for each constraint type)")
	fmt.Println("   • Rows for filled cells are pre-selected in the matrix")

	fmt.Printf("\n%s\n", color.HiYellowString("3. Dancing Links Operations:"))
	fmt.Println("   • Cover: Remove a column and all rows intersecting it")
	fmt.Println("   • Uncover: Restore a column and all intersecting rows (backtracking)")
	fmt.Println("   • Search: Recursively select rows and apply cover/uncover operations")

	fmt.Printf("\n%s\n", color.HiYellowString("4. Key Optimizations:"))
	fmt.Println("   • Minimum Remaining Values (MRV) heuristic: choose column with fewest options")
	fmt.Println("   • Doubly-linked circular lists enable O(1) cover/uncover operations")
	fmt.Println("   • Pre-filtering based on current candidate values")

	fmt.Printf("\n%s\n", color.HiYellowString("5. Advantages over other approaches:"))
	fmt.Println("   • Guaranteed to find solution if one exists")
	fmt.Println("   • Efficient backtracking with O(1) undo operations")
	fmt.Println("   • Naturally handles constraint propagation")
	fmt.Println("   • Works well for hard puzzles where logical deduction fails")

	// Create a small example to show matrix structure
	fmt.Printf("\n%s\n", color.HiGreenString("Example Matrix Structure:"))
	p := puzzle.NewPuzzle()
	p.Grid[0][0].PlaceValue(5) // R0C0 = 5

	dl := solver.NewDancingLinks(p)

	fmt.Println("For the constraint R0C0=5, the algorithm creates connections to:")
	fmt.Println("   • Column R0C0 (cell constraint)")
	fmt.Println("   • Column R0#5 (row constraint)")
	fmt.Println("   • Column C0#5 (column constraint)")
	fmt.Println("   • Column B0#5 (box constraint)")

	fmt.Printf("\nTotal columns created: ")
	columnCount := 0
	for col := dl.Header.Right; col != &dl.Header.Node; col = col.Right {
		columnCount++
	}
	fmt.Printf("%s\n", color.HiGreenString("%d", columnCount))

	fmt.Printf("Total rows created: %s\n", color.HiGreenString("%d", len(dl.Rows)))
}
