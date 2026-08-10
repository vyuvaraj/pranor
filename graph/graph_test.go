package graph

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/graph/api"
	graphctx "github.com/vyuvaraj/pranor/graph/pkg/context"
)

func TestQuery_NilProvider(t *testing.T) {
	orig := DefaultProvider
	DefaultProvider = nil
	defer func() { DefaultProvider = orig }()

	_, err := Query(context.Background(), api.ContextQuery{})
	if err != api.ErrGraphContextUnavailable {
		t.Errorf("expected ErrGraphContextUnavailable, got %v", err)
	}

	err = HealthCheck(context.Background())
	if err != api.ErrGraphContextUnavailable {
		t.Errorf("expected ErrGraphContextUnavailable for HealthCheck, got %v", err)
	}
}

func TestContextQueryFields(t *testing.T) {
	q := api.ContextQuery{
		EntityID:      "e1",
		TenantID:      "t1",
		AgentID:       "a1",
		UserID:        "u1",
		RequestedTier: api.TierWarm,
		MaxAgeSecs:    60,
	}
	if q.EntityID != "e1" {
		t.Errorf("expected e1")
	}
}

func TestThreeTierAssembler_HotHit(t *testing.T) {
	assembler := graphctx.NewThreeTierAssembler()
	assembler.SetHot("e1", "t1", api.ContextResult{
		EntityID: "e1",
		TenantID: "t1",
		Data:     map[string]any{"foo": "bar"},
	})

	q := api.ContextQuery{
		EntityID:      "e1",
		TenantID:      "t1",
		RequestedTier: api.TierHot,
	}

	res, err := assembler.Assemble(context.Background(), q)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if !res.CacheHit {
		t.Errorf("expected cache hit")
	}
	if res.Tier != api.TierHot {
		t.Errorf("expected TierHot")
	}
}

func TestThreeTierAssembler_WarmFallback(t *testing.T) {
	assembler := graphctx.NewThreeTierAssembler()
	q := api.ContextQuery{
		EntityID:      "e1",
		TenantID:      "t1",
		RequestedTier: api.TierWarm,
	}

	res, err := assembler.Assemble(context.Background(), q)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if res.Tier != api.TierWarm {
		t.Errorf("expected TierWarm")
	}
}

func TestThreeTierAssembler_ColdExhausted(t *testing.T) {
	assembler := graphctx.NewThreeTierAssembler()
	q := api.ContextQuery{
		EntityID:      "e1",
		TenantID:      "t1",
		RequestedTier: api.TierCold,
	}

	_, err := assembler.Assemble(context.Background(), q)
	if err != api.ErrGraphContextUnavailable {
		t.Errorf("expected ErrGraphContextUnavailable, got %v", err)
	}
}
