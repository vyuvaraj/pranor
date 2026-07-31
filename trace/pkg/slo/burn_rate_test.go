package import (
	"testing"
)

func TestSloBreachPredictor_EvaluateFastAndSlowBurn(t *testing.T) {
	cfg := SLOBurnRateConfig{
		SLOName:           "checkout-latency-slo",
		TargetSLO:         0.999, // 0.1% error budget
		FastBurnThreshold: 14.4,  // 14.4x burn rate = 1.44% error rate
		SlowBurnThreshold: 6.0,   // 6.0x burn rate = 0.6% error rate
	}

	predictor := NewSloBreachPredictor(cfg)

	// 1. Error rate 0.05% (0.5x burn rate) -> no alert
	alert, fired := predictor.EvaluateBurnRate(0.0005, true)
	if fired || alert != nil {
		t.Error("expected no alert for low burn rate")
	}

	// 2. Fast window error rate 2.0% (20x burn rate > 14.4x) -> fast_burn alert
	alert, fired = predictor.EvaluateBurnRate(0.02, true)
	if !fired || alert.AlertType != "fast_burn" {
		t.Fatalf("expected fast_burn alert, got fired=%v alert=%+v", fired, alert)
	}

	if alert.BurnRate < 19.9 || alert.BurnRate > 20.1 {
		t.Errorf("expected ~20.0x burn rate, got %f", alert.BurnRate)
	}
}
