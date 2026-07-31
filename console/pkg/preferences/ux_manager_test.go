package preferences

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUXPreferencesManager_SaveAndGet(t *testing.T) {
	manager := NewUXPreferencesManager()

	pref := UserUXPreferences{
		UserID:        "admin",
		Theme:         "dark",
		PinnedWidgets: []string{"flamegraph", "latency_slo"},
	}

	body, _ := json.Marshal(pref)

	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/console/preferences", bytes.NewBuffer(body))
	wPost := httptest.NewRecorder()

	manager.HTTPHandler().ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", wPost.Code)
	}

	saved, found := manager.GetPreferences("admin")
	if !found || saved.Theme != "dark" || len(saved.PinnedWidgets) != 2 {
		t.Fatalf("unexpected saved preferences: %+v", saved)
	}
}
