package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	evalapi "github.com/vyuvaraj/pranor/eval/api"
)

func TestReplayTrajectoryJSON(t *testing.T) {
	traj := evalapi.Trajectory{
		ID:         "tr-test-100",
		AgentID:    "agent-1",
		TenantID:   "tenant-1",
		RecordedAt: time.Now().UTC(),
		Spans: []evalapi.TrajectorySpan{
			{SpanName: "pranor.agent_execution", Module: "gate", Outcome: "ALLOW", DurationMs: 10},
			{SpanName: "pranor.decision.evaluate", Module: "decision", Outcome: "APPROVE", DurationMs: 15},
		},
	}

	data, err := json.Marshal(traj)
	if err != nil {
		t.Fatalf("failed to marshal trajectory: %v", err)
	}

	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "trajectory.json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		t.Fatalf("failed to write test trajectory: %v", err)
	}

	// Verify file exists and is valid JSON
	readData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	var parsed evalapi.Trajectory
	if err := json.Unmarshal(readData, &parsed); err != nil {
		t.Fatalf("failed to unmarshal trajectory: %v", err)
	}

	if parsed.ID != traj.ID {
		t.Errorf("expected ID %s, got %s", traj.ID, parsed.ID)
	}
}
