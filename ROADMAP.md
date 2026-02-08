# Sudoku Solver Roadmap

This document tracks the implementation status of various Sudoku solving techniques,
classified by human difficulty.

## Status Legend

- [x] Implemented
- [ ] Todo / Planned
- [-] Stubbed (Defined but not implemented)

## 1. Basic Techniques

*Fundamental logical deductions required for almost all puzzles.*

- [x] **Naked Single**: A cell has only one possible candidate.
- [x] **Hidden Single**: A candidate appears in only one cell within a house (row, col, box).
- [x] **Naked Pair / Triple / Quadruple**: $N$ cells in a house share $N$ candidates.
- [x] **Hidden Pair / Triple / Quadruple**: $N$ candidates are restricted to $N$ cells in a house.
- [x] **Locked Candidates**:
    - [x] **Pointing**: Candidates in a box are restricted to a single line (eliminates from line).
    - [x] **Claiming**: Candidates in a line are restricted to a single box (eliminates from box).

## 2. Intermediate Techniques

*Patterns involving intersections and simple chains.*

- [x] **X-Wing**: Two lines have a candidate in the same two positions.
- [x] **XY-Wing (Y-Wing)**: A "bent" triple of bivalue cells (Pivot + 2 Pincers).
- [x] **XYZ-Wing**: A variation of XY-Wing where the pivot has 3 candidates.
- [x] **Skyscraper**: A specific Turbot Fish pattern (offset X-Wing).
- [x] **2-String Kite**: Two conjugate pairs connected by a box.
- [-] **Empty Rectangle**: Intersection logic within a box.
- [-] **W-Wing**: Two identical bivalue cells connected by a strong link.
- [-] **Remote Pair**: A chain of identical bivalue cells.

## 3. Advanced Techniques

*Complex patterns and chaining logic.*

- [x] **Swordfish**: Three lines have candidates in the same three positions.
- [x] **Jellyfish**: Four lines have candidates in the same four positions.
- [-] **Simple Coloring (Color Chain)**: Single-digit conjugate chains (A/B logic).
- [-] **Multi-Coloring**: Multi-digit conjugate chains.
- [-] **Finned Fish**:
    - [-] Finned X-Wing
    - [-] Finned Swordfish
    - [-] Finned Jellyfish
- [-] **Unique Rectangle**:
    - [x] Type 1 (Naked pair in 3 corners).
    - [-] Type 2 (One corner has extra candidate).
    - [-] Type 3 (Subset / Pseudo-cells).
    - [-] Type 4 (Conjugate pair).
- [-] **Avoidable Rectangle**: Uses placed values to find contradictions.

## 4. Expert Techniques

*Generalized set theory and long chains.*

- [-] **X-Chain**: Chains that use only one candidate value.
- [-] **XY-Chain**: Chains that use only bivalue cells.
- [ ] **AIC (Alternating Inference Chains)**: Generalized chaining (Strong/Weak links).
- [ ] **ALS (Almost Locked Sets)**:
    - [ ] ALS-XZ
    - [ ] ALS-XY-Wing
- [-] **BUG+1**: Bivalue Universal Grave.
- [-] **Sue de Coq**: Set intersection logic.

## 5. Fallback (Brute Force)

- [x] **Dancing Links (DLX)**: Algorithm X brute-force solver (used when deductive logic fails).

## 6. Performance Optimizations

*Transitioning from object-oriented/pointer-heavy logic to high-performance data-oriented design.*

### Phase 1: Data Structures

- [x] **Bitmask Candidates**: Replace `set.Set[int]` with `uint16` bitmasks (O(1) operations).
- [x] **Flattened Board**: Replace `[9][9]*Cell` with `[81]Cell` (contiguous memory, cache locality).
- [x] **Zero-Allocation**: Remove heap allocations for `ValSet`, `LocSet` and `LocValMap`.

### Phase 2: Tables & Lookups

- [ ] **Pre-computed Peers**: Implement `PeerLookup [81][20]uint8` to replace dynamic peer calculation.
- [ ] **Pre-computed Houses**: Implement `HouseLookup [27][9]uint8` to replace `House` structs and maps.
- [ ] **Remove Maps**: Eliminate `House.Unsolved` map in favor of direct array iteration.

### Phase 3: Algorithm Optimization

- [ ] **Bitwise Techniques**: Rewrite all solver techniques to use bitwise logic (POPCNT, AND, OR, XOR).
- [ ] **Branchless Iteration**: Update loops to use fixed-size array iteration.
- [ ] **Stack Allocation**: Ensure the main `Solver` and `Board` structs fit entirely on the stack.
