package board

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
)

const (
	borderTop    = "┌───┬───┬───╥───┬───┬───╥───┬───┬───┐"
	borderBot    = "└───┴───┴───╨───┴───┴───╨───┴───┴───┘"
	dividerMinor = "├───┼───┼───╫───┼───┼───╫───┼───┼───┤"
	dividerMajor = "╞═══╪═══╪═══╬═══╪═══╪═══╬═══╪═══╪═══╡"
	edgeMinor    = "│"
	edgeMajor    = "║"
)

func (b *Board) Print() {
	color.HiWhite(borderTop)
	for r, row := range b.cells {
		if r != 0 {
			if r%3 == 0 {
				color.HiWhite(dividerMajor)
			} else {
				color.HiWhite(dividerMinor)
			}
		}
		printRow(row)
	}
	color.HiWhite(borderBot)
}

func printRow(row [9]*Cell) {
	for cr := range 3 {
		printCandidateRow(row, cr)
	}
}

func printCandidateRow(row [9]*Cell, candidateRow int) {
	for c, cell := range row {
		if c != 0 && c%3 == 0 {
			fmt.Print(color.HiWhiteString(edgeMajor))
		} else {
			fmt.Print(color.HiWhiteString(edgeMinor))
		}
		if cell.IsLocked() {
			if candidateRow == 1 {
				fmt.Print(color.HiWhiteString(" %s ", cell.lockedValueString()))
			} else {
				fmt.Print("   ")
			}
		} else {
			cell.printCandidates(candidateRow)
		}
	}
	color.HiWhite(edgeMinor)
}

func (c *Cell) lockedValueString() string {
	valStr := strconv.Itoa(int(c.value))
	if c.isFixed {
		return color.HiYellowString(valStr)
	} else {
		return color.HiWhiteString(valStr)

	}
}

func (c *Cell) printCandidates(candidateRow int) {
	candidateBase := candidateRow*3 + 1
	for col := range 3 {
		candidate := int8(candidateBase + col)
		if c.IsCandidate(candidate) {
			fmt.Print(color.HiBlackString("%d", candidate))
		} else {
			fmt.Print(" ")
		}
	}
}
