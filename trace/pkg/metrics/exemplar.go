package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MetricExemplar binds a W3C trace ID and span ID to a Prometheus metric observation.
type MetricExemplar struct {
	MetricName string    `json:"metric_name"`
	Value      float64   `json:"value"`
	TraceID    string    `json:"trace_id"`
	SpanID     string    `json:"span_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// ExemplarStore maintains Prometheus open-metrics format exemplar attachments.
type ExemplarStore struct {
	mu        sync.RWMutex
	exemplars []MetricExemplar
}

// NewExemplarStore creates an ExemplarStore instance.
func NewExemplarStore() *ExemplarStore {
	return &ExemplarStore{
		exemplars: make([]MetricExemplar, 0),
	}
}

// RecordExemplar attaches trace_id and span_id exemplars to a metric observation.
func (es *ExemplarStore) RecordExemplar(metricName string, val float64, traceID, spanID string) {
	if traceID == "" {
		return
	}
	es.mu.Lock()
	defer es.mu.Unlock()

	ex := MetricExemplar{
		MetricName: metricName,
		Value:      val,
		TraceID:    traceID,
		SpanID:     spanID,
		Timestamp:  time.Now(),
	}
	es.exemplars = append(es.exemplars, ex)
}

// RenderOpenMetrics returns Prometheus OpenMetrics text exposition with `# {trace_id="...",span_id="..."}` exemplar attachments.
func (es *ExemplarStore) RenderOpenMetrics() string {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("# TYPE http_request_duration_seconds histogram\n")

	for _, ex := range es.exemplars {
		sb.WriteString(fmt.Sprintf("%s_bucket{le=\"+Inf\"} 1 # {trace_id=\"%s\",span_id=\"%s\"} %f %d\n",
			ex.MetricName, ex.TraceID, ex.SpanID, ex.Value, ex.Timestamp.Unix()))
	}
	return sb.String()
}

// HTTPHandler exposes OpenMetrics text endpoint for Prometheus scraping.
func (es *ExemplarStore) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/openmetrics-text; version=1.0.0; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(es.RenderOpenMetrics()))
	})
}
