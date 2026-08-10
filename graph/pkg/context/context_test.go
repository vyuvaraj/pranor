package graphctx

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/graph/api"
)

func TestThreeTierAssembler_FallbackFailClosed(t *testing.T) {
	assembler := NewThreeTierAssembler()
	ctx := context.Background()

	q := api.ContextQuery{
		EntityID: "ent123",
		TenantID: "ten456",
	}

	// With empty cache, it should miss hot, warm, and cold, then return ErrGraphContextUnavailable
	res, err := assembler.Assemble(ctx, q)
	if err != api.ErrGraphContextUnavailable {
		t.Fatalf("expected ErrGraphContextUnavailable, got %v", err)
	}
	if res.EntityID != "" {
		t.Fatalf("expected empty result, got %v", res)
	}

	spans := assembler.GetMissSpanLog()
	if len(spans) != 3 {
		t.Fatalf("expected 3 miss spans, got %d", len(spans))
	}
	if spans[0] != "hot_cache_miss:ent123:ten456" {
		t.Errorf("unexpected span 0: %s", spans[0])
	}
	if spans[1] != "warm_cache_miss:ent123:ten456" {
		t.Errorf("unexpected span 1: %s", spans[1])
	}
	if spans[2] != "cold_cache_miss:ent123:ten456" {
		t.Errorf("unexpected span 2: %s", spans[2])
	}
}

func TestThreeTierAssembler_HotCacheHit(t *testing.T) {
	assembler := NewThreeTierAssembler()
	ctx := context.Background()

	q := api.ContextQuery{
		EntityID: "ent123",
		TenantID: "ten456",
	}

	assembler.SetHot("ent123", "ten456", api.ContextResult{
		EntityID: "ent123",
		TenantID: "ten456",
		Data:     map[string]any{"foo": "bar"},
	})

	res, err := assembler.Assemble(ctx, q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.CacheHit || res.Tier != api.TierHot {
		t.Fatalf("expected hot cache hit, got %v", res)
	}
	
	spans := assembler.GetMissSpanLog()
	if len(spans) != 0 {
		t.Fatalf("expected 0 miss spans on hot hit, got %d", len(spans))
	}
}
