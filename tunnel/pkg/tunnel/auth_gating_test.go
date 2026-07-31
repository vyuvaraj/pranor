package tunnel

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthGatingManager_APIKeyAndJWT(t *testing.T) {
	agm := NewAuthGatingManager("test-secret-123")
	agm.RegisterAPIKey("tn_key_999", "dev-user")

	// 1. Valid API Key
	req1 := httptest.NewRequest(http.MethodGet, "/tunnel?api_key=tn_key_999", nil)
	owner, err := agm.AuthenticateRequest(req1)
	if err != nil || owner != "dev-user" {
		t.Fatalf("expected API key auth success, got owner=%s err=%v", owner, err)
	}

	// 2. Valid JWT
	jwt := agm.GenerateJWT("jwt-user", 1*time.Hour)
	req2 := httptest.NewRequest(http.MethodGet, "/tunnel", nil)
	req2.Header.Set("Authorization", "Bearer "+jwt)

	owner, err = agm.AuthenticateRequest(req2)
	if err != nil || owner != "jwt-user" {
		t.Fatalf("expected JWT auth success, got owner=%s err=%v", owner, err)
	}

	// 3. Invalid Auth
	req3 := httptest.NewRequest(http.MethodGet, "/tunnel?api_key=invalid", nil)
	_, err = agm.AuthenticateRequest(req3)
	if err == nil {
		t.Error("expected error for invalid API key")
	}
}
