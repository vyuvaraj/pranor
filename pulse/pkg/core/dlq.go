package import (
	"fmt"
	"sync"
	"time"
)

type DLQPolicy struct {
	MaxRetries       int           `json:"max_retries"`
	InitialBackoff   time.Duration `json:"initial_backoff"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
}

type FailedMessage struct {
	LogEntry     LogEntry      `json:"log_entry"`
	RetryCount   int           `json:"retry_count"`
	NextRetryAt  time.Time     `json:"next_retry_at"`
	FailureReason string       `json:"failure_reason"`
}

type DLQManager struct {
	engine *Engine
	policy DLQPolicy
	mu     sync.Mutex
	dlq    map[string][]FailedMessage // topic -> failed messages
}

func NewDLQManager(engine *Engine, policy DLQPolicy) *DLQManager {
	if policy.MaxRetries <= 0 {
		policy.MaxRetries = 3
	}
	if policy.InitialBackoff <= 0 {
		policy.InitialBackoff = 1 * time.Second
	}
	if policy.BackoffMultiplier <= 1.0 {
		policy.BackoffMultiplier = 2.0
	}

	return &DLQManager{
		engine: engine,
		policy: policy,
		dlq:    make(map[string][]FailedMessage),
	}
}

// HandleFailure handles processing failure, applying exponential backoff or routing to DLQ.
func (d *DLQManager) HandleFailure(entry LogEntry, reason string) (LogEntry, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	dlqTopic := fmt.Sprintf("%s.dlq", entry.Topic)

	// Route to DLQ topic in engine storage
	dlqEntry, err := d.engine.Enqueue(dlqTopic, fmt.Sprintf(`{"original_payload": %q, "reason": %q}`, entry.Payload, reason))
	if err != nil {
		return LogEntry{}, fmt.Errorf("dlq: failed to enqueue to DLQ topic '%s': %w", dlqTopic, err)
	}

	failed := FailedMessage{
		LogEntry:      entry,
		RetryCount:    1,
		NextRetryAt:   time.Now().Add(d.policy.InitialBackoff),
		FailureReason: reason,
	}
	d.dlq[entry.Topic] = append(d.dlq[entry.Topic], failed)
	return dlqEntry, nil
}

// ReplayDLQ redelivers poison-pill messages from DLQ topic back to main topic.
func (d *DLQManager) ReplayDLQ(topic string) ([]LogEntry, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	messages, exists := d.dlq[topic]
	if !exists || len(messages) == 0 {
		return nil, nil
	}

	var replayed []LogEntry
	for _, failed := range messages {
		entry, err := d.engine.Enqueue(topic, failed.LogEntry.Payload)
		if err != nil {
			return replayed, fmt.Errorf("dlq: replay failed for topic '%s': %w", topic, err)
		}
		replayed = append(replayed, entry)
	}

	delete(d.dlq, topic)
	return replayed, nil
}

func (d *DLQManager) GetDLQMessages(topic string) []FailedMessage {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dlq[topic]
}
