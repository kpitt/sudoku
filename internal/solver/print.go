package solver

import (
	"fmt"

	"github.com/fatih/color"
)

func (s *Solver) printProgress(format string, a ...any) {
	if !s.EnableDebug {
		return
	}
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(color.Error, color.HiBlackString(">>> %s", msg))
}

func (s *Solver) printChecking(name string) {
	s.printProgress("Checking %q technique", name)
}

func (s *Solver) PrintStep(step *SolutionStep) {
	fmt.Println(s.FormatStep(step))
}

func (s *Solver) PrintSolution() {
	for i, step := range s.solution {
		fmt.Printf("%2d. %s\n", i+1, s.FormatStep(step))
	}
}
