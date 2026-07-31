package import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTimelineRecorder_CircularBufferAndHTTP(t *testing.T) {
	rec := NewTimelineRecorder(3)

	rec.Record(TimelineEntry{JobID: "j1", DurationMs: 10, Outcome: "success"})
	rec.Record(TimelineEntry{JobID: "j2", DurationMs: 20, Outcome: "success"})
	rec.Record(TimelineEntry{JobID: "j3", DurationMs: 30, Outcome: "failed"})

	entries := rec.GetEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Record 4th item — should evict j1
	rec.Record(TimelineEntry{JobID: "j4", DurationMs: 40, Outcome: "success"})

	entries = rec.GetEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after overflow, got %d", len(entries))
	}

	if entries[0].JobID != "j2" || entries[2].JobID != "j4" {
		t.Errorf("unexpected buffer contents: %+v", entries)
	}

	// Test HTTP Endpoint
	handler := rec.HTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/timeline", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp["count"].(float64) != 3 {
		t.Errorf("expected count 3, got %v", resp["count"])
	}
}
