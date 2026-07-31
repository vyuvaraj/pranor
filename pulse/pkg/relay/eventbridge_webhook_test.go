package relay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEventBridgeAndWebhookRelay(t *testing.T) {
	receivedWebhook := false
	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedWebhook = true
		w.WriteHeader(http.StatusOK)
	}))
	defer webhookServer.Close()

	receivedEB := false
	ebServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEB = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ebServer.Close()

	relay := NewEventBridgeRelay()
	relay.RegisterWebhook(WebhookConfig{TargetURL: webhookServer.URL})
	relay.RegisterEventBridge(EventBridgeConfig{EventBusName: "default", EndpointURL: ebServer.URL})

	err := relay.Dispatch(context.Background(), "orders", "evt-1", `{"order_id": 99}`)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if !receivedWebhook {
		t.Errorf("Expected webhook server to receive event")
	}
	if !receivedEB {
		t.Errorf("Expected EventBridge server to receive event")
	}
	if relay.GetDispatchedCount() != 1 {
		t.Errorf("Expected 1 dispatched count, got %d", relay.GetDispatchedCount())
	}
}
