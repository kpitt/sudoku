package puzzle

import (
	"fmt"
	"os"
	"strings"
)

func errPuzzleFormat(format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("invalid puzzle format: %s", msg)
}

func puzzleStateError(msg string) {
	fatalError("invalid puzzle state", msg)
}

func fatalError(msgs ...string) {
	msg := strings.Join(msgs, ": ")
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}
