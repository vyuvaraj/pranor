package import (
	"fmt"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/flow/pkg/storage"
)

// SubWorkflowLink tracks a parent-child workflow execution hierarchy.
type SubWorkflowLink struct {
	ParentInstanceID string `json:"parent_instance_id"`
	ParentTaskName   string `json:"parent_task_name"`
	ChildWorkflowID  string `json:"child_workflow_id"`
	ChildInstanceID  string `json:"child_instance_id"`
	Status           string `json:"status"` // "running", "completed", "failed"
}

// SubWorkflowManager handles nested workflow invocation and status propagation.
type SubWorkflowManager struct {
	mu    sync.RWMutex
	links map[string]*SubWorkflowLink // childInstanceID -> link
	store storage.WorkflowStore
}

// NewSubWorkflowManager initializes a sub-workflow manager.
func NewSubWorkflowManager(store storage.WorkflowStore) *SubWorkflowManager {
	return &SubWorkflowManager{
		links: make(map[string]*SubWorkflowLink),
		store: store,
	}
}

// InvokeSubWorkflow triggers a child workflow execution on behalf of a parent task step.
func (swm *SubWorkflowManager) InvokeSubWorkflow(parentInstanceID, parentTaskName, childWorkflowID string) (*SubWorkflowLink, error) {
	if parentInstanceID == "" || childWorkflowID == "" {
		return nil, fmt.Errorf("parentInstanceID and childWorkflowID are required")
	}

	defs, err := swm.store.LoadDefinitions()
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow definitions: %w", err)
	}

	childDef, ok := defs[childWorkflowID]
	if !ok {
		return nil, fmt.Errorf("child workflow definition '%s' not found", childWorkflowID)
	}

	childInstID := fmt.Sprintf("sub-%s-%d", childWorkflowID, time.Now().UnixNano())
	link := &SubWorkflowLink{
		ParentInstanceID: parentInstanceID,
		ParentTaskName:   parentTaskName,
		ChildWorkflowID:  childWorkflowID,
		ChildInstanceID:  childInstID,
		Status:           "running",
	}

	swm.mu.Lock()
	swm.links[childInstID] = link
	swm.mu.Unlock()

	// Execute child workflow instance
	childInst := &storage.WorkflowInstance{
		ID:         childInstID,
		WorkflowID: childWorkflowID,
		Status:     "running",
		TaskStates: make(map[string]*storage.TaskStatus),
		StartedAt:  time.Now(),
	}
	for _, task := range childDef.Tasks {
		childInst.TaskStates[task.Name] = &storage.TaskStatus{
			Name:   task.Name,
			Status: "pending",
		}
	}

	insts, _ := swm.store.LoadInstances()
	if insts == nil {
		insts = make(map[string]*storage.WorkflowInstance)
	}
	insts[childInstID] = childInst
	_ = swm.store.SaveInstances(insts)

	go func() {
		var mu sync.RWMutex
		RunWorkflow(childInst, childDef, swm.store, insts, &mu)
		swm.mu.Lock()
		if childInst.Status == "failed" {
			link.Status = "failed"
		} else {
			link.Status = "completed"
		}
		swm.mu.Unlock()
	}()

	return link, nil
}

// GetSubWorkflowLink returns the status link for a child workflow instance.
func (swm *SubWorkflowManager) GetSubWorkflowLink(childInstanceID string) (*SubWorkflowLink, bool) {
	swm.mu.RLock()
	defer swm.mu.RUnlock()
	link, ok := swm.links[childInstanceID]
	return link, ok
}
