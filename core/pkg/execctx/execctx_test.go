package execctx

import (
	"context"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	ec := New(ctx, "t1", "a1", "u1")
	if ec.TenantID != "t1" || ec.AgentID != "a1" || ec.UserID != "u1" {
		t.Errorf("New failed to set fields correctly")
	}
	if ec.Capabilities == nil || ec.PolicyContext == nil || ec.Metadata == nil {
		t.Errorf("New failed to initialize maps/slices")
	}
	if ec.CreatedAt.IsZero() {
		t.Errorf("New failed to set CreatedAt")
	}
}

func TestFromHTTP_OK(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderTenantID, "t1")
	req.Header.Set(HeaderAgentID, "a1")
	req.Header.Set(HeaderUserID, "u1")
	req.Header.Set(HeaderTraceID, "trace1")
	req.Header.Set(HeaderRequestID, "req1")
	req.Header.Set(HeaderParentAgentID, "pa1")

	ec, err := FromHTTP(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ec.TenantID != "t1" || ec.AgentID != "a1" || ec.UserID != "u1" || ec.TraceID != "trace1" || ec.RequestID != "req1" || ec.ParentAgentID != "pa1" {
		t.Errorf("FromHTTP did not extract fields correctly")
	}
}

func TestFromHTTP_MissingTenant(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	_, err := FromHTTP(context.Background(), req)
	if err != ErrMissingTenantID {
		t.Errorf("expected ErrMissingTenantID, got %v", err)
	}
}

func TestWithAgent(t *testing.T) {
	ec := New(context.Background(), "t1", "a1", "u1")
	ec2 := ec.WithAgent("a2")
	
	if ec2.AgentID != "a2" {
		t.Errorf("expected AgentID a2, got %s", ec2.AgentID)
	}
	if ec2.ParentAgentID != "a1" {
		t.Errorf("expected ParentAgentID a1, got %s", ec2.ParentAgentID)
	}
	// Original unchanged
	if ec.AgentID != "a1" || ec.ParentAgentID != "" {
		t.Errorf("original context modified")
	}
}

func TestWithCapability(t *testing.T) {
	ec := New(context.Background(), "t1", "a1", "u1")
	ec2 := ec.WithCapability("cap1")
	
	if len(ec2.Capabilities) != 1 || ec2.Capabilities[0] != "cap1" {
		t.Errorf("capability not appended correctly")
	}
	if len(ec.Capabilities) != 0 {
		t.Errorf("original context modified")
	}
}

func TestWithPolicy(t *testing.T) {
	ec := New(context.Background(), "t1", "a1", "u1")
	ec2 := ec.WithPolicy("k1", "v1")
	
	if ec2.PolicyContext["k1"] != "v1" {
		t.Errorf("policy not set correctly")
	}
	if _, ok := ec.PolicyContext["k1"]; ok {
		t.Errorf("original context modified")
	}
}

func TestWithBudget(t *testing.T) {
	ec := New(context.Background(), "t1", "a1", "u1")
	ec2 := ec.WithBudget(0.5, 100, 10.5)
	
	if ec2.RiskBudget != 0.5 || ec2.TokenBudget != 100 || ec2.CostBudgetUS != 10.5 {
		t.Errorf("budget not set correctly")
	}
	if ec.RiskBudget != 0 || ec.TokenBudget != 0 || ec.CostBudgetUS != 0 {
		t.Errorf("original context modified")
	}
}

func TestValidate_Empty(t *testing.T) {
	ec := &ExecutionContext{}
	err := ec.Validate()
	if err != ErrMissingTenantID {
		t.Errorf("expected ErrMissingTenantID, got %v", err)
	}
}

func TestValidate_OK(t *testing.T) {
	ec := &ExecutionContext{TenantID: "t1"}
	err := ec.Validate()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestHasCapability(t *testing.T) {
	ec := New(context.Background(), "t1", "a1", "u1").WithCapability("cap1")
	if !ec.HasCapability("cap1") {
		t.Errorf("expected true for cap1")
	}
	if ec.HasCapability("cap2") {
		t.Errorf("expected false for cap2")
	}
}

func TestInjectHTTP(t *testing.T) {
	ec := New(context.Background(), "t1", "a1", "u1")
	ec.TraceID = "trace1"
	ec.RequestID = "req1"
	ec.ParentAgentID = "pa1"

	req, _ := http.NewRequest("GET", "/", nil)
	ec.InjectHTTP(req)

	if req.Header.Get(HeaderTenantID) != "t1" ||
		req.Header.Get(HeaderAgentID) != "a1" ||
		req.Header.Get(HeaderUserID) != "u1" ||
		req.Header.Get(HeaderTraceID) != "trace1" ||
		req.Header.Get(HeaderRequestID) != "req1" ||
		req.Header.Get(HeaderParentAgentID) != "pa1" {
		t.Errorf("headers not injected correctly")
	}
}
