package solver

import "testing"

func TestHouseLookup(t *testing.T) {
	// Helper to check if a slice contains a value
	contains := func(slice *[9]int, val int) bool {
		for _, v := range slice {
			if v == val {
				return true
			}
		}
		return false
	}

	// Test a Row (House 0)
	row0 := &HouseLookup[0] // Row 0
	for c := 0; c < 9; c++ {
		idx := 0*9 + c
		if !contains(row0, idx) {
			t.Errorf("Row 0 missing cell %d", idx)
		}
	}

	// Test a Column (House 9)
	col0 := &HouseLookup[9] // Col 0 is 9th house (0-8 are rows)
	for r := 0; r < 9; r++ {
		idx := r*9 + 0
		if !contains(col0, idx) {
			t.Errorf("Col 0 missing cell %d", idx)
		}
	}

	// Test a Box (House 18)
	box0 := &HouseLookup[18] // Box 0 is 18th house (0-8 rows, 9-17 cols)
	// Box 0 cells: r0-2, c0-2
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			idx := r*9 + c
			if !contains(box0, idx) {
				t.Errorf("Box 0 missing cell %d", idx)
			}
		}
	}
}
