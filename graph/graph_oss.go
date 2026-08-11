//go:build !enterprise

package graph

import (
	"context"

	"github.com/vyuvaraj/pranor/graph/api"
	graphctx "github.com/vyuvaraj/pranor/graph/pkg/context"
)

type ossGraphProvider struct {
	assembler *graphctx.ThreeTierAssembler
}

func init() {
	DefaultProvider = &ossGraphProvider{
		assembler: graphctx.NewThreeTierAssembler(),
	}
}

func (p *ossGraphProvider) Query(ctx context.Context, q api.ContextQuery) (api.ContextResult, error) {
	return p.assembler.Assemble(ctx, q)
}

func (p *ossGraphProvider) Invalidate(ctx context.Context, entityID, tenantID string) error {
	return nil
}

func (p *ossGraphProvider) HealthCheck(ctx context.Context) error {
	return nil
}
