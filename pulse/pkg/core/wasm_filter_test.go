package import (
	"testing"
)

func TestWASMFilterEngineScrubber(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	filterEngine := NewWASMFilterEngine()
	filterEngine.RegisterBuiltinScrubber("auth_logs")

	// Enqueue with sensitive text
	rawPayload := `{"user": "admin", "password": "supersecretpassword123"}`
	entry, err := engine.EnqueueWithFilter("auth_logs", rawPayload, filterEngine)
	if err != nil {
		t.Fatalf("Failed to enqueue with filter: %v", err)
	}

	if entry.Payload == rawPayload {
		t.Errorf("Expected payload to be scrubbed, but got unscrubbed: %s", entry.Payload)
	}

	if !testing.Verbose() && entry.Payload == "" {
		t.Errorf("Unexpected empty payload")
	}
}

func TestWASMFilterEngineDropper(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	filterEngine := NewWASMFilterEngine()
	filterEngine.RegisterRule("metrics", func(topic, payload string) (string, bool, error) {
		if payload == "drop_me" {
			return "", true, nil // Drop record
		}
		return payload, false, nil
	})

	_, err := engine.EnqueueWithFilter("metrics", "drop_me", filterEngine)
	if err == nil {
		t.Errorf("Expected record to be dropped by filter rule, but got no error")
	}
}
