package hitl

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrRequestNotFound  = errors.New("approval request not found")
	ErrAlreadyProcessed = errors.New("approval request already processed")
	ErrEERequired       = errors.New("enterprise edition required")
)

type ApprovalRequest struct {
	ID        string
	AgentID   string
	StepName  string
	Payload   map[string]any
	Status    string
	CreatedAt time.Time
}

type ApprovalQueue struct {
	mu       sync.RWMutex
	requests map[string]*ApprovalRequest
}

func NewApprovalQueue() *ApprovalQueue {
	return &ApprovalQueue{
		requests: make(map[string]*ApprovalRequest),
	}
}

func (q *ApprovalQueue) Submit(req ApprovalRequest) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	req.Status = "PENDING"
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}
	q.requests[req.ID] = &req
	return nil
}

func (q *ApprovalQueue) Approve(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	req, ok := q.requests[id]
	if !ok {
		return ErrRequestNotFound
	}
	if req.Status != "PENDING" {
		return ErrAlreadyProcessed
	}
	req.Status = "APPROVED"
	return nil
}

func (q *ApprovalQueue) Reject(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	req, ok := q.requests[id]
	if !ok {
		return ErrRequestNotFound
	}
	if req.Status != "PENDING" {
		return ErrAlreadyProcessed
	}
	req.Status = "REJECTED"
	return nil
}

func (q *ApprovalQueue) ListPending() []ApprovalRequest {
	q.mu.RLock()
	defer q.mu.RUnlock()
	var pending []ApprovalRequest
	for _, req := range q.requests {
		if req.Status == "PENDING" {
			pending = append(pending, *req)
		}
	}
	return pending
}
