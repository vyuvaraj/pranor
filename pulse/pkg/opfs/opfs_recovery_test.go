package opfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOPFSTruncationAutoRecovery(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "opfs_corruption_test")
	walFile := filepath.Join(tmpDir, "opfs_wal.log")
	_ = os.MkdirAll(tmpDir, 0755)

	// Write valid line followed by corrupted partial line
	corruptedData := `{"offset":1,"topic":"orders","payload":"valid_1","timestamp":100,"synced":false}` + "\n" + `{"offset":2,"topic":"orders","payload":"corrupt_partial_json`

	if err := os.WriteFile(walFile, []byte(corruptedData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	driver, err := NewOPFSDriver(tmpDir)
	if err != nil {
		t.Fatalf("Failed to initialize driver on corrupted WAL: %v", err)
	}
	defer driver.Close()

	entries, err := driver.ReadRange("orders", 1, 10)
	if err != nil {
		t.Fatalf("Failed to read range: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 recovered valid entry, got %d", len(entries))
	}

	if entries[0].Payload != "valid_1" {
		t.Errorf("Unexpected recovered payload: %s", entries[0].Payload)
	}
}
