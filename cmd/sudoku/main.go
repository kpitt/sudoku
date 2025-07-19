package main

import (
	"github.com/fatih/color"
	"github.com/kpitt/sudoku/internal/board"
)

func main() {
	b := board.NewBoard()
	b.Read()

	b.Solve()

	if b.IsSolved() {
		color.HiWhite("\nSolution:")
	} else {
		color.HiWhite("\nPartial Solution:")
	}
	b.Print()
}
