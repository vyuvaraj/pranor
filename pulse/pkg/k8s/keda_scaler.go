//go:build !enterprise

package import (
	"context"
	"fmt"
)

type KEDAScaler struct {
	TargetTopic        string
	TargetLagThreshold int64
}

type ScalerMetricValue struct {
	MetricName  string `json:"metric_name"`
	MetricValue int64  `json:"metric_value"`
}

func NewKEDAScaler(targetTopic string, targetLagThreshold int64) *KEDAScaler {
	if targetLagThreshold <= 0 {
		targetLagThreshold = 100
	}
	return &KEDAScaler{
		TargetTopic:        targetTopic,
		TargetLagThreshold: targetLagThreshold,
	}
}

func (s *KEDAScaler) IsActive(ctx context.Context, currentLag int64) (bool, error) {
	return currentLag > 0, nil
}

func (s *KEDAScaler) GetMetricSpec() ScalerMetricValue {
	return ScalerMetricValue{
		MetricName:  fmt.Sprintf("pranorPulse_lag_%s", s.TargetTopic),
		MetricValue: s.TargetLagThreshold,
	}
}

func (s *KEDAScaler) GetMetrics(ctx context.Context, currentLag int64) (ScalerMetricValue, error) {
	return ScalerMetricValue{
		MetricName:  fmt.Sprintf("pranorPulse_lag_%s", s.TargetTopic),
		MetricValue: currentLag,
	}, nil
}
