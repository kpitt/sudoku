package puzzle

import (
	"bufio"
	"os"
)

func ReadBoard(f *os.File) *Board {
	b := NewBoard()
	scanner := bufio.NewScanner(f)

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
		fatalError("error reading standard input", err.Error())
	}

	return b
}

func (b *Board) processRow(row int, line string) {
	for col := range 9 {
		val := line[col] - 48
		if val >= 1 && val <= 9 {
			b.setFixedValue(row, col, int8(val))
		}
	}
}
