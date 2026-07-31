package import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExemplarStore_RecordAndRender(t *testing.T) {
	store := NewExemplarStore()

	traceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	spanID := "00f067aa0ba902b7"

	store.RecordExemplar("http_request_duration_seconds", 0.125, traceID, spanID)

	output := store.RenderOpenMetrics()
	if !strings.Contains(output, `trace_id="4bf92f3577b34da6a3ce929d0e0e4736"`) {
		t.Errorf("expected OpenMetrics exemplar trace_id, got: %s", output)
	}

	// Test HTTP Handler
	handler := store.HTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/openmetrics-text") {
		t.Errorf("expected openmetrics content-type, got %s", contentType)
	}
}
