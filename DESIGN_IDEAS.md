# Design Ideas for Efficient Sudoku Solving

This document outlines a proposed architectural shift to improve the efficiency of detecting advanced Sudoku patterns like Swordfish, Jellyfish, XYZ-Wing, and General Forcing Chains.

## 1. Core Data Structure: BitwiseConstraintMap

To efficiently detect advanced Sudoku patterns, we need to move beyond a simple "Grid of Cells" representation. While a grid is excellent for simple iteration, it is inefficient for pattern matching because it requires scanning memory to rebuild the state of specific candidates constantly.

We recommend a dual-state architecture: maintaining the **Grid** (Spatial View) synchronized with an **Inverted Bitwise Candidate Map** (Constraint View).

### The Core Types

The core idea is to represent the board not just as "Cells containing Candidates," but as "Candidates distributed across Constraints." We utilize 16-bit integers as bitmasks to represent positions within a house (Row, Column, or Box).

We leverage bitsets for fast "PopCount" (population count) and bitwise logic operations.

```go
type BitMask uint16 // Represents 9 positions in a row/col/box

type CandidateMap struct {
    // The "Spatial View" (Standard Grid)
    // Fast access to: "What candidates are in cell (r,c)?"
    // Used for: XYZ-Wing, Naked Subsets
    Grid [9][9]BitMask 

    // The "Constraint View" (Inverted Index)
    // Fast access to: "Where can Digit 'd' go in Row 'r'?"
    // Dimensions: [Digit 1-9][House Index 0-8]
    // Used for: Swordfish, Jellyfish, Skyscraper
    Rows  [10][9]BitMask 
    Cols  [10][9]BitMask
    Boxes [10][9]BitMask
    
    // Metadata for heuristics
    PopCounts [10][9]int // Cached number of bits set in the maps above
}
```

### Why This Structure?

1.  **Bitwise Parallelism:** A CPU can compare the candidacy of a number across an entire row in a single CPU cycle using bitwise `AND`/`OR`/`XOR`.
2.  **Pattern Isolation:** Swordfish and Jellyfish care only about the topology of a *single digit*. The `Rows[d]` array isolates that digit immediately, removing the noise of other numbers.
3.  **Low Memory Footprint:** The entire structure fits easily in the L1 cache, making traversal extremely fast.

### Comparison: Standard Grid vs. Bitwise Map

| Feature | Standard Grid | Inverted Bitwise Map |
| :--- | :--- | :--- |
| **Storage** | `[9][9]Cell` | `[10][9]uint16` |
| **Single Value Lookup** | Fast | Fast |
| **Row Intersection** | Slow (Loop + Array) | **Instant** (Bitwise `&`) |
| **Union of 3 Rows** | Very Slow | **Instant** (Bitwise `|`) |
| **Memory Locality** | Scattered pointers | Contiguous memory |

---

## 2. Advanced Pattern Detection

This structure allows "Human" techniques to be run hundreds of times per second.

### Detecting a Swordfish (and Jellyfish)

A **Swordfish** occurs for digit `d` when candidates for `d` appear in exactly 3 rows, and those candidates line up such that they occupy only 3 columns.

**Traditional Approach:** Iterate all cells, build lists of coordinates, try to find intersections. Slow and complex.

**Bitwise Approach:**
We iterate combinations of 3 rows (R1, R2, R3). We check the `Rows` map for digit `d`.

```go
// Example logic for Swordfish on Digit 'd'
func (s *CandidateMap) FindSwordfish(d int) {
    // 1. Filter: Only look at rows that have 2 or 3 candidates for 'd' (Pre-calculated PopCounts)
    viableRows := s.getRowsWithCount(d, 2, 3) 

    // 2. Iterate combinations of 3 rows (triplets)
    for i := 0; i < len(viableRows)-2; i++ {
        r1 := viableRows[i]
        mask1 := s.Rows[d][r1]
        
        for j := i+1; j < len(viableRows)-1; j++ {
            r2 := viableRows[j]
            mask2 := s.Rows[d][r2]
            
            // Optimization: If union of 2 rows > 3 cols, 3rd row won't help
            if PopCount(mask1 | mask2) > 3 { continue } 

            for k := j+1; k < len(viableRows); k++ {
                r3 := viableRows[k]
                mask3 := s.Rows[d][r3]

                // The Magic: Union the columns
                unionMask := mask1 | mask2 | mask3
                
                // If the union only covers 3 columns, we found a Swordfish!
                if PopCount(unionMask) == 3 {
                    s.eliminateSwordfish(d, unionMask, r1, r2, r3)
                    return
                }
            }
        }
    }
}
```
*Complexity:* The bitwise OR is constant time. We avoid allocating arrays or lists during the search.

### Detecting a Skyscraper

A **Skyscraper** is a specific arrangement of a digit in two rows. Each row has exactly two candidates for that digit. Two of the candidates are in the same column (the "base"), and the other two are in different columns (the "towers").

**Bitwise Approach:**
1.  Look for rows where `PopCounts[d][row] == 2`.
2.  Check for a "Base": Do `RowA & RowB` (bitwise AND).
3.  If `PopCount(RowA & RowB) == 1`, they share exactly one column. This is a potential Skyscraper.
4.  The "Towers" are the bits remaining: `(RowA | RowB) ^ (RowA & RowB)`. Eliminating candidates from the peers of the towers is then a simple lookup.

### Detecting XYZ-Wing

**XYZ-Wing** relies on geometry (Pivot and Pincers) and specific candidate combinations (e.g., Pivot: `123`, Pincer1: `12`, Pincer2: `13`).

**Data Structure Usage:**
Here we switch to the **Grid (Spatial View)** side of our structure, but use the bitmasks to check validity instantly.

1.  Find a **Pivot**: Scan `Grid` for cells with `PopCount == 3`.
2.  Let Pivot candidates be `XYZ` (mask).
3.  Find **Pincers**: Look at the neighbors (Peers). A valid pincer must have `PopCount == 2` and its mask must be a subset of the Pivot's mask (`(PincerMask & ^PivotMask) == 0` is false, but they share candidates).
4.  **Verification**: We can check the intersection of the constraints of the "Z" value using the `Boxes` or `Rows` view to see if they share a common cell seen by all three.

---

## 3. General Forcing Chains & AIC

The **BitwiseConstraintMap** is highly efficient for *state queries* (e.g., "Is there a candidate 5 in this row?"), but Forcing Chains require *graph traversal* (e.g., "If 5 is false here, where must 5 be true?").

To solve General Forcing Chains (GFC) or Alternating Inference Chains (AIC) efficiently, we treat the `BitwiseConstraintMap` as our **Link Validation Oracle**, but we need an additional **Adjacency Cache** to drive the traversal.

### Using the BitwiseConstraintMap for Chain Traversal

A forcing chain consists of alternating **Strong Links** and **Weak Links**.

#### A. Finding Strong Links (The "If Not A, then B" logic)
A Strong Link exists between two candidates if they are the **only two** possibilities in a specific region (Cell or House).

*   **Intra-Cell Strong Links (Bivalue Cells):**
    *   *Logic:* If a cell has exactly two candidates (e.g., 3 and 7), then `Not 3 => 7`.
    *   *Using the Structure:* We check `PopCount(CandidateMap.Grid[r][c]) == 2`. If true, the two bits set in the mask are strongly linked.
*   **Intra-House Strong Links (Bilocation):**
    *   *Logic:* If a Digit `d` appears only twice in a Row (e.g., Col 1 and Col 5), then `Row(d)@C1` is strongly linked to `Row(d)@C5`.
    *   *Using the Structure:* We check `PopCount(CandidateMap.Rows[d][r]) == 2`. The two bits set in that 16-bit integer define the link.

#### B. Finding Weak Links (The "If A, then Not B" logic)
A Weak Link exists between any two candidates that "see" each other. If one is true, the other *must* be false.

*   *Logic:* Any peer in the same Row, Column, or Box, or any other candidate in the same cell.
*   *Using the Structure:* This is where `BitwiseConstraintMap` shines. We do not need to store weak links; we calculate them implicitly and instantly.
    *   To find weak links for `(r, c, digit)`:
        *   Fetch `Rows[digit][r]`. Every bit set (except `c`) is a weak link.
        *   Fetch `Cols[digit][c]`. Every bit set (except `r`) is a weak link.
        *   Fetch `Boxes[digit][b]`. Every bit set (except `i`) is a weak link.
        *   Fetch `Grid[r][c]`. Every bit set (except `digit`) is a weak link inside the cell.

### Additional Data Structures Required

While the map handles queries, traversing a chain (BFS or DFS) requires knowing "Where can I go next?" without scanning the whole board every microsecond. Since Weak Links are abundant and easy to find, we essentially only need to index the **Strong Links**.

#### A. The Strong Link Graph (Adjacency Cache)
We should generate a directed graph of Strong Links at the beginning of the chain search phase.

```go
type NodeID uint16 // Packed: (Row << 7) | (Col << 3) | Digit

// The Adjacency Cache
// Maps a candidate to the list of candidates it strongly implies.
// Usage: StrongLinks[fromNode] returns a list of toNodes.
type StrongLinkGraph map[NodeID][]NodeID

func BuildStrongLinkGraph(bm *BitwiseConstraintMap) StrongLinkGraph {
    g := make(StrongLinkGraph)
    
    // 1. Scan Cells for Bivalue Strong Links
    for r in 0..8 {
        for c in 0..8 {
            mask := bm.Grid[r][c]
            if PopCount(mask) == 2 {
                // Decode the two bits (d1, d2)
                // Add edge: (r,c,d1) -> (r,c,d2)
                // Add edge: (r,c,d2) -> (r,c,d1)
            }
        }
    }

    // 2. Scan Houses for Bilocation Strong Links
    for d in 1..9 {
        for r in 0..8 {
            mask := bm.Rows[d][r]
            if PopCount(mask) == 2 {
                 // Decode the two columns (c1, c2)
                 // Add edge: (r,c1,d) -> (r,c2,d)
                 // Add edge: (r,c2,d) -> (r,c1,d)
            }
            // Repeat for Cols and Boxes...
        }
    }
    return g
}
```
*Why this is needed:* Without this cache, every step of the DFS would require re-scanning rows, cols, and boxes to check `PopCount == 2`. Doing this once per "Solver Step" is much faster.

#### B. The Inference Map (Visited History)
To detect forcing chains (implications), we need to track what we have proven true/false during the current search to detect contradictions or loops.

```go
type ChainState struct {
    // Determine if we visited this node and what its forced state is
    // 0 = Unvisited, 1 = True, 2 = False
    Status map[NodeID]byte 
    
    // For reconstructing the chain (User Interface/Explanation)
    Parent map[NodeID]NodeID
}
```

### The Combined Algorithm (Putting it together)

To find a chain (e.g., an XY-Chain or Forcing Chain):

1.  **Snapshot:** Create the `StrongLinkGraph` from the current `BitwiseConstraintMap`.
2.  **Start:** Pick a candidate `startNode` (e.g., a candidate in a cell with 3 options).
3.  **Traverse (BFS/DFS):**
    *   **Step 1 (Strong):** Look up `startNode` in `StrongLinkGraph`. Let's say it links to `nodeB`. (Meaning: If startNode is False, nodeB is True).
    *   **Step 2 (Weak):** From `nodeB` (which is now "True"), use `BitwiseConstraintMap` to find all peers.
        *   `Rows[nodeB.digit][nodeB.row]` gives all row peers.
        *   `Grid[nodeB.row][nodeB.col]` gives all cell peers.
    *   **Step 3:** These peers are now "False". Check if any of these peers is our original `startNode` or triggers a contradiction.
    *   **Recurse:** Treat these "False" peers as new starting points for Strong Links.
