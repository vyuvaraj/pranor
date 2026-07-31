package profiler

import (
	"sync"
	"time"
)

// TraceFlamegraphCorrelation links an OTel trace ID with eBPF profile stack samples.
type TraceFlamegraphCorrelation struct {
	TraceID     string        `json:"trace_id"`
	SpanID      string        `json:"span_id"`
	Service     string        `json:"service"`
	SampleCount int           `json:"sample_count"`
	TopFrame    string        `json:"top_frame"`
	Timestamp   time.Time     `json:"timestamp"`
	Flamegraph  string        `json:"flamegraph_text"`
}

// TraceCorrelator handles automatic correlation between OTel trace spans and eBPF stack samples.
type TraceCorrelator struct {
	mu           sync.RWMutex
	correlations map[string]*TraceFlamegraphCorrelation // trace_id -> correlation
}

// NewTraceCorrelator creates a TraceCorrelator instance.
func NewTraceCorrelator() *TraceCorrelator {
	return &TraceCorrelator{
		correlations: make(map[string]*TraceFlamegraphCorrelation),
	}
}

// CorrelateTraceSpan binds a trace ID and span ID to sampled eBPF stack traces.
func (tc *TraceCorrelator) CorrelateTraceSpan(traceID, spanID, service string, samples []StackSample) *TraceFlamegraphCorrelation {
	if traceID == "" {
		return nil
	}

	topFrame := "unknown"
	if len(samples) > 0 && len(samples[0].StackFrames) > 0 {
		topFrame = samples[0].StackFrames[len(samples[0].StackFrames)-1]
	}

	corr := &TraceFlamegraphCorrelation{
		TraceID:     traceID,
		SpanID:      spanID,
		Service:     service,
		SampleCount: len(samples),
		TopFrame:    topFrame,
		Timestamp:   time.Now(),
	}

	tc.mu.Lock()
	tc.correlations[traceID] = corr
	tc.mu.Unlock()

	return corr
}

// GetCorrelation retrieves flamegraph correlation by trace ID.
func (tc *TraceCorrelator) GetCorrelation(traceID string) (*TraceFlamegraphCorrelation, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	corr, ok := tc.correlations[traceID]
	return corr, ok
}
