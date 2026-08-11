package tenant

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

func TestOSSEnforcer(t *testing.T) {
	enforcer := NewEnforcer()
	tenantID := "tenant-1"
	
	enforcer.SetQuota(tenantID, Quota{
		MaxRequestsPerMin:   2,
		MaxConcurrentAgents: 1,
		MaxTokensPerDay:     100,
		MaxCostUSDPerDay:    1.0,
	})

	ec := execctx.New(context.Background(), tenantID, "agent-1", "user-1")

	// Test Enforce
	if err := enforcer.Enforce(ec); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test rate limit on concurrent agents
	ec2 := execctx.New(context.Background(), tenantID, "agent-2", "user-1")
	if err := enforcer.Enforce(ec2); err != ErrTenantRateLimited {
		t.Fatalf("expected rate limited error, got %v", err)
	}

	// Release agent
	enforcer.ReleaseAgent(ec)

	// Test request rate limit
	if err := enforcer.Enforce(ec); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Third request should be rate limited by requests per min
	if err := enforcer.Enforce(ec); err != ErrTenantRateLimited {
		t.Fatalf("expected rate limited error on request count, got %v", err)
	}

	enforcer.ReleaseAgent(ec)

	// Test usage recording
	if err := enforcer.RecordUsage(ec, 50, 0.5); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := enforcer.RecordUsage(ec, 60, 0.6); err != ErrTenantQuotaExceeded {
		t.Fatalf("expected quota exceeded, got %v", err)
	}
}
