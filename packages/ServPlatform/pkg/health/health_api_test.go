package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUnifiedHealthAPI_RollupAndEndpoints(t *testing.T) {
	api := NewUnifiedHealthAPI()

	api.ReportHealth(ComponentHealth{Name: "servgate", Healthy: true, Latency: 2 * time.Millisecond})
	api.ReportHealth(ComponentHealth{Name: "servqueue", Healthy: true, Latency: 3 * time.Millisecond})
	api.ReportHealth(ComponentHealth{Name: "servstore", Healthy: false, Message: "disk full"})

	rollup := api.GetRollup()
	if rollup.Healthy {
		t.Error("expected rollup to be unhealthy when any component fails")
	}
	if rollup.Failing != 1 || rollup.Passing != 2 {
		t.Errorf("unexpected rollup counts: %+v", rollup)
	}

	// /health endpoint - should return 503 because servstore is failing
	reqH := httptest.NewRequest(http.MethodGet, "/health", nil)
	wH := httptest.NewRecorder()
	api.HTTPHandler().ServeHTTP(wH, reqH)
	if wH.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 from /health, got %d", wH.Code)
	}

	// /api/v1/platform/health/rollup
	reqR := httptest.NewRequest(http.MethodGet, "/api/v1/platform/health/rollup", nil)
	wR := httptest.NewRecorder()
	api.HTTPHandler().ServeHTTP(wR, reqR)
	if wR.Code != http.StatusOK {
		t.Fatalf("expected 200 from rollup endpoint, got %d", wR.Code)
	}
	var resp HealthRollup
	_ = json.NewDecoder(wR.Body).Decode(&resp)
	if resp.Total != 3 {
		t.Errorf("expected 3 components in rollup, got %d", resp.Total)
	}
}
