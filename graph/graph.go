package graph

import (
	"context"

	"github.com/vyuvaraj/pranor/graph/api"
)

var DefaultProvider api.GraphProvider

func Query(ctx context.Context, q api.ContextQuery) (api.ContextResult, error) {
	if DefaultProvider == nil {
		return api.ContextResult{}, api.ErrGraphContextUnavailable
	}
	return DefaultProvider.Query(ctx, q)
}

func HealthCheck(ctx context.Context) error {
	if DefaultProvider == nil {
		return api.ErrGraphContextUnavailable
	}
	return DefaultProvider.HealthCheck(ctx)
}
