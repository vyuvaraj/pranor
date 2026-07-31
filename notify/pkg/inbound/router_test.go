package inbound

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestInboundRouter_RouteEmail(t *testing.T) {
	var receivedPayload InboundEmailPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	router := NewInboundRouter()
	router.RegisterRoute(InboundWebhookRoute{
		ID:         "support-route",
		Recipient:  "*@support.dev",
		WebhookURL: ts.URL,
	})

	email := InboundEmailPayload{
		ID:         "msg-123",
		From:       "user@example.com",
		To:         []string{"help@support.dev"},
		Subject:    "Need Assistance",
		TextBody:   "Hello support team",
		ReceivedAt: time.Now(),
	}

	errs, count := router.RouteEmail(context.Background(), email)
	if len(errs) != 0 || count != 1 {
		t.Fatalf("expected 1 successful webhook dispatch, got count=%d, errs=%v", count, errs)
	}

	if receivedPayload.Subject != "Need Assistance" || receivedPayload.From != "user@example.com" {
		t.Errorf("received payload mismatch: %+v", receivedPayload)
	}
}
