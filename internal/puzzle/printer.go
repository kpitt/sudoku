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
	givenColor     = color.New(color.FgHiBlue)
	givenLegend    = givenColor.Sprint("Blue")
	solvedColor    = color.New(color.FgHiGreen)
	solvedLegend   = solvedColor.Sprint("Green")
	unsolvedColor  = color.New(color.FgHiBlack)
	unsolvedLegend = unsolvedColor.Sprint("Gray")
)

func (p *Puzzle) Print() {
	fmt.Println("┌───────┬───────┬───────┐")
	for r := range 9 {
		if r == 3 || r == 6 {
			fmt.Println("├───────┼───────┼───────┤")
		}
		fmt.Print("│ ")
		for c := range 9 {
			if c == 3 || c == 6 {
				fmt.Print("│ ")
			}
			cell := p.Grid[r][c]
			if cell.IsSolved() {
				if cell.IsGiven {
					givenColor.Printf("%d ", cell.Value())
				} else {
					solvedColor.Printf("%d ", cell.Value())
				}
			} else {
				unsolvedColor.Print("· ")
			}
		}
		fmt.Println("│")
	}
	fmt.Println("└───────┴───────┴───────┘")
	fmt.Printf("Legend: %s = Given, %s = Solved, %s = Empty\n",
		givenLegend, solvedLegend, unsolvedLegend)
}

func (p *Puzzle) PrintUnsolvedCounts() {
	color.HiWhite("Unsolved Digits (%d cells):", p.unsolvedCounts[0])
	for i := range 9 {
		digit := i + 1
		if !p.IsDigitSolved(digit) {
			fmt.Printf("%d: %d remaining\n", digit, p.unsolvedCounts[digit])
		} else {
			fmt.Printf("%d: complete\n", digit)
		}
	}
	fmt.Println()
}

func (p *Puzzle) PrintCandidateGrid() {
	fmt.Println(borderTop)
	for r, row := range p.Grid {
		if r != 0 {
			if r%3 == 0 {
				fmt.Println(dividerMajor)
			} else {
				fmt.Println(dividerMinor)
			}
		}
		printRow(row)
	}
	fmt.Println(borderBot)
	fmt.Printf("Legend: %s = Given, %s = Solved, %s = Candidate\n",
		givenLegend, solvedLegend, unsolvedLegend)
}

func printRow(row [9]*Cell) {
	for cr := range 3 {
		printCandidateRow(row, cr)
	}
}

func printCandidateRow(row [9]*Cell, candidateRow int) {
	for c, cell := range row {
		if c != 0 && c%3 == 0 {
			fmt.Print(edgeMajor)
		} else {
			fmt.Print(edgeMinor)
		}
		if cell.IsSolved() {
			if candidateRow == 1 {
				if cell.IsGiven {
					givenColor.Printf(" [%d] ", cell.Value())
				} else {
					solvedColor.Printf("  %d  ", cell.Value())
				}
			} else {
				fmt.Print("     ")
			}
		} else {
			cell.printCandidates(candidateRow)
		}
	}
	fmt.Println(edgeMinor)
}

func (c *Cell) printCandidates(candidateRow int) {
	candidateBase := candidateRow*3 + 1
	for col := range 3 {
		if col > 0 {
			// Add a space between candidates.
			fmt.Print(" ")
		}
		candidate := candidateBase + col
		if c.HasCandidate(candidate) {
			unsolvedColor.Printf("%d", candidate)
		} else {
			fmt.Print(" ")
		}
	}
}
