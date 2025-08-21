package puzzle

import (
	"fmt"

	"github.com/fatih/color"
)

const (
	borderTop    = "┌─────┬─────┬─────╥─────┬─────┬─────╥─────┬─────┬─────┐"
	borderBot    = "└─────┴─────┴─────╨─────┴─────┴─────╨─────┴─────┴─────┘"
	dividerMinor = "├─────┼─────┼─────╫─────┼─────┼─────╫─────┼─────┼─────┤"
	dividerMajor = "╞═════╪═════╪═════╬═════╪═════╪═════╬═════╪═════╪═════╡"
	edgeMinor    = "│"
	edgeMajor    = "║"
)

var (
	lockedValueColor = color.New(color.Bold, color.FgHiWhite)
	fixedValueColor  = color.New(color.Bold, color.FgHiYellow, color.BgHiBlack)
)

func (p *Puzzle) Print() {
	color.HiWhite(borderTop)
	for r, row := range p.Grid {
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

func (p *Puzzle) PrintUnsolvedCounts() {
	color.HiWhite("Unsolved Digits:")
	for i := range 9 {
		digit := int8(i + 1)
		if !p.IsDigitSolved(digit) {
			fmt.Printf("%d: %d remaining\n", digit, p.unsolvedCounts[digit])
		} else {
			fmt.Printf("%d: complete\n", digit)
		}
	}
	fmt.Printf("\n%s %d\n",
		color.HiWhiteString("Total Unsolved Cells:"),
		p.unsolvedCounts[0])
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
		if cell.IsSolved() {
			cellColor := lockedValueColor
			if cell.IsGiven {
				cellColor = fixedValueColor
			}
			if candidateRow == 1 {
				cellColor.Printf("  %d  ", cell.Value())
			} else {
				cellColor.Print("     ")
			}
		} else {
			cell.printCandidates(candidateRow)
		}
	}
	color.HiWhite(edgeMinor)
}

func (c *Cell) printCandidates(candidateRow int) {
	candidateBase := candidateRow*3 + 1
	for col := range 3 {
		if col > 0 {
			// Add a space between candidates.
			fmt.Print(" ")
		}
		candidate := int8(candidateBase + col)
		if c.HasCandidate(candidate) {
			fmt.Print(color.HiBlackString("%d", candidate))
		} else {
			fmt.Print(" ")
		}
	}
}
