package import (
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/core"
	"github.com/vyuvaraj/pranor/flow/pkg/storage"
)

type MockStore struct {
	defs  map[string]storage.WorkflowDef
	insts map[string]*storage.WorkflowInstance
}

func NewMockStore() *MockStore {
	return &MockStore{
		defs:  make(map[string]storage.WorkflowDef),
		insts: make(map[string]*storage.WorkflowInstance),
	}
}

func (m *MockStore) LoadDefinitions() (map[string]storage.WorkflowDef, error) {
	return m.defs, nil
}
func (m *MockStore) SaveDefinitions(defs map[string]storage.WorkflowDef) error {
	m.defs = defs
	return nil
}
func (m *MockStore) LoadInstances() (map[string]*storage.WorkflowInstance, error) {
	return m.insts, nil
}
func (m *MockStore) SaveInstances(insts map[string]*storage.WorkflowInstance) error {
	m.insts = insts
	return nil
}
func (m *MockStore) GetClient() *Pranor Core.StoreClient { return nil }

func TestSubWorkflowManager_Invoke(t *testing.T) {
	mockStore := NewMockStore()

	childDef := storage.WorkflowDef{
		ID: "child-flow",
		Tasks: []storage.Task{
			{Name: "child-step-1", Action: "mock-success"},
		},
	}
	mockStore.defs["child-flow"] = childDef

	swm := NewSubWorkflowManager(mockStore)

	link, err := swm.InvokeSubWorkflow("parent-inst-1", "subflow-task", "child-flow")
	if err != nil {
		t.Fatalf("InvokeSubWorkflow failed: %v", err)
	}

	if link.ParentInstanceID != "parent-inst-1" || link.ChildWorkflowID != "child-flow" {
		t.Errorf("unexpected subworkflow link: %+v", link)
	}

	// Wait for async completion
	time.Sleep(100 * time.Millisecond)

	retrieved, found := swm.GetSubWorkflowLink(link.ChildInstanceID)
	if !found {
		t.Fatalf("subworkflow link not found by ID %s", link.ChildInstanceID)
	}
	if retrieved.Status != "completed" {
		t.Errorf("expected subworkflow status 'completed', got %s", retrieved.Status)
	}
}

func TestSubWorkflowManager_NotFound(t *testing.T) {
	mockStore := NewMockStore()
	swm := NewSubWorkflowManager(mockStore)

	_, err := swm.InvokeSubWorkflow("parent-1", "task-1", "non-existent-child")
	if err == nil {
		t.Error("expected error for non-existent child workflow definition")
	}
}
