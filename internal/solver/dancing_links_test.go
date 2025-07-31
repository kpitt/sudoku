package solver

import (
	"testing"

	"github.com/kpitt/sudoku/internal/puzzle"
)

func TestDancingLinksBasic(t *testing.T) {
	// Create a simple sudoku puzzle
	b := puzzle.NewPuzzle()

	// Set up a partially filled puzzle
	testPuzzle := [][]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}

	// Fill the puzzle grid with the test puzzle values
	givenCount := 0
	for r := range 9 {
		for c := range 9 {
			if testPuzzle[r][c] != 0 {
				b.Grid[r][c].GivenValue(testPuzzle[r][c])
				givenCount++
			}
		}
	}

	// Create Dancing Links solver
	dl := NewDancingLinks(b)

	// Test that we can create the matrix
	if dl.Header == nil {
		t.Error("Dancing Links header not initialized")
	}

	// Test column creation
	columnCount := 0
	for col := dl.Header.Right; col != &dl.Header.Node; col = col.Right {
		columnCount++
	}

	// Columns for given values should not appear in the matrix.  There are 4
	// constraints per given value, so we expect 324 - givenCount*4 columns.
	expected := 324 - givenCount*4
	if columnCount != expected {
		t.Errorf("Expected %d columns, got %d", expected, columnCount)
	}
}

func TestDancingLinksColumnNaming(t *testing.T) {
	p := puzzle.NewPuzzle()
	dl := NewDancingLinks(p)

	tests := []struct {
		index    int
		expected string
	}{
		{0, "R1C1"},   // Cell constraint
		{80, "R9C9"},  // Last cell constraint
		{81, "R1#1"},  // First row constraint
		{161, "R9#9"}, // Last row constraint
		{162, "C1#1"}, // First column constraint
		{242, "C9#9"}, // Last column constraint
		{243, "B1#1"}, // First box constraint
		{323, "B9#9"}, // Last box constraint
	}

	for _, test := range tests {
		result := dl.getColumnName(test.index)
		if result != test.expected {
			t.Errorf("getColumnName(%d) = %s, expected %s", test.index, result, test.expected)
		}
	}
}

func TestDancingLinksNodeCreation(t *testing.T) {
	p := puzzle.NewPuzzle()

	// Set a single value to test node creation
	p.Grid[0][0].PlaceValue(5)

	dl := NewDancingLinks(p)

	// Should have at least one row
	if len(dl.Rows) == 0 {
		t.Error("No rows created")
	}

	// Test that rows are not created for solved cells
	foundSolvedCell := false
	for _, row := range dl.Rows {
		if row != nil {
			// Check if this row corresponds to our solved cell
			r, c, val := dl.decodeRow(row.RowID)
			if r == 0 && c == 0 && val == 5 {
				foundSolvedCell = true
				break
			}
		}
	}

	if foundSolvedCell {
		t.Error("Constraint for solved cell found in Dancing Links matrix")
	}
}

func TestDancingLinksCoverUncover(t *testing.T) {
	p := puzzle.NewPuzzle()
	dl := NewDancingLinks(p)

	// Get first column
	firstCol := dl.Header.Right.Column
	originalSize := firstCol.Size

	// Cover the column
	dl.cover(firstCol)

	// Check that column is removed from header list
	if dl.Header.Right == &firstCol.Node {
		t.Error("Column not properly covered")
	}

	// Uncover the column
	dl.uncover(firstCol)

	// Check that column is restored
	if dl.Header.Right != &firstCol.Node {
		t.Error("Column not properly uncovered")
	}

	// Check that size is restored
	if firstCol.Size != originalSize {
		t.Errorf("Column size not restored: expected %d, got %d", originalSize, firstCol.Size)
	}
}

func TestDancingLinksChooseColumn(t *testing.T) {
	p := puzzle.NewPuzzle()

	// Create a puzzle with some constraints to make column sizes different
	p.Grid[0][0].PlaceValue(1)
	p.Grid[0][1].PlaceValue(2)

	dl := NewDancingLinks(p)

	chosen := dl.chooseColumn()
	if chosen == nil {
		t.Error("chooseColumn returned nil")
	}

	// Verify that chosen column has minimum size
	minSize := chosen.Size
	for col := dl.Header.Right; col != &dl.Header.Node; col = col.Right {
		if col.Column.Size < minSize {
			t.Errorf("chooseColumn didn't choose minimum: found %d < %d", col.Column.Size, minSize)
		}
	}
}

func TestDancingLinksEmptyPuzzle(t *testing.T) {
	// Test with completely empty puzzle
	p := puzzle.NewPuzzle()
	dl := NewDancingLinks(p)

	// Should create 729 rows (9*9*9)
	expectedRows := 9 * 9 * 9
	if len(dl.Rows) != expectedRows {
		t.Errorf("Expected %d rows for empty puzzle, got %d", expectedRows, len(dl.Rows))
	}
}

func TestDancingLinksFullyConstrainedPuzzle(t *testing.T) {
	// Test with a fully solved puzzle
	p := puzzle.NewPuzzle()

	// Fill with a valid solution
	solution := [][]int{
		{5, 3, 4, 6, 7, 8, 9, 1, 2},
		{6, 7, 2, 1, 9, 5, 3, 4, 8},
		{1, 9, 8, 3, 4, 2, 5, 6, 7},
		{8, 5, 9, 7, 6, 1, 4, 2, 3},
		{4, 2, 6, 8, 5, 3, 7, 9, 1},
		{7, 1, 3, 9, 2, 4, 8, 5, 6},
		{9, 6, 1, 5, 3, 7, 2, 8, 4},
		{2, 8, 7, 4, 1, 9, 6, 3, 5},
		{3, 4, 5, 2, 8, 6, 1, 7, 9},
	}

	for r := range 9 {
		for c := range 9 {
			p.Grid[r][c].PlaceValue(solution[r][c])
		}
	}

	dl := NewDancingLinks(p)

	// Should not create any rows
	if len(dl.Rows) != 0 {
		t.Errorf("Expected 0 rows for fully solved puzzle, got %d", len(dl.Rows))
	}
}

// Benchmark tests
func BenchmarkDancingLinksCreation(b *testing.B) {
	p := puzzle.NewPuzzle()

	for b.Loop() {
		_ = NewDancingLinks(p)
	}
}

func BenchmarkDancingLinksColumnChoice(b *testing.B) {
	p := puzzle.NewPuzzle()
	dl := NewDancingLinks(p)

	for b.Loop() {
		_ = dl.chooseColumn()
	}
}
