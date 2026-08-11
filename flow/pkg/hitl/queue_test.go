package hitl

import (
	"testing"
)

func TestApprovalQueue_SubmitAndListPending(t *testing.T) {
	q := NewApprovalQueue()
	req := ApprovalRequest{
		ID:       "req-1",
		AgentID:  "agent-1",
		StepName: "step-1",
	}
	
	err := q.Submit(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pending := q.ListPending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending request, got %d", len(pending))
	}
	if pending[0].ID != "req-1" {
		t.Fatalf("expected pending request ID 'req-1', got '%s'", pending[0].ID)
	}
}

func TestApprovalQueue_Approve(t *testing.T) {
	q := NewApprovalQueue()
	req := ApprovalRequest{ID: "req-1"}
	q.Submit(req)

	err := q.Approve("req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pending := q.ListPending()
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending requests, got %d", len(pending))
	}

	err = q.Approve("req-1")
	if err != ErrAlreadyProcessed {
		t.Fatalf("expected ErrAlreadyProcessed, got %v", err)
	}
}

func TestApprovalQueue_Reject(t *testing.T) {
	q := NewApprovalQueue()
	req := ApprovalRequest{ID: "req-2"}
	q.Submit(req)

	err := q.Reject("req-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pending := q.ListPending()
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending requests, got %d", len(pending))
	}

	err = q.Reject("req-2")
	if err != ErrAlreadyProcessed {
		t.Fatalf("expected ErrAlreadyProcessed, got %v", err)
	}
}

func TestApprovalQueue_NotFound(t *testing.T) {
	q := NewApprovalQueue()
	
	err := q.Approve("req-missing")
	if err != ErrRequestNotFound {
		t.Fatalf("expected ErrRequestNotFound, got %v", err)
	}

	err = q.Reject("req-missing")
	if err != ErrRequestNotFound {
		t.Fatalf("expected ErrRequestNotFound, got %v", err)
	}
}

func TestApprovalQueue_OSSNotifications(t *testing.T) {
	q := NewApprovalQueue()
	req := ApprovalRequest{ID: "req-1"}
	
	err := q.NotifySlack(req)
	if err != ErrEERequired {
		t.Fatalf("expected ErrEERequired, got %v", err)
	}

	err = q.NotifyTeams(req)
	if err != ErrEERequired {
		t.Fatalf("expected ErrEERequired, got %v", err)
	}
}
