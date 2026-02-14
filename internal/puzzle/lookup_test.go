package puzzle

import "testing"

func TestPeerLookup(t *testing.T) {
	// Pick a few cells and verify their peers.
	// Index 0 (r0c0) should have peers:
	// Row 0: 1,2,3,4,5,6,7,8
	// Col 0: 9,18,27,36,45,54,63,72
	// Box 0: 10,11,19,20 (since 0,1,2,9,10,11,18,19,20 are in box 0, and 1,2,9,18 are already row/col peers)
	// Wait, let's just count unique peers.
	// Row peers: 8
	// Col peers: 8
	// Box peers: 4 (remaining)
	// Total: 20

	peers := GetPeers(0)
	if len(peers) != 20 {
		t.Errorf("Expected 20 peers, got %d", len(peers))
	}

	expectedPeers := map[int]bool{
		// Row 0
		1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true,
		// Col 0
		9: true, 18: true, 27: true, 36: true, 45: true, 54: true, 63: true, 72: true,
		// Box 0 (excluding row 0 and col 0)
		10: true, 11: true,
		19: true, 20: true,
	}

	for _, p := range peers {
		if !expectedPeers[p] {
			t.Errorf("Unexpected peer %d for cell 0", p)
		}
		delete(expectedPeers, p)
	}
	if len(expectedPeers) > 0 {
		t.Errorf("Missing peers for cell 0: %v", expectedPeers)
	}
}
