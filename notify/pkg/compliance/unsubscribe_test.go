package import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUnsubscribeManager_GenerateAndVerify(t *testing.T) {
	um := NewUnsubscribeManager("https://mail.example.com", "secret-key-123")

	email := "subscriber@example.com"
	headers := um.GenerateHeaders(email)

	if !strings.Contains(headers.ListUnsubscribe, "https://mail.example.com/api/v1/unsubscribe?email=subscriber@example.com&token=") {
		t.Errorf("unexpected List-Unsubscribe header: %s", headers.ListUnsubscribe)
	}
	if headers.ListUnsubscribePost != "List-Unsubscribe=One-Click" {
		t.Errorf("unexpected List-Unsubscribe-Post header: %s", headers.ListUnsubscribePost)
	}

	// Extract token from URL for test
	tokenStart := strings.Index(headers.ListUnsubscribe, "token=") + 6
	tokenEnd := strings.Index(headers.ListUnsubscribe[tokenStart:], ">")
	token := headers.ListUnsubscribe[tokenStart : tokenStart+tokenEnd]

	// Trigger HTTP Handler
	req := httptest.NewRequest(http.MethodPost, "/api/v1/unsubscribe?email="+email+"&token="+token, nil)
	w := httptest.NewRecorder()
	um.HTTPHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	if !um.IsUnsubscribed(email) {
		t.Error("expected email to be recorded as unsubscribed")
	}
}
