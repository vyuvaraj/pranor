package prometheus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/trace/pkg/store"
)

func TestPrometheusRemoteWrite(t *testing.T) {
	ts := store.NewStore(100)
	recv := NewReceiver(ts)

	payload := RemoteWritePayload{
		Timeseries: []struct {
			Labels []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"labels"`
			Samples []struct {
				Value     float64 `json:"value"`
				Timestamp int64   `json:"timestamp"`
			} `json:"samples"`
		}{
			{
				Labels: []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				}{
					{Name: "__name__", Value: "http_requests_total"},
					{Name: "job", Value: "api-server"},
					{Name: "handler", Value: "/api/v1/users"},
				},
				Samples: []struct {
					Value     float64 `json:"value"`
					Timestamp int64   `json:"timestamp"`
				}{
					{Value: 42.0, Timestamp: time.Now().UnixMilli()},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/prometheus/write", bytes.NewReader(data))
	w := httptest.NewRecorder()

	recv.HandleRemoteWrite(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	traces := ts.ListTraces()
	if len(traces) == 0 {
		t.Fatalf("Expected ingested metric trace in store, got 0")
	}

	found := false
	for _, tr := range traces {
		if tr.Service == "api-server" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected trace with service 'api-server', traces: %+v", traces)
	}
}
