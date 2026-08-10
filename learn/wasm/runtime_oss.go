//go:build !enterprise

package wasm

import (
	"context"
	"time"

	"github.com/vyuvaraj/pranor/learn/api"
)

func (w *WASMPredictor) Predict(ctx context.Context, input api.PredictInput) (api.PredictOutput, error) {
	if input.BudgetMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(input.BudgetMs)*time.Millisecond)
		defer cancel()
	}

	// Stub/mock logic for OSS
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return api.PredictOutput{}, api.ErrSidecarTimeout
		}
		return api.PredictOutput{}, ctx.Err()
	default:
	}

	return api.PredictOutput{}, api.ErrEERequired
}

func (w *WASMPredictor) HealthCheck(ctx context.Context) error {
	return api.ErrEERequired
}
