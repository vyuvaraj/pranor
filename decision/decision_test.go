package decision

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/decision/api"
	"github.com/vyuvaraj/pranor/decision/pkg/engine"
	graphapi "github.com/vyuvaraj/pranor/graph/api"
)

type mockGraphProvider struct {
	err error
}

func (m *mockGraphProvider) Query(ctx context.Context, q graphapi.ContextQuery) (graphapi.ContextResult, error) {
	return graphapi.ContextResult{}, m.err
}

func (m *mockGraphProvider) Invalidate(ctx context.Context, entityID, tenantID string) error {
	return nil
}

func (m *mockGraphProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func TestDecisionEngine_GraphContextUnavailable(t *testing.T) {
	eng := engine.NewVetoLadderEngine(&mockGraphProvider{err: graphapi.ErrGraphContextUnavailable})
	req := api.DecisionRequest{}
	res, err := eng.Evaluate(context.Background(), req)
	if err != api.ErrContextUnavailable {
		t.Fatalf("expected ErrContextUnavailable, got %v", err)
	}
	if res.Action != api.ActionDeny {
		t.Fatalf("expected DENY action, got %v", res.Action)
	}
	if res.PriorityLevel != api.PriorityAuth {
		t.Fatalf("expected priority %d, got %d", api.PriorityAuth, res.PriorityLevel)
	}
}

func TestDecisionEngine_AuthHardDeny(t *testing.T) {
	eng := engine.NewVetoLadderEngine(&mockGraphProvider{})
	req := api.DecisionRequest{Capability: "FORBIDDEN_AUTH"}
	res, err := eng.Evaluate(context.Background(), req)
	if err != api.ErrDecisionDenied {
		t.Fatalf("expected ErrDecisionDenied, got %v", err)
	}
	if res.Action != api.ActionDeny {
		t.Fatalf("expected DENY action, got %v", res.Action)
	}
	if res.PriorityLevel != api.PriorityAuth {
		t.Fatalf("expected PriorityAuth, got %d", res.PriorityLevel)
	}
}

func TestDecisionEngine_BudgetHardDeny(t *testing.T) {
	eng := engine.NewVetoLadderEngine(&mockGraphProvider{})
	req := api.DecisionRequest{Capability: "EXCEEDS_BUDGET"}
	res, err := eng.Evaluate(context.Background(), req)
	if err != api.ErrDecisionDenied {
		t.Fatalf("expected ErrDecisionDenied, got %v", err)
	}
	if res.Action != api.ActionDeny {
		t.Fatalf("expected DENY action, got %v", res.Action)
	}
	if res.PriorityLevel != api.PriorityBudget {
		t.Fatalf("expected PriorityBudget, got %d", res.PriorityLevel)
	}
}

func TestDecisionEngine_DefaultApprove(t *testing.T) {
	eng := engine.NewVetoLadderEngine(&mockGraphProvider{})
	req := api.DecisionRequest{Capability: "ALLOWED"}
	res, err := eng.Evaluate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Action != api.ActionApprove {
		t.Fatalf("expected APPROVE action, got %v", res.Action)
	}
	if res.PriorityLevel != api.PriorityDefault {
		t.Fatalf("expected PriorityDefault, got %d", res.PriorityLevel)
	}
}

func TestDecisionEngine_LearnSoftAdvisory(t *testing.T) {
	// Learn is priority 5. If it fails, it's skipped. Since our OSS stub
	// skips it anyway, let's just make sure it passes through to Default.
	eng := engine.NewVetoLadderEngine(&mockGraphProvider{})
	req := api.DecisionRequest{Capability: "LEARN_ADVISORY"}
	res, err := eng.Evaluate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Action != api.ActionApprove {
		t.Fatalf("expected APPROVE action, got %v", res.Action)
	}
	if res.PriorityLevel != api.PriorityDefault {
		t.Fatalf("expected PriorityDefault, got %d", res.PriorityLevel)
	}
}
