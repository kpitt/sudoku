# Dancing Links Algorithm Implementation

This document describes the Dancing Links algorithm implementation for solving Sudoku puzzles using Knuth's Algorithm X.

## Overview

The Dancing Links algorithm, also known as Algorithm X, is a recursive backtracking algorithm that efficiently solves exact cover problems. For Sudoku, we model the puzzle as an exact cover problem where we must select exactly one row from a constraint matrix to satisfy all constraints.

## Algorithm Details

### Exact Cover Problem

A Sudoku puzzle can be represented as an exact cover problem with the following constraints:

1. **Cell Constraints** (81): Each cell must contain exactly one digit
2. **Row Constraints** (81): Each row must contain each digit 1-9 exactly once
3. **Column Constraints** (81): Each column must contain each digit 1-9 exactly once
4. **Box Constraints** (81): Each 3×3 box must contain each digit 1-9 exactly once

**Total: 324 constraints**

### Matrix Structure

The constraint matrix has:
- **324 columns**: One for each constraint
- **Up to 729 rows**: One for each possible (row, col, value) combination
- Each row has exactly **4 nodes**: One satisfying each constraint type

### Data Structure

The implementation uses doubly-linked circular lists for efficient operations:

```go
type Node struct {
    Left, Right, Up, Down *Node
    Column                *ColumnNode
    RowID                 int
}

type ColumnNode struct {
    Node
    Size int    // number of nodes in this column
    Name string // column identifier
}
```

### Core Operations

#### Cover Operation
Removes a column and all rows that intersect with it:
```go
func (dl *DancingLinks) cover(col *ColumnNode) {
    // Remove column header from list
    col.Right.Left = col.Left
    col.Left.Right = col.Right

    // Remove all intersecting rows
    for i := col.Down; i != &col.Node; i = i.Down {
        for j := i.Right; j != i; j = j.Right {
            j.Down.Up = j.Up
            j.Up.Down = j.Down
            j.Column.Size--
        }
    }
}
```

#### Uncover Operation
Restores a column and all intersecting rows (for backtracking):
```go
func (dl *DancingLinks) uncover(col *ColumnNode) {
    // Restore all intersecting rows
    for i := col.Up; i != &col.Node; i = i.Up {
        for j := i.Left; j != i; j = j.Left {
            j.Column.Size++
            j.Down.Up = j
            j.Up.Down = j
        }
    }

    // Restore column header
    col.Right.Left = &col.Node
    col.Left.Right = &col.Node
}
```

#### Search Algorithm
Recursive search with backtracking:
```go
func (dl *DancingLinks) search() bool {
    if dl.Header.Right == &dl.Header.Node {
        return dl.applySolution() // All columns covered
    }

    col := dl.chooseColumn() // MRV heuristic
    dl.cover(col)

    for r := col.Down; r != &col.Node; r = r.Down {
        dl.solution = append(dl.solution, r.RowID)

        // Cover intersecting columns
        for j := r.Right; j != r; j = j.Right {
            dl.cover(j.Column)
        }

        if dl.search() {
            return true
        }

        // Backtrack
        for j := r.Left; j != r; j = j.Left {
            dl.uncover(j.Column)
        }

        dl.solution = dl.solution[:len(dl.solution)-1]
    }

    dl.uncover(col)
    return false
}
```

## Key Optimizations

### 1. Minimum Remaining Values (MRV) Heuristic
Always choose the column with the fewest remaining options to minimize branching factor:

```go
func (dl *DancingLinks) chooseColumn() *ColumnNode {
    var chosen *ColumnNode
    minSize := int(^uint(0) >> 1) // max int

    for col := dl.Header.Right; col != &dl.Header.Node; col = col.Right {
        columnNode := col.Column
        if columnNode.Size < minSize {
            chosen = columnNode
            minSize = columnNode.Size
        }
    }

    return chosen
}
```

### 2. O(1) Operations
Doubly-linked circular lists enable constant-time cover/uncover operations, making backtracking very efficient.

### 3. Constraint Propagation
The covering operation automatically handles constraint propagation by removing incompatible choices.

## Column Naming Convention

Columns are named for easy debugging:

- **R0C0** to **R8C8**: Cell constraints (row 0-8, column 0-8)
- **R0#1** to **R8#9**: Row constraints (row 0-8, digit 1-9)
- **C0#1** to **C8#9**: Column constraints (column 0-8, digit 1-9)
- **B0#1** to **B8#9**: Box constraints (box 0-8, digit 1-9)

Box numbering: 0-2 (top row), 3-5 (middle row), 6-8 (bottom row)

## Usage

### Basic Usage
```go
// Create puzzle and set initial values
p := puzzle.NewPuzzle()
// ... set up puzzle ...

// Create Dancing Links solver
dl := solver.NewDancingLinks(p)

// Solve the puzzle
if dl.Solve() {
    fmt.Println("Puzzle solved!")
} else {
    fmt.Println("No solution found")
}
```

### Integration with Existing Solver
```go
// Use as part of the main solver
s := solver.NewSolver(p)
solved := s.SolveDancingLinks()
```

## Performance Characteristics

### Time Complexity
- **Best case**: O(1) when puzzle is already solved
- **Average case**: Exponential, but with excellent pruning
- **Worst case**: O(9^k) where k is the number of empty cells

### Space Complexity
- **Matrix storage**: O(729 × 4) = O(2916) nodes maximum
- **Recursion depth**: O(81) maximum (one per empty cell)
- **Total**: O(n) where n is puzzle size

### Practical Performance
- **Easy puzzles**: < 1ms
- **Medium puzzles**: 1-10ms
- **Hard puzzles**: 10-100ms
- **Extreme puzzles**: 100ms-1s

The algorithm typically outperforms logical solvers on hard puzzles where human-style deduction becomes insufficient.

## Advantages

1. **Guaranteed Solution**: Will find a solution if one exists
2. **Efficient Backtracking**: O(1) undo operations
3. **Natural Constraint Handling**: Automatic constraint propagation
4. **Hard Puzzle Performance**: Excels where logical methods fail
5. **Exact Cover Generality**: Can solve other exact cover problems

## Disadvantages

1. **Memory Usage**: Requires significant memory for the matrix
2. **Setup Overhead**: Matrix construction takes time
3. **Cache Performance**: Pointer-heavy structure may have poor cache locality
4. **Debugging Complexity**: Complex data structure makes debugging harder

## Testing

Run the tests to verify correctness:

```bash
go test ./internal/solver -v -run TestDancingLinks
```

Run the demonstration program:

```bash
go run ./cmd/dancing_links_demo
```

## Implementation Files

- `dancing_links.go`: Core algorithm implementation
- `dancing_links_test.go`: Comprehensive test suite
- `cmd/dancing_links_demo/main.go`: Interactive demonstration

## References

1. Knuth, Donald E. "Dancing Links" (2000)
2. "The Art of Computer Programming, Volume 4A" - Donald Knuth
3. "Exact Cover Problems and Dancing Links" - Various implementations
