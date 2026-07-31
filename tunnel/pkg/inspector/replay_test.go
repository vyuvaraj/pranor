package import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestInspectorBuffer_CaptureAndReplay(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("replayed: " + string(body)))
	}))
	defer ts.Close()

	rib := NewRequestInspectorBuffer(10)

	tx := CapturedTransaction{
		ID:             "tx-100",
		TunnelID:       "tn-1",
		Method:         http.MethodPost,
		URL:            "http://placeholder",
		RequestBody:    "hello replay",
		ResponseStatus: 200,
		Timestamp:      time.Now(),
	}

	rib.CaptureTransaction(tx)

	history := rib.GetTransactions()
	if len(history) != 1 || history[0].ID != "tx-100" {
		t.Fatalf("unexpected history: %+v", history)
	}

	// Replay transaction against test server
	replayed, err := rib.ReplayTransaction(context.Background(), "tx-100", ts.URL)
	if err != nil {
		t.Fatalf("ReplayTransaction failed: %v", err)
	}

	if replayed.ResponseStatus != http.StatusCreated || replayed.ResponseBody != "replayed: hello replay" {
		t.Errorf("unexpected replayed transaction: %+v", replayed)
	}
}
