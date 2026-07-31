package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/flow/pkg/storage"
)

// DLQItem captures permanently failed workflow execution context.
type DLQItem struct {
	ID           string                         `json:"id"`
	WorkflowID   string                         `json:"workflow_id"`
	InstanceID   string                         `json:"instance_id"`
	FailedTask   string                         `json:"failed_task"`
	ErrorMessage string                         `json:"error_message"`
	FailedAt     time.Time                      `json:"failed_at"`
	TaskStates   map[string]*storage.TaskStatus `json:"task_states"`
	RetryCount   int                            `json:"retry_count"`
}

// DeadLetterQueue manages failed workflow instances for operator inspection and manual retry.
type DeadLetterQueue struct {
	mu    sync.RWMutex
	items map[string]*DLQItem
	store storage.WorkflowStore
}

// NewDeadLetterQueue creates a new DeadLetterQueue instance.
func NewDeadLetterQueue(store storage.WorkflowStore) *DeadLetterQueue {
	return &DeadLetterQueue{
		items: make(map[string]*DLQItem),
		store: store,
	}
}

// EnqueueFailedInstance records a failed workflow instance into the DLQ.
func (dlq *DeadLetterQueue) EnqueueFailedInstance(inst *storage.WorkflowInstance, failedTask string, errStr string) *DLQItem {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	item := &DLQItem{
		ID:           fmt.Sprintf("dlq-%s-%d", inst.ID, time.Now().UnixNano()),
		WorkflowID:   inst.WorkflowID,
		InstanceID:   inst.ID,
		FailedTask:   failedTask,
		ErrorMessage: errStr,
		FailedAt:     time.Now(),
		TaskStates:   inst.TaskStates,
		RetryCount:   0,
	}
	dlq.items[item.ID] = item
	return item
}

// GetItems returns all dead-lettered workflow items.
func (dlq *DeadLetterQueue) GetItems() []*DLQItem {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	list := make([]*DLQItem, 0, len(dlq.items))
	for _, item := range dlq.items {
		list = append(list, item)
	}
	return list
}

// RetryItem triggers manual re-execution of a dead-lettered workflow from the failed task step.
func (dlq *DeadLetterQueue) RetryItem(dlqID string) error {
	dlq.mu.Lock()
	item, ok := dlq.items[dlqID]
	if !ok {
		dlq.mu.Unlock()
		return fmt.Errorf("dlq item '%s' not found", dlqID)
	}
	item.RetryCount++
	dlq.mu.Unlock()

	insts, err := dlq.store.LoadInstances()
	if err != nil || insts == nil {
		return fmt.Errorf("failed to load instances: %w", err)
	}

	inst, exists := insts[item.InstanceID]
	if !exists {
		return fmt.Errorf("instance '%s' not found", item.InstanceID)
	}

	defs, err := dlq.store.LoadDefinitions()
	if err != nil {
		return fmt.Errorf("failed to load definitions: %w", err)
	}
	wfDef, exists := defs[item.WorkflowID]
	if !exists {
		return fmt.Errorf("definition '%s' not found", item.WorkflowID)
	}

	inst.Mu.Lock()
	if ts, ok := inst.TaskStates[item.FailedTask]; ok {
		ts.Status = "pending"
		ts.Error = ""
	}
	inst.Status = "running"
	inst.Mu.Unlock()
	_ = dlq.store.SaveInstances(insts)

	go func() {
		var mu sync.RWMutex
		RunWorkflow(inst, wfDef, dlq.store, insts, &mu)
	}()

	return nil
}

// HTTPHandler exposes REST endpoints for Pranor Console DLQ inspection and manual retry.
func (dlq *DeadLetterQueue) HTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/dlq", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(dlq.GetItems())
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/v1/dlq/retry", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		dlqID := r.URL.Query().Get("id")
		if dlqID == "" {
			http.Error(w, "query param 'id' is required", http.StatusBadRequest)
			return
		}

		if err := dlq.RetryItem(dlqID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})

	return mux
}
