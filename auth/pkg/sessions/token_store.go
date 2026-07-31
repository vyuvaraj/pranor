package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrInvalidUserID is returned when an empty userID is passed to Issue.
	ErrInvalidUserID = errors.New("user ID cannot be empty")
	// ErrInvalidToken is returned when an empty token is passed.
	ErrInvalidToken = errors.New("token cannot be empty")
	// ErrTokenNotFound is returned when looking up a token that does not exist.
	ErrTokenNotFound = errors.New("token not found")
	// ErrTokenExpired is returned when a token has passed its TTL.
	ErrTokenExpired = errors.New("token expired")
	// ErrTokenRevoked is returned when a token has been explicitly revoked.
	ErrTokenRevoked = errors.New("token revoked")
)

// DefaultTTL is the default time-to-live for issued tokens (7 days).
const DefaultTTL = 7 * 24 * time.Hour

type tokenEntry struct {
	userID    string
	createdAt time.Time
	expiresAt time.Time
	revoked   bool
}

// TokenStore is an in-memory, thread-safe opaque token store with auto-expiry and revocation.
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*tokenEntry
	ttl    time.Duration
}

// NewTokenStore creates a new TokenStore. If custom TTL is provided and > 0, it overrides DefaultTTL.
func NewTokenStore(ttl ...time.Duration) *TokenStore {
	storeTTL := DefaultTTL
	if len(ttl) > 0 && ttl[0] > 0 {
		storeTTL = ttl[0]
	}
	return &TokenStore{
		tokens: make(map[string]*tokenEntry),
		ttl:    storeTTL,
	}
}

// Issue generates a cryptographically random 32-byte hex token for the given userID.
func (ts *TokenStore) Issue(userID string) (string, error) {
	if userID == "" {
		return "", ErrInvalidUserID
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	token := hex.EncodeToString(bytes)
	now := time.Now()

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.tokens[token] = &tokenEntry{
		userID:    userID,
		createdAt: now,
		expiresAt: now.Add(ts.ttl),
		revoked:   false,
	}

	return token, nil
}

// Validate checks if a token is valid, active, non-expired, and non-revoked, returning the associated userID.
func (ts *TokenStore) Validate(token string) (string, error) {
	if token == "" {
		return "", ErrInvalidToken
	}

	ts.mu.RLock()
	entry, exists := ts.tokens[token]
	if !exists {
		ts.mu.RUnlock()
		return "", ErrTokenNotFound
	}

	if entry.revoked {
		ts.mu.RUnlock()
		return "", ErrTokenRevoked
	}

	expired := time.Now().After(entry.expiresAt)
	userID := entry.userID
	ts.mu.RUnlock()

	if expired {
		ts.mu.Lock()
		delete(ts.tokens, token)
		ts.mu.Unlock()
		return "", ErrTokenExpired
	}

	return userID, nil
}

// Revoke invalidates a token server-side.
func (ts *TokenStore) Revoke(token string) error {
	if token == "" {
		return ErrInvalidToken
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	entry, exists := ts.tokens[token]
	if !exists {
		return ErrTokenNotFound
	}

	entry.revoked = true
	return nil
}

// CleanExpired removes all expired tokens from the store and returns the count removed.
func (ts *TokenStore) CleanExpired() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now()
	cleaned := 0
	for token, entry := range ts.tokens {
		if now.After(entry.expiresAt) {
			delete(ts.tokens, token)
			cleaned++
		}
	}
	return cleaned
}
