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

	opts := &solver.Options{EnableDebug: true, LiveLog: true}
	s := solver.NewSolver(p, opts)
	s.Solve()

	if opts.LiveLog {
		// Add a line break between the live-log and the solution.
		fmt.Println()
	}
	if p.IsSolved() {
		color.HiBlue("Solution:")
		p.Print()
	} else {
		color.HiBlue("Partial Solution:")
		p.PrintCandidateGrid()
	}

	fmt.Println()
	color.HiYellow("Total Checks:     %d", s.NumChecks)
	color.HiYellow("Total Solve Time: %v", s.SolveTime)

	if !p.IsSolved() {
		fmt.Println()
		p.PrintUnsolvedCounts()
	} else if !opts.LiveLog {
		// Only print solution if steps were not already live-logged.
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
