package telemetry

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// FlamegraphProfilePayload represents an eBPF profiling sample payload.
type FlamegraphProfilePayload struct {
	Service        string    `json:"service"`
	ProfileType    string    `json:"profile_type"` // "cpu" or "memory"
	FoldedStack    string    `json:"folded_stack"`
	SampleCount    int64     `json:"sample_count"`
	CorrelatedTrace string   `json:"correlated_trace_id,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// FlamegraphTelemetryEndpoint collects eBPF profile samples for ServConsole visualization.
type FlamegraphTelemetryEndpoint struct {
	mu       sync.RWMutex
	profiles map[string][]FlamegraphProfilePayload // service -> profiles
}

// NewFlamegraphTelemetryEndpoint creates a FlamegraphTelemetryEndpoint instance.
func NewFlamegraphTelemetryEndpoint() *FlamegraphTelemetryEndpoint {
	return &FlamegraphTelemetryEndpoint{
		profiles: make(map[string][]FlamegraphProfilePayload),
	}
}

// RecordProfile ingests a flamegraph profile sample.
func (fte *FlamegraphTelemetryEndpoint) RecordProfile(payload FlamegraphProfilePayload) {
	fte.mu.Lock()
	defer fte.mu.Unlock()

	if payload.Timestamp.IsZero() {
		payload.Timestamp = time.Now()
	}
	fte.profiles[payload.Service] = append(fte.profiles[payload.Service], payload)
}

// GetProfiles returns captured flamegraph profiles for a service.
func (fte *FlamegraphTelemetryEndpoint) GetProfiles(service string) []FlamegraphProfilePayload {
	fte.mu.RLock()
	defer fte.mu.RUnlock()

	list, ok := fte.profiles[service]
	if !ok {
		return []FlamegraphProfilePayload{}
	}
	res := make([]FlamegraphProfilePayload, len(list))
	copy(res, list)
	return res
}

// HTTPHandler exposes REST endpoint `/api/v1/console/profiler/flamegraph`.
func (fte *FlamegraphTelemetryEndpoint) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var payload FlamegraphProfilePayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid JSON payload", http.StatusBadRequest)
				return
			}
			fte.RecordProfile(payload)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
			return
		}

		service := r.URL.Query().Get("service")
		profiles := fte.GetProfiles(service)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":    len(profiles),
			"profiles": profiles,
		})
	})
}
