package import (
	"fmt"
	"strings"
)

// StreamFilterFunc defines a client-side stream transformation function.
type StreamFilterFunc func(topic string, payload string) (transformed string, drop bool, err error)

// WASMFilterEngine manages client-side stream processing rules before OPFS commit.
type WASMFilterEngine struct {
	filters map[string]StreamFilterFunc
}

func NewWASMFilterEngine() *WASMFilterEngine {
	return &WASMFilterEngine{
		filters: make(map[string]StreamFilterFunc),
	}
}

// RegisterRule registers a topic transformation function.
func (w *WASMFilterEngine) RegisterRule(topic string, filter StreamFilterFunc) {
	w.filters[topic] = filter
}

// RegisterBuiltinScrubber registers a standard PII scrubbing transformation.
func (w *WASMFilterEngine) RegisterBuiltinScrubber(topic string) {
	w.filters[topic] = func(t, payload string) (string, bool, error) {
		// Scrubber masks sensitive keywords
		cleaned := strings.ReplaceAll(payload, "password", "*****")
		cleaned = strings.ReplaceAll(cleaned, "secret", "*****")
		return cleaned, false, nil
	}
}

// Process applies registered stream processing filters to a payload.
func (w *WASMFilterEngine) Process(topic string, payload string) (string, bool, error) {
	filter, exists := w.filters[topic]
	if !exists {
		return payload, false, nil // Pass-through
	}
	return filter(topic, payload)
}

// ProcessPayload applies filter engine rules on Engine Enqueue
func (e *Engine) EnqueueWithFilter(topic, payload string, filterEngine *WASMFilterEngine) (LogEntry, error) {
	if filterEngine != nil {
		processed, drop, err := filterEngine.Process(topic, payload)
		if err != nil {
			return LogEntry{}, fmt.Errorf("filter error: %w", err)
		}
		if drop {
			return LogEntry{}, fmt.Errorf("record dropped by filter rule")
		}
		payload = processed
	}
	return e.Enqueue(topic, payload)
}
