# Changes Report — M1: Pranor Auth Implementation

## Files Modified & Created

### 1. `packages/Pranor Auth/pkg/sessions/token_store.go`
- **Purpose**: Implementation of opaque session token store (SA.G1).
- **Exported Structs & Methods**:
  - `TokenStore` struct
  - `NewTokenStore(ttl ...time.Duration) *TokenStore`
  - `Issue(userID string) (string, error)`
  - `Validate(token string) (string, error)`
  - `Revoke(token string) error`
  - `CleanExpired() int`
- **Key Features**:
  - Cryptographically random 32-byte hex token generation via `crypto/rand`.
  - In-memory storage using `map[string]*tokenEntry`.
  - Default 7-day TTL with configurable custom TTL.
  - Full thread safety protected by `sync.RWMutex`.
  - Explicit server-side revocation resulting in `ErrTokenRevoked` on `Validate`.
  - Auto-expiry handling returning `ErrTokenExpired`.

### 2. `packages/Pranor Auth/pkg/sessions/token_store_test.go`
- **Purpose**: Unit test suite for `TokenStore`.
- **Test Coverage**:
  - Token issuance and 32-byte hex encoding verification.
  - Validation of active session tokens.
  - Validation of revoked session tokens (`ErrTokenRevoked`).
  - Validation of expired session tokens (`ErrTokenExpired`).
  - Validation error handling for empty/non-existent tokens.
  - Concurrent usage safety across multiple goroutines (`go test -race`).

### 3. `packages/Pranor Auth/pkg/security/velocity_limiter.go`
- **Purpose**: Implementation of credential stuffing sliding-window rate limiter (SA.G6).
- **Exported Structs & Methods**:
  - `VelocityLimiter` struct
  - `NewVelocityLimiter(windowDuration time.Duration, maxAttempts int, blockDuration time.Duration) *VelocityLimiter`
  - `RecordFailure(key string)`
  - `IsBlocked(key string) bool`
  - `Reset(key string)`
  - Config getters (`GetWindowDuration`, `GetMaxAttempts`, `GetBlockDuration`)
- **Key Features**:
  - Sliding-window tracking of failed login attempts per key (IP / username).
  - In-memory counters tracking timestamps within sliding window.
  - Configurable window duration, max attempt threshold, and block duration with safe defaults.
  - Thread safety using `sync.RWMutex`.

### 4. `packages/Pranor Auth/pkg/security/velocity_limiter_test.go`
- **Purpose**: Unit test suite for `VelocityLimiter`.
- **Test Coverage**:
  - Threshold blocking after N failure attempts.
  - Immediate unblocking via `Reset(key)`.
  - Sliding-window expiration of old failed attempt timestamps.
  - Block duration expiration unblocking.
  - Key isolation (separate tracking for different IPs/usernames).
  - Default parameter fallbacks and empty key safety.
  - Thread safety under concurrent load (`go test -race`).
