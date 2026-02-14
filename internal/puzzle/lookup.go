package puzzle

// PeerLookup is a pre-computed table of the peers for each cell index.
// A peer is a cell that "sees" the cell at the given index, and therefore
// cannot contain the same value.
// Each cell has 20 peers: 8 cells in the same row, 8 cells in the same column,
// and 4 additional cells in the same box.
var PeerLookup [81][20]int

func init() {
	InitPeerLookup()
}

// InitPeerLookup initializes the PeerLookup table.
func InitPeerLookup() {
	for i := range 81 {
		r, c := i/9, i%9
		boxStartR, boxStartC := (r/3)*3, (c/3)*3
		peerMap := make(map[int]bool)

		// Add Row and Column peers
		for k := range 9 {
			if k != c {
				peerMap[r*9+k] = true // Row
			}
			if k != r {
				peerMap[k*9+c] = true // Col
			}
		}

		// Add Box peers
		for br := range 3 {
			for bc := range 3 {
				peerIdx := (boxStartR+br)*9 + (boxStartC + bc)
				if peerIdx != i {
					peerMap[peerIdx] = true
				}
			}
		}

		idx := 0
		for p := range peerMap {
			PeerLookup[i][idx] = p
			idx++
		}
	}
}

// GetPeers returns the peers of the cell at the given index.
// The `PeerLookup` table is a global array, so we can return a stable pointer
// to avoid copying the peer array.
func GetPeers(idx int) *[20]int {
	return &PeerLookup[idx]
}
