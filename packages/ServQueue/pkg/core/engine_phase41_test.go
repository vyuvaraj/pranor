package core

import (
	"testing"
)

func TestSchemaRegistryValidation(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	registry := NewSchemaRegistry()
	registry.RegisterSchema("orders", SchemaRule{
		RequiredFields: []string{"order_id", "amount"},
	})

	// 1. Valid payload
	validPayload := `{"order_id": "ord_1001", "amount": 299.99, "currency": "USD"}`
	_, err := engine.EnqueueWithSchema("orders", validPayload, registry)
	if err != nil {
		t.Fatalf("Expected valid payload to pass schema validation, got error: %v", err)
	}

	// 2. Invalid payload (missing required 'amount' field)
	invalidPayload := `{"order_id": "ord_1002"}`
	_, err = engine.EnqueueWithSchema("orders", invalidPayload, registry)
	if err == nil {
		t.Errorf("Expected schema validation error for missing field, but got nil")
	}
}

func TestAtomicMultiTopicTransactions(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	// Begin atomic transaction
	tx := engine.BeginTx("tx_checkout_1")

	if err := tx.Enqueue("orders", `{"order_id": "ord_55"}`); err != nil {
		t.Fatalf("Failed to enqueue in tx: %v", err)
	}
	if err := tx.Enqueue("inventory", `{"item_id": "prod_12", "delta": -1}`); err != nil {
		t.Fatalf("Failed to enqueue in tx: %v", err)
	}

	// Commit atomic transaction
	entries, err := tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 committed entries, got %d", len(entries))
	}

	if tx.State != TxCommitted {
		t.Errorf("Expected tx state COMMITTED, got %s", tx.State)
	}

	// Verify entries exist in storage
	orderEntries, _ := engine.Dequeue("orders", 1, 10)
	if len(orderEntries) != 1 {
		t.Errorf("Expected 1 order entry in engine, got %d", len(orderEntries))
	}
}
