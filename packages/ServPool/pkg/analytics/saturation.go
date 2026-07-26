package analytics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// PoolSaturationAlert represents a pool connection saturation alert.
type PoolSaturationAlert struct {
	PoolName       string    `json:"pool_name"`
	ActiveConns    int       `json:"active_conns"`
	MaxConns       int       `json:"max_conns"`
	UtilizationPct float64   `json:"utilization_pct"`
	WaitCount      int64     `json:"wait_count"`
	AlertTriggered bool      `json:"alert_triggered"`
	Timestamp      time.Time `json:"timestamp"`
}

// SaturationMonitor evaluates connection pool capacity saturation and queue wait pressure.
type SaturationMonitor struct {
	mu                   sync.RWMutex
	poolName             string
	highWatermarkPct     float64 // e.g. 85.0 (85%)
	lastAlert            *PoolSaturationAlert
}

// NewSaturationMonitor creates a SaturationMonitor instance.
func NewSaturationMonitor(poolName string, highWatermarkPct float64) *SaturationMonitor {
	if highWatermarkPct <= 0 {
		highWatermarkPct = 85.0
	}
	return &SaturationMonitor{
		poolName:         poolName,
		highWatermarkPct: highWatermarkPct,
	}
}

// Evaluate evaluate pool utilization percentage and wait pressure.
func (sm *SaturationMonitor) Evaluate(activeConns, maxConns int, waitCount int64) PoolSaturationAlert {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	utilPct := 0.0
	if maxConns > 0 {
		utilPct = (float64(activeConns) / float64(maxConns)) * 100.0
	}

	triggered := utilPct >= sm.highWatermarkPct || waitCount > 10

	alert := PoolSaturationAlert{
		PoolName:       sm.poolName,
		ActiveConns:    activeConns,
		MaxConns:       maxConns,
		UtilizationPct: utilPct,
		WaitCount:      waitCount,
		AlertTriggered: triggered,
		Timestamp:      time.Now(),
	}

	sm.lastAlert = &alert
	return alert
}

// HTTPHandler exposes `/api/v1/pool/saturation` for ServConsole visual status gauges.
func (sm *SaturationMonitor) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sm.mu.RLock()
		alert := sm.lastAlert
		sm.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if alert == nil {
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no data recorded"})
			return
		}
		_ = json.NewEncoder(w).Encode(alert)
	})
}
