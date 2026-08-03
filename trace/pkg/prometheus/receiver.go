package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vyuvaraj/pranor/trace/pkg/store"
)

// MetricSample represents a single Prometheus metric sample
type MetricSample struct {
	MetricName string            `json:"metric"`
	Labels     map[string]string `json:"labels"`
	Value      float64           `json:"value"`
	Timestamp  int64             `json:"timestamp"`
}

// RemoteWritePayload represents JSON-formatted or simple RemoteWrite payload
type RemoteWritePayload struct {
	Timeseries []struct {
		Labels []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"labels"`
		Samples []struct {
			Value     float64 `json:"value"`
			Timestamp int64   `json:"timestamp"`
		} `json:"samples"`
	} `json:"timeseries"`
}

// Receiver handles incoming Prometheus remote_write requests
type Receiver struct {
	traceStore *store.Store
}

func NewReceiver(ts *store.Store) *Receiver {
	return &Receiver{traceStore: ts}
}

// HandleRemoteWrite accepts Prometheus remote write requests
func (r *Receiver) HandleRemoteWrite(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var payload RemoteWritePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		// Fallback: parse custom simple JSON lines or snappy-like plain text format if needed
		// For standardized HTTP endpoint, JSON payload is parsed
		http.Error(w, fmt.Sprintf("Invalid remote write payload: %v", err), http.StatusBadRequest)
		return
	}

	spansAdded := 0
	for _, ts := range payload.Timeseries {
		labels := make(map[string]string)
		metricName := ""
		serviceName := "prometheus-remote-write"

		for _, lbl := range ts.Labels {
			labels[lbl.Name] = lbl.Value
			if lbl.Name == "__name__" {
				metricName = lbl.Value
			} else if lbl.Name == "service" || lbl.Name == "job" || lbl.Name == "app" {
				serviceName = lbl.Value
			}
		}

		for _, s := range ts.Samples {
			// Convert Prometheus metric into trace/span representation or store metric
			attrs := make(map[string]interface{})
			for k, v := range labels {
				attrs[k] = v
			}
			attrs["metric.name"] = metricName
			attrs["metric.value"] = s.Value
			attrs["source"] = "prometheus_remote_write"

			span := store.Span{
				TraceID:   fmt.Sprintf("prom-%s-%d", strings.ReplaceAll(metricName, ":", "_"), s.Timestamp),
				SpanID:    fmt.Sprintf("prom-sample-%d", s.Timestamp),
				Name:      fmt.Sprintf("metric:%s", metricName),
				Kind:      1,
				StartTime: s.Timestamp * 1000000, // convert ms to ns if ms
				EndTime:   s.Timestamp * 1000000,
				Status:    1,
				Attributes: attrs,
				Service:   serviceName,
			}
			r.traceStore.AddSpans([]store.Span{span})
			spansAdded++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status":"success","samples_ingested":%d}`, spansAdded)))
}
