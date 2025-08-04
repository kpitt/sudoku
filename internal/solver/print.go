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
	printProgress("Trying %q technique", name)
}

func printFound(name string, r, c int, val int8) {
	fmt.Fprintf(os.Stderr, "%s: (%d,%d) = %d\n", name, r+1, c+1, val)
}

func printEliminate(name string, r, c int, val int8) {
	fmt.Fprintf(os.Stderr, "%s: eliminate %d at (%d,%d)\n", name, val, r+1, c+1)
}
