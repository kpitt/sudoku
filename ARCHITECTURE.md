# Sudoku Project Architecture

## Project Overview

This project is a Sudoku solver that employs a hybrid approach: it prioritizes deductive reasoning (human-like strategies) for speed and educational value, falling back to a brute-force algorithm (Dancing Links) only when necessary.

## Directory Structure

- **`cmd/sudoku/`**: Application entry point.
- **`internal/puzzle/`**: The data model representing the grid and its state.
- **`internal/solver/`**: The core logic engine containing deductive strategies
  and the backtracking solver.
- **`internal/set/`**: A generic Set data structure utility.
- **`internal/bitset/`**: A specialized set implementation using bitmasks for efficient bitwise operations.
- **`test/`**: A comprehensive collection of Sudoku puzzles of varying difficulty (beginner, advanced, expert, impossible) used for testing and benchmarking.
- **`test/techniques/`**: A collection of Sudoku puzzles designed to test specific solving techniques.
- **`bench/`**: Stores baseline benchmark results for performance tracking.
- **`bench.sh`**: Script to run standardized performance benchmarks.

## Core Components

### 1. `internal/puzzle` (The Data Model)

- **`Puzzle`**: The central data structure representing the 9x9 grid.
    - **Responsibilities**: Maintains the 2D array of `Cell` objects and tracks `unsolvedCounts` (both total and per-digit) to allow for O(1) checks on whether the puzzle or specific digits are solved.
    - **State Management**: It is the single source of truth for the board state.
- **`Cell`**: Represents an individual square (0-8, 0-8).
    - **Responsibilities**: Holds the confirmed value (or 0 if empty) and a `set.Set` of valid `Candidates`.
    - **Logic**: Handles its own mutations, such as `PlaceValue` (which clears candidates) and `RemoveCandidate`.
- **IO**: `reader.go` and `printer.go` handle parsing standard Sudoku formats and rendering the grid to the console.

### 2. `internal/solver` (The Logic)

- **`Solver`**: The orchestrator.
    - **Responsibilities**: Holds a reference to the `Puzzle` and a list of `Technique` objects. It manages the solving loop.
- **`House`**: A critical optimization component.
    - **Concept**: Represents a Row, Column, or Box (3x3 area).
    - **Mechanism**: Maintains a `ValLocMap` (Value Location Map), which maps a digit to the set of cell indices in that house where the digit can still be placed. This allows techniques to efficiently find "Hidden Singles" or other patterns without constantly re-scanning the grid.
- **`Technique`**: A functional interface for deductive strategies.
    - **Examples**: Naked Pairs, X-Wing, Swordfish, Jellyfish.
    - **Flow**: The solver iterates through these techniques. If a technique deduces a move or elimination, it updates the `Puzzle` state, and the loop restarts.
- **`DancingLinks`**: The "nuclear option."
    - **Algorithm**: Implements Knuth's Algorithm X using the Dancing Links (DLX) data structure.
    - **Role**: Used as a fallback when all deductive techniques fail to find a solution. It performs a highly efficient exact cover search.

## Architectural Interaction

1. **Initialization**:
    - `main.go` reads input and initializes a `solver.Solver`.
    - The solver creates 27 `House` objects (9 rows, 9 columns, 9 boxes) to index the initial state.

2. **Deductive Solve Loop**:
    - The `Solver.Solve()` method runs a priority loop of `Technique` functions.
    - When a technique finds a move, it calls `Solver.applyStep()`.
    - `applyStep()` invokes `Puzzle.PlaceValue()` or `Cell.RemoveCandidate()`.

3. **State Propagation**:
    - `Puzzle.PlaceValue()` triggers `updatePuzzleState()`.
    - This removes the placed value from the candidate sets of all peer cells (same row, col, box).
    - The `Solver` listens to these changes to keep its `House` maps synchronized.

4. **Fallback**:
    - If the deductive loop completes a full pass without changes and the puzzle remains unsolved, the `Solver` invokes `DancingLinks`.
    - DLX converts the current state into an Exact Cover matrix, solves it, and writes the result back to the `Puzzle`.

## Key Entry Points

- **`cmd/sudoku/main.go`**: `main()` - Executable entry point.
- **`internal/puzzle/puzzle.go`**: `FromFile(io.Reader)` - Loading logic.
- **`internal/solver/solver.go`**: `NewSolver(p, opts).Solve()` - Primary solving routine.
