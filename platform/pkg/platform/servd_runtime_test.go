package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServdRuntime_RegisterAndStart(t *testing.T) {
	rt := NewServdRuntime()

	components := []string{"Pranor Gate", "Pranor Pulse", "Pranor Vault", "Pranor Mesh", "Pranor Trace"}
	for _, c := range components {
		rt.RegisterComponent(c)
	}

	if err := rt.StartComponent("Pranor Gate"); err != nil {
		t.Fatalf("failed to start Pranor Gate: %v", err)
	}

	list := rt.ListComponents()
	if len(list) != 5 {
		t.Fatalf("expected 5 components, got %d", len(list))
	}

	var gateStatus *ComponentStatus
	for _, c := range list {
		if c.Name == "Pranor Gate" {
			c := c
			gateStatus = &c
		}
	}
	if gateStatus == nil || !gateStatus.Running {
		t.Errorf("expected Pranor Gate to be running")
	}

	// HTTP status endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pranord/components", nil)
	w := httptest.NewRecorder()
	rt.HTTPHandler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if _, ok := resp["components"]; !ok {
		t.Error("expected 'components' key in JSON response")
	}

	// Shutdown
	rt.Shutdown(context.Background())
	for _, c := range rt.ListComponents() {
		if c.Running {
			t.Errorf("expected all components stopped after Shutdown, but %s is still running", c.Name)
		}
	}
}
