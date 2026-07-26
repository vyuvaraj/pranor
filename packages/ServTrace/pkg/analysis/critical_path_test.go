package analysis

import (
	"testing"
	"time"
)

func TestCriticalPathAnalyzer_AnalyzeTrace(t *testing.T) {
	cpa := NewCriticalPathAnalyzer()

	now := time.Now()
	spans := []TraceSpan{
		{SpanID: "root", ParentSpanID: "", Name: "HTTP GET /checkout", Service: "gateway", StartTime: now, DurationMs: 120},
		{SpanID: "s1", ParentSpanID: "root", Name: "Auth Check", Service: "auth-svc", StartTime: now.Add(5 * time.Millisecond), DurationMs: 20},
		{SpanID: "s2", ParentSpanID: "root", Name: "Process Payment", Service: "payment-svc", StartTime: now.Add(30 * time.Millisecond), DurationMs: 80},
		{SpanID: "s3", ParentSpanID: "s2", Name: "DB Insert", Service: "payment-db", StartTime: now.Add(40 * time.Millisecond), DurationMs: 50},
	}

	result, err := cpa.AnalyzeTrace("trace-abc-123", spans)
	if err != nil {
		t.Fatalf("AnalyzeTrace failed: %v", err)
	}

	if result.TraceID != "trace-abc-123" {
		t.Errorf("unexpected trace ID: %s", result.TraceID)
	}

	// Critical path should be root -> s2 -> s3 (duration = 120 + 80 + 50 = 250)
	if len(result.CriticalPath) != 3 {
		t.Fatalf("expected 3 nodes in critical path, got %d", len(result.CriticalPath))
	}

	if result.CriticalPath[0].SpanID != "root" || result.CriticalPath[1].SpanID != "s2" || result.CriticalPath[2].SpanID != "s3" {
		t.Errorf("unexpected critical path sequence: %+v", result.CriticalPath)
	}
}
