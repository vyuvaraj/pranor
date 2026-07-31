package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vyuvaraj/pranor/flow/pkg/storage"
)

func TestDeadLetterQueue_EnqueueAndRetry(t *testing.T) {
	mockStore := NewMockStore()
	wfDef := storage.WorkflowDef{
		ID: "failed-wf",
		Tasks: []storage.Task{
			{Name: "step-1", Action: "mock-success"},
		},
	}
	mockStore.defs["failed-wf"] = wfDef

	inst := &storage.WorkflowInstance{
		ID:         "inst-fail-1",
		WorkflowID: "failed-wf",
		Status:     "failed",
		TaskStates: map[string]*storage.TaskStatus{
			"step-1": {Name: "step-1", Status: "failed", Error: "network failure"},
		},
	}
	mockStore.insts["inst-fail-1"] = inst

	dlq := NewDeadLetterQueue(mockStore)

	// 1. Enqueue failed instance
	item := dlq.EnqueueFailedInstance(inst, "step-1", "network failure")
	if item.WorkflowID != "failed-wf" || item.FailedTask != "step-1" {
		t.Fatalf("unexpected DLQ item: %+v", item)
	}

	items := dlq.GetItems()
	if len(items) != 1 {
		t.Fatalf("expected 1 DLQ item, got %d", len(items))
	}

	// 2. HTTP Handler List
	handler := dlq.HTTPHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dlq", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	// 3. HTTP Retry Endpoint
	retryReq := httptest.NewRequest(http.MethodPost, "/api/v1/dlq/retry?id="+item.ID, nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, retryReq)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 for retry, got %d body=%s", rw.Code, rw.Body.String())
	}

	var resp map[string]bool
	_ = json.NewDecoder(rw.Body).Decode(&resp)
	if !resp["success"] {
		t.Errorf("expected success=true in retry response")
	}
}
