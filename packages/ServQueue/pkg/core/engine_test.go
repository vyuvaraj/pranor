package core

import (
	"testing"
)

func TestEngineMemoryDriver(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	// 1. Enqueue events
	e1, err := engine.Enqueue("orders", `{"id": 101, "item": "laptop"}`)
	if err != nil {
		t.Fatalf("Failed to enqueue event: %v", err)
	}
	if e1.Offset != 1 {
		t.Errorf("Expected offset 1, got %d", e1.Offset)
	}

	e2, err := engine.Enqueue("orders", `{"id": 102, "item": "mouse"}`)
	if err != nil {
		t.Fatalf("Failed to enqueue event: %v", err)
	}
	if e2.Offset != 2 {
		t.Errorf("Expected offset 2, got %d", e2.Offset)
	}

	// 2. Dequeue range
	entries, err := engine.Dequeue("orders", 1, 10)
	if err != nil {
		t.Fatalf("Failed to dequeue range: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// 3. Pending sync check
	pending, err := engine.GetPendingSync(10)
	if err != nil {
		t.Fatalf("Failed to get pending sync: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("Expected 2 pending sync entries, got %d", len(pending))
	}

	// 4. Acknowledge sync
	if err := engine.AcknowledgeSync([]uint64{1, 2}); err != nil {
		t.Fatalf("Failed to acknowledge sync: %v", err)
	}

	pendingAfter, _ := engine.GetPendingSync(10)
	if len(pendingAfter) != 0 {
		t.Errorf("Expected 0 pending sync entries after ACK, got %d", len(pendingAfter))
	}
}
