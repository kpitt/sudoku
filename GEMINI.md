# Sudoku Solver

A robust Sudoku solver written in Go, featuring a hybrid approach that combines human-like deductive techniques with an optimized Dancing Links (Algorithm X) brute-force fallback.

## Overview

This project implements a Sudoku solver that aims to solve puzzles efficiently while also providing insights into the solving process.

*   **Hybrid Solving Engine:**
    *   **Deductive Techniques:** First attempts to solve the puzzle using logic patterns (e.g., Naked Singles, Hidden Singles) similar to how a human plays.
    *   **Dancing Links (Algorithm X):** If logic fails, it seamlessly switches to a highly optimized exact cover algorithm (Dancing Links) to find the solution via backtracking.
*   **Input:** Accepts puzzles via standard input as 9 lines of 9 characters (digits 1-9 for values, other characters for empty cells).
*   **Performance:** optimized for speed with careful memory management and efficient data structures (bitsets).

## Getting Started

### Prerequisites

*   **Go:** Version 1.24 or higher.
*   **Make:** For running build and test commands.

### Building

To build the executable:

```bash
make build
```

This will create the `bin/sudoku` binary.

### Running

The solver reads from `stdin`. You can pipe a puzzle file to it:

```bash
# Run with a sample puzzle
cat test/expert1.txt | bin/sudoku
```

Or run interactively:

```bash
bin/sudoku
# Paste your puzzle grid here...
```

### Testing

Run the test suite:

```bash
make test
```

Run benchmarks:

```bash
make bench
```

## Project Structure

*   `cmd/sudoku/`: Contains the `main.go` entry point for the CLI application.
*   `internal/`: Private application code.
    *   `puzzle/`: Core data structures for the Sudoku grid, cells, and parsing logic.
    *   `solver/`: The heart of the application. Contains the `Solver` struct, deductive techniques, and the `DancingLinks` implementation.
    *   `bitset/`: Helper package for efficient bitwise operations used in constraint tracking.
    *   `set/`: General set implementations.
*   `test/`: A comprehensive collection of Sudoku puzzles of varying difficulty (beginner, advanced, expert, impossible) used for testing and benchmarking.
*   `bench/`: Stores baseline benchmark results for performance tracking.
*   `bench.sh`: Script to run standardized performance benchmarks.

## Development

The `Makefile` provides several useful targets for development:

*   `make fmt`: Format code using `go fmt`.
*   `make lint`: Run linters (requires `golangci-lint`).
*   `make vet`: Run `go vet`.
*   `make test-coverage`: Generate and view a test coverage report.
*   `make dev`: Runs format, vet, and tests in one go.
