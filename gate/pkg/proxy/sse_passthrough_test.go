package import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSSEPassthroughProxy_ProxySSEStream(t *testing.T) {
	// Upstream LLM Server emitting SSE events
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher := w.(http.Flusher)
		_, _ = fmt.Fprintln(w, "data: {\"token\": \"Hello\"}")
		flusher.Flush()
		_, _ = fmt.Fprintln(w, "data: {\"token\": \" World\"}")
		flusher.Flush()
		_, _ = fmt.Fprintln(w, "data: [DONE]")
		flusher.Flush()
	}))
	defer upstream.Close()

	proxy := NewSSEPassthroughProxy()

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()

	err := proxy.ProxySSEStream(w, req, upstream.URL)
	if err != nil {
		t.Fatalf("ProxySSEStream failed: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "data: {\"token\": \"Hello\"}") || !strings.Contains(body, "data: [DONE]") {
		t.Errorf("unexpected SSE response body: %s", body)
	}
}
