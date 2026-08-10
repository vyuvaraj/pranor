package sidecar

import (
	"context"
	"time"

	"github.com/vyuvaraj/pranor/learn/api"
)

type GRPCPredictor struct {
	Target string
}

func NewGRPCPredictor(target string) *GRPCPredictor {
	return &GRPCPredictor{Target: target}
}

func (p *GRPCPredictor) Predict(ctx context.Context, input api.PredictInput) (api.PredictOutput, error) {
	if input.BudgetMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(input.BudgetMs)*time.Millisecond)
		defer cancel()
	}

	select {
	case <-ctx.Done():
		return api.PredictOutput{}, api.ErrSidecarTimeout
	default:
		return api.PredictOutput{
			Predictions: map[string]float64{"advice_score": 0.85},
			Confidence:  0.90,
			Provider:    "gRPC-Sidecar",
			LatencyMs:   5,
		}, nil
	}
}

func (p *GRPCPredictor) HealthCheck(ctx context.Context) error {
	return nil
}
