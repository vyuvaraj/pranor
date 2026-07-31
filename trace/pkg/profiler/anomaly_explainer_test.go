package import (
	"strings"
	"testing"
	"time"
)

func TestAnomalyExplainer_ExplainAnomaly(t *testing.T) {
	explainer := NewAnomalyExplainer()

	samples := []StackSample{
		{
			StackFrames: []string{"main", "http.ServeHTTP", "db.QuerySQL"},
			Count:       80,
			SampleType:  "cpu",
			Timestamp:   time.Now(),
		},
		{
			StackFrames: []string{"main", "http.ServeHTTP", "json.Unmarshal"},
			Count:       20,
			SampleType:  "cpu",
			Timestamp:   time.Now(),
		},
	}

	explainer.RecordSamples("payment-service", samples)

	explanation, err := explainer.ExplainAnomaly("payment-service")
	if err != nil {
		t.Fatalf("ExplainAnomaly failed: %v", err)
	}

	if explanation.TopHotspotFunction != "db.QuerySQL" {
		t.Errorf("expected top hotspot db.QuerySQL, got %s", explanation.TopHotspotFunction)
	}
	if explanation.PercentageCPU != 80.0 {
		t.Errorf("expected 80%% CPU allocation, got %f", explanation.PercentageCPU)
	}

	flamegraph := explainer.FormatFlamegraphText("payment-service")
	if !strings.Contains(flamegraph, "main;http.ServeHTTP;db.QuerySQL 80") {
		t.Errorf("unexpected flamegraph output: %s", flamegraph)
	}
}
