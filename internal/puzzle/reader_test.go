package puzzle

import (
	"os"
	"testing"
)

func TestFromFile_UnboundedRead(t *testing.T) {
	// Create a large file to simulate a DOS attack
	f, err := os.CreateTemp("", "large_puzzle")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Write 2MB of data (greater than our intended 1MB limit)
	data := make([]byte, 2*1024*1024)
	for i := range data {
		data[i] = ' '
	}
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	// This should fail or be limited
	_, err = FromFile(f)
	if err == nil {
		t.Error("Expected error for large input, but got nil")
	}
}
