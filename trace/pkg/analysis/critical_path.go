package analysis

import (
	"fmt"
	"sort"
	"time"
)

// TraceSpan represents a span in an OpenTelemetry trace hierarchy.
type TraceSpan struct {
	SpanID        string        `json:"span_id"`
	ParentSpanID  string        `json:"parent_span_id,omitempty"`
	Name          string        `json:"name"`
	Service       string        `json:"service"`
	StartTime     time.Time     `json:"start_time"`
	DurationMs    int64         `json:"duration_ms"`
	SelfDuration  int64         `json:"self_duration_ms"`
	IsCritical    bool          `json:"is_critical"`
}

// CriticalPathResult breaks down the longest sequential path in a trace execution.
type CriticalPathResult struct {
	TraceID            string      `json:"trace_id"`
	TotalDurationMs    int64       `json:"total_duration_ms"`
	CriticalDurationMs int64       `json:"critical_duration_ms"`
	CriticalPath       []TraceSpan `json:"critical_path"`
}

// CriticalPathAnalyzer analyzes distributed span trees to calculate the critical bottleneck path.
type CriticalPathAnalyzer struct{}

// NewCriticalPathAnalyzer creates a CriticalPathAnalyzer instance.
func NewCriticalPathAnalyzer() *CriticalPathAnalyzer {
	return &CriticalPathAnalyzer{}
}

// AnalyzeTrace builds the dependency tree and extracts the critical path (longest time path).
func (cpa *CriticalPathAnalyzer) AnalyzeTrace(traceID string, spans []TraceSpan) (*CriticalPathResult, error) {
	if len(spans) == 0 {
		return nil, fmt.Errorf("spans list cannot be empty for trace %s", traceID)
	}

	spanMap := make(map[string]*TraceSpan)
	childrenMap := make(map[string][]*TraceSpan)
	var root *TraceSpan

	for i := range spans {
		s := &spans[i]
		spanMap[s.SpanID] = s
		if s.ParentSpanID == "" {
			root = s
		} else {
			childrenMap[s.ParentSpanID] = append(childrenMap[s.ParentSpanID], s)
		}
	}

	if root == nil {
		root = &spans[0] // Fallback to first span if root not explicitly unset
	}

	// Dynamic programming / DFS to find path with maximum cumulative duration
	var findLongestPath func(s *TraceSpan) ([]*TraceSpan, int64)
	findLongestPath = func(curr *TraceSpan) ([]*TraceSpan, int64) {
		children := childrenMap[curr.SpanID]
		if len(children) == 0 {
			return []*TraceSpan{curr}, curr.DurationMs
		}

		var bestChildPath []*TraceSpan
		var maxChildDur int64

		for _, child := range children {
			cPath, cDur := findLongestPath(child)
			if cDur > maxChildDur {
				maxChildDur = cDur
				bestChildPath = cPath
			}
		}

		res := append([]*TraceSpan{curr}, bestChildPath...)
		return res, curr.DurationMs + maxChildDur
	}

	longestPathNodes, totalDur := findLongestPath(root)

	criticalPath := make([]TraceSpan, len(longestPathNodes))
	for i, node := range longestPathNodes {
		cpNode := *node
		cpNode.IsCritical = true
		criticalPath[i] = cpNode
	}

	// Sort critical path chronologically by StartTime
	sort.Slice(criticalPath, func(i, j int) bool {
		return criticalPath[i].StartTime.Before(criticalPath[j].StartTime)
	})

	return &CriticalPathResult{
		TraceID:            traceID,
		TotalDurationMs:    root.DurationMs,
		CriticalDurationMs: totalDur,
		CriticalPath:       criticalPath,
	}, nil
}
