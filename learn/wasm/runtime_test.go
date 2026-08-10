package wasm

import (
	"context"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/learn/api"
)

func TestWASMPredictor_OSSStub(t *testing.T) {
	p := NewWASMPredictor()
	ctx := context.Background()

	_, err := p.Predict(ctx, api.PredictInput{ModelID: "test-model"})
	if err != api.ErrEERequired {
		t.Fatalf("expected ErrEERequired, got %v", err)
	}

	if err := p.HealthCheck(ctx); err != api.ErrEERequired {
		t.Fatalf("expected ErrEERequired for HealthCheck, got %v", err)
	}
}

func TestWASMPredictor_Timeout(t *testing.T) {
	p := NewWASMPredictor()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	time.Sleep(2 * time.Millisecond)
	defer cancel()

	_, err := p.Predict(ctx, api.PredictInput{ModelID: "test-model"})
	if err != api.ErrSidecarTimeout && err != context.DeadlineExceeded {
		t.Fatalf("expected timeout error, got %v", err)
	}
}
