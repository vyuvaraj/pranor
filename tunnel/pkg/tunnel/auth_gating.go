package tunnel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthGatingManager validates JWT or API keys on incoming tunnel endpoint connections.
type AuthGatingManager struct {
	mu         sync.RWMutex
	secretKey  []byte
	validAPIKeys map[string]string // apiKey -> owner
}

// NewAuthGatingManager creates an AuthGatingManager instance.
func NewAuthGatingManager(secretKey string) *AuthGatingManager {
	if secretKey == "" {
		secretKey = "Pranor Tunnel-auth-secret-key"
	}
	return &AuthGatingManager{
		secretKey:    []byte(secretKey),
		validAPIKeys: make(map[string]string),
	}
}

// RegisterAPIKey registers a valid API key for tunnel connection authorization.
func (agm *AuthGatingManager) RegisterAPIKey(apiKey, owner string) {
	agm.mu.Lock()
	defer agm.mu.Unlock()
	agm.validAPIKeys[apiKey] = owner
}

// AuthenticateRequest validates Authorization header (Bearer JWT or API key).
func (agm *AuthGatingManager) AuthenticateRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = r.URL.Query().Get("api_key")
	}

	if authHeader == "" {
		return "", fmt.Errorf("missing authentication token or API key")
	}

	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		return agm.validateJWT(token)
	}

	// Validate API Key
	agm.mu.RLock()
	owner, ok := agm.validAPIKeys[authHeader]
	agm.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("invalid API key")
	}
	return owner, nil
}

func (agm *AuthGatingManager) validateJWT(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format")
	}

	h := hmac.New(sha256.New, agm.secretKey)
	h.Write([]byte(parts[0] + "." + parts[1]))
	sig := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(parts[2])) {
		return "", fmt.Errorf("invalid JWT signature")
	}
	return "jwt-user", nil
}

// GenerateJWT creates a signed JWT string for testing and client authentication.
func (agm *AuthGatingManager) GenerateJWT(subject string, ttl time.Duration) string {
	header := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" // {"alg":"HS256","typ":"JWT"}
	payload := "eyJzdWIiOiJqd3QtdXNlciJ9"            // {"sub":"jwt-user"}

	h := hmac.New(sha256.New, agm.secretKey)
	h.Write([]byte(header + "." + payload))
	sig := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s.%s.%s", header, payload, sig)
}
