# Sudoku Solver with Dancing Links Algorithm

A comprehensive Sudoku solver implementation in Go featuring both traditional logical solving patterns and the Dancing Links algorithm for exact cover problems.

## Features

- **Traditional Solver**: Implements human-style solving patterns including Naked Singles, Hidden Singles, X-Wings, XY-Wings, and more
- **Dancing Links Algorithm**: Knuth's Algorithm X implementation for guaranteed solutions to any valid Sudoku puzzle
- **Comprehensive Testing**: Full test coverage for all solving algorithms
- **Performance Optimized**: Efficient data structures and algorithms for fast solving
- **Interactive Demonstrations**: Example programs showing different solving approaches

## Project Structure

```
sudoku/
├── cmd/
│   └── dancing_links_demo/     # Interactive Dancing Links demonstration
├── internal/
│   ├── puzzle/                 # Puzzle and cell data structures
│   ├── set/                    # Set operations for candidates
│   └── solver/                 # Solving algorithms
│       ├── solver.go           # Traditional pattern-based solver
│       ├── dancing_links.go    # Dancing Links algorithm implementation
│       ├── dancing_links_util.go # Utilities and statistics
│       └── patterns.go         # Human-style solving patterns
├── examples/                   # Usage examples
└── test/                      # Test data and fixtures
```

## Dancing Links Algorithm

The Dancing Links algorithm (Algorithm X) models Sudoku as an exact cover problem with 324 constraints:

- **81 Cell Constraints**: Each cell must contain exactly one digit
- **81 Row Constraints**: Each row must contain digits 1-9 exactly once
- **81 Column Constraints**: Each column must contain digits 1-9 exactly once
- **81 Box Constraints**: Each 3×3 box must contain digits 1-9 exactly once

### Key Features

- **Guaranteed Solution**: Will find a solution if one exists
- **Efficient Backtracking**: O(1) cover/uncover operations using doubly-linked circular lists
- **MRV Heuristic**: Chooses columns with minimum remaining values first
- **Automatic Constraint Propagation**: Covering operations naturally handle constraint propagation

### Performance

- **Easy Puzzles**: < 1ms
- **Medium Puzzles**: 1-10ms
- **Hard Puzzles**: 10-100ms
- **Extreme Puzzles**: 100ms-1s

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/kpitt/sudoku/internal/puzzle"
    "github.com/kpitt/sudoku/internal/solver"
)

func main() {
    // Create a new puzzle
    p := puzzle.NewPuzzle()

    // Set up your puzzle (0 for empty cells)
    givens := [][]int{
        {5, 3, 0, 0, 7, 0, 0, 0, 0},
        {6, 0, 0, 1, 9, 5, 0, 0, 0},
        // ... more rows
    }

    // Fill the puzzle grid with given values
    for r := 0; r < 9; r++ {
        for c := 0; c < 9; c++ {
            if givens[r][c] != 0 {
                p.Grid[r][c].GivenValue(givens[r][c])
            }
        }
    }

    // Solve using Dancing Links
    dl := solver.NewDancingLinks(b)
    if dl.Solve() {
        fmt.Println("Puzzle solved!")
        // Access solution through b.Cells
    }
}
```

### Using the Traditional Solver

```go
// Create solver with logical patterns
s := solver.NewSolver(p)
s.Solve() // Uses human-style solving patterns

// Or use Dancing Links as fallback
solved := s.SolveDancingLinks()
```

### Advanced Usage with Statistics

```go
options := &solver.DancingLinksOptions{
    EnableDebugging: true,
    TimeLimit:       5 * time.Second,
    MaxSolutions:    1,
}

solved, stats, err := solver.SolveWithDancingLinks(p, options)
if solved {
    stats.PrintStats() // Shows detailed solving statistics
}
```

## Running Examples

### Basic Dancing Links Example
```bash
go run examples/dancing_links_basic.go
```

### Interactive Demonstration
```bash
go run ./cmd/dancing_links_demo
```

## Testing

Run all tests:
```bash
go test ./...
```

Run Dancing Links specific tests:
```bash
go test ./internal/solver -v -run TestDancingLinks
```

Run benchmarks:
```bash
go test ./internal/solver -bench=BenchmarkDancingLinks
```

## Implementation Details

### Data Structures

The Dancing Links implementation uses doubly-linked circular lists for efficient operations:

```go
type Node struct {
    Left, Right, Up, Down *Node
    Column                *ColumnNode
    RowID                 int
}

type ColumnNode struct {
    Node
    Size int    // number of nodes in this column
    Name string // column identifier for debugging
}
```

### Algorithm Complexity

- **Time**: O(9^k) worst case, where k is the number of empty cells
- **Space**: O(n) where n is the puzzle size
- **Practical Performance**: Excellent due to constraint propagation and MRV heuristic

### Matrix Structure

- **Columns**: 324 constraints (81 each for cells, rows, columns, boxes)
- **Rows**: Up to 729 possible (row, column, value) combinations
- **Density**: Typically 1-2% (very sparse matrix)

## Comparison with Traditional Approaches

| Aspect | Traditional Solver | Dancing Links |
|--------|-------------------|---------------|
| **Approach** | Pattern-based logical deduction | Exact cover with backtracking |
| **Guarantees** | May get stuck on hard puzzles | Always finds solution if exists |
| **Performance** | Fast on easy/medium puzzles | Consistent across all difficulties |
| **Memory** | Low memory usage | Higher due to matrix structure |
| **Human-like** | Mimics human solving | Pure algorithmic approach |

## Dependencies

- `github.com/fatih/color` - Terminal colors for output
- `github.com/mattn/go-isatty` - TTY detection
- Go 1.24+

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -am 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## References

- Knuth, Donald E. "Dancing Links" (2000)
- "The Art of Computer Programming, Volume 4A" - Donald Knuth
- Various Sudoku solving techniques and patterns

## Acknowledgments

- Donald Knuth for the Dancing Links algorithm
- The Sudoku community for documenting solving patterns
- Contributors to this implementation
