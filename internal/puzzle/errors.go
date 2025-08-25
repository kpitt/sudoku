package puzzle

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func errPuzzleFormat(format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("invalid puzzle format: %s", msg)
}

func errPuzzleState(format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("invalid puzzle state: %s", msg)
}

func puzzleStateError(msg string) {
	fatalError("invalid puzzle state", msg)
}

func fatalError(msgs ...string) {
	msg := strings.Join(msgs, ": ")
	fmt.Fprintln(os.Stderr, color.HiRedString("error: %s", msg))
	os.Exit(1)
}
