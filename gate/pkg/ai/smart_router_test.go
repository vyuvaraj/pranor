package ai

import (
	"context"
	"testing"
)

func TestSmartAIRouter(t *testing.T) {
	router := NewSmartAIRouter(SmartAIRouterConfig{
		EnablePrefetch: true,
	})

	// Low complexity prompt
	resLow, compLow, saved, err := router.RouteAndExecute(context.Background(), "What is 2 + 2?")
	if err != nil || compLow != ComplexityLow || saved <= 0 {
		t.Fatalf("low complexity routing failed: %v", err)
	}
	if resLow == "" {
		t.Errorf("empty response from smart router")
	}

	// High complexity prompt
	_, compHigh, _, err := router.RouteAndExecute(context.Background(), "Please refactor this architecture to use distributed saga pattern with proof of correctness.")
	if err != nil || compHigh != ComplexityHigh {
		t.Errorf("high complexity routing failed")
	}

	low, high, savedTotal := router.GetTelemetryStats()
	if low != 1 || high != 1 || savedTotal <= 0 {
		t.Errorf("telemetry stats mismatch: low=%d, high=%d, saved=%.4f", low, high, savedTotal)
	}
}
