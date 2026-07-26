package main_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/vyuvaraj/serv/packages/ServAuth/pkg/security"
	"github.com/vyuvaraj/serv/packages/ServAuth/pkg/sessions"
)

// ============================================================================
// SA.G1: Opaque Session Token Store Tests
// ============================================================================

// --- Tier 1: Feature Coverage (SA.G1) ---

func TestE2E_SA_G1_Tier1_IssueToken_Success(t *testing.T) {
	store := sessions.NewTokenStore()
	userID := "usr_alice_123"

	token, err := store.Issue(userID)
	if err != nil {
		t.Fatalf("expected Issue to succeed, got error: %v", err)
	}
	if len(token) != 64 { // 32-byte hex string is 64 hex characters
		t.Errorf("expected 64-char hex token, got length %d (%s)", len(token), token)
	}
}

func TestE2E_SA_G1_Tier1_ValidateToken_Active(t *testing.T) {
	store := sessions.NewTokenStore()
	userID := "usr_bob_456"

	token, err := store.Issue(userID)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	gotUserID, err := store.Validate(token)
	if err != nil {
		t.Fatalf("expected Validate to succeed, got: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("expected userID %q, got %q", userID, gotUserID)
	}
}

func TestE2E_SA_G1_Tier1_RevokeToken_Success(t *testing.T) {
	store := sessions.NewTokenStore()
	userID := "usr_charlie_789"

	token, err := store.Issue(userID)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	err = store.Revoke(token)
	if err != nil {
		t.Fatalf("expected Revoke to succeed, got: %v", err)
	}

	_, err = store.Validate(token)
	if !errors.Is(err, sessions.ErrTokenRevoked) {
		t.Errorf("expected ErrTokenRevoked after revocation, got: %v", err)
	}
}

func TestE2E_SA_G1_Tier1_MultipleUsers_Isolation(t *testing.T) {
	store := sessions.NewTokenStore()
	users := []string{"usr_1", "usr_2", "usr_3", "usr_4", "usr_5"}
	userTokens := make(map[string]string)

	for _, u := range users {
		token, err := store.Issue(u)
		if err != nil {
			t.Fatalf("failed to issue token for %s: %v", u, err)
		}
		userTokens[u] = token
	}

	for u, token := range userTokens {
		validatedUser, err := store.Validate(token)
		if err != nil {
			t.Errorf("failed to validate token for %s: %v", u, err)
		}
		if validatedUser != u {
			t.Errorf("expected %s, got %s", u, validatedUser)
		}
	}
}

func TestE2E_SA_G1_Tier1_CleanExpired_RemovesTokens(t *testing.T) {
	shortTTL := 50 * time.Millisecond
	store := sessions.NewTokenStore(shortTTL)

	token1, err := store.Issue("usr_expired_1")
	if err != nil {
		t.Fatalf("failed to issue token 1: %v", err)
	}
	token2, err := store.Issue("usr_expired_2")
	if err != nil {
		t.Fatalf("failed to issue token 2: %v", err)
	}

	time.Sleep(70 * time.Millisecond)

	cleaned := store.CleanExpired()
	if cleaned < 2 {
		t.Errorf("expected at least 2 tokens cleaned, got %d", cleaned)
	}

	_, err = store.Validate(token1)
	if !errors.Is(err, sessions.ErrTokenNotFound) && !errors.Is(err, sessions.ErrTokenExpired) {
		t.Errorf("expected token1 to be missing or expired after clean, got %v", err)
	}
	_, err = store.Validate(token2)
	if !errors.Is(err, sessions.ErrTokenNotFound) && !errors.Is(err, sessions.ErrTokenExpired) {
		t.Errorf("expected token2 to be missing or expired after clean, got %v", err)
	}
}

// --- Tier 2: Boundary & Corner Cases (SA.G1) ---

func TestE2E_SA_G1_Tier2_Issue_EmptyUserID(t *testing.T) {
	store := sessions.NewTokenStore()
	_, err := store.Issue("")
	if !errors.Is(err, sessions.ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID when issuing with empty userID, got: %v", err)
	}
}

func TestE2E_SA_G1_Tier2_Validate_EmptyToken(t *testing.T) {
	store := sessions.NewTokenStore()
	_, err := store.Validate("")
	if !errors.Is(err, sessions.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for empty token, got: %v", err)
	}
}

func TestE2E_SA_G1_Tier2_Validate_NonExistentToken(t *testing.T) {
	store := sessions.NewTokenStore()
	dummyToken := "0000000000000000000000000000000000000000000000000000000000000000"
	_, err := store.Validate(dummyToken)
	if !errors.Is(err, sessions.ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound for nonexistent token, got: %v", err)
	}
}

func TestE2E_SA_G1_Tier2_Revoke_EmptyOrInvalidToken(t *testing.T) {
	store := sessions.NewTokenStore()
	err := store.Revoke("")
	if !errors.Is(err, sessions.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken when revoking empty string, got: %v", err)
	}

	err = store.Revoke("nonexistent_token_string")
	if !errors.Is(err, sessions.ErrTokenNotFound) {
		t.Errorf("expected ErrTokenNotFound when revoking nonexistent token, got: %v", err)
	}
}

func TestE2E_SA_G1_Tier2_Validate_ExpiredToken(t *testing.T) {
	store := sessions.NewTokenStore(20 * time.Millisecond)
	token, err := store.Issue("usr_short_lived")
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	_, err = store.Validate(token)
	if !errors.Is(err, sessions.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired for expired token, got: %v", err)
	}
}

// ============================================================================
// SA.G6: Credential Stuffing Velocity Limiter Tests
// ============================================================================

// --- Tier 1: Feature Coverage (SA.G6) ---

func TestE2E_SA_G6_Tier1_RecordFailure_UnderThreshold(t *testing.T) {
	limiter := security.NewVelocityLimiter(1*time.Minute, 5, 15*time.Minute)
	key := "ip:192.168.1.100"

	for i := 0; i < 4; i++ {
		limiter.RecordFailure(key)
	}

	if limiter.IsBlocked(key) {
		t.Errorf("key should not be blocked with 4 failures (threshold 5)")
	}
}

func TestE2E_SA_G6_Tier1_Block_AtThreshold(t *testing.T) {
	limiter := security.NewVelocityLimiter(1*time.Minute, 5, 15*time.Minute)
	key := "usr:attacker"

	for i := 0; i < 5; i++ {
		limiter.RecordFailure(key)
	}

	if !limiter.IsBlocked(key) {
		t.Errorf("key should be blocked after 5 failures")
	}
}

func TestE2E_SA_G6_Tier1_Reset_Unblocks(t *testing.T) {
	limiter := security.NewVelocityLimiter(1*time.Minute, 3, 10*time.Minute)
	key := "usr:locked_out"

	for i := 0; i < 3; i++ {
		limiter.RecordFailure(key)
	}
	if !limiter.IsBlocked(key) {
		t.Fatalf("expected key to be blocked before reset")
	}

	limiter.Reset(key)

	if limiter.IsBlocked(key) {
		t.Errorf("expected key to be unblocked after Reset()")
	}
}

func TestE2E_SA_G6_Tier1_KeyIsolation_IPvsUser(t *testing.T) {
	limiter := security.NewVelocityLimiter(1*time.Minute, 3, 10*time.Minute)
	ipKey := "ip:10.0.0.1"
	userKey := "usr:alice"

	for i := 0; i < 3; i++ {
		limiter.RecordFailure(ipKey)
	}

	if !limiter.IsBlocked(ipKey) {
		t.Errorf("ipKey should be blocked")
	}
	if limiter.IsBlocked(userKey) {
		t.Errorf("userKey should not be affected by failures on ipKey")
	}
}

func TestE2E_SA_G6_Tier1_BlockDuration_Expiry(t *testing.T) {
	shortBlock := 40 * time.Millisecond
	limiter := security.NewVelocityLimiter(1*time.Minute, 2, shortBlock)
	key := "usr:temp_block"

	limiter.RecordFailure(key)
	limiter.RecordFailure(key)

	if !limiter.IsBlocked(key) {
		t.Fatalf("expected key to be blocked immediately")
	}

	time.Sleep(55 * time.Millisecond)

	if limiter.IsBlocked(key) {
		t.Errorf("expected key block to expire after %v", shortBlock)
	}
}

// --- Tier 2: Boundary & Corner Cases (SA.G6) ---

func TestE2E_SA_G6_Tier2_EmptyKey_Ignored(t *testing.T) {
	limiter := security.NewVelocityLimiter(1*time.Minute, 3, 10*time.Minute)
	limiter.RecordFailure("")
	if limiter.IsBlocked("") {
		t.Errorf("empty key should never be blocked")
	}
	limiter.Reset("")
}

func TestE2E_SA_G6_Tier2_SlidingWindow_Expiry(t *testing.T) {
	shortWindow := 50 * time.Millisecond
	limiter := security.NewVelocityLimiter(shortWindow, 3, 10*time.Minute)
	key := "ip:slow_bruteforce"

	limiter.RecordFailure(key)
	limiter.RecordFailure(key)

	time.Sleep(65 * time.Millisecond) // wait for window to expire

	limiter.RecordFailure(key) // 1 failure in new window

	if limiter.IsBlocked(key) {
		t.Errorf("key should not be blocked since old failures expired outside sliding window")
	}
}

func TestE2E_SA_G6_Tier2_ZeroOrNegativeConfig_Defaults(t *testing.T) {
	limiter := security.NewVelocityLimiter(0, -1, -5*time.Minute)
	if limiter.GetWindowDuration() != 1*time.Minute {
		t.Errorf("expected default window duration 1m, got %v", limiter.GetWindowDuration())
	}
	if limiter.GetMaxAttempts() != 5 {
		t.Errorf("expected default max attempts 5, got %d", limiter.GetMaxAttempts())
	}
	if limiter.GetBlockDuration() != 15*time.Minute {
		t.Errorf("expected default block duration 15m, got %v", limiter.GetBlockDuration())
	}
}

func TestE2E_SA_G6_Tier2_RecordFailure_WhileBlocked(t *testing.T) {
	limiter := security.NewVelocityLimiter(1*time.Minute, 2, 10*time.Minute)
	key := "ip:persistent_attacker"

	limiter.RecordFailure(key)
	limiter.RecordFailure(key)
	if !limiter.IsBlocked(key) {
		t.Fatalf("expected key to be blocked")
	}

	// Record further failure while already blocked
	limiter.RecordFailure(key)
	if !limiter.IsBlocked(key) {
		t.Errorf("key should remain blocked after recording failure while blocked")
	}
}

func TestE2E_SA_G6_Tier2_ConcurrentFailures_Safety(t *testing.T) {
	limiter := security.NewVelocityLimiter(1*time.Minute, 50, 10*time.Minute)
	key := "ip:concurrent_hammer"

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.RecordFailure(key)
			_ = limiter.IsBlocked(key)
		}()
	}
	wg.Wait()

	if !limiter.IsBlocked(key) {
		t.Errorf("expected key to be blocked after 100 concurrent failure recordings")
	}
}

// ============================================================================
// Tier 3: Cross-Feature Combinations (SA.G1 + SA.G6)
// ============================================================================

func TestE2E_SA_Tier3_Cross_TokenStoreAndVelocityLimiter(t *testing.T) {
	store := sessions.NewTokenStore()
	limiter := security.NewVelocityLimiter(1*time.Minute, 3, 5*time.Minute)

	userID := "usr_hybrid_1"
	ipKey := "ip:198.51.100.42"

	// Simulate 2 failed login attempts
	limiter.RecordFailure(ipKey)
	limiter.RecordFailure(ipKey)
	if limiter.IsBlocked(ipKey) {
		t.Fatalf("IP should not be blocked after 2 failures")
	}

	// 3rd attempt succeeds -> issue token
	token, err := store.Issue(userID)
	if err != nil {
		t.Fatalf("token issue failed: %v", err)
	}

	// Validate token works
	valUser, err := store.Validate(token)
	if err != nil || valUser != userID {
		t.Fatalf("token validation failed: %v", err)
	}

	// Later, user revokes token (logout)
	if err := store.Revoke(token); err != nil {
		t.Fatalf("token revocation failed: %v", err)
	}

	// An attacker tries using revoked token -> validation fails, record failure on IP
	_, err = store.Validate(token)
	if !errors.Is(err, sessions.ErrTokenRevoked) {
		t.Fatalf("expected ErrTokenRevoked, got %v", err)
	}
	limiter.RecordFailure(ipKey)

	// Now IP exceeds threshold (3 failures total) -> blocked
	if !limiter.IsBlocked(ipKey) {
		t.Errorf("expected IP %s to be blocked after reaching 3 failures", ipKey)
	}
}

// ============================================================================
// Tier 4: Real-World Application Scenarios (ServAuth)
// ============================================================================

func TestE2E_SA_Tier4_Scenario_UserSessionLifecycle(t *testing.T) {
	// Scenario: User Authentication & Session Security Flow
	// 1. Attacker attempts login with wrong credentials -> recorded in VelocityLimiter
	// 2. Legitimate user logs in successfully -> gets session token from TokenStore
	// 3. User performs authenticated actions -> token validated
	// 4. Attacker attempts credential stuffing on same username -> velocity limiter blocks username
	// 5. Legitimate user logs out -> session token revoked
	// 6. Reuse of revoked session token is rejected

	limiter := security.NewVelocityLimiter(5*time.Minute, 3, 15*time.Minute)
	store := sessions.NewTokenStore(1*time.Hour)

	username := "user_enterprise_corp"
	clientIP := "203.0.113.5"

	// Step 1: Failed attempt
	limiter.RecordFailure(clientIP)
	if limiter.IsBlocked(clientIP) {
		t.Fatalf("Step 1 failed: client IP prematurely blocked")
	}

	// Step 2: Successful auth
	token, err := store.Issue(username)
	if err != nil {
		t.Fatalf("Step 2 failed: token issue error: %v", err)
	}

	// Step 3: API Request with valid token
	activeUser, err := store.Validate(token)
	if err != nil || activeUser != username {
		t.Fatalf("Step 3 failed: session validation error: %v", err)
	}

	// Step 4: Attacker brute-forces account
	limiter.RecordFailure(username)
	limiter.RecordFailure(username)
	limiter.RecordFailure(username)
	if !limiter.IsBlocked(username) {
		t.Errorf("Step 4 failed: account %s should be blocked after 3 failures", username)
	}

	// Step 5: User logs out
	if err := store.Revoke(token); err != nil {
		t.Fatalf("Step 5 failed: token revoke error: %v", err)
	}

	// Step 6: Verify revoked token rejected
	_, err = store.Validate(token)
	if !errors.Is(err, sessions.ErrTokenRevoked) {
		t.Errorf("Step 6 failed: expected ErrTokenRevoked, got: %v", err)
	}
}
