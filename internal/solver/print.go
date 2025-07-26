package solver

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func printProgress(format string, a ...any) {
	color.Yellow(format, a...)
}

func printChecking(pattern string) {
	printProgress("Checking %q pattern", pattern)
}

func printFound(pattern string, r, c int, val int8) {
	fmt.Fprintf(os.Stderr, "%s: (%d,%d) = %d\n", pattern, r+1, c+1, val)
}

func printEliminate(pattern string, r, c int, val int8) {
	fmt.Fprintf(os.Stderr, "%s: eliminate %d at (%d,%d)\n", pattern, val, r+1, c+1)
}
