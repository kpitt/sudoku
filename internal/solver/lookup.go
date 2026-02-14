package solver

// HouseLookup is a pre-computed table of cell indices for each house.
// There are 27 houses:
// - Indices 0-8: Rows 0-8
// - Indices 9-17: Columns 0-8
// - Indices 18-26: Boxes 0-8
var HouseLookup [27][9]int

func init() {
	InitHouseLookup()
}

// InitHouseLookup initializes the HouseLookup table.
func InitHouseLookup() {
	var houseIdx int

	// 1. Rows
	for r := range 9 {
		for c := range 9 {
			HouseLookup[houseIdx][c] = r*9 + c
		}
		houseIdx++
	}

	// 2. Columns
	for c := range 9 {
		for r := range 9 {
			HouseLookup[houseIdx][r] = r*9 + c
		}
		houseIdx++
	}

	// 3. Boxes
	for br := range 3 {
		for bc := range 3 {
			// Internal box index
			for i := range 9 {
				r := br*3 + i/3
				c := bc*3 + i%3
				HouseLookup[houseIdx][i] = r*9 + c
			}
			houseIdx++
		}
	}
}
