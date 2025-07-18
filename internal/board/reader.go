package board

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

func (b *Board) Read() {
	scanner := bufio.NewScanner(os.Stdin)

	if isStdinTTY() {
		fmt.Println("Enter initial board as 9 lines of 9 characters.")
		fmt.Println("Use any character other than the digits 1-9 for empty cells.")
		fmt.Println("(Ctrl+D to finish on Unix/Linux, Ctrl+Z then Enter on Windows):")
	}

	r := 0
	for scanner.Scan() {
		if r >= 9 {
			boardStateError("too many input lines")
		}
		line := scanner.Text()
		if len(line) < 9 {
			boardStateError("input line too short")
		}
		b.processRow(r, line[:9])
		r = r + 1
	}
	if r < 9 {
		boardStateError("not enough input lines")
	}

	if err := scanner.Err(); err != nil {
		fatalError("error reading standard input: " + err.Error())
	}
}

func (b *Board) processRow(row int, line string) {
	for col := range 9 {
		val := line[col] - 48
		if val >= 1 && val <= 9 {
			b.FixValue(row, col, int8(val))
		}
	}
}

func boardStateError(msg string) {
	fatalError("invalid board state: " + msg)
}

func fatalError(msg string) {
	fmt.Fprintf(os.Stderr, "error: invalid board state: %s\n", msg)
	os.Exit(1)
}

func isStdinTTY() bool {
	return isTerminal(os.Stdin)
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
