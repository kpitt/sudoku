package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/solver"
)

func main() {
	fmt.Println("Dancing Links Algorithm - Basic Example")
	fmt.Println("======================================")

	// Create a simple Sudoku givens
	givens := [][]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}

	// Create and setup puzzle
	p := puzzle.NewPuzzle()
	for r := range 9 {
		for c := range 9 {
			if givens[r][c] != 0 {
				p.Grid[r][c].GivenValue(givens[r][c])
			}
		}
	}

	fmt.Println("Original puzzle:")
	printPuzzle(p)

	// Method 1: Basic Dancing Links solving
	fmt.Println("\n--- Method 1: Basic Solving ---")
	start := time.Now()

	dl := solver.NewDancingLinks(p)
	solved := dl.Solve()

	duration := time.Since(start)

	if solved {
		fmt.Printf("✓ Solved in %v\n", duration)
		fmt.Println("Solution:")
		printPuzzle(p)
	} else {
		fmt.Println("✗ Failed to solve")
	}

	// Method 2: Using solver integration
	fmt.Println("\n--- Method 2: Solver Integration ---")

	// Reset puzzle for second test
	p2 := puzzle.NewPuzzle()
	for r := range 9 {
		for c := range 9 {
			if givens[r][c] != 0 {
				p2.Grid[r][c].PlaceValue(givens[r][c])
			}
		}
	}

	start = time.Now()
	s := solver.NewSolver(p2)
	solved = s.SolveDancingLinks()
	duration = time.Since(start)

	if solved {
		fmt.Printf("✓ Solved in %v\n", duration)
		fmt.Println("Solution matches:", puzzlesEqual(p, p2))
	} else {
		fmt.Println("✗ Failed to solve")
	}

	// Method 3: Using advanced options with statistics
	fmt.Println("\n--- Method 3: Advanced with Statistics ---")

	// Reset puzzle for third test
	p3 := puzzle.NewPuzzle()
	for r := range 9 {
		for c := range 9 {
			if givens[r][c] != 0 {
				p3.Grid[r][c].PlaceValue(givens[r][c])
			}
		}
	}

	options := &solver.DancingLinksOptions{
		EnableDebugging: false,
		TimeLimit:       5 * time.Second,
		MaxSolutions:    1,
	}

	solved, stats, err := solver.SolveWithDancingLinks(p3, options)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	if solved {
		fmt.Println("✓ Solved with statistics:")
		stats.PrintStats()
	} else {
		fmt.Println("✗ Failed to solve")
	}

	// Demonstrate matrix information
	fmt.Println("\n--- Matrix Information ---")
	dl4 := solver.NewDancingLinks(p3)

	// Count empty cells in original puzzle
	emptyCells := 0
	for r := range 9 {
		for c := range 9 {
			if givens[r][c] == 0 {
				emptyCells++
			}
		}
	}

	fmt.Printf("Empty cells in puzzle: %d\n", emptyCells)
	fmt.Printf("Matrix rows created: %d\n", len(dl4.Rows))
	fmt.Printf("Theoretical maximum rows: %d (9×9×9)\n", 9*9*9)

	// Count columns
	columnCount := 0
	for col := dl4.Header.Right; col != &dl4.Header.Node; col = col.Right {
		columnCount++
	}
	fmt.Printf("Matrix columns: %d\n", columnCount)
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
				fmt.Printf("%d ", cell.Value())
			} else {
				fmt.Print("· ")
			}
		}
		fmt.Println("│")
	}
	fmt.Println("└───────┴───────┴───────┘")
}

func puzzlesEqual(p1, p2 *puzzle.Puzzle) bool {
	for r := range 9 {
		for c := range 9 {
			cell1, cell2 := p1.Grid[r][c], p2.Grid[r][c]
			if cell1.IsSolved() != cell2.IsSolved() {
				return false
			}
			if cell1.IsSolved() && cell1.Value() != cell2.Value() {
				return false
			}
		}
	}
	return true
}
