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

func TestEngineAppendWithTraceparent(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	tp := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	meta := map[string]string{"traceparent": tp}

	entry, err := engine.Append("logs", "test payload", meta)
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	if entry.Traceparent != tp {
		t.Errorf("Expected traceparent %s, got %s", tp, entry.Traceparent)
	}

	// Verify Dequeue retrieves entry with Traceparent intact
	dequeued, err := engine.Dequeue("logs", 1, 10)
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}
	if len(dequeued) != 1 {
		t.Fatalf("Expected 1 dequeued entry, got %d", len(dequeued))
	}
	if dequeued[0].Traceparent != tp {
		t.Errorf("Dequeued entry expected traceparent %s, got %s", tp, dequeued[0].Traceparent)
	}
}

func TestEngineAppendWithoutTraceparent(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	entry, err := engine.Append("logs", "payload without traceparent")
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	if entry.Traceparent != "" {
		t.Errorf("Expected empty traceparent, got %s", entry.Traceparent)
	}
}

func TestEngineAppendCaseInsensitiveTraceparentKey(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	tp := "00-11111111111111111111111111111111-2222222222222222-01"
	meta := map[string]string{"TraceParent": tp}

	entry, err := engine.Append("logs", "case test", meta)
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	if entry.Traceparent != tp {
		t.Errorf("Expected traceparent %s, got %s", tp, entry.Traceparent)
	}
}
