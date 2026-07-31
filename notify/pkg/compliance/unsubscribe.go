package compliance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// UnsubscribeHeaders contains RFC 8058 compliant email headers.
type UnsubscribeHeaders struct {
	ListUnsubscribe     string `json:"List-Unsubscribe"`
	ListUnsubscribePost string `json:"List-Unsubscribe-Post"`
}

// UnsubscribeManager handles RFC 8058 one-click unsubscribe links and HMAC verification.
type UnsubscribeManager struct {
	mu           sync.RWMutex
	secretKey    []byte
	baseURL      string
	unsubscribed map[string]bool // email -> status
}

// NewUnsubscribeManager creates an UnsubscribeManager instance.
func NewUnsubscribeManager(baseURL, secretKey string) *UnsubscribeManager {
	if baseURL == "" {
		baseURL = "https://mail.Pranor Deploy.dev"
	}
	if secretKey == "" {
		secretKey = "default-Pranor Notify-unsubscribe-secret"
	}
	return &UnsubscribeManager{
		secretKey:    []byte(secretKey),
		baseURL:      strings.TrimRight(baseURL, "/"),
		unsubscribed: make(map[string]bool),
	}
}

// GenerateHeaders creates RFC 8058 List-Unsubscribe and List-Unsubscribe-Post headers.
func (um *UnsubscribeManager) GenerateHeaders(email string) UnsubscribeHeaders {
	token := um.generateToken(email)
	unsubURL := fmt.Sprintf("%s/api/v1/unsubscribe?email=%s&token=%s", um.baseURL, email, token)
	mailto := fmt.Sprintf("<mailto:unsubscribe@%s?subject=unsubscribe>", getDomainFromURL(um.baseURL))

	return UnsubscribeHeaders{
		ListUnsubscribe:     fmt.Sprintf("%s, <%s>", mailto, unsubURL),
		ListUnsubscribePost: "List-Unsubscribe=One-Click",
	}
}

// VerifyAndUnsubscribe verifies HMAC token and records recipient unsubscribe.
func (um *UnsubscribeManager) VerifyAndUnsubscribe(email, token string) bool {
	if !um.verifyToken(email, token) {
		return false
	}

	um.mu.Lock()
	defer um.mu.Unlock()
	um.unsubscribed[strings.ToLower(email)] = true
	return true
}

// IsUnsubscribed checks if recipient has unsubscribed.
func (um *UnsubscribeManager) IsUnsubscribed(email string) bool {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.unsubscribed[strings.ToLower(email)]
}

// HTTPHandler exposes POST /api/v1/unsubscribe handler for RFC 8058 POST requests.
func (um *UnsubscribeManager) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		token := r.URL.Query().Get("token")

		if email == "" || token == "" {
			http.Error(w, "missing email or token", http.StatusBadRequest)
			return
		}

		if !um.VerifyAndUnsubscribe(email, token) {
			http.Error(w, "invalid or expired token", http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Unsubscribed successfully"))
	})
}

func (um *UnsubscribeManager) generateToken(email string) string {
	h := hmac.New(sha256.New, um.secretKey)
	h.Write([]byte(strings.ToLower(email)))
	return hex.EncodeToString(h.Sum(nil))
}

func (um *UnsubscribeManager) verifyToken(email, token string) bool {
	expected := um.generateToken(email)
	return hmac.Equal([]byte(expected), []byte(token))
}

func getDomainFromURL(url string) string {
	parts := strings.Split(url, "//")
	if len(parts) > 1 {
		return strings.Split(parts[1], "/")[0]
	}
	return "Pranor Deploy.dev"
}
