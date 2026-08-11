package monitor

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/eval/api"
)

type RegressionAlert struct {
	AgentID        string
	EvaluatorName  string
	CurrentScore   float64
	Threshold      float64
	DropPercentage float64
	AlertedAt      time.Time
}

type Config struct {
	SampleRate     float64 // 0.0 to 1.0 (1.0 = 100% sampling)
	WindowSize     int     // max trajectories stored per agent window (default 100)
	DropRatioAlert float64 // threshold drop to trigger alert e.g. 0.10 (10%)
}

type OnlineMonitor struct {
	mu         sync.RWMutex
	cfg        Config
	evaluators []api.Evaluator
	history    map[string][]api.EvalResult // key: agentID
	alertChan  chan RegressionAlert
}

func NewOnlineMonitor(cfg Config, evs ...api.Evaluator) *OnlineMonitor {
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = 100
	}
	return &OnlineMonitor{
		cfg:        cfg,
		evaluators: evs,
		history:    make(map[string][]api.EvalResult),
		alertChan:  make(chan RegressionAlert, 100),
	}
}

func (m *OnlineMonitor) RegisterEvaluator(e api.Evaluator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evaluators = append(m.evaluators, e)
}

func (m *OnlineMonitor) Sample(t api.Trajectory) {
	if m.cfg.SampleRate < 1.0 {
		if rand.Float64() > m.cfg.SampleRate {
			return
		}
	}

	go func(traj api.Trajectory) {
		m.mu.RLock()
		evs := make([]api.Evaluator, len(m.evaluators))
		copy(evs, m.evaluators)
		m.mu.RUnlock()

		if len(evs) == 0 {
			return
		}

		result := api.EvalResult{
			TrajectoryID: traj.ID,
			EvaluatedAt:  time.Now().UTC(),
			OverallPass:  true,
		}

		for _, ev := range evs {
			score, err := ev.Evaluate(context.Background(), traj)
			if err != nil {
				score = api.EvalScore{
					Evaluator: ev.Name(),
					Score:     0,
					MaxScore:  1.0,
					Pass:      false,
					Detail:    err.Error(),
				}
			}
			result.Scores = append(result.Scores, score)
			if !score.Pass {
				result.OverallPass = false
			}
		}

		m.mu.Lock()
		agentID := traj.AgentID
		m.history[agentID] = append(m.history[agentID], result)
		if len(m.history[agentID]) > m.cfg.WindowSize {
			m.history[agentID] = m.history[agentID][len(m.history[agentID])-m.cfg.WindowSize:]
		}

		scores := make(map[string]float64)
		counts := make(map[string]int)
		maxes := make(map[string]float64)
		for _, res := range m.history[agentID] {
			for _, s := range res.Scores {
				scores[s.Evaluator] += s.Score
				counts[s.Evaluator]++
				maxes[s.Evaluator] = s.MaxScore
			}
		}

		var alerts []RegressionAlert
		for evName, totalScore := range scores {
			avg := totalScore / float64(counts[evName])
			maxScore := maxes[evName]
			if maxScore == 0 {
				maxScore = 1.0
			}
			drop := (maxScore - avg) / maxScore
			if drop >= m.cfg.DropRatioAlert {
				alerts = append(alerts, RegressionAlert{
					AgentID:        agentID,
					EvaluatorName:  evName,
					CurrentScore:   avg,
					Threshold:      maxScore,
					DropPercentage: drop,
					AlertedAt:      time.Now().UTC(),
				})
			}
		}
		m.mu.Unlock()

		for _, alert := range alerts {
			select {
			case m.alertChan <- alert:
			default:
			}
		}
	}(t)
}

func (m *OnlineMonitor) GetRollingScore(agentID string) (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, ok := m.history[agentID]
	if !ok || len(history) == 0 {
		return nil, nil
	}

	scores := make(map[string]float64)
	counts := make(map[string]int)

	for _, res := range history {
		for _, s := range res.Scores {
			scores[s.Evaluator] += s.Score
			counts[s.Evaluator]++
		}
	}

	result := make(map[string]float64)
	for k, v := range scores {
		result[k] = v / float64(counts[k])
	}
	return result, nil
}

func (m *OnlineMonitor) Alerts() <-chan RegressionAlert {
	return m.alertChan
}
