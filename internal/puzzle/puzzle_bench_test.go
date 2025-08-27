package puzzle

import (
	"testing"
)

// BenchmarkSolveMemory benchmarks memory allocations for different difficulty levels.
func BenchmarkParseString(b *testing.B) {
	var testCases = []struct {
		name   string
		puzzle string
	}{
		{"Beginner", ".942......5.....29.2.3.6.1..6..89.4.7....3..6.3..7..8..7.5.2..421....5.8.....83.."},
		{"Advanced", ".26..39...154.........5.13.2...6.5.8....2..6.8.3.9...4.8...2....31..7492...5..7.."},
		{"Expert", "..7...41..6.4.....1...5..6.2....4..8..8.3.1..9..8....2.5..9.2.66....7.8..2......."},
		{"Pro", "......63..5...27.4...1.6.2......9.81...4.3.5.37.5......9.6.4...7.18...4..4......."},
		{"Impossible", "47...9....9.1..3......8.27....5.19.4....7....6.4..8....58.9....3....6......8...2."},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, err := FromString(tc.puzzle)
				if err != nil {
					b.Fatalf("Failed to load %s puzzle: %v", tc.name, err)
				}
			}
		})
	}
}
