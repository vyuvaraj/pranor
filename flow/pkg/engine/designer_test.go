package engine

import (
	"testing"

	"github.com/vyuvaraj/pranor/flow/pkg/storage"
)

func TestDesignerWorkflowCanvas_ExportAndImport(t *testing.T) {
	canvas := DesignerWorkflowCanvas{
		WorkflowID: "wf-order-process",
		Name:       "Order Processing Workflow",
		Nodes: []DesignerNode{
			{
				ID:       "payment",
				Type:     "http_task",
				Label:    "Process Payment",
				Position: DesignerNodePosition{X: 100, Y: 100},
				Config:   map[string]interface{}{"target_url": "http://payment-service/charge"},
			},
			{
				ID:       "fulfillment",
				Type:     "http_task",
				Label:    "Ship Order",
				Position: DesignerNodePosition{X: 350, Y: 100},
				Config:   map[string]interface{}{"target_url": "http://ship-service/pack"},
			},
		},
		Edges: []DesignerEdge{
			{ID: "e1", Source: "payment", Target: "fulfillment", Condition: "success"},
		},
	}

	// 1. Export canvas to storage.WorkflowDef
	def, err := canvas.ExportToWorkflowDef()
	if err != nil {
		t.Fatalf("ExportToWorkflowDef failed: %v", err)
	}

	if def.ID != "wf-order-process" || len(def.Tasks) != 2 {
		t.Fatalf("unexpected exported WorkflowDef: %+v", def)
	}

	// Verify fulfillment depends on payment
	var fulfillmentTask *storage.Task
	for i := range def.Tasks {
		if def.Tasks[i].Name == "fulfillment" {
			fulfillmentTask = &def.Tasks[i]
		}
	}
	if fulfillmentTask == nil || len(fulfillmentTask.DependsOn) != 1 || fulfillmentTask.DependsOn[0] != "payment" {
		t.Errorf("expected fulfillment to depend on payment, got: %+v", fulfillmentTask)
	}

	// 2. Import back from WorkflowDef to canvas
	importedCanvas := ImportFromWorkflowDef(def)
	if importedCanvas.WorkflowID != "wf-order-process" || len(importedCanvas.Nodes) != 2 || len(importedCanvas.Edges) != 1 {
		t.Errorf("imported canvas mismatch: %+v", importedCanvas)
	}
}

func TestDesignerWorkflowCanvas_CycleDetectionOnExport(t *testing.T) {
	// Create cyclic canvas: A -> B -> A
	cyclicCanvas := DesignerWorkflowCanvas{
		WorkflowID: "wf-cycle",
		Name:       "Cycle Workflow",
		Nodes: []DesignerNode{
			{ID: "nodeA", Position: DesignerNodePosition{X: 100, Y: 100}},
			{ID: "nodeB", Position: DesignerNodePosition{X: 300, Y: 100}},
		},
		Edges: []DesignerEdge{
			{ID: "e1", Source: "nodeA", Target: "nodeB"},
			{ID: "e2", Source: "nodeB", Target: "nodeA"},
		},
	}

	_, err := cyclicCanvas.ExportToWorkflowDef()
	if err == nil {
		t.Error("expected error exporting workflow canvas with a dependency cycle")
	}
}
