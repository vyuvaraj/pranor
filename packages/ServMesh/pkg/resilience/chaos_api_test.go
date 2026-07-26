package resilience

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChaosEngine_EvaluationAndPartition(t *testing.T) {
	ce := NewChaosEngine()

	req := httptest.NewRequest(http.MethodGet, "http://serv/order-service/checkout", nil)

	// No rule -> no fault
	resp, hit := ce.Evaluate("order-service", req)
	if hit || resp != nil {
		t.Error("expected no chaos evaluation hit when no rule is configured")
	}

	// Inject Network Partition
	ce.InjectRule(ChaosRule{
		TargetService:   "order-service",
		EnablePartition: true,
	})

	resp, hit = ce.Evaluate("order-service", req)
	if !hit || resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503 partition hit, got hit=%v status=%d", hit, resp.StatusCode)
	}
}

func TestChaosEngine_HTTPHandler(t *testing.T) {
	ce := NewChaosEngine()
	handler := ce.HTTPHandler()

	rule := ChaosRule{
		TargetService: "payment-svc",
		ErrorStatus:   500,
		ErrorRatio:    1.0,
	}

	body, _ := json.Marshal(rule)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chaos/inject", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on chaos inject, got %d", w.Code)
	}

	evalReq := httptest.NewRequest(http.MethodGet, "http://serv/payment-svc", nil)
	resp, hit := ce.Evaluate("payment-svc", evalReq)
	if !hit || resp.StatusCode != 500 {
		t.Errorf("expected 500 error fault hit, got hit=%v status=%d", hit, resp.StatusCode)
	}
}
