package profiler

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// StackSample represents an eBPF CPU or memory stack trace sample.
type StackSample struct {
	StackFrames []string  `json:"stack_frames"` // e.g. ["main", "http.Handler", "db.Query"]
	Count       int64     `json:"count"`
	SampleType  string    `json:"sample_type"` // "cpu" or "memory"
	Timestamp   time.Time `json:"timestamp"`
}

// AnomalyExplanation details identified performance degradation root causes.
type AnomalyExplanation struct {
	Service            string   `json:"service"`
	AnomalyType        string   `json:"anomaly_type"` // "cpu_spike", "memory_leak"
	TopHotspotFunction string   `json:"top_hotspot_function"`
	PercentageCPU      float64  `json:"percentage_cpu"`
	SuspectStack       []string `json:"suspect_stack"`
	Recommendation     string   `json:"recommendation"`
}

// AnomalyExplainer analyzes continuous eBPF stack trace samples to explain performance anomalies.
type AnomalyExplainer struct {
	mu      sync.RWMutex
	samples map[string][]StackSample // service -> samples
}

// NewAnomalyExplainer creates an AnomalyExplainer instance.
func NewAnomalyExplainer() *AnomalyExplainer {
	return &AnomalyExplainer{
		samples: make(map[string][]StackSample),
	}
}

// RecordSamples ingests eBPF stack trace samples for a service.
func (ae *AnomalyExplainer) RecordSamples(service string, samples []StackSample) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.samples[service] = append(ae.samples[service], samples...)
}

// ExplainAnomaly performs stack frequency analysis to identify CPU/memory hotspots during an anomaly.
func (ae *AnomalyExplainer) ExplainAnomaly(service string) (*AnomalyExplanation, error) {
	ae.mu.RLock()
	samples, ok := ae.samples[service]
	ae.mu.RUnlock()

	if !ok || len(samples) == 0 {
		return nil, fmt.Errorf("no eBPF profile samples recorded for service '%s'", service)
	}

	frameCounts := make(map[string]int64)
	var totalSamples int64
	frameStacks := make(map[string][]string)

	for _, sample := range samples {
		totalSamples += sample.Count
		if len(sample.StackFrames) > 0 {
			topFrame := sample.StackFrames[len(sample.StackFrames)-1]
			frameCounts[topFrame] += sample.Count
			frameStacks[topFrame] = sample.StackFrames
		}
	}

	if totalSamples == 0 {
		return nil, fmt.Errorf("total sample count is zero")
	}

	type frameStat struct {
		Frame string
		Count int64
	}
	var stats []frameStat
	for frame, count := range frameCounts {
		stats = append(stats, frameStat{Frame: frame, Count: count})
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	top := stats[0]
	pct := (float64(top.Count) / float64(totalSamples)) * 100.0

	explanation := &AnomalyExplanation{
		Service:            service,
		AnomalyType:        "cpu_spike",
		TopHotspotFunction: top.Frame,
		PercentageCPU:      pct,
		SuspectStack:       frameStacks[top.Frame],
		Recommendation:     fmt.Sprintf("Optimize function '%s' which accounts for %.1f%% of total CPU execution samples", top.Frame, pct),
	}

	return explanation, nil
}

// FormatFlamegraphText converts stack samples to fold-stack format for flamegraph rendering.
func (ae *AnomalyExplainer) FormatFlamegraphText(service string) string {
	ae.mu.RLock()
	samples := ae.samples[service]
	ae.mu.RUnlock()

	var sb strings.Builder
	for _, s := range samples {
		sb.WriteString(strings.Join(s.StackFrames, ";"))
		sb.WriteString(fmt.Sprintf(" %d\n", s.Count))
	}
	return sb.String()
}
