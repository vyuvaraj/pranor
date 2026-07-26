package storage

import (
	"testing"
	"time"
)

func TestLogCompactionEngine_CompactSegment(t *testing.T) {
	engine := NewLogCompactionEngine(24*time.Hour, 10*time.Minute)

	now := time.Now()

	entries := []CompactLogEntry{
		{Offset: 1, Key: "user-10", Value: []byte("v1"), Timestamp: now.Add(-1 * time.Hour)},
		{Offset: 2, Key: "user-10", Value: []byte("v2"), Timestamp: now.Add(-30 * time.Minute)}, // Latest user-10
		{Offset: 3, Key: "user-20", Value: []byte("v1"), Timestamp: now.Add(-5 * time.Minute)},
		{Offset: 4, Key: "user-20", Value: nil, Timestamp: now.Add(-15 * time.Minute)}, // Expired Tombstone (>10m)
	}

	compacted := engine.CompactSegment(entries, now)

	if len(compacted) != 1 {
		t.Fatalf("expected 1 compacted entry, got %d: %+v", len(compacted), compacted)
	}

	if compacted[0].Key != "user-10" || string(compacted[0].Value) != "v2" {
		t.Errorf("unexpected compacted entry: %+v", compacted[0])
	}
}
