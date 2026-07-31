package import (
	"bytes"
	"strings"
	"testing"
)

func TestAES256EncryptionAtRest(t *testing.T) {
	key := bytes.Repeat([]byte("a"), 32)
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	if err := engine.SetEncryptionKey(key); err != nil {
		t.Fatalf("Failed to set encryption key: %v", err)
	}

	// 1. Enqueue encrypted payload
	payload := `{"user_id": 9901, "ssn": "123-45-6789"}`
	entry, err := engine.Enqueue("sensitive_events", payload)
	if err != nil {
		t.Fatalf("Failed to enqueue payload: %v", err)
	}

	// Check underlying raw storage driver payload starts with ENC: prefix
	rawEntries, err := driver.ReadRange("sensitive_events", entry.Offset, 1)
	if err != nil || len(rawEntries) == 0 {
		t.Fatalf("Failed to read raw entry: %v", err)
	}
	if !strings.HasPrefix(rawEntries[0].Payload, "ENC:") {
		t.Errorf("Expected raw payload to start with 'ENC:', got: %s", rawEntries[0].Payload)
	}

	// 2. Dequeue decrypted payload through Engine
	decryptedEntries, err := engine.Dequeue("sensitive_events", entry.Offset, 1)
	if err != nil || len(decryptedEntries) == 0 {
		t.Fatalf("Failed to dequeue entry: %v", err)
	}

	if decryptedEntries[0].Payload != payload {
		t.Errorf("Expected decrypted payload '%s', got '%s'", payload, decryptedEntries[0].Payload)
	}
}

func TestAutoCompactionQuotaManager(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)
	defer engine.Close()

	// Enqueue multiple records
	for i := 1; i <= 10; i++ {
		_, _ = engine.Enqueue("logs", `{"log_level": "info", "msg": "system health OK"}`)
	}

	// Mark first 5 records as synced
	var syncedOffsets []uint64
	for i := uint64(1); i <= 5; i++ {
		syncedOffsets = append(syncedOffsets, i)
	}
	_ = engine.AcknowledgeSync(syncedOffsets)

	// Trigger Compaction with tight quota policy
	policy := QuotaPolicy{
		MaxBytes:         50, // Small quota budget to force purge of synced records
		MaxSyncedRecords: 2,
	}

	purged, err := engine.Compact(policy)
	if err != nil {
		t.Fatalf("Compaction failed: %v", err)
	}

	if purged == 0 {
		t.Errorf("Expected purged records > 0, got %d", purged)
	}
}
