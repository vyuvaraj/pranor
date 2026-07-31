package import (
	"fmt"
	"strings"
)

// ConcurrentBlockAST represents a `concurrent { ... }` parallel execution block in AST.
type ConcurrentBlockAST struct {
	ID        string    `json:"id"`
	Tasks     []AsyncTask `json:"tasks"`
	IsJoinAll bool      `json:"is_join_all"`
}

// AsyncTask represents an async function call or task expression.
type AsyncTask struct {
	TaskName string `json:"task_name"`
	CallExpr string `json:"call_expr"`
}

// ConcurrentParser parses `async` functions and `concurrent {}` block primitives into execution AST nodes.
type ConcurrentParser struct{}

// NewConcurrentParser creates a ConcurrentParser instance.
func NewConcurrentParser() *ConcurrentParser {
	return &ConcurrentParser{}
}

// ParseConcurrentBlock parses a `concurrent {}` block string into a ConcurrentBlockAST.
func (cp *ConcurrentParser) ParseConcurrentBlock(blockContent string) (*ConcurrentBlockAST, error) {
	trimmed := strings.TrimSpace(blockContent)
	if !strings.HasPrefix(trimmed, "concurrent {") || !strings.HasSuffix(trimmed, "}") {
		return nil, fmt.Errorf("invalid concurrent block syntax: must start with 'concurrent {' and end with '}'")
	}

	inner := strings.TrimPrefix(trimmed, "concurrent {")
	inner = strings.TrimSuffix(inner, "}")

	lines := strings.Split(inner, "\n")
	var tasks []AsyncTask

	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" || strings.HasPrefix(l, "//") {
			continue
		}

		if strings.HasPrefix(l, "async ") {
			callExpr := strings.TrimPrefix(l, "async ")
			callExpr = strings.TrimSuffix(callExpr, ";")
			taskName := fmt.Sprintf("task_%d", len(tasks)+1)
			tasks = append(tasks, AsyncTask{
				TaskName: taskName,
				CallExpr: callExpr,
			})
		}
	}

	return &ConcurrentBlockAST{
		ID:        fmt.Sprintf("conc_block_%d", len(tasks)),
		Tasks:     tasks,
		IsJoinAll: true,
	}, nil
}
