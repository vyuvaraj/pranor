package tunnel

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ShareToken holds time-bounded or single-use tunnel access token metadata.
type ShareToken struct {
	Token     string    `json:"token"`
	TunnelID  string    `json:"tunnel_id"`
	SingleUse bool      `json:"single_use"`
	Used      bool      `json:"used"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ShareTokenManager provisions and validates shareable tunnel access URLs.
type ShareTokenManager struct {
	mu     sync.RWMutex
	tokens map[string]*ShareToken // tokenStr -> ShareToken
}

// NewShareTokenManager creates a ShareTokenManager instance.
func NewShareTokenManager() *ShareTokenManager {
	return &ShareTokenManager{
		tokens: make(map[string]*ShareToken),
	}
}

// CreateShareToken generates a shareable tunnel URL token with TTL and single-use options.
func (stm *ShareTokenManager) CreateShareToken(tunnelID string, ttl time.Duration, singleUse bool) (*ShareToken, string) {
	if ttl <= 0 {
		ttl = 1 * time.Hour
	}

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	tokenStr := hex.EncodeToString(b)

	now := time.Now()
	st := &ShareToken{
		Token:     tokenStr,
		TunnelID:  tunnelID,
		SingleUse: singleUse,
		Used:      false,
		ExpiresAt: now.Add(ttl),
	}

	stm.mu.Lock()
	stm.tokens[tokenStr] = st
	stm.mu.Unlock()

	shareURL := fmt.Sprintf("https://tn-%s.Pranor Deploy.dev?share_token=%s", tunnelID, tokenStr)
	return st, shareURL
}

// ValidateToken checks token validity, expiration, and enforces single-use constraints.
func (stm *ShareTokenManager) ValidateToken(tokenStr string) (string, error) {
	stm.mu.Lock()
	defer stm.mu.Unlock()

	st, ok := stm.tokens[tokenStr]
	if !ok {
		return "", fmt.Errorf("invalid share token")
	}

	if time.Now().After(st.ExpiresAt) {
		delete(stm.tokens, tokenStr)
		return "", fmt.Errorf("share token expired")
	}

	if st.SingleUse {
		if st.Used {
			return "", fmt.Errorf("one-time share token already consumed")
		}
		st.Used = true
	}

	return st.TunnelID, nil
}

// Middleware returns HTTP middleware enforcing share token verification.
func (stm *ShareTokenManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.URL.Query().Get("share_token")
		if tokenStr == "" {
			tokenStr = r.Header.Get("X-Share-Token")
		}

		tunnelID, err := stm.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("unauthorized tunnel access: %v", err), http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-Tunnel-Id", tunnelID)
		next.ServeHTTP(w, r)
	})
}
