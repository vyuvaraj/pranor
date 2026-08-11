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

func TestContextQuery_MaxTokenBudget_PruningFIFO(t *testing.T) {
	assembler := graphctx.NewThreeTierAssembler()
	
	data := map[string]any{
		"a": "long string data that consumes some tokens ...",
		"b": "more long string data ...",
		"c": "even more string data ...",
	}
	assembler.SetHot("e1", "t1", api.ContextResult{
		EntityID: "e1",
		TenantID: "t1",
		Data:     data,
	})

	q := api.ContextQuery{
		EntityID:        "e1",
		TenantID:        "t1",
		RequestedTier:   api.TierHot,
		MaxTokenBudget:  15,
		PruningStrategy: api.PruningStrategyFIFO,
	}

	res, err := assembler.Assemble(context.Background(), q)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if res.PrunedNodeCount == 0 {
		t.Errorf("expected nodes to be pruned")
	}
	if res.TokenCount > q.MaxTokenBudget {
		t.Errorf("expected token count %d to be <= budget %d", res.TokenCount, q.MaxTokenBudget)
	}
	
	_, hasA := res.Data["a"]
	_, hasC := res.Data["c"]
	if hasA && !hasC {
		t.Errorf("expected 'a' to be pruned before 'c' (FIFO lexicographical)")
	}
}

func TestContextQuery_MaxTokenBudget_PruningSemantic(t *testing.T) {
	assembler := graphctx.NewThreeTierAssembler()
	
	data := map[string]any{
		"user_profile": "very important user profile ...",
		"debug_logs":   "some large debug logs ...",
		"history_data": "some history data ...",
		"intent_data":  "very important intent ...",
	}
	assembler.SetHot("e2", "t2", api.ContextResult{
		EntityID: "e2",
		TenantID: "t2",
		Data:     data,
	})

	q := api.ContextQuery{
		EntityID:        "e2",
		TenantID:        "t2",
		RequestedTier:   api.TierHot,
		MaxTokenBudget:  25,
		PruningStrategy: api.PruningStrategySemanticRelevance,
	}

	res, err := assembler.Assemble(context.Background(), q)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if res.PrunedNodeCount == 0 {
		t.Errorf("expected nodes to be pruned")
	}
	if res.TokenCount > q.MaxTokenBudget {
		t.Errorf("expected token count %d to be <= budget %d", res.TokenCount, q.MaxTokenBudget)
	}
	
	_, hasDebug := res.Data["debug_logs"]
	_, hasHistory := res.Data["history_data"]
	_, hasUser := res.Data["user_profile"]
	_, hasIntent := res.Data["intent_data"]
	
	if hasDebug || hasHistory {
		t.Errorf("expected low-relevance nodes (debug, history) to be pruned first")
	}
	if !hasUser || !hasIntent {
		t.Errorf("expected high-relevance nodes (user, intent) to be kept")
	}
}
