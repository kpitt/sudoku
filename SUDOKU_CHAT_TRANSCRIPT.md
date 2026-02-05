# Transcript: High-Performance Go Sudoku Solver Design

## 1. Data Structure Design

**User:**
I'm designing a program to solve Sudoku puzzles using well-known deductive solving techniques.
Suggest an appropriate data structure to represent the current puzzle board state and remaining candidates.

**Consultant:**

### Conceptual Analysis

Solving Sudoku deductively (as opposed to brute-force backtracking) is a **Constraint Satisfaction Problem (CSP)**.
The core of your logic will revolve around reducing the search space by eliminating possibilities.

To do this efficiently, your data structure must optimize for two specific operations:

1. **Intersection/Union:** Algorithms like "Hidden Pairs" or "X-Wing" require comparing the possibilities of multiple cells in a row, column, or box.
   You need to quickly answer questions like, "Which numbers appear as candidates in exactly two cells in this row?"
2. **State Management:** As you deduce a number, you must propagate that constraint to peers (removing that number from their candidate lists).

### Data Structure Selection

We will use a **Flat Fixed-Size Array (`[81]Cell`)** combined with **Bitmasks**.

1. **Bitmasks (`uint16`) for Candidates:**
   Instead of a `[]int` or `map[int]bool` to store candidates (e.g., `{1, 5, 9}`), we use a single `uint16`.
   Each bit position corresponds to a number.
   If the $k$-th bit is set, the number $k+1$ is a candidate.
   **Why?** Set operations (Union, Intersection, Difference) become single CPU instructions (bitwise OR `|`, AND `&`, AND NOT `&^`).
   This transforms O(N) set logic into O(1).

2. **Flattened Array (`[81]` vs `[9][9]`):**
   We use a single array of size 81.
   **Why?** It improves cache locality and simplifies iteration.
   You can loop 0–80 to check every cell without nested loops.
   Converting `(row, col)` to index is simple math: `i = row*9 + col`.

### Go Implementation

This implementation defines a memory-efficient structure with helper methods leveraging the `math/bits` standard library for population counts.

```go
package sudoku

import (
    "fmt"
    "math/bits"
)

// Constants for board dimensions
const (
    Size      = 9
    TotalCells = Size * Size
    // AllCandidates represents binary 111111111 (numbers 1-9)
    AllCandidates CandidateMask = 0x1FF 
)

// CandidateMask represents the set of possible values for a cell.
// Bit 0 = 1, Bit 1 = 2, ..., Bit 8 = 9.
type CandidateMask uint16

// Cell represents a single square on the board.
// Memory: 2 bytes (Mask) + 1 byte (Val) + 1 byte (padding) = 4 bytes per cell.
type Cell struct {
    Candidates CandidateMask // The set of possible numbers (if unsolved)
    Value      uint8         // The solved number (1-9), or 0 if empty
}

// Board represents the state of the puzzle.
// Memory: 81 * 4 bytes = 324 bytes. Extremely lightweight.
type Board struct {
    Cells [TotalCells]Cell
}

// NewBoard initializes a board with all candidates enabled for empty cells.
func NewBoard(initialGrid [Size][Size]uint8) *Board {
    b := &Board{}
    for r := 0; r < Size; r++ {
        for c := 0; c < Size; c++ {
            val := initialGrid[r][c]
            idx := r*Size + c
            
            if val != 0 {
                // Cell is pre-filled
                b.Cells[idx].Value = val
                b.Cells[idx].Candidates = 0 // No candidates needed for solved cells
            } else {
                // Cell is empty, all numbers are theoretically possible initially
                b.Cells[idx].Value = 0
                b.Cells[idx].Candidates = AllCandidates
            }
        }
    }
    // Note: A real solver would run an initial constraint propagation here
    // to remove invalid candidates based on the initialGrid values.
    return b
}

// IsSolved checks if the cell has a determined value.
func (c *Cell) IsSolved() bool {
    return c.Value != 0
}

// HasCandidate checks if a specific number (1-9) is still possible.
func (c *Cell) HasCandidate(num uint8) bool {
    if num < 1 || num > 9 {
        return false
    }
    bit := uint16(1 << (num - 1))
    return uint16(c.Candidates)&bit != 0
}

// RemoveCandidate removes a number from the possibilities.
// Returns true if the mask changed.
func (c *Cell) RemoveCandidate(num uint8) bool {
    if num < 1 || num > 9 {
        return false
    }
    bit := uint16(1 << (num - 1))
    if uint16(c.Candidates)&bit == 0 {
        return false // Candidate wasn't there
    }
    // Use Go's "AND NOT" operator (&^) to clear the bit
    c.Candidates = c.Candidates &^ CandidateMask(bit)
    return true
}

// CandidateCount returns how many possibilities remain.
// Useful for finding "Naked Singles" (count == 1).
func (c *Cell) CandidateCount() int {
    return bits.OnesCount16(uint16(c.Candidates))
}

// Example usage showing how to iterate a specific row
func (b *Board) GetRowMask(row int) CandidateMask {
    var union CandidateMask
    start := row * Size
    for i := start; i < start+Size; i++ {
        // Union of all candidates in this row (useful for deductive steps)
        union |= b.Cells[i].Candidates
    }
    return union
}
```

### Performance Note

1. **Time Complexity:**
   - **Access:** O(1) for getting/setting values.
   - **Set Operations:** Intersection/Union of candidates is O(1) (bitwise `&`, `|`).
   - **Counting:** `bits.OnesCount16` compiles to a POPCNT CPU instruction, making it O(1).
2. **Space Complexity:**
   - The entire `Board` struct is **~324 bytes**.
   - This is small enough to be allocated entirely on the **Stack**.
   - This dramatically reduces Garbage Collector (GC) pressure compared to slice-based implementations.
3. **Memory Layout:**
   - Because `Cell` is 4 bytes (after alignment), the `[81]Cell` array is contiguous in memory.
   - Iterating through it for board-wide validation is extremely cache-friendly.

---

## 2. Constraint Propagation (Naked Singles)

**User:**
Show the Constraint Propagation function.
For the remainder of this session, starting with this prompt, format all Go code using 4 spaces for indentation instead of the idiomatic hard tabs.
This will improve code readability in the chat interface.

**Consultant:**

### Conceptual Analysis

Constraint Propagation (specifically for "Naked Singles") acts as a chain reaction.

1. **Trigger:** When a cell is solved (e.g., Row 0, Col 0 becomes **5**), that number is no longer valid for any of its **peers** (cells in the same row, column, or 3x3 box).
2. **Action:** We iterate through all peers and mask out the bit corresponding to **5**.
3. **Ripple Effect:** If removing **5** from a peer leaves that peer with only **one candidate remaining** (a "Naked Single"), that peer is now effectively solved.
   We add it to a queue to process *its* peers, continuing until the board stabilizes or a contradiction is found.

### Peer Optimization

Calculating peers (Row, Col, Box) inside the hot loop is inefficient.
In a high-performance solver, we pre-compute these relationships into a static lookup table.

### Go Implementation

```go
package sudoku

import (
    "errors"
    "fmt"
    "math/bits"
)

// Pre-computed lookup table for peers.
// Global variable or singleton in a real app.
// Index [0-80] -> Slice of 20 peer indices.
var PeerLookup [81][]int

// InitializePeerLookup should be called once at startup.
func InitializePeerLookup() {
    for i := 0; i < 81; i++ {
        r, c := i/9, i%9
        boxStartR, boxStartC := (r/3)*3, (c/3)*3
        
        peerMap := make(map[int]bool)
        
        // Add Row and Column peers
        for k := 0; k < 9; k++ {
            peerMap[r*9 + k] = true // Row
            peerMap[k*9 + c] = true // Col
        }
        
        // Add Box peers
        for br := 0; br < 3; br++ {
            for bc := 0; bc < 3; bc++ {
                peerMap[(boxStartR+br)*9 + (boxStartC+bc)] = true
            }
        }
        
        // Remove self and convert to slice
        delete(peerMap, i)
        peers := make([]int, 0, 20)
        for p := range peerMap {
            peers = append(peers, p)
        }
        PeerLookup[i] = peers
    }
}

// Solver wraps the board to track propagation state.
type Solver struct {
    Board *Board
    // Queue for cells that have become solved but haven't propagated yet.
    // Fixed size buffer to avoid allocations in the hot loop.
    WorkQueue      [81]int 
    QueueHead      int
    QueueTail      int
}

func NewSolver(b *Board) *Solver {
    return &Solver{
        Board: b,
    }
}

// SetCell sets a value and triggers constraint propagation.
func (s *Solver) SetCell(idx int, val uint8) error {
    // 1. Set the value
    s.Board.Cells[idx].Value = val
    s.Board.Cells[idx].Candidates = 0
    
    // 2. Add to queue to start propagation
    s.push(idx)
    
    // 3. Process the ripple effect
    return s.propagate()
}

// propagate processes the WorkQueue until empty.
func (s *Solver) propagate() error {
    for s.QueueHead != s.QueueTail {
        // Pop current cell index
        currIdx := s.pop()
        currVal := s.Board.Cells[currIdx].Value
        
        // Get pre-computed peers
        peers := PeerLookup[currIdx]
        
        for _, peerIdx := range peers {
            peerCell := &s.Board.Cells[peerIdx]
            
            // Skip already solved cells
            if peerCell.IsSolved() {
                if peerCell.Value == currVal {
                    return errors.New("contradiction: duplicate value in peer group")
                }
                continue
            }
            
            // Try to remove the candidate
            if peerCell.RemoveCandidate(currVal) {
                // Check if this reduction created a new Naked Single
                count := peerCell.CandidateCount()
                
                if count == 0 {
                    return errors.New("contradiction: cell has no candidates left")
                }
                
                if count == 1 {
                    // We found a Naked Single! 
                    // Convert the remaining bit mask back to an integer (1-9).
                    // TrailingZeros16 returns the index of the set bit (0-8).
                    remainingVal := uint8(bits.TrailingZeros16(uint16(peerCell.Candidates))) + 1
                    
                    peerCell.Value = remainingVal
                    peerCell.Candidates = 0
                    
                    // Add to queue to propagate this new deduction
                    s.push(peerIdx)
                }
            }
        }
    }
    return nil
}

// Simple ring-buffer queue methods
func (s *Solver) push(idx int) {
    s.WorkQueue[s.QueueTail] = idx
    s.QueueTail = (s.QueueTail + 1) % 81
}

func (s *Solver) pop() int {
    val := s.WorkQueue[s.QueueHead]
    s.QueueHead = (s.QueueHead + 1) % 81
    return val
}
```

### Performance Note

1. **Complexity:** In the worst case (a cascading solve of the entire board), we visit every edge in the constraint graph once. This is roughly proportional to $81 \times 20$ operations.
2. **Bit Manipulation:** The line `bits.TrailingZeros16` is critical. It converts the candidate mask (e.g., `000010000`) to the integer value (`5`) in a single CPU instruction (O(1)), avoiding a loop over bits 0-8.
3. **Queue Allocation:** We use a fixed-size ring buffer `[81]int` inside the `Solver` struct. This generates **zero garbage** during the solve process, as opposed to using a `slice` that might grow/shrink.

---

## 3. Loop Optimization: Indices vs Pointers

**User:**
In the `peerIdx` loop, we have to look up the cell at that index and take its address each time through the loop.
Considering any optimizations that would be applied by the Go compiler, would there be any performance advantage to storing the pointers directly in the peer map instead of the indexes?

**Consultant:**

### Conceptual Analysis

The short answer is **no**.
In this specific use case (a highly iterative solver with a small, flat memory footprint), using pointers instead of indices would likely **degrade performance** rather than improve it.

### Why Indices Are Superior Here

1. **The "Static Topology" vs. "Dynamic Instance" Problem:**
   The relationships between cells are **immutable topology**.
   Indices allow us to compute `PeerLookup` once at program startup (globally) and reuse it for millions of different board instances.
   If we used pointers, every time you call `NewBoard`, you would be forced to rebuild the peer graph (generating 1,620 pointers) for *every single puzzle* you attempt to solve.

2. **CPU Instructions & Bounds Checking:**
   Modern CPUs are optimized for `&Cells[i]`.
   The instruction is a simple bit-shift and add, effectively "free" in the pipeline.
   The Go compiler (1.21+) is excellent at **Bounds Check Elimination (BCE)** since `peerIdx` comes from a controlled lookup table.

3. **Cache Locality (The "Fat Struct" Issue):**
   `Cell` is 4 bytes.
   The whole board fits in ~5 cache lines.
   Adding 20 pointers (`*Cell` is 8 bytes) per cell would increase the board size from **324 bytes** to **~13 KB**, thrashing the L1 cache.

### A Better Optimization: Compressed Indices

If you want to optimize the loop, the better approach is to reduce the memory bandwidth of the lookup table itself.
Since we only have 81 cells, we can use `uint8` (1 byte) instead of `int` (8 bytes) for the lookup table.

```go
// Use uint8 to save cache space (indices are 0-80)
// Use a fixed-size array [20]uint8 to avoid slice header overhead
var PeerLookup [81][20]uint8

func InitializePeerLookup() {
    for i := 0; i < 81; i++ {
        // ... (logic to find peers) ...
        
        // Fill the fixed array
        for k, p := range peers {
            PeerLookup[i][k] = uint8(p)
        }
    }
}

// Optimized loop in Solver
func (s *Solver) propagate() error {
    // ...
    // The compiler can now vectorize this access more easily
    // and the table stays in L1 cache.
    peers := &PeerLookup[currIdx] // Take pointer to avoid array copy
    
    for i := 0; i < 20; i++ {
        peerIdx := peers[i] 
        peerCell := &s.Board.Cells[peerIdx]
        // ...
    }
    // ...
}
```

---

## 4. Hidden Singles

**User:**
Describe the "Hidden Singles" implementation.

**Consultant:**

### Conceptual Analysis

**Hidden Singles** are the inverse of Naked Singles.
- **Naked Single:** A specific *cell* has only one possible candidate.
- **Hidden Single:** A specific *unit* (row, column, or box) has only one cell that can hold a specific number.

To find these, we must pivot our perspective from "Cell-centric" to "Unit-centric".
We iterate through every Row, Column, and Box to count the frequency of every candidate number.

### Data Structure Selection

1. **Unit Lookup Table:** Just like `PeerLookup`, we pre-compute `UnitLookup`—a static array `[27][9]uint8`.
2. **Stack-Allocated Counters:** We will use small fixed-size arrays `[10]uint8` to count occurrences, avoiding heap allocation.

### Go Implementation

```go
package sudoku

import (
    "fmt"
    "math/bits"
)

// Pre-computed lookup for the 27 units (9 Rows, 9 Cols, 9 Boxes).
// Dimensions: [27 units][9 cell indices]
var UnitLookup [27][9]uint8

func InitializeUnitLookup() {
    var unitIdx int
    
    // 1. Rows
    for r := 0; r < 9; r++ {
        for c := 0; c < 9; c++ {
            UnitLookup[unitIdx][c] = uint8(r*9 + c)
        }
        unitIdx++
    }
    
    // 2. Columns
    for c := 0; c < 9; c++ {
        for r := 0; r < 9; r++ {
            UnitLookup[unitIdx][r] = uint8(r*9 + c)
        }
        unitIdx++
    }
    
    // 3. Boxes
    for br := 0; br < 3; br++ {
        for bc := 0; bc < 3; bc++ {
            // Internal box index
            for i := 0; i < 9; i++ {
                r := br*3 + i/3
                c := bc*3 + i%3
                UnitLookup[unitIdx][i] = uint8(r*9 + c)
            }
            unitIdx++
        }
    }
}

// FindHiddenSingles scans all units for numbers that appear only once.
// It applies the first one it finds and returns true (triggering propagation).
// Returns false if no hidden singles are found.
func (s *Solver) FindHiddenSingles() (bool, error) {
    // Buffers to track counts and positions for numbers 1-9.
    // Index 0 is unused for clarity (1-based mapping).
    var counts [10]uint8
    var positions [10]uint8 
    
    // Iterate through all 27 units (Rows, Cols, Boxes)
    for u := 0; u < 27; u++ {
        // Reset counters for this unit
        // (Compiler optimizes this into a block zeroing instruction)
        counts = [10]uint8{}
        
        // Pass 1: Tally candidates in this unit
        for i := 0; i < 9; i++ {
            cellIdx := UnitLookup[u][i]
            cell := &s.Board.Cells[cellIdx]
            
            // Skip if cell is already solved
            if cell.IsSolved() {
                continue
            }
            
            // Iterate bits in the mask
            mask := uint16(cell.Candidates)
            for mask != 0 {
                // Get index of the lowest set bit (0-8)
                bitIdx := bits.TrailingZeros16(mask)
                num := bitIdx + 1
                
                // Increment count and record position
                counts[num]++
                positions[num] = uint8(cellIdx)
                
                // Clear the bit to continue to the next candidate
                mask &= mask - 1
            }
        }
        
        // Pass 2: Check for counts of exactly 1
        for num := 1; num <= 9; num++ {
            if counts[num] == 1 {
                targetIdx := int(positions[num])
                
                // We found a Hidden Single!
                // The number 'num' can ONLY go in 'targetIdx' for this unit.
                
                // Apply the move
                err := s.SetCell(targetIdx, uint8(num))
                if err != nil {
                    return false, err
                }
                
                // Return true immediately.
                // Setting a cell triggers 'propagate()', which might 
                // reveal more Naked Singles.
                return true, nil
            }
        }
    }
    
    return false, nil
}

// SolveStrategy combines Naked and Hidden singles.
func (s *Solver) SolveStrategy() error {
    for {
        // 1. Run basic constraint propagation (Naked Singles)
        if err := s.propagate(); err != nil {
            return err
        }
        
        // 2. If stuck, look for Hidden Singles
        found, err := s.FindHiddenSingles()
        if err != nil {
            return err
        }
        
        // If we didn't find any hidden singles, we are truly stuck.
        if !found {
            break
        }
    }
    return nil
}
```

---

## 5. Locked Candidates and Subsets

**User:**
Let's take a look at Locked Candidates, and Naked and Hidden Subsets (Pairs, Triples, Quadruples).

**Consultant:**

### Conceptual Analysis

1. **Locked Candidates (Intersection Removal):**
   - **Pointing:** If a candidate `N` in a **Box** is restricted to a single **Row**, then `N` cannot appear anywhere else in that Row outside the Box.
   - **Claiming:** If `N` in a **Row** is restricted to a single **Box**, then `N` cannot appear anywhere else in that Box.

2. **Subsets (Pairs, Triples, Quads):**
   - **Naked Subsets:** If $k$ cells in a unit contain *only* candidates from a set of size $k$ (e.g., two cells both have candidates `{2, 7}`), then `{2, 7}` can be removed from all other cells in that unit.

### Go Implementation

#### 1. Locked Candidates (Pointing & Claiming)

```go
// FindLockedCandidates searches for Pointing and Claiming reductions.
func (s *Solver) FindLockedCandidates() (bool, error) {
    changed := false
    
    // Iterate over all 9 Boxes to find Pointing pairs/triples
    for b := 0; b < 9; b++ {
        boxStartR, boxStartC := (b/3)*3, (b%3)*3
        
        // Check each number 1-9
        for num := uint8(1); num <= 9; num++ {
            var rMask, cMask uint16 // Bitmasks tracking which rows/cols contain 'num' in this box
            count := 0
            
            // 1. Scan the box to see where 'num' is possible
            for i := 0; i < 9; i++ {
                r, c := boxStartR + i/3, boxStartC + i%3
                idx := r*9 + c
                
                if s.Board.Cells[idx].HasCandidate(num) {
                    count++
                    // We map the row/col index (0-8) to a bit position
                    rMask |= 1 << r
                    cMask |= 1 << c
                }
            }
            
            // If the number appears 2 or 3 times (forming a pair/triple), check alignment
            if count >= 2 {
                // Check Row alignment (Pointing)
                // bits.OnesCount16(rMask) == 1 means all candidates are in the same Row
                if bits.OnesCount16(rMask) == 1 {
                    rowIdx := bits.TrailingZeros16(rMask)
                    // Eliminate 'num' from the rest of this Row (outside this Box)
                    if s.eliminateFromLine(rowIdx, num, b, true) {
                        changed = true
                    }
                }
                
                // Check Column alignment (Pointing)
                if bits.OnesCount16(cMask) == 1 {
                    colIdx := bits.TrailingZeros16(cMask)
                    // Eliminate 'num' from the rest of this Col (outside this Box)
                    if s.eliminateFromLine(colIdx, num, b, false) {
                        changed = true
                    }
                }
            }
        }
    }
    
    return changed, nil
}

// eliminateFromLine removes 'num' from a Row or Col, skipping a specific Box.
// isRow: true for Row, false for Column.
// skipBox: the box index (0-8) to protect from elimination.
func (s *Solver) eliminateFromLine(lineIdx int, num uint8, skipBox int, isRow bool) bool {
    changed := false
    for k := 0; k < 9; k++ {
        // Calculate cell index based on line type
        var r, c int
        if isRow {
            r, c = lineIdx, k
        } else {
            r, c = k, lineIdx
        }
        
        // Calculate which box this cell belongs to
        b := (r/3)*3 + (c/3)
        
        // If this cell is inside the "Pointing" box, SKIP it.
        if b == skipBox {
            continue
        }
        
        idx := r*9 + c
        if s.Board.Cells[idx].RemoveCandidate(num) {
            changed = true
             s.push(idx) // Ensure propagation picks this up
        }
    }
    return changed
}
```

#### 2. Naked Subsets (Pairs, Triples, Quads)

```go
// FindNakedSubsets looks for k cells in a unit that share a union of k candidates.
// k = 2 (Pairs), 3 (Triples), 4 (Quads).
func (s *Solver) FindNakedSubsets(k int) (bool, error) {
    changed := false
    
    // Iterate all 27 units
    for u := 0; u < 27; u++ {
        // 1. Collect interesting cells in this unit
        // We only care about cells with candidate count <= k (and > 1)
        type candidateInfo struct {
            idx  int
            mask CandidateMask
        }
        candidates := make([]candidateInfo, 0, 9)
        
        for i := 0; i < 9; i++ {
            cellIdx := int(UnitLookup[u][i])
            cell := &s.Board.Cells[cellIdx]
            
            if !cell.IsSolved() {
                count := cell.CandidateCount()
                if count > 1 && count <= k {
                    candidates = append(candidates, candidateInfo{cellIdx, cell.Candidates})
                }
            }
        }
        
        if len(candidates) < k {
            continue
        }

        // 2. Check Combinations
        // We need to find a subset of size 'k' from 'candidates'.
        var matchFound bool
        var processCombination func(start int, current []candidateInfo)
        
        processCombination = func(start int, current []candidateInfo) {
            if matchFound { return }
            
            // Base case: we have selected k cells
            if len(current) == k {
                // Check the union of their masks
                var unionMask CandidateMask
                for _, c := range current {
                    unionMask |= c.mask
                }
                
                // CRITICAL LOGIC:
                // If the union of masks has exactly k bits set, 
                // these k cells MUST hold these k numbers.
                if bits.OnesCount16(uint16(unionMask)) == k {
                    // Eliminate these bits from ALL OTHER cells in the unit
                    if s.applyNakedSubset(u, current, unionMask) {
                        changed = true
                        matchFound = true // Optimization: Move to next unit
                    }
                }
                return
            }
            
            // Recursive step
            for i := start; i < len(candidates); i++ {
                processCombination(i+1, append(current, candidates[i]))
            }
        }
        
        processCombination(0, make([]candidateInfo, 0, k))
    }
    
    return changed, nil
}

// applyNakedSubset removes the 'unionMask' candidates from all cells in unit 'u'
// EXCEPT for the cells that form the subset.
func (s *Solver) applyNakedSubset(unitIdx int, subset []candidateInfo, unionMask CandidateMask) bool {
    changed := false
    
    // Create a quick lookup map for the indices in the subset
    subsetMap := make(map[int]bool)
    for _, c := range subset {
        subsetMap[c.idx] = true
    }
    
    for i := 0; i < 9; i++ {
        targetIdx := int(UnitLookup[unitIdx][i])
        
        // Skip the cells that are part of the subset
        if subsetMap[targetIdx] {
            continue
        }
        
        // Attempt to remove candidates found in the subset from this outsider cell
        cell := &s.Board.Cells[targetIdx]
        if !cell.IsSolved() {
            // Logic: mask &^ unionMask
            // If the cell had one of the locked numbers, remove it.
            original := cell.Candidates
            cell.Candidates = cell.Candidates &^ unionMask
            
            if cell.Candidates != original {
                changed = true
                s.push(targetIdx) // Queue for propagation
            }
        }
    }
    return changed
}
```

### Updated Strategy

```go
func (s *Solver) SolveStrategy() error {
    for {
        // 1. Low Hanging Fruit: Naked Singles (Propagate)
        if err := s.propagate(); err != nil { return err }
        
        // 2. Hidden Singles (Unit-based singles)
        if found, err := s.FindHiddenSingles(); err != nil { return err } else if found { continue }
        
        // 3. Locked Candidates (Intersection)
        // O(81) check, fairly cheap
        if found, err := s.FindLockedCandidates(); err != nil { return err } else if found { continue }
        
        // 4. Naked Pairs
        if found, err := s.FindNakedSubsets(2); err != nil { return err } else if found { continue }

        // 5. Naked Triples
        if found, err := s.FindNakedSubsets(3); err != nil { return err } else if found { continue }

        // 6. Naked Quads
        // Expensive: O(Units * Combinations)
        if found, err := s.FindNakedSubsets(4); err != nil { return err } else if found { continue }
        
        // If we reach here, we are stuck.
        break
    }
    return nil
}
```

---

## 6. Reverse Mapping

**User:**
The current board structure allows us to efficiently answer the question "What candidate values are possible in this cell?"
I've seen implementations that maintain a reverse map for answering the question "What candidate cells are possible locations for this value?"
How does this approach compare in performance and implementation complexity to calculating the candidate cells every time?

**Consultant:**

### Conceptual Analysis

You are asking about the classic "Space-Time Trade-off."
- **Cell-First (Current):** `Board[cell] -> mask`. Optimized for "What can go here?"
- **Value-First (Proposed):** `Candidates[value][unit] -> mask`. Optimized for "Where can '5' go in Row 3?"

### The Verdict for Go

**Stick to On-the-Fly Calculation.**

In Go, compiling to machine code, the "calculation" is trivial.
Scanning a row for a specific candidate in your current structure is a tight loop that the compiler unwraps.
It is often faster than the memory latency of fetching a "cached" value from a complex data structure.
Maintaining a reverse map also introduces "cache invalidation" complexity.

### Optimized "On-the-Fly" Implementation

```go
// GetValuePositions returns a bitmask (0-511) where each set bit represents 
// a cell index (0-8) within the unit that holds the candidate 'num'.
func (s *Solver) GetValuePositions(unitIdx int, num uint8) uint16 {
    var positions uint16
    bit := uint16(1 << (num - 1))
    
    // Unrolling this loop (manually or by compiler) is extremely fast
    // because we are accessing contiguous memory if unit is a Row.
    for i := 0; i < 9; i++ {
        cellIdx := UnitLookup[unitIdx][i]
        // & operation is extremely cheap
        if s.Board.Cells[cellIdx].Candidates & bit != 0 {
            positions |= 1 << i
        }
    }
    return positions
}
```

---

## 7. Fish Patterns (X-Wing, Swordfish, Jellyfish)

**User:**
Let's take a look at X-Wings and other basic fish patterns (Swordfish, Jellyfish).

**Consultant:**

### Conceptual Analysis

Fish patterns rely on the relationship between **Base Sets** and **Cover Sets**.
- **Logic:** If a candidate `N` appears in $k$ different Rows (Base Sets), and within those rows, `N` is strictly confined to $k$ Columns (Cover Sets), then `N` is "locked" into that grid intersection.
- **Deduction:** Since `N` *must* exist in the Base Sets (Rows) at those specific column intersections, it **cannot** exist anywhere else in those Columns (Cover Sets).

- **X-Wing:** Size 2
- **Swordfish:** Size 3
- **Jellyfish:** Size 4

### Go Implementation

```go
// FindFish searches for X-Wing (size 2), Swordfish (3), and Jellyfish (4).
func (s *Solver) FindFish(size int) (bool, error) {
    // Check both orientations:
    // 1. Base = Rows, Cover = Cols
    if found, err := s.findFishInOrientation(size, true); err != nil {
        return false, err
    } else if found {
        return true, nil
    }

    // 2. Base = Cols, Cover = Rows
    if found, err := s.findFishInOrientation(size, false); err != nil {
        return false, err
    } else if found {
        return true, nil
    }

    return false, nil
}

// findFishInOrientation looks for the pattern for all numbers 1-9.
// isRowBase: true if we are looking for Rows locked into Cols.
func (s *Solver) findFishInOrientation(size int, isRowBase bool) (bool, error) {
    for num := uint8(1); num <= 9; num++ {
        // 1. Build the "Pattern Map" for this number.
        var patterns [9]uint16
        var validLines []int

        for i := 0; i < 9; i++ {
            // Reusing logic similar to GetValuePositions, but adapted for orientation
            mask := s.getLinePattern(i, num, isRowBase)
            
            // Optimization: A line must have at least 2 candidates to form a basic Fish.
            if bits.OnesCount16(mask) >= 2 {
                patterns[i] = mask
                validLines = append(validLines, i)
            }
        }

        // We need at least 'size' valid lines to form a Fish
        if len(validLines) < size {
            continue
        }

        // 2. Find Combinations of 'size' lines
        var found bool
        var err error
        
        var recurse func(start int, currentLines []int, unionMask uint16)
        recurse = func(start int, currentLines []int, unionMask uint16) {
            if found || err != nil { return }

            // Base case: we picked k lines
            if len(currentLines) == size {
                // CORE LOGIC:
                // If the Union of positions has exactly 'size' bits set, 
                // it means these 'size' lines are perfectly covered by 'size' orthogonal lines.
                if bits.OnesCount16(unionMask) == size {
                    // We found a Fish!
                    // 'unionMask' tells us which orthogonal lines (Cover Sets) are involved.
                    if s.eliminateFish(num, currentLines, unionMask, isRowBase) {
                        found = true
                    }
                }
                return
            }

            // Recursive step
            for i := start; i < len(validLines); i++ {
                lineIdx := validLines[i]
                newMask := unionMask | patterns[lineIdx]
                
                // Pruning: If union grows larger than 'size', this path is dead.
                if bits.OnesCount16(newMask) <= size {
                    recurse(i+1, append(currentLines, lineIdx), newMask)
                }
            }
        }

        // Start search with empty set
        recurse(0, make([]int, 0, size), 0)
        
        if err != nil { return false, err }
        if found { return true, nil }
    }
    return false, nil
}

// getLinePattern returns a bitmask of where 'num' appears in a Row or Col.
func (s *Solver) getLinePattern(idx int, num uint8, isRow bool) uint16 {
    var mask uint16
    bit := uint16(1 << (num - 1))
    
    for k := 0; k < 9; k++ {
        var cellIdx int
        if isRow {
            cellIdx = idx*9 + k
        } else {
            cellIdx = k*9 + idx
        }
        
        if s.Board.Cells[cellIdx].Candidates & bit != 0 {
            mask |= 1 << k
        }
    }
    return mask
}

// eliminateFish removes candidates from the Cover Sets.
func (s *Solver) eliminateFish(num uint8, baseLines []int, coverMask uint16, isRowBase bool) bool {
    changed := false
    
    // Quick lookup for base lines to skip
    baseMap := make(map[int]bool)
    for _, l := range baseLines {
        baseMap[l] = true
    }
    
    // Iterate over the orthogonal lines (Cover Sets) identified by the mask
    for i := 0; i < 9; i++ {
        // Check if index 'i' is in the cover set
        if coverMask & (1 << i) != 0 {
            // This is a Cover Line. Remove 'num' from this line, EXCEPT where it intersects Base Lines.
            
            for k := 0; k < 9; k++ {
                // Skip if 'k' is one of our defining Base Lines
                if baseMap[k] {
                    continue
                }
                
                var cellIdx int
                if isRowBase {
                    // Cover is Col 'i', iterating Rows 'k'
                    cellIdx = k*9 + i
                } else {
                    // Cover is Row 'i', iterating Cols 'k'
                    cellIdx = i*9 + k
                }
                
                if s.Board.Cells[cellIdx].RemoveCandidate(num) {
                    changed = true
                    s.push(cellIdx)
                }
            }
        }
    }
    return changed
}
```
