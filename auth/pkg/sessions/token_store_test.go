package import (
	"encoding/hex"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestTokenStore_IssueAndValidate(t *testing.T) {
	ts := NewTokenStore()

	// Happy path: Issue token
	userID := "user-123"
	token, err := ts.Issue(userID)
	if err != nil {
		t.Fatalf("expected no error issuing token, got: %v", err)
	}

	if len(token) != 64 { // 32 bytes hex encoded = 64 hex characters
		t.Errorf("expected 64 char hex string token, got len %d (%s)", len(token), token)
	}

	decoded, err := hex.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		t.Errorf("expected 32 decoded bytes, got err=%v len=%d", err, len(decoded))
	}

	// Validate valid token
	gotUserID, err := ts.Validate(token)
	if err != nil {
		t.Fatalf("expected no error validating token, got: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("expected userID %s, got %s", userID, gotUserID)
	}
}

func TestTokenStore_Issue_InvalidUserID(t *testing.T) {
	ts := NewTokenStore()

	_, err := ts.Issue("")
	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got: %v", err)
	}
}

func TestTokenStore_Validate_Errors(t *testing.T) {
	ts := NewTokenStore()

	// Empty token
	_, err := ts.Validate("")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got: %v", err)
	}

	// Non-existent token
	_, err = ts.Validate("nonexistenttoken123456789012345678901234567890123456789012345678")
	if !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound, got: %v", err)
	}
}

func TestTokenStore_Revoke(t *testing.T) {
	ts := NewTokenStore()

	userID := "user-456"
	token, err := ts.Issue(userID)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Revoke empty token error
	if err := ts.Revoke(""); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken on empty revoke, got: %v", err)
	}

	// Revoke nonexistent token error
	if err := ts.Revoke("unknown"); !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound on unknown revoke, got: %v", err)
	}

	// Revoke valid token
	err = ts.Revoke(token)
	if err != nil {
		t.Fatalf("expected no error revoking token, got: %v", err)
	}

	// Subsequent validate must fail with ErrTokenRevoked
	_, err = ts.Validate(token)
	if !errors.Is(err, ErrTokenRevoked) {
		t.Errorf("expected ErrTokenRevoked after revocation, got: %v", err)
	}
}

func TestTokenStore_TTLExpiry(t *testing.T) {
	// Create token store with very short TTL
	shortTTL := 50 * time.Millisecond
	ts := NewTokenStore(shortTTL)

	token, err := ts.Issue("user-789")
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Immediately valid
	userID, err := ts.Validate(token)
	if err != nil || userID != "user-789" {
		t.Fatalf("expected valid token immediately, got err=%v, userID=%s", err, userID)
	}

	// Wait for TTL to expire
	time.Sleep(70 * time.Millisecond)

	// Subsequent validate should fail with ErrTokenExpired
	_, err = ts.Validate(token)
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired after TTL, got: %v", err)
	}

	// Re-check validate returns ErrTokenNotFound since it was purged
	_, err = ts.Validate(token)
	if !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound after purge, got: %v", err)
	}
}

func TestTokenStore_CleanExpired(t *testing.T) {
	ts := NewTokenStore(30 * time.Millisecond)

	t1, _ := ts.Issue("u1")
	t2, _ := ts.Issue("u2")

	time.Sleep(50 * time.Millisecond)

	cleaned := ts.CleanExpired()
	if cleaned != 2 {
		t.Errorf("expected 2 cleaned tokens, got %d", cleaned)
	}

	_, err := ts.Validate(t1)
	if !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound for t1, got %v", err)
	}
	_, err = ts.Validate(t2)
	if !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound for t2, got %v", err)
	}
}

func TestTokenStore_Concurrency(t *testing.T) {
	ts := NewTokenStore(1 * time.Second)
	const numGoroutines = 20
	const numOps = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				token, err := ts.Issue("user-concurrent")
				if err != nil {
					t.Errorf("concurrent Issue failed: %v", err)
					return
				}
				_, _ = ts.Validate(token)
				if j%2 == 0 {
					_ = ts.Revoke(token)
				}
			}
		}(i)
	}

	wg.Wait()
}
