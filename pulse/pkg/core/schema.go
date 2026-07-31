package import (
	"encoding/json"
	"fmt"
	"sync"
)

type SchemaRule struct {
	RequiredFields []string `json:"required_fields"`
}

type SchemaRegistry struct {
	mu    sync.RWMutex
	rules map[string]SchemaRule
}

func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		rules: make(map[string]SchemaRule),
	}
}

// RegisterSchema registers mandatory payload structure rules for a topic.
func (s *SchemaRegistry) RegisterSchema(topic string, rule SchemaRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[topic] = rule
}

// ValidatePayload checks payload validity against registered schema rules.
func (s *SchemaRegistry) ValidatePayload(topic string, payload string) error {
	s.mu.RLock()
	rule, exists := s.rules[topic]
	s.mu.RUnlock()

	if !exists || len(rule.RequiredFields) == 0 {
		return nil // No schema constraint for topic
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return fmt.Errorf("schema: invalid JSON payload: %w", err)
	}

	for _, field := range rule.RequiredFields {
		if _, present := data[field]; !present {
			return fmt.Errorf("schema: missing required field '%s' for topic '%s'", field, topic)
		}
	}

	return nil
}

// EnqueueWithSchema validates schema before appending event to storage log.
func (e *Engine) EnqueueWithSchema(topic, payload string, registry *SchemaRegistry) (LogEntry, error) {
	if registry != nil {
		if err := registry.ValidatePayload(topic, payload); err != nil {
			return LogEntry{}, fmt.Errorf("enqueue schema validation failed: %w", err)
		}
	}
	return e.Enqueue(topic, payload)
}
