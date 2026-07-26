package cache

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsCollector_RecordingAndSnapshot(t *testing.T) {
	m := NewMetricsCollector()

	m.RecordHit()
	m.RecordHit()
	m.RecordHit()
	m.RecordMiss()
	m.RecordEviction()
	m.RecordSet()
	m.RecordDelete()

	snap := m.Snapshot()

	if snap.Hits != 3 {
		t.Errorf("expected 3 hits, got %d", snap.Hits)
	}
	if snap.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", snap.Misses)
	}
	if snap.TotalRequests != 4 {
		t.Errorf("expected 4 total requests, got %d", snap.TotalRequests)
	}
	if snap.HitRate != 0.75 {
		t.Errorf("expected hit rate 0.75, got %f", snap.HitRate)
	}
	if snap.Evictions != 1 {
		t.Errorf("expected 1 eviction, got %d", snap.Evictions)
	}
}

func TestMetricsCollector_HTTPHandler(t *testing.T) {
	m := NewMetricsCollector()
	m.RecordHit()
	m.RecordMiss()

	handler := m.HTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var snap MetricsSnapshot
	if err := json.NewDecoder(w.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if snap.Hits != 1 || snap.Misses != 1 || snap.HitRate != 0.5 {
		t.Errorf("unexpected metrics snapshot: %+v", snap)
	}
}
