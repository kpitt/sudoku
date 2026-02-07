package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/kpitt/sudoku/internal/puzzle"
	"github.com/kpitt/sudoku/internal/solver"
	"github.com/mattn/go-isatty"
)

func main() {
	if isStdinTTY() {
		fmt.Println("Enter initial puzzle as 9 lines of 9 characters.")
		fmt.Println("Use any character other than the digits 1-9 for empty cells.")
		fmt.Println("(Ctrl+D to finish on Unix/Linux, Ctrl+Z then Enter on Windows):")
	}

	p, err := puzzle.FromFile(os.Stdin)
	if err != nil {
		fatalError(err.Error())
	}

	color.HiBlue("Original Puzzle:")
	p.Print()
	fmt.Println()

	opts := &solver.Options{
		// EnableDebug:      true,
		// LiveLog:          true,
		EnableBruteForce: true,
	}
	s := solver.NewSolver(p, opts)
	s.Solve()

	if p.IsSolved() {
		fmt.Printf("%s (%v)\n\n",
			color.HiGreenString("✓ Solved successfully"), s.SolveTime)
		color.HiBlue("Solution:")
		p.Print()
	} else {
		fmt.Printf("%s (%v)\n\n",
			color.HiRedString("✗ Failed to solve"), s.SolveTime)
		color.HiBlue("Partial Solution:")
		p.PrintCandidateGrid()
		fmt.Println()
		p.PrintUnsolvedCounts()
	}

	// Only print solution if steps were not already live-logged.
	if !opts.LiveLog {
		fmt.Println()
		s.PrintSolution()
	}
}

func isStdinTTY() bool {
	return isTerminal(os.Stdin)
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

func fatalError(msg string) {
	fmt.Fprintln(os.Stderr, color.HiRedString("error: %s", msg))
	os.Exit(1)
}
