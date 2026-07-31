package import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSaturationMonitor_EvaluateAndAlert(t *testing.T) {
	sm := NewSaturationMonitor("master-db-pool", 80.0)

	// 1. Normal load (50% util) -> No alert
	alert := sm.Evaluate(50, 100, 0)
	if alert.AlertTriggered || alert.UtilizationPct != 50.0 {
		t.Errorf("expected no alert for 50%% load, got %+v", alert)
	}

	// 2. High load (90% util > 80%) -> Alert triggered
	alert = sm.Evaluate(90, 100, 5)
	if !alert.AlertTriggered {
		t.Error("expected saturation alert to trigger at 90% load")
	}

	// HTTP Handler test
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pool/saturation", nil)
	w := httptest.NewRecorder()
	sm.HTTPHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var resp PoolSaturationAlert
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.PoolName != "master-db-pool" || !resp.AlertTriggered {
		t.Errorf("unexpected alert JSON response: %+v", resp)
	}
}
