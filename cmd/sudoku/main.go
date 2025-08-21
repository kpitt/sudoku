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

	p := puzzle.PuzzleFromFile(os.Stdin)
	s := solver.NewSolver(p)
	s.Solve()

	if p.IsSolved() {
		color.HiWhite("\nSolution:")
	} else {
		color.HiWhite("\nPartial Solution:")
	}
	p.Print()

	if !p.IsSolved() {
		fmt.Println()
		p.PrintUnsolvedCounts()
	}
}

func isStdinTTY() bool {
	return isTerminal(os.Stdin)
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
