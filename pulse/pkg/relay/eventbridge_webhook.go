//go:build !enterprise

package import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type WebhookConfig struct {
	TargetURL  string `json:"target_url"`
	AuthHeader string `json:"auth_header"`
	MaxRetries int    `json:"max_retries"`
}

type EventBridgeConfig struct {
	EventBusName string `json:"event_bus_name"`
	Region       string `json:"region"`
	EndpointURL  string `json:"endpoint_url"`
}

type RelayEvent struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

type EventBridgeRelay struct {
	Webhooks    []WebhookConfig
	EventBridge []EventBridgeConfig
	HTTPClient  *http.Client
	mu          sync.Mutex
	Dispatched  uint64
}

func NewEventBridgeRelay() *EventBridgeRelay {
	return &EventBridgeRelay{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *EventBridgeRelay) RegisterWebhook(cfg WebhookConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	r.Webhooks = append(r.Webhooks, cfg)
}

func (r *EventBridgeRelay) RegisterEventBridge(cfg EventBridgeConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.EventBridge = append(r.EventBridge, cfg)
}

func (r *EventBridgeRelay) Dispatch(ctx context.Context, topic, id, payload string) error {
	r.mu.Lock()
	r.Dispatched++
	r.mu.Unlock()

	event := RelayEvent{
		ID:        id,
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	eventData, _ := json.Marshal(event)

	for _, wh := range r.Webhooks {
		req, err := http.NewRequestWithContext(ctx, "POST", wh.TargetURL, bytes.NewBuffer(eventData))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if wh.AuthHeader != "" {
			req.Header.Set("Authorization", wh.AuthHeader)
		}

		resp, err := r.HTTPClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
	}

	for _, eb := range r.EventBridge {
		if eb.EndpointURL != "" {
			ebBody := map[string]interface{}{
				"DetailType": "PranorPulseTopicEvent",
				"Source":     "Pranor Pulse.engine",
				"EventBus":   eb.EventBusName,
				"Detail":     string(eventData),
			}
			ebBytes, _ := json.Marshal(ebBody)
			req, err := http.NewRequestWithContext(ctx, "POST", eb.EndpointURL, bytes.NewBuffer(ebBytes))
			if err == nil {
				req.Header.Set("Content-Type", "application/x-amz-json-1.1")
				resp, err := r.HTTPClient.Do(req)
				if err == nil {
					_ = resp.Body.Close()
				}
			}
		}
	}

	return nil
}

func (r *EventBridgeRelay) GetDispatchedCount() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Dispatched
}
