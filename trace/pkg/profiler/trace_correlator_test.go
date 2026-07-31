package import (
	"testing"
	"time"
)

func TestTraceCorrelator_CorrelateAndRetrieve(t *testing.T) {
	tc := NewTraceCorrelator()

	traceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	spanID := "00f067aa0ba902b7"

	samples := []StackSample{
		{
			StackFrames: []string{"main", "runtime.mallocgc"},
			Count:       50,
			SampleType:  "memory",
			Timestamp:   time.Now(),
		},
	}

	corr := tc.CorrelateTraceSpan(traceID, spanID, "auth-svc", samples)
	if corr.TraceID != traceID || corr.TopFrame != "runtime.mallocgc" {
		t.Fatalf("unexpected correlation object: %+v", corr)
	}

	retrieved, found := tc.GetCorrelation(traceID)
	if !found || retrieved.SpanID != spanID {
		t.Errorf("failed to retrieve correlation by trace ID: %+v", retrieved)
	}
}
