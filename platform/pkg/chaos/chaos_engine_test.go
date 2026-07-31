package chaos

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUnifiedChaosEngine_InjectAndAbort(t *testing.T) {
	engine := NewUnifiedChaosEngine()

	fault := ChaosFault{
		Kind:       FaultNetwork,
		TargetNode: "node-1",
		Intensity:  0.3,
		Duration:   10 * time.Second,
	}

	body, _ := json.Marshal(fault)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/platform/chaos/faults", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	engine.HTTPHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var injected ChaosFault
	_ = json.NewDecoder(w.Body).Decode(&injected)
	if !injected.Active || injected.ID == "" {
		t.Fatalf("expected active injected fault, got %+v", injected)
	}

	if err := engine.AbortFault(injected.ID); err != nil {
		t.Fatalf("AbortFault failed: %v", err)
	}

	faults := engine.ListFaults()
	if len(faults) != 1 || faults[0].Active {
		t.Errorf("expected fault to be inactive after abort")
	}
}
