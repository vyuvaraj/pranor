package core

import (
	"testing"
	"time"
)

func TestDLQManagerRoutingAndReplay(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	policy := DLQPolicy{
		MaxRetries:       3,
		InitialBackoff:   100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	dlq := NewDLQManager(engine, policy)

	entry, _ := engine.Enqueue("orders", `{"order_id": "ord_990"}`)

	// Simulate processing failure
	dlqEntry, err := dlq.HandleFailure(entry, "database connection timeout")
	if err != nil {
		t.Fatalf("HandleFailure failed: %v", err)
	}

	if dlqEntry.Topic != "orders.dlq" {
		t.Errorf("Expected DLQ topic 'orders.dlq', got '%s'", dlqEntry.Topic)
	}

	messages := dlq.GetDLQMessages("orders")
	if len(messages) != 1 {
		t.Fatalf("Expected 1 DLQ message, got %d", len(messages))
	}

	// Replay DLQ messages back to main orders topic
	replayed, err := dlq.ReplayDLQ("orders")
	if err != nil {
		t.Fatalf("ReplayDLQ failed: %v", err)
	}

	if len(replayed) != 1 {
		t.Fatalf("Expected 1 replayed entry, got %d", len(replayed))
	}

	if len(dlq.GetDLQMessages("orders")) != 0 {
		t.Errorf("Expected 0 DLQ messages after replay")
	}
}
