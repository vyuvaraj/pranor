package import (
	"testing"
	"time"
)

func TestTelemetryTracker_StartAndEndSpan(t *testing.T) {
	tracker := NewTelemetryTracker()

	span, traceparent := tracker.StartTaskSpan("inst-123", "payment-task", "")
	if traceparent == "" {
		t.Fatal("expected non-empty traceparent string")
	}

	time.Sleep(10 * time.Millisecond)

	rec := tracker.EndTaskSpan(span, "inst-123", "payment-task", nil)

	if rec.InstanceID != "inst-123" || rec.TaskName != "payment-task" {
		t.Errorf("unexpected cost record: %+v", rec)
	}
	if rec.CostUSD <= 0 {
		t.Errorf("expected positive estimated cost USD, got %f", rec.CostUSD)
	}

	records := tracker.GetRecords()
	if len(records) != 1 {
		t.Errorf("expected 1 recorded telemetry entry, got %d", len(records))
	}
}
