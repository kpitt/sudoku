package main

import (
	"fmt"

	"github.com/kpitt/sudoku/internal/board"
)

func main() {
	b := board.NewBoard()
	b.Read()

	b.Solve()

	if b.IsSolved() {
		fmt.Println("Solution:")
	} else {
		fmt.Println("Partial Solution:")
	}
	b.Print()
}
