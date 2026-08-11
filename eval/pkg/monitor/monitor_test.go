package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/eval/api"
)

type mockEvaluator struct {
	name  string
	score float64
}

func (m *mockEvaluator) Name() string { return m.name }
func (m *mockEvaluator) Evaluate(ctx context.Context, t api.Trajectory) (api.EvalScore, error) {
	return api.EvalScore{
		Evaluator: m.name,
		Score:     m.score,
		MaxScore:  1.0,
		Pass:      m.score >= 0.5,
	}, nil
}

func TestOnlineMonitor(t *testing.T) {
	cfg := Config{
		SampleRate:     1.0,
		WindowSize:     5,
		DropRatioAlert: 0.1,
	}

	monitor := NewOnlineMonitor(cfg, &mockEvaluator{name: "acc", score: 0.8})

	traj := api.Trajectory{
		ID:      "t1",
		AgentID: "agent1",
	}

	monitor.Sample(traj)

	// Wait for goroutine to finish
	time.Sleep(100 * time.Millisecond)

	scores, err := monitor.GetRollingScore("agent1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scores) == 0 {
		t.Fatalf("expected scores, got none")
	}

	if scores["acc"] != 0.8 {
		t.Errorf("expected 0.8, got %v", scores["acc"])
	}

	select {
	case alert := <-monitor.Alerts():
		if alert.EvaluatorName != "acc" {
			t.Errorf("unexpected evaluator: %s", alert.EvaluatorName)
		}
		if alert.DropPercentage < 0.19 || alert.DropPercentage > 0.21 {
			t.Errorf("unexpected drop: %f", alert.DropPercentage)
		}
	default:
		t.Error("expected an alert")
	}
}

func TestOnlineMonitorEviction(t *testing.T) {
	cfg := Config{
		SampleRate:     1.0,
		WindowSize:     2,
		DropRatioAlert: 0.5,
	}

	monitor := NewOnlineMonitor(cfg, &mockEvaluator{name: "acc", score: 1.0})
	
	monitor.Sample(api.Trajectory{ID: "t1", AgentID: "agent1"})
	monitor.Sample(api.Trajectory{ID: "t2", AgentID: "agent1"})
	monitor.Sample(api.Trajectory{ID: "t3", AgentID: "agent1"})
	
	time.Sleep(100 * MillisecondHelper())

	monitor.mu.RLock()
	hist := monitor.history["agent1"]
	monitor.mu.RUnlock()

	if len(hist) != 2 {
		t.Errorf("expected history to be truncated to 2, got %d", len(hist))
	}
}

func MillisecondHelper() time.Duration {
	return time.Millisecond
}
