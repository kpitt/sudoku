package solver

import (
	"testing"

	"github.com/kpitt/sudoku/internal/puzzle"
)

// Benchmark test cases organized by difficulty level
var benchmarkCases = []struct {
	name       string
	difficulty string
	puzzle     string
}{
	// Beginner level puzzles
	{"Beginner1", "beginner", ".942......5.....29.2.3.6.1..6..89.4.7....3..6.3..7..8..7.5.2..421....5.8.....83.."},
	{"Beginner2", "beginner", "6..8..7...1.7....5..5.426.....2..91.96.....42.4..89....2..9....1....5.7.7.4..82.9"},
	{"Beginner3", "beginner", "..5.9...882.3.54.61.......3...42.67.7.2..3..5.691....29...3.8.7..39.1.645.1..8..."},
	{"Beginner4", "beginner", ".24..9...7...41629.8..7...387...596...........46397.18..7..6.8.36871.5........47."},

	// Advanced level puzzles
	{"Advanced1", "advanced", ".26..39...154.........5.13.2...6.5.8....2..6.8.3.9...4.8...2....31..7492...5..7.."},
	{"Advanced2", "advanced", "9.....1...3.2......27.8.46..518....4.894.265.2....5.1...8...59...31.6.8.........1"},
	{"Advanced3", "advanced", ".5..7..2.718.9.4.3...4....9.3185....9.......6....4193.5....2.67..7.......8.93..4."},
	{"Advanced4", "advanced", "4..5.618.3......9..51..3...7....9.6...2.483...9.7....4...98.4...3.6...1..154....8"},
	{"Advanced5", "advanced", "..7....6.41.8....5..85.7..1....945...3...5.4...218..7.7..4...8.3..2.6.1..2....3.6"},

	// Expert level puzzles
	{"Expert1", "expert", "..7...41..6.4.....1...5..6.2....4..8..8.3.1..9..8....2.5..9.2.66....7.8..2......."},
	{"Expert2", "expert", "...8.4....9....574....9..1.1....8...64..7..38...5....6....8.....56....4....6.2..3"},
	{"Expert3", "expert", "..2....6.96...1...4.8.7.1.....7.42.1....3....2.49.8.....5.9...3...2...18.7....92."},
	{"Expert4", "expert", "......7..28.6..5.....8..62...3.6.....5.7.3...4...5.9...76.9...8..9..1.4.........."},
	{"Expert5", "expert", ".1489..........5..7..3......8......5.32...6..9......8......2..8..7.6...1....7426."},

	// Pro level puzzles
	{"Pro1", "pro", "......63..5...27.4...1.6.2......9.81...4.3.5.37.5......9.6.4...7.18...4..4......."},
	{"Pro2", "pro", ".81.......9......4...97..3.4..8..6...78..245...2..3..7.3..46...1..............21."},
	{"Pro3", "pro", ".....7.929..6.25..3......8...24......7..6..5.......1...5.....3...69.1..414.7....."},
	{"Pro4", "pro", "...5..93......9.........81.2..96..5..1.4.5.2..3..82..41.9...............4571.8..."},
	{"Pro5", "pro", ".6.....7.9..7...8..38..695.2.....3.....2.8.....6..4..9.936..14..7...5..6.1......."},

	// Impossible level puzzles
	{"Impossible1", "impossible", "..15..8...6.....4.5.8.......8..51.....58673.....3...6.1.......3.9.....1...7..62.."},
	{"Impossible2", "impossible", "47...9....9.1..3......8.27....5.19.4....7....6.4..8....58.9....3....6......8...2."},
	{"Impossible3", "impossible", "4....2..8..284..1...8..6........7.523.......786.....3....1..2...7...96..9..6....5"},
	// Impossible4 and Impossible5 currently require brute-force to complete.
	// {"Impossible4", "impossible", "...723.....3...6.........219....4.3.1...8...9.84.1....36.........8...7.....236..."},
	// {"Impossible5", "impossible", ".48.......3.2.85.....5.......9.75...8......6...53..8.......6.....47.3.5..7....31."},

	// Extra puzzles with unknown difficulty
	// TurbotFish requires Skyscraper and 2-String Kite techniques, which have not been implemented yet,
	// so it currently requires brute-force to complete the solution.
	{"TurbotFish", "", "000400100000705032032000700001080605070000020503010800008000560650803000007001000"},
	{"Extended1", "", ".93.1...4.5...7.....14.....6...9.....3.5.8.7.....7...97.....1.....1...5.5...4..6."},
}

// BenchmarkSolve benchmarks the Solve method across different difficulty levels.
func BenchmarkSolve(b *testing.B) {
	opts := &Options{EnableBruteForce: true}

	for _, tc := range benchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			// Load the puzzle once before the benchmark loop
			originalPuzzle, err := puzzle.FromString(tc.puzzle)
			if err != nil {
				b.Fatalf("Failed to load puzzle %s: %v", tc.name, err)
			}

			for b.Loop() {
				// Create a fresh copy of the puzzle for each iteration
				puzzleCopy := copyPuzzle(originalPuzzle)

				solver := NewSolver(puzzleCopy, opts)
				solver.Solve()

				// Verify the puzzle was solved (optional sanity check)
				if !puzzleCopy.IsSolved() {
					b.Fatalf("Puzzle %s was not fully solved", tc.name)
				}
			}
		})
	}
}

// BenchmarkSolveByDifficulty benchmarks puzzles grouped by difficulty level.
func BenchmarkSolveByDifficulty(b *testing.B) {
	difficultyGroups := make(map[string][]*puzzle.Puzzle)

	// Group puzzles by difficulty
	for _, tc := range benchmarkCases {
		// Ignore puzzles with unknown difficulty
		if tc.difficulty == "" {
			continue
		}

		if _, exists := difficultyGroups[tc.difficulty]; !exists {
			difficultyGroups[tc.difficulty] = make([]*puzzle.Puzzle, 0)
		}

		puzzle, err := puzzle.FromString(tc.puzzle)
		if err != nil {
			b.Fatalf("Failed to load puzzle %s: %v", tc.name, err)
		}
		difficultyGroups[tc.difficulty] = append(difficultyGroups[tc.difficulty], puzzle)
	}

	opts := &Options{}

	// Benchmark each difficulty level
	for difficulty, puzzles := range difficultyGroups {
		b.Run(difficulty, func(b *testing.B) {
			for i := 0; b.Loop(); i++ {
				// Cycle through all puzzles in this difficulty level
				puzzleIndex := i % len(puzzles)
				// Create a fresh copy of the puzzle for each iteration
				puzzleCopy := copyPuzzle(puzzles[puzzleIndex])

				solver := NewSolver(puzzleCopy, opts)
				solver.Solve()
			}
		})
	}
}

// copyPuzzle creates a deep copy of a puzzle for benchmarking.
// This ensures each benchmark iteration starts with a fresh puzzle state
func copyPuzzle(original *puzzle.Puzzle) *puzzle.Puzzle {
	// Create a new puzzle
	newPuzzle := puzzle.NewPuzzle()

	// Copy the grid state
	for i := range 81 {
		originalCell := original.Cell(i)

		// Copy the given values
		if originalCell.IsGiven {
			newPuzzle.GivenValue(i, originalCell.Value())
		}
	}

	return newPuzzle
}

// One puzzle from each difficulty level for memory profiling and quick
// performance comparisons.
var comparisonCases = []struct {
	name   string
	puzzle string
}{
	{"Beginner", ".942......5.....29.2.3.6.1..6..89.4.7....3..6.3..7..8..7.5.2..421....5.8.....83.."},
	{"Advanced", ".26..39...154.........5.13.2...6.5.8....2..6.8.3.9...4.8...2....31..7492...5..7.."},
	{"Expert", "..7...41..6.4.....1...5..6.2....4..8..8.3.1..9..8....2.5..9.2.66....7.8..2......."},
	{"Pro", "......63..5...27.4...1.6.2......9.81...4.3.5.37.5......9.6.4...7.18...4..4......."},
	{"Impossible", "47...9....9.1..3......8.27....5.19.4....7....6.4..8....58.9....3....6......8...2."},
}

// BenchmarkNewSolver benchmarks performance and memory allocations for initializing
// a new solver.
func BenchmarkNewSolver(b *testing.B) {
	for _, tc := range comparisonCases {
		b.Run(tc.name, func(b *testing.B) {
			thePuzzle, err := puzzle.FromString(tc.puzzle)
			if err != nil {
				b.Fatalf("Failed to load %s puzzle: %v", tc.name, err)
			}

			b.ReportAllocs()
			for b.Loop() {
				NewSolver(thePuzzle, nil)
			}
		})
	}
}

// BenchmarkSolveMemory benchmarks memory allocations for different difficulty levels.
func BenchmarkSolveMemory(b *testing.B) {
	opts := &Options{}

	for _, tc := range comparisonCases {
		b.Run(tc.name, func(b *testing.B) {
			originalPuzzle, err := puzzle.FromString(tc.puzzle)
			if err != nil {
				b.Fatalf("Failed to load %s puzzle: %v", tc.name, err)
			}

			b.ReportAllocs()
			for b.Loop() {
				// Create a fresh copy of the puzzle for each iteration
				puzzleCopy := copyPuzzle(originalPuzzle)

				solver := NewSolver(puzzleCopy, opts)
				solver.Solve()
			}
		})
	}
}

// BenchmarkComparison runs a quick comparison benchmark across all difficulty levels.
// This is useful for getting a quick overview of performance characteristics.
func BenchmarkComparison(b *testing.B) {
	opts := &Options{}

	for _, tc := range comparisonCases {
		b.Run(tc.name, func(b *testing.B) {
			originalPuzzle, err := puzzle.FromString(tc.puzzle)
			if err != nil {
				b.Fatalf("Failed to load %s puzzle: %v", tc.name, err)
			}

			b.ReportAllocs()

			for b.Loop() {
				// Create a fresh copy of the puzzle for each iteration
				puzzleCopy := copyPuzzle(originalPuzzle)

				solver := NewSolver(puzzleCopy, opts)
				solver.Solve()
			}
		})
	}
}

// BenchmarkBruteForce runs a quick comparison benchmark for a brute-force only
// solution across all difficulty levels.  This is useful for performance
// comparisons between the deductive solver and the brute-force solver.
func BenchmarkBruteForce(b *testing.B) {
	opts := &Options{}

	for _, tc := range comparisonCases {
		b.Run(tc.name, func(b *testing.B) {
			originalPuzzle, err := puzzle.FromString(tc.puzzle)
			if err != nil {
				b.Fatalf("Failed to load %s puzzle: %v", tc.name, err)
			}

			b.ReportAllocs()

			for b.Loop() {
				// Create a fresh copy of the puzzle for each iteration
				puzzleCopy := copyPuzzle(originalPuzzle)

				solver := NewSolver(puzzleCopy, opts)
				solver.Solve()
			}
		})
	}
}
