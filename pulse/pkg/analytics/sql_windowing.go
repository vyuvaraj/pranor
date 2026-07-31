package analytics

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/pulse/pkg/core"
)

type WindowResult struct {
	Topic     string             `json:"topic"`
	Count     int64              `json:"count"`
	Sum       float64            `json:"sum"`
	Avg       float64            `json:"avg"`
	Min       float64            `json:"min"`
	Max       float64            `json:"max"`
	WindowEnd time.Time          `json:"window_end"`
}

type StreamSQLEngine struct {
	mu            sync.Mutex
	windowSize    time.Duration
	windowRecords map[string][]float64
}

func NewStreamSQLEngine(windowSize time.Duration) *StreamSQLEngine {
	if windowSize <= 0 {
		windowSize = 10 * time.Second
	}
	return &StreamSQLEngine{
		windowSize:    windowSize,
		windowRecords: make(map[string][]float64),
	}
}

// RecordEvent ingests a numeric metric field from an event payload into the sliding window.
func (s *StreamSQLEngine) RecordEvent(topic string, payload string, numericField string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return fmt.Errorf("analytics: payload unmarshal error: %w", err)
	}

	val, exists := data[numericField]
	if !exists {
		return nil // Numeric field not present in payload
	}

	var numVal float64
	switch v := val.(type) {
	case float64:
		numVal = v
	case int:
		numVal = float64(v)
	case int64:
		numVal = float64(v)
	default:
		return nil
	}

	s.windowRecords[topic] = append(s.windowRecords[topic], numVal)
	return nil
}

// EvaluateWindow computes window aggregations (COUNT, SUM, AVG, MIN, MAX).
func (s *StreamSQLEngine) EvaluateWindow(topic string) (WindowResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	values, exists := s.windowRecords[topic]
	if !exists || len(values) == 0 {
		return WindowResult{Topic: topic, WindowEnd: time.Now()}, nil
	}

	var sum float64
	min := values[0]
	max := values[0]

	for _, v := range values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	count := int64(len(values))
	avg := sum / float64(count)

	result := WindowResult{
		Topic:     topic,
		Count:     count,
		Sum:       sum,
		Avg:       avg,
		Min:       min,
		Max:       max,
		WindowEnd: time.Now(),
	}

	// Reset window after evaluation
	s.windowRecords[topic] = nil
	return result, nil
}

// ProcessStream evaluates live queue events for Stream SQL Windowing
func (s *StreamSQLEngine) ProcessStream(entries []core.LogEntry, numericField string) map[string]WindowResult {
	for _, entry := range entries {
		_ = s.RecordEvent(entry.Topic, entry.Payload, numericField)
	}

	results := make(map[string]WindowResult)
	s.mu.Lock()
	topics := make([]string, 0, len(s.windowRecords))
	for topic := range s.windowRecords {
		topics = append(topics, topic)
	}
	s.mu.Unlock()

	for _, topic := range topics {
		res, err := s.EvaluateWindow(topic)
		if err == nil {
			results[topic] = res
		}
	}
	return results
}
