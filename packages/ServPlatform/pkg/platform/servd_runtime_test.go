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

	components := []string{"servgate", "servqueue", "servstore", "servmesh", "servtrace"}
	for _, c := range components {
		rt.RegisterComponent(c)
	}

	if err := rt.StartComponent("servgate"); err != nil {
		t.Fatalf("failed to start servgate: %v", err)
	}

	list := rt.ListComponents()
	if len(list) != 5 {
		t.Fatalf("expected 5 components, got %d", len(list))
	}

	var gateStatus *ComponentStatus
	for _, c := range list {
		if c.Name == "servgate" {
			c := c
			gateStatus = &c
		}
	}
	if gateStatus == nil || !gateStatus.Running {
		t.Errorf("expected servgate to be running")
	}

	// HTTP status endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/v1/servd/components", nil)
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
