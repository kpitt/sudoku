package solver

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func printProgress(format string, a ...any) {
	color.Yellow(format, a...)
}

func printChecking(name string) {
	printProgress("Checking %q technique", name)
}

func (s *Solver) PrintStep(step *SolutionStep) {
	fmt.Fprintln(os.Stderr, s.FormatStep(step))
}
