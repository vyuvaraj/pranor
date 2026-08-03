package grafana

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vyuvaraj/pranor/trace/pkg/store"
)

func TestGrafanaPlugin(t *testing.T) {
	ts := store.NewStore(100)
	ts.AddSpans([]store.Span{
		{
			TraceID:   "tr-graf-1",
			SpanID:    "sp-1",
			Name:      "root",
			StartTime: time.Now().UnixNano(),
			EndTime:   time.Now().Add(50 * time.Millisecond).UnixNano(),
			Service:   "test-svc",
		},
	})

	plugin := NewGrafanaPlugin(ts)

	// Test Connection
	reqTest := httptest.NewRequest("GET", "/", nil)
	wTest := httptest.NewRecorder()
	plugin.HandleTestConnection(wTest, reqTest)
	if wTest.Code != http.StatusOK {
		t.Errorf("Expected 200 for test connection, got %d", wTest.Code)
	}

	// Query
	qReq := GrafanaQueryRequest{
		Targets: []struct {
			Target string `json:"target"`
			RefID  string `json:"refId"`
			Type   string `json:"type"`
		}{
			{Target: "latency", RefID: "A", Type: "timeseries"},
		},
	}
	body, _ := json.Marshal(qReq)
	reqQuery := httptest.NewRequest("POST", "/query", bytes.NewReader(body))
	wQuery := httptest.NewRecorder()
	plugin.HandleQuery(wQuery, reqQuery)

	if wQuery.Code != http.StatusOK {
		t.Errorf("Expected 200 for query, got %d", wQuery.Code)
	}
}
