package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/core"
)

// StepCostRecord captures telemetry span attribution and estimated token/execution cost.
type StepCostRecord struct {
	InstanceID string        `json:"instance_id"`
	TaskName   string        `json:"task_name"`
	TraceID    string        `json:"trace_id"`
	SpanID     string        `json:"span_id"`
	Duration   time.Duration `json:"duration"`
	CostUSD    float64       `json:"cost_usd"`
}

// TelemetryTracker manages OpenTelemetry span context propagation and step cost tracking.
type TelemetryTracker struct {
	mu      sync.RWMutex
	records []StepCostRecord
}

// NewTelemetryTracker creates a TelemetryTracker.
func NewTelemetryTracker() *TelemetryTracker {
	return &TelemetryTracker{
		records: make([]StepCostRecord, 0),
	}
}

// StartTaskSpan creates a W3C traceparent context and starts an OTel span for a workflow step.
func (tt *TelemetryTracker) StartTaskSpan(instanceID, taskName, parentTraceparent string) (*core.Span, string) {
	var traceID, spanID string
	if parentTraceparent != "" {
		parts := parseTraceparent(parentTraceparent)
		if len(parts) >= 3 {
			traceID = parts[1]
		}
	}
	if traceID == "" {
		traceID = randomHex(16)
	}
	spanID = randomHex(8)

	span := core.StartSpan(fmt.Sprintf("Pranor Flow:task %s %s", instanceID, taskName), parentTraceparent)
	if span != nil {
		span.TraceID = traceID
		span.SpanID = spanID
	}

	traceparent := fmt.Sprintf("00-%s-%s-01", traceID, spanID)
	return span, traceparent
}

// EndTaskSpan records span completion, duration, and computes estimated execution cost.
func (tt *TelemetryTracker) EndTaskSpan(span *core.Span, instanceID, taskName string, err error) StepCostRecord {
	var duration time.Duration
	var traceID, spanID string
	if span != nil {
		traceID = span.TraceID
		spanID = span.SpanID
		core.EndSpan(span, err, map[string]interface{}{
			"Pranor Flow.instance_id": instanceID,
			"Pranor Flow.task_name":   taskName,
		})
		duration = time.Duration(time.Now().UnixNano() - span.StartTime)
	} else {
		traceID = randomHex(16)
		spanID = randomHex(8)
		duration = 10 * time.Millisecond
	}

	// Cost estimation formula: base $0.0001 per step + $0.00005 per second compute
	costUSD := 0.0001 + (duration.Seconds() * 0.00005)

	record := StepCostRecord{
		InstanceID: instanceID,
		TaskName:   taskName,
		TraceID:    traceID,
		SpanID:     spanID,
		Duration:   duration,
		CostUSD:    costUSD,
	}

	tt.mu.Lock()
	tt.records = append(tt.records, record)
	tt.mu.Unlock()

	return record
}

// GetRecords returns all recorded step telemetry cost snapshots.
func (tt *TelemetryTracker) GetRecords() []StepCostRecord {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	out := make([]StepCostRecord, len(tt.records))
	copy(out, tt.records)
	return out
}

func randomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func parseTraceparent(tp string) []string {
	// Format: 00-traceid-spanid-01
	return splitString(tp, "-")
}

func splitString(s, sep string) []string {
	var res []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			res = append(res, s[start:i])
			start = i + len(sep)
		}
	}
	res = append(res, s[start:])
	return res
}
