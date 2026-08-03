package otel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vyuvaraj/pranor/trace/pkg/store"
)

func TestOTelExporter(t *testing.T) {
	received := false
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode OTLP payload: %v", err)
		}
		if _, ok := payload["resourceSpans"]; !ok {
			t.Errorf("Expected resourceSpans in OTLP payload")
		}
		received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	exporter := NewExporter(mockServer.URL)

	spans := []store.Span{
		{
			TraceID:   "trace-123",
			SpanID:    "span-456",
			Name:      "GET /api/v1/orders",
			Kind:      1,
			StartTime: 1000,
			EndTime:   2000,
			Status:    1,
			Service:   "order-service",
		},
	}

	err := exporter.ExportTraces(context.Background(), spans)
	if err != nil {
		t.Fatalf("ExportTraces returned error: %v", err)
	}

	if !received {
		t.Errorf("Mock server did not receive exported trace")
	}
}
