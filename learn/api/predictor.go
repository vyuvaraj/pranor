package api

import (
	"context"
	"errors"
)

var (
	ErrSidecarTimeout      = errors.New("sidecar timeout")
	ErrModelBudgetExceeded = errors.New("model budget exceeded")
	ErrEERequired          = errors.New("enterprise edition required")
	ErrInvalidModelInput   = errors.New("invalid model input")
)

type PredictInput struct {
	ModelID  string
	Inputs   map[string]any
	BudgetMs int64
}

type PredictOutput struct {
	Predictions map[string]float64
	Confidence  float64
	Provider    string
	LatencyMs   int64
}

type Predictor interface {
	Predict(ctx context.Context, input PredictInput) (PredictOutput, error)
	HealthCheck(ctx context.Context) error
}
