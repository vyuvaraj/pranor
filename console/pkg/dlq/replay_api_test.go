package dlq

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDLQReplayManager_AddAndReplay(t *testing.T) {
	republished := false
	mockRepublish := func(ctx context.Context, topic, key, payload string) error {
		if topic == "orders-topic" && payload == "payload-data" {
			republished = true
		}
		return nil
	}

	drm := NewDLQReplayManager(mockRepublish)

	drm.AddDeadLetter(DLQMessage{
		ID:            "dlq-1",
		OriginalTopic: "orders-topic",
		Key:           "order-99",
		Payload:       "payload-data",
		FailureReason: "unhandled exception",
		FailedAt:      time.Now(),
	})

	msgs := drm.GetDLQMessages()
	if len(msgs) != 1 || msgs[0].ID != "dlq-1" {
		t.Fatalf("unexpected DLQ messages: %+v", msgs)
	}

	// Replay via HTTP Handler
	body, _ := json.Marshal(map[string]string{"message_id": "dlq-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/console/dlq/replay", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	drm.HTTPHandler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	if !republished {
		t.Error("expected message to be re-published to topic")
	}
}
