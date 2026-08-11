package sidecar

import (
	"context"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/learn/api"
)

func TestGRPCPredictor(t *testing.T) {
	p := NewGRPCPredictor("localhost:50051")
	ctx := context.Background()

	_, err := p.Predict(ctx, api.PredictInput{
		ModelID:  "test-model",
		BudgetMs: 1000,
	})
	if err != api.ErrEERequired {
		t.Fatalf("expected ErrEERequired, got %v", err)
	}

	if err := p.HealthCheck(ctx); err != api.ErrEERequired {
		t.Errorf("expected ErrEERequired, got %v", err)
	}
}

func TestGRPCPredictor_Timeout(t *testing.T) {
	p := NewGRPCPredictor("localhost:50051")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	time.Sleep(2 * time.Millisecond)
	defer cancel()

	_, err := p.Predict(ctx, api.PredictInput{ModelID: "test-model"})
	// The stub implementation might return context.DeadlineExceeded or ErrSidecarTimeout
	if err != api.ErrSidecarTimeout && err != context.DeadlineExceeded {
		t.Fatalf("expected timeout error, got %v", err)
	}
}
