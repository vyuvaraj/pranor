package mirror

import (
	"context"
	"testing"
	"time"
)

func TestGeoReplicationCRDT(t *testing.T) {
	mw := NewMirrorWorker("us-east-1", []string{"http://eu-west-1.servqueue:8082"})

	ok, err := mw.SyncEvent(context.Background(), "orders", "evt-100", "payload-a")
	if err != nil || !ok {
		t.Fatalf("SyncEvent failed: %v, ok=%t", err, ok)
	}

	if mw.GetSyncedCount() != 1 {
		t.Errorf("Expected 1 synced count, got %d", mw.GetSyncedCount())
	}

	// Conflict resolution (older timestamp rejected)
	oldRec := EventRecord{
		ID:        "evt-100",
		Topic:     "orders",
		Payload:   "payload-old",
		Timestamp: time.Now().Add(-1 * time.Hour),
		NodeID:    "eu-west-1",
	}

	merged := mw.State.Merge(oldRec)
	if merged {
		t.Errorf("Expected older event to be rejected by CRDT LWW rule")
	}
}
