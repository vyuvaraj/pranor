package import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQueryHistogram_RecordAndSlowQueryLog(t *testing.T) {
	qh := NewQueryHistogram(100 * time.Millisecond)

	qh.RecordQuery("SELECT * FROM users WHERE id = 1", 2*time.Millisecond)
	qh.RecordQuery("SELECT * FROM orders WHERE status = 'pending'", 300*time.Millisecond) // Slow

	buckets, slowLogs := qh.GetSnapshot()

	if buckets["le_5ms"] != 1 || buckets["gt_200ms"] != 1 {
		t.Errorf("unexpected bucket counts: %+v", buckets)
	}

	if len(slowLogs) != 1 || slowLogs[0].Query != "SELECT * FROM orders WHERE status = 'pending'" {
		t.Fatalf("unexpected slow query log: %+v", slowLogs)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pool/query-stats", nil)
	w := httptest.NewRecorder()
	qh.HTTPHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w.Code)
	}
}
