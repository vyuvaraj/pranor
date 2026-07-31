package import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleQueueInspector(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queue/inspector", nil)
	w := httptest.NewRecorder()

	HandleQueueInspector(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHandleQueueTailStream(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/queue/tail?topic=test.topic", nil)
	w := httptest.NewRecorder()

	HandleQueueTailStream(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
