package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vyuvaraj/serv/packages/ServFlow/pkg/storage"
)

// DesignerNodePosition holds 2D canvas coordinates for visual workflow rendering.
type DesignerNodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// DesignerNode represents a single visual node on the ServConsole drag-and-drop canvas.
type DesignerNode struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"` // "http_task", "wasm_task", "human_gate", "sub_flow"
	Label    string               `json:"label"`
	Position DesignerNodePosition `json:"position"`
	Config   map[string]interface{}`json:"config"`
}

// DesignerEdge represents a visual dependency connection between canvas nodes.
type DesignerEdge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Condition string `json:"condition,omitempty"` // "success", "failure", "always"
}

// DesignerWorkflowCanvas represents a full drag-and-drop visual workflow state.
type DesignerWorkflowCanvas struct {
	WorkflowID  string         `json:"workflow_id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Nodes       []DesignerNode `json:"nodes"`
	Edges       []DesignerEdge `json:"edges"`
}

// ExportToWorkflowDef converts a visual canvas layout into a runnable ServFlow WorkflowDef.
func (dwc *DesignerWorkflowCanvas) ExportToWorkflowDef() (storage.WorkflowDef, error) {
	if dwc.WorkflowID == "" {
		return storage.WorkflowDef{}, fmt.Errorf("workflow_id cannot be empty")
	}

	def := storage.WorkflowDef{
		ID:    dwc.WorkflowID,
		Tasks: make([]storage.Task, 0, len(dwc.Nodes)),
	}

	// Map incoming edges to build DependsOn slice for each node
	dependsOnMap := make(map[string][]string)
	for _, edge := range dwc.Edges {
		dependsOnMap[edge.Target] = append(dependsOnMap[edge.Target], edge.Source)
	}

	for _, node := range dwc.Nodes {
		taskName := node.ID
		if taskName == "" {
			continue
		}

		action, _ := node.Config["target_url"].(string)
		if action == "" {
			action, _ = node.Config["action"].(string)
		}
		if action == "" && node.Type == "http_task" {
			action = fmt.Sprintf("http://localhost:8080/tasks/%s", taskName)
		}

		task := storage.Task{
			Name:      taskName,
			Action:    action,
			DependsOn: dependsOnMap[taskName],
		}
		def.Tasks = append(def.Tasks, task)
	}

	// Verify the exported workflow DAG has no cycles
	if HasCycle(def) {
		return storage.WorkflowDef{}, fmt.Errorf("exported workflow definition contains a dependency cycle")
	}

	return def, nil
}

// ImportFromWorkflowDef converts an existing WorkflowDef into a visual canvas layout.
func ImportFromWorkflowDef(def storage.WorkflowDef) DesignerWorkflowCanvas {
	canvas := DesignerWorkflowCanvas{
		WorkflowID: def.ID,
		Name:       def.ID,
		Nodes:      make([]DesignerNode, 0, len(def.Tasks)),
		Edges:      make([]DesignerEdge, 0),
	}

	for i, task := range def.Tasks {
		node := DesignerNode{
			ID:    task.Name,
			Type:  "http_task",
			Label: strings.Title(strings.ReplaceAll(task.Name, "_", " ")),
			Position: DesignerNodePosition{
				X: float64(100 + (i%3)*250),
				Y: float64(100 + (i/3)*150),
			},
			Config: map[string]interface{}{
				"action": task.Action,
			},
		}
		canvas.Nodes = append(canvas.Nodes, node)

		for j, dep := range task.DependsOn {
			edge := DesignerEdge{
				ID:        fmt.Sprintf("edge-%s-%s-%d", dep, task.Name, j),
				Source:    dep,
				Target:    task.Name,
				Condition: "success",
			}
			canvas.Edges = append(canvas.Edges, edge)
		}
	}

	return canvas
}

// CanvasToJSON serializes a canvas to formatted JSON string.
func (dwc *DesignerWorkflowCanvas) CanvasToJSON() (string, error) {
	b, err := json.MarshalIndent(dwc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
