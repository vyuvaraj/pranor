package import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestShareTokenManager_CreateAndValidate(t *testing.T) {
	stm := NewShareTokenManager()

	st, shareURL := stm.CreateShareToken("tunnel-prod-1", 1*time.Hour, true)
	if st.TunnelID != "tunnel-prod-1" || !st.SingleUse {
		t.Fatalf("unexpected share token: %+v", st)
	}

	if shareURL == "" {
		t.Fatal("expected non-empty share URL")
	}

	// First validation -> Success
	tunnelID, err := stm.ValidateToken(st.Token)
	if err != nil || tunnelID != "tunnel-prod-1" {
		t.Fatalf("expected 1st validation success, got tunnelID=%s err=%v", tunnelID, err)
	}

	// Second validation -> Fails due to single-use
	_, err = stm.ValidateToken(st.Token)
	if err == nil {
		t.Error("expected error validating single-use token second time")
	}

	// Middleware test
	st2, _ := stm.CreateShareToken("tunnel-dev-2", 1*time.Hour, false)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := stm.Middleware(next)
	req := httptest.NewRequest(http.MethodGet, "/?share_token="+st2.Token, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 from share token middleware, got %d", w.Code)
	}
}
