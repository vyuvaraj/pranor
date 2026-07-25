package opfs

import (
	"path/filepath"
	"testing"

	"github.com/vyuvaraj/serv/packages/ServQueue/pkg/core"
)

func TestOPFSDriverPersistence(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "opfs_test")
	driver, err := NewOPFSDriver(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create OPFSDriver: %v", err)
	}

	engine := core.NewEngine(driver)

	// Enqueue test records
	_, err = engine.Enqueue("metrics", `{"cpu": 45}`)
	if err != nil {
		t.Fatalf("Failed to enqueue record: %v", err)
	}

	_, err = engine.Enqueue("metrics", `{"cpu": 82}`)
	if err != nil {
		t.Fatalf("Failed to enqueue record: %v", err)
	}

	if err := engine.Close(); err != nil {
		t.Fatalf("Failed to close engine: %v", err)
	}

	// Reopen OPFS driver to verify crash recovery / persistent WAL reading
	reopenedDriver, err := NewOPFSDriver(tmpDir)
	if err != nil {
		t.Fatalf("Failed to reopen OPFSDriver: %v", err)
	}
	reopenedEngine := core.NewEngine(reopenedDriver)
	defer reopenedEngine.Close()

	recovered, err := reopenedEngine.Dequeue("metrics", 1, 10)
	if err != nil {
		t.Fatalf("Failed to dequeue recovered records: %v", err)
	}

	if len(recovered) != 2 {
		t.Fatalf("Expected 2 recovered records from OPFS WAL, got %d", len(recovered))
	}

	if recovered[0].Payload != `{"cpu": 45}` {
		t.Errorf("Unexpected recovered payload: %s", recovered[0].Payload)
	}
}
