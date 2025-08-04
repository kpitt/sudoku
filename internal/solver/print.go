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

func printFound(name string, r, c int, val int) {
	fmt.Fprintf(os.Stderr, "%s: %s=%d\n", name, formatCellRef(r, c), val)
}

func printEliminate(name string, r, c int, val int) {
	fmt.Fprintf(os.Stderr, "%s: => %s<>%d\n", name, formatCellRef(r, c), val)
}

func formatCellRef(r, c int) string {
	return fmt.Sprintf("r%dc%d", r+1, c+1)
}
