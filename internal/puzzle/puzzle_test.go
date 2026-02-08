package puzzle

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/kpitt/sudoku/internal/bitset"
)

func TestNewPuzzle(t *testing.T) {
	p := NewPuzzle()
	if p.IsSolved() {
		t.Error("New puzzle should not be solved")
	}

	// Check initial cell state
	// TODO: r,c -> index?
	for r := range 9 {
		for c := range 9 {
			cell := p.Get(r, c)
			if cell.Row != r || cell.Col != c {
				t.Errorf("cell at (%d, %d) has incorrect coordinates (%d, %d)", r, c, cell.Row, cell.Col)
			}
			if cell.IsSolved() {
				t.Errorf("cell at (%d,%d) should be empty", r, c)
			}
			if cell.Candidates != bitset.BitSet16(allDigitBits) { // All candidates should be present
				t.Errorf("cell at (%d, %d) should have all candidates", r, c)
			}
		}
	}
}

func TestPlaceValue(t *testing.T) {
	p := NewPuzzle()

	// Place 5 at index 0 (row 0, col 0)
	success := p.PlaceValue(0, 5)
	if !success {
		t.Error("PlaceValue failed for valid move")
	}

	c := p.Cell(0)
	if !c.IsSolved() || c.Value() != 5 {
		t.Error("Cell value not set correctly")
	}
	if c.NumCandidates() != 0 {
		t.Error("Solved cell should have 0 candidates")
	}

	// Check row peer (0, 1) - should not have 5 as candidate
	if p.Cell(1).HasCandidate(5) {
		t.Error("Row peer should verify candidate 5 removed")
	}

	// Check col peer (1, 0)
	if p.Cell(9).HasCandidate(5) {
		t.Error("Col peer should verify candidate 5 removed")
	}
	// Check box peer (1, 1)
	if p.Cell(10).HasCandidate(5) {
		t.Error("Box peer should verify candidate 5 removed")
	}
}

func TestPlaceValue_AlreadySolved(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		p := NewPuzzle()
		p.PlaceValue(0, 5)
		p.PlaceValue(0, 6)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestPlaceValue_AlreadySolved")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		expected := "cell r1c1 is already solved (value=5)"
		if !strings.Contains(stderr.String(), expected) {
			t.Errorf("Expected error message containing %q, got %q", expected, stderr.String())
		}
		return
	}
	t.Fatalf("Process ran with err %v, want exit status 1", err)
}

func TestValidateSolution(t *testing.T) {
	p := NewPuzzle()

	// Empty puzzle is invalid solution
	if err := p.ValidateSolution(); err == nil {
		t.Error("ValidateSolution should fail for empty puzzle")
	}

	// Partially filled with conflict
	p.PlaceValue(0, 5)
	p.PlaceValue(9, 5) // Row conflict (though PlaceValue removes candidate, let's force check logic if we could)

	// Note: PlaceValue removes candidates, so it prevents us from easily creating an invalid state
	// via standard methods unless we manipulate internals or ignore candidate checks (which PlaceValue doesn't enforce strictly on "Is valid move", it just does it).
	// However, ValidateSolution checks for duplicates.

	// Let's manually inject a bad value to test ValidateSolution logic purely
	p.Cell(2).value = 5 // Force a 3rd 5 in the row

	err := p.ValidateSolution()
	if err == nil {
		t.Error("Should detect duplicate in row")
	}
}

func TestCell_Box(t *testing.T) {
	tests := []struct {
		r, c, expected int
	}{
		{0, 0, 0},
		{0, 8, 2},
		{4, 4, 4},
		{8, 0, 6},
		{8, 8, 8},
	}

	for _, tt := range tests {
		c := NewCell(tt.r, tt.c)
		if got := c.Box(); got != tt.expected {
			t.Errorf("Cell(%d,%d).Box() = %d; want %d", tt.r, tt.c, got, tt.expected)
		}
	}
}
