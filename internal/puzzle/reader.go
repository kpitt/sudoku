package puzzle

import (
	"io"
	"os"
	"strings"
	"unicode"
)

func FromFile(f *os.File) (*Puzzle, error) {
	var buf strings.Builder
	if data, err := io.ReadAll(f); err != nil {
		return nil, err
	} else {
		_, _ = buf.Write(data)
	}
	return FromString(buf.String())
}

// FromString creates a new puzzle from the given puzzle string.
// Empty lines and comments starting with a '#' are ignored.
// The following formats are supported:
//
// A single 81-digit string ('0' or '.' for an empty cell)
//
// A 9x9 grid of cells, with optional whitespace and border characters.
// Possible formats include:
//
//	```
//	16.54..7.   or   *-----------*   or   +-----+-----+-----+
//	..8..1.3.        |16.|54.|.7.|        |1 6 .|5 4 .|. 7 .|
//	.3.8.....        |..8|..1|.3.|        |. . 8|. . 1|. 3 .|
//	7...5..69        |.3.|8..|...|        |. 3 .|8 . .|. . .|
//	6..9.2.57        |---+---+---|        +-----+-----+-----+
//	.........        |7..|.5.|.69|        |7 . .|. 5 .|. 6 9|
//	....3..4.        |6..|9.2|.57|        |6 . .|9 . 2|. 5 7|
//	.......16        |...|...|...|        |. . .|. . .|. . .|
//	...1645..        |---+---+---|        +-----+-----+-----+
//	                 |...|.3.|.4.|        |. . .|. 3 .|. 4 .|
//	                 |...|...|.16|        |. . .|. . .|. 1 6|
//	                 |...|164|5..|        |. . .|1 6 4|5 . .|
//	                 *-----------*        +-----+-----+-----+
//	```
func FromString(s string) (*Puzzle, error) {
	p := NewPuzzle()
	// Determine the newline sequence used in the input string, and split the
	// string into lines based on that sequence.
	sep := "\n"
	if strings.Contains(s, "\r\n") {
		sep = "\r\n"
	} else if strings.Contains(s, "\r") {
		sep = "\r"
	}
	lines := strings.Split(s, sep)
	i := 0
	for _, line := range lines {
		// Remove comments and extra whitespace from the line.
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		// Skip empty lines and border lines.
		// Any line with a sequence of 3 dashes ("---") is a border line.
		if line == "" || strings.Contains(line, "---") {
			continue
		}

		for pos, c := range line {
			// Discard whitespace and '|' border characters.
			if unicode.IsSpace(c) || c == '|' {
				continue
			}
			if i >= 81 {
				// We've already filled all 81 cells and skipped any whitespace
				// and border characters that might follow the last cell, so we
				// shouldn't have any more characters to process.
				return nil, errPuzzleFormat("extraneous characters: %q", line[pos:])
			}
			if c != '.' && !unicode.IsDigit(c) {
				return nil, errPuzzleFormat("invalid character: %c", c)
			}
			// Place a given for digits '1'-'9' and advance the index.
			// A '0' or '.' is an empty cell, so just advance the index without
			// placing a given.
			if c >= '1' && c <= '9' {
				p.GivenValue(i/9, i%9, int(c-'0'))
			}
			i++
		}

		if i >= 81 {
			// If we've filled all 81 cells, then we're done.
			// Just ignore any remaining lines.
			break
		}
	}

	if i < 81 {
		return nil, errPuzzleFormat("not enough cells")
	}

	return p, nil
}

func FromArray(grid [][]int) *Puzzle {
	b := NewPuzzle()
	for r := range 9 {
		for c := range 9 {
			if grid[r][c] != 0 {
				b.GivenValue(r, c, grid[r][c])
			}
		}
	}
	return b
}
