package relay

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vyuvaraj/serv/packages/ServQueue/pkg/core"
)

func TestOutboxRelaySyncNow(t *testing.T) {
	driver := core.NewMemoryDriver()
	engine := core.NewEngine(driver)
	defer engine.Close()

	// Enqueue 2 test entries
	_, _ = engine.Enqueue("orders", `{"id": 1}`)
	_, _ = engine.Enqueue("orders", `{"id": 2}`)

	webTransportHeaderReceived := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-WebTransport-Mode") == "quic-multiplex" {
			webTransportHeaderReceived = true
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ack"}`))
	}))
	defer ts.Close()

	relay := NewOutboxRelay(engine, ts.URL)
	if err := relay.SyncNow(); err != nil {
		t.Fatalf("SyncNow failed: %v", err)
	}

	if !webTransportHeaderReceived {
		t.Errorf("Expected X-WebTransport-Mode header to be sent")
	}

	// Verify pending sync is now 0 after ACK
	pending, _ := engine.GetPendingSync(10)
	if len(pending) != 0 {
		t.Errorf("Expected 0 pending sync entries after relay sync, got %d", len(pending))
	}
}
