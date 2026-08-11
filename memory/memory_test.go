package memory_test

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
	"github.com/vyuvaraj/pranor/memory"
)

func TestWorkingMemory_CRUD(t *testing.T) {
	ctx := context.Background()
	ec := execctx.New(ctx, "tenant-1", "agent-1", "user-1")
	sessionID := "sess-100"

	wm := memory.Working()

	// Set
	err := wm.Set(ctx, ec, sessionID, "scratch_note", "process user order #402")
	if err != nil {
		t.Fatalf("unexpected error on Set: %v", err)
	}

	// Get
	val, found, err := wm.Get(ctx, ec, sessionID, "scratch_note")
	if err != nil || !found {
		t.Fatalf("expected key to be found, got found=%v err=%v", found, err)
	}
	if val != "process user order #402" {
		t.Errorf("expected 'process user order #402', got %v", val)
	}

	// Delete
	err = wm.Delete(ctx, ec, sessionID, "scratch_note")
	if err != nil {
		t.Fatalf("unexpected error on Delete: %v", err)
	}

	_, found, _ = wm.Get(ctx, ec, sessionID, "scratch_note")
	if found {
		t.Errorf("expected key to be deleted")
	}

	// Flush
	_ = wm.Set(ctx, ec, sessionID, "k1", "v1")
	_ = wm.Set(ctx, ec, sessionID, "k2", "v2")
	err = wm.Flush(ctx, ec, sessionID)
	if err != nil {
		t.Fatalf("unexpected error on Flush: %v", err)
	}

	_, found, _ = wm.Get(ctx, ec, sessionID, "k1")
	if found {
		t.Errorf("expected k1 to be flushed")
	}
}

func TestEpisodicMemory_StoreAndRecall(t *testing.T) {
	ctx := context.Background()
	ec := execctx.New(ctx, "tenant-1", "agent-1", "user-1")

	em := memory.Episodic()

	_, err := em.StoreEpisode(ctx, ec, "s1", "user", "User requested refund for order #12345", []string{"refund", "order"})
	if err != nil {
		t.Fatalf("unexpected error storing episode: %v", err)
	}

	_, _ = em.StoreEpisode(ctx, ec, "s1", "assistant", "Refund of $50 processed successfully", []string{"refund", "success"})
	_, _ = em.StoreEpisode(ctx, ec, "s2", "user", "User asked about shipping options", []string{"shipping"})

	// Recall query for "refund"
	results, err := em.Recall(ctx, ec, "refund", 5)
	if err != nil {
		t.Fatalf("unexpected error on Recall: %v", err)
	}

	if len(results) < 2 {
		t.Fatalf("expected at least 2 results for 'refund', got %d", len(results))
	}
}

func TestEpisodicMemory_RecallSemantic(t *testing.T) {
	ctx := context.Background()
	ec := execctx.New(ctx, "tenant-1", "agent-vector", "user-1")

	em := memory.Episodic()

	// Store episode with parallel vector [1.0, 0.0, 0.0]
	entry1, err := em.StoreEpisode(ctx, ec, "s1", "user", "Payment processing query", []string{"payment"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry1.Vector = []float32{1.0, 0.0, 0.0}

	// Store episode with orthogonal vector [0.0, 1.0, 0.0]
	entry2, _ := em.StoreEpisode(ctx, ec, "s1", "user", "UI theme configuration", []string{"ui"})
	entry2.Vector = []float32{0.0, 1.0, 0.0}

	// Query with vector close to entry1 [0.9, 0.1, 0.0]
	queryVec := []float32{0.9, 0.1, 0.0}
	results, err := em.RecallSemantic(ctx, ec, queryVec, 2)
	if err != nil {
		t.Fatalf("unexpected error on RecallSemantic: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected semantic recall results, got 0")
	}
	if results[0].Content != "Payment processing query" {
		t.Errorf("expected closest vector content 'Payment processing query', got '%s'", results[0].Content)
	}
}

func TestTenantIsolation(t *testing.T) {
	ctx := context.Background()
	ecTenantA := execctx.New(ctx, "tenant-a", "agent-1", "user-1")
	ecTenantB := execctx.New(ctx, "tenant-b", "agent-1", "user-1")

	wm := memory.Working()
	em := memory.Episodic()

	// Working memory isolation
	_ = wm.Set(ctx, ecTenantA, "sess-1", "secret", "tenant-a-data")
	_, found, _ := wm.Get(ctx, ecTenantB, "sess-1", "secret")
	if found {
		t.Errorf("tenant-b accessed tenant-a working memory!")
	}

	// Episodic memory isolation
	_, _ = em.StoreEpisode(ctx, ecTenantA, "sess-1", "user", "Confidential financial report for Tenant A", []string{"finance"})

	resultsB, err := em.Recall(ctx, ecTenantB, "financial", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resultsB) != 0 {
		t.Errorf("tenant-b recalled tenant-a episodic memory!")
	}
}
