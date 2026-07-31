package import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChaosControlPanel_TriggerAndAbort(t *testing.T) {
	ccp := NewChaosControlPanel()

	exp := ChaosExperiment{
		TargetService: "order-service",
		FaultType:     "latency",
		Duration:      30 * time.Second,
	}

	body, _ := json.Marshal(exp)

	// Trigger experiment via POST
	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/console/chaos/experiments", bytes.NewBuffer(body))
	wPost := httptest.NewRecorder()
	ccp.HTTPHandler().ServeHTTP(wPost, reqPost)

	if wPost.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", wPost.Code)
	}

	var created ChaosExperiment
	_ = json.NewDecoder(wPost.Body).Decode(&created)

	if created.ID == "" || created.Status != "running" {
		t.Errorf("unexpected created experiment: %+v", created)
	}

	// Abort experiment
	err := ccp.AbortExperiment(created.ID)
	if err != nil {
		t.Fatalf("AbortExperiment failed: %v", err)
	}

	exps := ccp.GetExperiments()
	if len(exps) != 1 || exps[0].Status != "aborted" {
		t.Errorf("expected aborted experiment status, got %+v", exps)
	}
}
