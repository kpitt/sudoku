package solver

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func printProgress(format string, a ...any) {
	color.Yellow(format, a...)
}

func printChecking(tk techniqueKind) {
	printProgress("Trying %q technique", techniqueName(tk))
}

func (step *SolutionStep) Print() {
	fmt.Fprintln(os.Stderr, step.Format())
}
