package analytics

import (
	"testing"
	"time"
)

func TestStreamSQLWindowingAggregations(t *testing.T) {
	engine := NewStreamSQLEngine(5 * time.Second)

	_ = engine.RecordEvent("orders", `{"amount": 100.0}`, "amount")
	_ = engine.RecordEvent("orders", `{"amount": 200.0}`, "amount")
	_ = engine.RecordEvent("orders", `{"amount": 300.0}`, "amount")

	res, err := engine.EvaluateWindow("orders")
	if err != nil {
		t.Fatalf("EvaluateWindow failed: %v", err)
	}

	if res.Count != 3 {
		t.Errorf("Expected count 3, got %d", res.Count)
	}

	if res.Sum != 600.0 {
		t.Errorf("Expected sum 600.0, got %f", res.Sum)
	}

	if res.Avg != 200.0 {
		t.Errorf("Expected avg 200.0, got %f", res.Avg)
	}

	if res.Min != 100.0 || res.Max != 300.0 {
		t.Errorf("Expected min 100.0 max 300.0, got min %f max %f", res.Min, res.Max)
	}
}
