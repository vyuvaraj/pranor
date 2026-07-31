package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFlamegraphTelemetryEndpoint_RecordAndQuery(t *testing.T) {
	endpoint := NewFlamegraphTelemetryEndpoint()

	payload := FlamegraphProfilePayload{
		Service:         "payment-service",
		ProfileType:     "cpu",
		FoldedStack:     "main;http.Handler;db.Query 100",
		SampleCount:     100,
		CorrelatedTrace: "4bf92f3577b34da6a3ce929d0e0e4736",
		Timestamp:       time.Now(),
	}

	body, _ := json.Marshal(payload)

	// Test POST ingest
	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/console/profiler/flamegraph", bytes.NewBuffer(body))
	wPost := httptest.NewRecorder()

	endpoint.HTTPHandler().ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on POST, got %d", wPost.Code)
	}

	// Test GET retrieval
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/console/profiler/flamegraph?service=payment-service", nil)
	wGet := httptest.NewRecorder()

	endpoint.HTTPHandler().ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on GET, got %d", wGet.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(wGet.Body).Decode(&resp)

	if resp["count"].(float64) != 1 {
		t.Errorf("expected 1 flamegraph profile, got %v", resp["count"])
	}
}
