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
		fmt.Println("Enter initial board as 9 lines of 9 characters.")
		fmt.Println("Use any character other than the digits 1-9 for empty cells.")
		fmt.Println("(Ctrl+D to finish on Unix/Linux, Ctrl+Z then Enter on Windows):")
	}

	b := puzzle.ReadBoard(os.Stdin)
	s := solver.NewSolver(b)
	s.Solve()

	if b.IsSolved() {
		color.HiWhite("\nSolution:")
	} else {
		color.HiWhite("\nPartial Solution:")
	}
	b.Print()

	if !b.IsSolved() {
		fmt.Println()
		b.PrintUnsolvedCounts()
	}
}

func isStdinTTY() bool {
	return isTerminal(os.Stdin)
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
