# Sudoku Solver Design

This document describes the architecture of a high-performance Sudoku solver written in Go. The design prioritizes CPU cache locality, minimal memory allocation, and bitwise parallelism to achieve maximum speed.

## 1. Data Structures & Architectural Analysis

The core of the solver is built upon two fundamental data structures: **Bitmasks** and a **Flattened Fixed-Size Array**.

### 1.1 Bitmasks (`uint16`)

Instead of representing cell candidates as a list or map of integers (e.g., `[1, 5, 9]`), we use a single `uint16` bitmask.

- **Structure:** Bit $k$ corresponds to the number $k+1$.
    - Example: Binary `0000 0001 0000 0010` (Hex `0x0102`) represents candidates {2, 9}.
- **Why it is efficient:**
    - **O(1) Set Operations:** Union, Intersection, and Difference of candidate sets become single CPU instructions (`OR`, `AND`, `AND NOT`). This is magnitudes faster than iterating over slices or hash map lookups.
    - **POPCNT Instruction:** Counting remaining candidates is performed using the hardware `POPCNT` (Population Count) instruction (via `math/bits.OnesCount16`), which executes in a single clock cycle on modern x86_64/ARM64 architectures.
    - **Memory Density:** A set of candidates occupies only 2 bytes, allowing the entire board state to fit comfortably in CPU registers and L1 cache.

```go
// CandidateMask represents the set of possible values for a cell.
// Bit 0 = 1, Bit 1 = 2, ..., Bit 8 = 9.
type CandidateMask uint16
```

### 1.2 Flattened Array (`[81]Cell`)

The board is represented as a single contiguous array of 81 `Cell` structs, rather than a 2D `[9][9]Cell` array or a slice.

- **Structure:** `[81]Cell` where index $i = row \times 9 + col$.
- **Why it is efficient:**
    - **Cache Locality:** Modern CPUs fetch memory in cache lines (typically 64 bytes). A contiguous array ensures that when one cell is accessed, its neighbors are likely already pre-fetched into the L1 cache. Linked structures (like graphs of pointers) cause "pointer chasing" which stalls the CPU waiting for main memory.
    - **Stack Allocation:** The entire `Board` struct size is small (~324 bytes). This allows it to be allocated entirely on the **stack**. Stack allocation is effectively free (a pointer decrement) and creates zero pressure on the Go Garbage Collector (GC), avoiding GC pauses during heavy solving.
    - **Loop Unrolling:** Iterating 0–80 eliminates nested loops, allowing the compiler to more easily unroll loops and use SIMD instructions.

```go
const (
    // Constants for board dimensions
    Size      = 9
    TotalCells = Size * Size
    // AllCandidates represents binary 111111111 (numbers 1-9)
    AllCandidates CandidateMask = 0x1FF 
)

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
```

### 1.3 Pre-computed Peer Lookup

Constraint propagation requires frequent access to a cell's "peers" (the 20 other cells in its Row, Column, and Box).

- **Design:** A global, immutable lookup table `PeerLookup [81][20]uint8` references peers by index (not pointer).
- **Efficiency Analysis:**
    - **Memory Size:** Storing 20 pointers per cell (`81 * 20 * 8 bytes`) would consume ~13KB. Storing 20 indices (`81 * 20 * 1 byte`) consumes ~1.6KB. The smaller footprint stays hot in L1 cache.
    - **Immutable Topology:** The relationships between cells never change. Calculating them once at startup avoids redundant runtime arithmetic.

```go
// Pre-computed lookup table for peers.
// Global variable or singleton in a real app.
// Index [0-80] -> Fixed array of 20 peer indices.
// Using uint8 saves cache space (indices are 0-80).
var PeerLookup [81][20]uint8

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

        delete(peerMap, i) // Remove self
        
        idx := 0
        for p := range peerMap {
            PeerLookup[i][idx] = uint8(p)
            idx++
        }
    }
}
```

### 1.4 Design Trade-off: On-the-Fly vs. Reverse Mapping

**Decision:** We calculate "Where is candidate X?" on-the-fly rather than maintaining a cached "Reverse Map" (e.g., `Candidate -> List[Cell]`).

- **Reasoning:**
    - **Memory Latency Dominates:** Fetching a random memory location (a cached list) is slow (tens of nanoseconds). Scanning 9 array slots effectively takes nanoseconds due to CPU pre-fetching and branch prediction.
    - **Complexity & Sync:** Maintaining a secondary data structure doubles the bookkeeping overhead for every state change. The CPU cost of re-scanning a row is lower than the logic glue required to update a "cache".

```go
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
```

### 1.5 Pre-computed House Lookup

Many algorithms (Hidden Singles, Locked Candidates, Subsets) function by scanning a "House" (Row, Column, or Box).

- **Design:** We treat all 27 houses (9 Rows, 9 Cols, 9 Boxes) uniformly using a `HouseLookup [27][9]uint8` table.
- **Efficiency Analysis:**
    - **Uniform Iteration:** Instead of writing three separate loops (one for rows, one for cols, one for boxes) with different index arithmetic, we write a single loop `for h := 0; h < 27; h++`.
    - **Branchless Logic:** The complex math to map "Cell 5 in Box 4" to "Board Index 41" is done once at initialization. Runtime lookups are just `HouseLookup[h][i]`.
    - **Cache Usage:** Like `PeerLookup`, this table uses `uint8` indices to stay small and fast.

```go
// Pre-computed lookup for the 27 houses (9 Rows, 9 Cols, 9 Boxes).
// Dimensions: [27 houses][9 cell indices]
var HouseLookup [27][9]uint8

func InitializeHouseLookup() {
    var houseIdx int
    
    // 1. Rows
    for r := 0; r < 9; r++ {
        for c := 0; c < 9; c++ {
            HouseLookup[houseIdx][c] = uint8(r*9 + c)
        }
        houseIdx++
    }
    
    // 2. Columns
    for c := 0; c < 9; c++ {
        for r := 0; r < 9; r++ {
            HouseLookup[houseIdx][r] = uint8(r*9 + c)
        }
        houseIdx++
    }
    
    // 3. Boxes
    for br := 0; br < 3; br++ {
        for bc := 0; bc < 3; bc++ {
            // Internal box index
            for i := 0; i < 9; i++ {
                r := br*3 + i/3
                c := bc*3 + i%3
                HouseLookup[houseIdx][i] = uint8(r*9 + c)
            }
            houseIdx++
        }
    }
}
```

---

### Basic Cell Operations

```go
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

---

## 3. General Solving & Propagation

### 3.1 Constraint Propagation

**Pattern:** *Naked Single*.
This is the foundational logic of Sudoku. When a cell takes a value (e.g., "5"), that value becomes impossible for all its peers.

- **Mechanism:**
    1. **Elimination:** We iterate the 20 peers of the solved cell and unset the bit corresponding to "5" using `AND NOT`.
    2. **Ripple Effect:** If a peer loses a candidate and is left with only 1 bit set (a **Naked Single**), it becomes solved.
    3. **Queue:** These newly solved cells are added to a `WorkQueue` to propagate *their* constraints, continuing until the board stabilizes.

```go
import "errors"

// Solver wraps the board to to track propagation state.
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
        
        // Take pointer to avoid array copy and keep in L1 cache
        peers := &PeerLookup[currIdx]

        for i := 0; i < 20; i++ {
            peerIdx := int(peers[i])
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
                    // Naked Single Found
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

---

## 4. Specific Algorithms

### 4.1 Hidden Singles

**Pattern:** Within a House (Row, Col, Box), a specific candidate $N$ appears in the **candidate set** of **only one cell**.
Even if that cell has other candidates ($N$ is "hidden" among them), it *must* be $N$.

- **Detection Strategy:**
    1. Iterate through a unit (e.g., Row 0).
    2. Maintain a frequency count of each candidate bit.
    3. If any candidate has a frequency of exactly 1, identifying the holding cell solves it.
- **Efficiency:** We use bit manipulation to iterate only set bits (`mask & -mask`), avoiding checking empty candidate slots.

```go
// FindHiddenSingles scans all houses for numbers that appear only once.
// It applies the first one it finds and returns true (triggering propagation).
// Returns false if no hidden singles are found.
func (s *Solver) FindHiddenSingles() (bool, error) {
    // Buffers to track counts and positions for numbers 1-9.
    // Index 0 is unused for clarity (1-based mapping).
    var counts [10]uint8
    var positions [10]uint8 
    
    // Iterate through all 27 houses (Rows, Cols, Boxes)
    for h := 0; h < 27; h++ {
        // Reset counters for this house
        // (Compiler optimizes this into a block zeroing instruction)
        counts = [10]uint8{}
        
        // Pass 1: Tally candidates in this house
        for i := 0; i < 9; i++ {
            cellIdx := HouseLookup[h][i]
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
                // The number 'num' can ONLY go in 'targetIdx' for this house.
                
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
```

### 4.2 Locked Candidates (Pointing / Claiming)

**Pattern:** A **Strong Link** between a Box and a Line (Row/Col).

1. **Pointing:** If all instances of candidate $N$ in a Box are confined to a single Row, then $N$ cannot exist elsewhere in that Row (outside the box).
2. **Claiming:** If all instances of candidate $N$ in a Row are confined to a single Box, then $N$ cannot exist elsewhere in that Box.

- **Detection Strategy:**
    - We use **Bitwise OR Accumulation**. By ORing the masks of all cells in a row/box, we quickly determine if a candidate exists.
    - We check alignment by verifying that the set of positions for candidate $N$ fits entirely within a specific peer group intersection.

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

### 4.3 Naked Subsets (Pairs, Triples, Quads)

**Pattern:** If $N$ cells in a house contain candidates exclusively from a set of $N$ numbers, those numbers is **locked** into those cells and can be eliminated from all other cells in that house.

- Example (Naked Pair): Two cells in a row act as `{2, 4}` and `{2, 4}`. No other cell in that row can be 2 or 4.

- **Detection Strategy:**
    - Recursive search (backtracking) within the small scope of a single house (9 cells).
    - **Union Check:** We accumulate the union of candidates for the selected subset. If `PopCount(UnionMask) == Subset_Size`, a Naked Subset is found.
    - **Efficiency:** Bit-parallelism allows us to check the condition ("is the union size equal to k?") in constant time.

```go
// FindNakedSubsets looks for k cells in a house that share a union of k candidates.
// k = 2 (Pairs), 3 (Triples), 4 (Quads).
func (s *Solver) FindNakedSubsets(k int) (bool, error) {
    changed := false
    
    // Iterate all 27 houses
    for h := 0; h < 27; h++ {
        // 1. Collect interesting cells in this house
        // We only care about cells with candidate count <= k (and > 1)
        type candidateInfo struct {
            idx  int
            mask CandidateMask
        }
        candidates := make([]candidateInfo, 0, 9)
        
        for i := 0; i < 9; i++ {
            cellIdx := int(HouseLookup[h][i])
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
                    // Eliminate these bits from ALL OTHER cells in the house
                    if s.applyNakedSubset(h, current, unionMask) {
                        changed = true
                        matchFound = true // Optimization: Move to next house
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

// applyNakedSubset removes the 'unionMask' candidates from all cells in house 'h'
// EXCEPT for the cells that form the subset.
func (s *Solver) applyNakedSubset(houseIdx int, subset []candidateInfo, unionMask CandidateMask) bool {
    changed := false
    
    // Create a quick lookup map for the indices in the subset
    subsetMap := make(map[int]bool)
    for _, c := range subset {
        subsetMap[c.idx] = true
    }
    
    for i := 0; i < 9; i++ {
        targetIdx := int(HouseLookup[houseIdx][i])
        
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

### 4.4 Fish Patterns (X-Wing, Swordfish, Jellyfish)

**Pattern:** Constraints involving multiple rows and columns based on **Base Sets** and **Cover Sets**.

- **Logic:** If a candidate `N` appears in $k$ different Rows (Base Sets), and within those rows, `N` is strictly confined to $k$ Columns (Cover Sets), then `N` is "locked" into that grid intersection. Since `N` *must* exist in the Base Sets (Rows) at those specific column intersections, it **cannot** exist anywhere else in those Columns (Cover Sets).

- **Detection Strategy:**
    - We treat rows as bitmasks identifying columns where candidate $N$ exists.
    - We look for $k$ rows whose combined bitmask (OR-sum) has a Population Count of exactly $k$.
    - This transforms a geometric pattern search into a pure subset-sum / bit-matching problem.

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

### 4.5 XY-Wing and XYZ-Wing Patterns

**Pattern:** Both XY-Wing and XYZ-Wing are "bent" logic patterns that rely on a central Pivot cell and two Pincer cells.

1. **XY-Wing (Y-Wing)**: 3 cells, each with exactly 2 candidates. The Pivot cell has candidates `{X, Y}`, and two Pincer cells that are peers of the Pivot cell have candidates `{X, Z}` and `{Y, Z}`. Any cell that sees both Pincer cells cannot be `Z`.

2. **XYZ-Wing**: Pivot has 3 candidates `{X, Y, Z}`, Pincers have 2 candidates `{X, Z}` and `{Y, Z}`. Any cell that sees the Pivot and both Pincers cannot be `Z`.

- **Detection Strategy**:
    - Iterate through every cell and find cells with exactly 2 (XY-Wing) or 3 (XYZ-Wing) candidates. These are potential Pivots.
    - Iterate through the Pivot's peers to find potential Pincers with exactly 2 candidates (the 2 Pincers must *not* be peers of each other for XY-Wing).
    - Check potential Pincers for the required `{X, Z}` and `{Y, Z}` candidate structure.
    - Check if the elimination target `Z` exists in the intersection of the Pincers' peers (for XY-Wing) or the intersection of the Pivot and Pincers' peers (for XYZ-Wing).
    - **Efficiency:** Bitwise operations are used to identify `X`, `Y`, and `Z` candidates. Finding the intersection of peer sets is a fast linear scan (or double loop) over small arrays using fast zero-allocation lookup tables.

```go
// FindXYWing looks for the XY-Wing pattern (3 cells, 2 candidates each).
func (s *Solver) FindXYWing() (bool, error) {
    // Iterate every cell to act as the potential "Pivot"
    for pivotIdx := 0; pivotIdx < 81; pivotIdx++ {
        pivotCell := &s.Board.Cells[pivotIdx]
        
        // Pivot must have exactly 2 candidates
        if pivotCell.CandidateCount() != 2 {
            continue
        }
        
        // Extract Pivot candidates (X and Y)
        pivotMask := uint16(pivotCell.Candidates)
        xBit := bits.TrailingZeros16(pivotMask)
        yBit := bits.TrailingZeros16(pivotMask &^ (1 << xBit))
        
        // Iterate Pivot's peers to find potential Pincers
        pivotPeers := &PeerLookup[pivotIdx] // Take a pointer to avoid array copy

        // We need to find two distinct pincers among the peers
        for i := 0; i < 20; i++ {
            pincerAIdx := int(pivotPeers[i]) // Cast uint8 to int
            pincerA := &s.Board.Cells[pincerAIdx]
            
            if pincerA.CandidateCount() != 2 { continue }
            
            // Check if Pincer A is valid (must share exactly 1 candidate with Pivot)
            // It must look like {X, Z} or {Y, Z}
            matchX := pincerA.HasCandidate(uint8(xBit + 1))
            matchY := pincerA.HasCandidate(uint8(yBit + 1))
            
            // XOR: It must match X or Y, but not both (that would be identical to Pivot)
            if matchX == matchY { continue }
            
            // Determine Z (the non-shared candidate in Pincer A)
            // Mask of PincerA minus Pivot Mask leaves only Z
            zMask := uint16(pincerA.Candidates) &^ pivotMask
            if bits.OnesCount16(zMask) != 1 { continue } // Should not happen if count is 2
            zBit := bits.TrailingZeros16(zMask)
            zVal := uint8(zBit + 1)

            // Now look for Pincer B
            // If Pincer A matched X, Pincer B must match Y (and share Z).
            // If Pincer A matched Y, Pincer B must match X (and share Z).
            targetSharedBit := yBit
            if matchY {
                targetSharedBit = xBit
            }
            
            // Inner loop to find Pincer B
            for j := i + 1; j < 20; j++ {
                pincerBIdx := pivotPeers[j]
                pincerB := &s.Board.Cells[pincerBIdx]
                
                if pincerB.CandidateCount() != 2 { continue }
                
                // Pincer B must have {targetSharedBit, Z}
                expectedMask := (uint16(1) << targetSharedBit) | (uint16(1) << zBit)
                if uint16(pincerB.Candidates) != expectedMask {
                    continue
                }
                
                // Found a valid XY-Wing!
                // Pivot: {X,Y}, PincerA: {X,Z}, PincerB: {Y,Z}
                // Elimination: Remove Z from common peers of Pincer A and Pincer B
                if s.eliminateFromIntersection(pincerAIdx, pincerBIdx, -1, zVal) {
                    return true, nil
                }
            }
        }
    }
    return false, nil
}

// FindXYZWing looks for the XYZ-Wing pattern (Pivot has 3 candidates).
func (s *Solver) FindXYZWing() (bool, error) {
    for pivotIdx := 0; pivotIdx < 81; pivotIdx++ {
        pivotCell := &s.Board.Cells[pivotIdx]
        
        if pivotCell.CandidateCount() != 3 { continue }
        
        pivotMask := uint16(pivotCell.Candidates)
        
        // OPTIMIZATION: Take pointer to avoid copying [20]uint8
        peers := &PeerLookup[pivotIdx]
        
        // Search for Pincer A
        for i := 0; i < 20; i++ {
            pAIdx := int(peers[i])
            pA := &s.Board.Cells[pAIdx]
            
            if pA.CandidateCount() != 2 { continue }
            
            maskA := uint16(pA.Candidates)
            if (maskA & pivotMask) != maskA { continue }
            
            // Search for Pincer B
            for j := i + 1; j < 20; j++ {
                pBIdx := int(peers[j])
                pB := &s.Board.Cells[pBIdx]
                
                if pB.CandidateCount() != 2 { continue }
                maskB := uint16(pB.Candidates)
                if (maskB & pivotMask) != maskB { continue }
                
                union := maskA | maskB
                intersection := maskA & maskB
                
                if union == pivotMask && bits.OnesCount16(intersection) == 1 {
                    zBit := bits.TrailingZeros16(intersection)
                    zVal := uint8(zBit + 1)
                    
                    if s.eliminateFromIntersection(pAIdx, pBIdx, pivotIdx, zVal) {
                        return true, nil
                    }
                }
            }
        }
    }
    return false, nil
}

// eliminateFromIntersection removes 'num' from any cell that sees ALL provided indices.
// If pivotIdx is -1, it finds intersection of (idx1, idx2).
// If pivotIdx is >= 0, it finds intersection of (idx1, idx2, pivotIdx).
func (s *Solver) eliminateFromIntersection(idx1, idx2, pivotIdx int, num uint8) bool {
    changed := false
    
    // OPTIMIZATION: Use pointer to iterate peers of idx1
    peers1 := &PeerLookup[idx1]
    
    // Build fast lookup tables (these helper calls already use pointer optimizations internally)
    isPeerOf2 := s.getPeerSet(idx2)
    
    var isPeerOfPivot [81]bool 
    hasPivot := false
    if pivotIdx != -1 {
        isPeerOfPivot = s.getPeerSet(pivotIdx)
        hasPivot = true
    }
    
    for i := 0; i < 20; i++ {
        candidateIdx := int(peers1[i])
        
        // Must be peer of idx2
        if !isPeerOf2[candidateIdx] { continue }
        
        // Must be peer of pivot
        if hasPivot && !isPeerOfPivot[candidateIdx] { continue }
        
        // Don't eliminate from the pattern cells themselves
        if candidateIdx == idx1 || candidateIdx == idx2 || candidateIdx == pivotIdx {
            continue
        }
        
        // Remove the candidate
        if s.Board.Cells[candidateIdx].RemoveCandidate(num) {
            changed = true
            s.push(candidateIdx)
        }
    }
    return changed
}

// getPeerSet returns a fast lookup table for peers.
// optimization: returns [81]bool instead of map to stay on the stack.
func (s *Solver) getPeerSet(idx int) [81]bool {
    var lookup [81]bool
    // PeerLookup[idx] is a fixed array [20]uint8, so we iterate 0..19
    // using pointer to iterate global table in place.
    peers := &PeerLookup[idx]
    
    for i := 0; i < 20; i++ {
        peerIdx := peers[i]
        lookup[peerIdx] = true
    }
    return lookup
}
```

---

## 5. Solving Strategy Loop

```go
// SolveStrategy combines Naked and Hidden singles, Locked Candidates, and Subsets.
func (s *Solver) SolveStrategy() error {
    for {
        // 1. Low Hanging Fruit: Naked Singles (Propagate)
        if err := s.propagate(); err != nil { return err }
        
        // 2. Hidden Singles (House-based singles)
        if found, err := s.FindHiddenSingles(); err != nil { return err } else if found { continue }
        
        // 3. Locked Candidates (Intersection)
        // O(81) check, fairly cheap
        if found, err := s.FindLockedCandidates(); err != nil { return err } else if found { continue }
        
        // 4. Naked Pairs
        if found, err := s.FindNakedSubsets(2); err != nil { return err } else if found { continue }

        // 5. Naked Triples
        if found, err := s.FindNakedSubsets(3); err != nil { return err } else if found { continue }

        // 6. Naked Quads
        // Expensive: O(Houses * Combinations)
        if found, err := s.FindNakedSubsets(4); err != nil { return err } else if found { continue }

        // 7. X-Wing
        // Moderate: 9 nums * 2 orientations * C(9,2) combinations
        if found, err := s.FindFish(2); err != nil { return err } else if found { continue }

        // 8. Swordfish
        // High: 9 nums * 2 orientations * C(9,3) combinations
        if found, err := s.FindFish(3); err != nil { return err } else if found { continue }

        // 9. Jellyfish
        // High: 9 nums * 2 orientations * C(9,4) combinations
        if found, err := s.FindFish(4); err != nil { return err } else if found { continue }

        // 10. XY-Wing
        // O(81 * 20 * 20 / 2). Roughly 16,000 iterations. Very fast.
        if found, err := s.FindXYWing(); err != nil { return err } else if found { continue }

        // 11. XYZ-Wing
        // O(81 * 20 * 20 / 2). Same complexity as XY-Wing.
        if found, err := s.FindXYZWing(); err != nil { return err } else if found { continue }
        
        // If we reach here, we are stuck.
        break
    }
    return nil
}
```
