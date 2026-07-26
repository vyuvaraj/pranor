package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DLQMessage represents a dead-lettered topic message.
type DLQMessage struct {
	ID            string    `json:"id"`
	OriginalTopic string    `json:"original_topic"`
	Key           string    `json:"key"`
	Payload       string    `json:"payload"`
	FailureReason string    `json:"failure_reason"`
	FailedAt      time.Time `json:"failed_at"`
	Replayed      bool      `json:"replayed"`
}

// DLQReplayManager manages dead-letter message inspection and re-publishing.
type DLQReplayManager struct {
	mu           sync.RWMutex
	messages     map[string]*DLQMessage // id -> DLQMessage
	republishFn  func(ctx context.Context, topic, key, payload string) error
}

// NewDLQReplayManager creates a DLQReplayManager instance.
func NewDLQReplayManager(republishFn func(ctx context.Context, topic, key, payload string) error) *DLQReplayManager {
	return &DLQReplayManager{
		messages:    make(map[string]*DLQMessage),
		republishFn: republishFn,
	}
}

// AddDeadLetter records a message into the DLQ store.
func (drm *DLQReplayManager) AddDeadLetter(msg DLQMessage) {
	if msg.ID == "" {
		return
	}
	drm.mu.Lock()
	defer drm.mu.Unlock()
	drm.messages[msg.ID] = &msg
}

// ReplayMessage re-publishes a dead-lettered message back to its original topic.
func (drm *DLQReplayManager) ReplayMessage(ctx context.Context, msgID string) error {
	drm.mu.Lock()
	msg, ok := drm.messages[msgID]
	if !ok {
		drm.mu.Unlock()
		return fmt.Errorf("DLQ message ID '%s' not found", msgID)
	}
	drm.mu.Unlock()

	if drm.republishFn != nil {
		if err := drm.republishFn(ctx, msg.OriginalTopic, msg.Key, msg.Payload); err != nil {
			return fmt.Errorf("failed to re-publish DLQ message: %w", err)
		}
	}

	drm.mu.Lock()
	msg.Replayed = true
	drm.mu.Unlock()

	return nil
}

// GetDLQMessages returns active dead-lettered messages.
func (drm *DLQReplayManager) GetDLQMessages() []DLQMessage {
	drm.mu.RLock()
	defer drm.mu.RUnlock()

	res := make([]DLQMessage, 0, len(drm.messages))
	for _, m := range drm.messages {
		res = append(res, *m)
	}
	return res
}

// HTTPHandler exposes `/api/v1/console/dlq` and `/api/v1/console/dlq/replay`.
func (drm *DLQReplayManager) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/console/dlq/replay" {
			var body struct {
				MessageID string `json:"message_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MessageID == "" {
				http.Error(w, "invalid payload or missing message_id", http.StatusBadRequest)
				return
			}
			if err := drm.ReplayMessage(r.Context(), body.MessageID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]bool{"replayed": true})
			return
		}

		w.WriteHeader(http.StatusOK)
		msgs := drm.GetDLQMessages()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":    len(msgs),
			"messages": msgs,
		})
	})
}
