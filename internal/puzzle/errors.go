package puzzle

import (
	"fmt"
	"os"
	"strings"
)

func boardStateError(msg string) {
	fatalError("invalid board state", msg)
}

func fatalError(msgs ...string) {
	msg := strings.Join(msgs, ": ")
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}
