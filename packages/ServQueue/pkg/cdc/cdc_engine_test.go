package cdc

import (
	"encoding/json"
	"testing"

	"github.com/vyuvaraj/serv/packages/ServQueue/pkg/core"
)

func TestCDCMutationStreaming(t *testing.T) {
	driver := core.NewMemoryDriver()
	engine := core.NewEngine(driver)
	defer engine.Close()

	cdcEngine := NewCDCEngine(engine)

	// Stream DB INSERT mutation
	before := map[string]interface{}{}
	after := map[string]interface{}{"id": 1001, "name": "Acme Corp", "status": "active"}

	entry, err := cdcEngine.StreamMutation("users", OpInsert, before, after)
	if err != nil {
		t.Fatalf("Failed to stream CDC mutation: %v", err)
	}

	if entry.Topic != "users.cdc" {
		t.Errorf("Expected topic 'users.cdc', got '%s'", entry.Topic)
	}

	var event CDCEvent
	if err := json.Unmarshal([]byte(entry.Payload), &event); err != nil {
		t.Fatalf("Failed to unmarshal CDC payload: %v", err)
	}

	if event.Operation != OpInsert {
		t.Errorf("Expected operation INSERT, got %s", event.Operation)
	}

	if event.After["name"] != "Acme Corp" {
		t.Errorf("Unexpected payload content: %v", event.After)
	}
}
