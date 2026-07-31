package import (
	"testing"
)

func TestDiskOffsetLog(t *testing.T) {
	tmpDir := t.TempDir()

	logEngine, err := NewDiskOffsetLog(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create DiskOffsetLog: %v", err)
	}

	// 1. Acknowledge and persist offset
	err = logEngine.AcknowledgeOffset("order_processor", "orders", 1050)
	if err != nil {
		t.Fatalf("Failed to acknowledge offset: %v", err)
	}

	// 2. Read committed offset
	offset, found := logEngine.GetCommittedOffset("order_processor", "orders")
	if !found || offset != 1050 {
		t.Fatalf("Expected committed offset 1050, got %d (found=%v)", offset, found)
	}

	_ = logEngine.Close()

	// 3. Recover offset after crash/restart
	recoveredLog, err := NewDiskOffsetLog(tmpDir)
	if err != nil {
		t.Fatalf("Failed to recover DiskOffsetLog: %v", err)
	}
	defer recoveredLog.Close()

	recOffset, recFound := recoveredLog.GetCommittedOffset("order_processor", "orders")
	if !recFound || recOffset != 1050 {
		t.Fatalf("Expected recovered offset 1050 after restart, got %d (found=%v)", recOffset, recFound)
	}
}
