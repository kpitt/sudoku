package puzzle

import (
	"bufio"
	"os"
)

func PuzzleFromFile(f *os.File) *Puzzle {
	b := NewPuzzle()
	scanner := bufio.NewScanner(f)

	r := 0
	for scanner.Scan() {
		if r >= 9 {
			puzzleStateError("too many input lines")
		}
		line := scanner.Text()
		if len(line) < 9 {
			puzzleStateError("input line too short")
		}
		b.processRow(r, line[:9])
		r = r + 1
	}
	if r < 9 {
		puzzleStateError("not enough input lines")
	}

	if err := scanner.Err(); err != nil {
		fatalError("error reading standard input", err.Error())
	}

	return b
}

func (p *Puzzle) processRow(row int, line string) {
	for col := range 9 {
		val := line[col] - 48
		if val >= 1 && val <= 9 {
			p.GivenValue(row, col, int(val))
		}
	}
}
