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

	out, err := p.Predict(ctx, api.PredictInput{
		ModelID:  "test-model",
		BudgetMs: 1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Provider != "gRPC-Sidecar" {
		t.Errorf("expected provider gRPC-Sidecar, got %s", out.Provider)
	}

	if err := p.HealthCheck(ctx); err != nil {
		t.Errorf("expected healthy, got %v", err)
	}
}

func TestGRPCPredictor_Timeout(t *testing.T) {
	p := NewGRPCPredictor("localhost:50051")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	time.Sleep(2 * time.Millisecond)
	defer cancel()

	_, err := p.Predict(ctx, api.PredictInput{ModelID: "test-model"})
	if err != api.ErrSidecarTimeout && err != context.DeadlineExceeded {
		t.Fatalf("expected timeout error, got %v", err)
	}
}
