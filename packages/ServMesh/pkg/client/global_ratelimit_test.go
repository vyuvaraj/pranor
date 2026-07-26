package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type MockCacheStore struct {
	mu   sync.Mutex
	data map[string]interface{}
}

func NewMockCacheStore() *MockCacheStore {
	return &MockCacheStore{data: make(map[string]interface{})}
}

func (m *MockCacheStore) Get(key string) (interface{}, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	return val, ok, nil
}

func (m *MockCacheStore) Set(key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func TestGlobalRateLimiter_Allow(t *testing.T) {
	cacheStore := NewMockCacheStore()
	limiter := NewGlobalRateLimiter(cacheStore, 2, 1000) // max 2 tokens

	ctx := context.Background()
	pair := "serviceA->serviceB"

	// 1st request - allowed (tokens left: 1)
	ok, err := limiter.Allow(ctx, pair)
	if err != nil || !ok {
		t.Fatalf("expected 1st request allowed, got ok=%v, err=%v", ok, err)
	}

	// 2nd request - allowed (tokens left: 0)
	ok, err = limiter.Allow(ctx, pair)
	if err != nil || !ok {
		t.Fatalf("expected 2nd request allowed, got ok=%v, err=%v", ok, err)
	}

	// 3rd request - rate limited
	ok, err = limiter.Allow(ctx, pair)
	if err != nil || ok {
		t.Errorf("expected 3rd request to be blocked by global rate limiter")
	}
}

func TestGlobalRateLimiter_Middleware(t *testing.T) {
	cacheStore := NewMockCacheStore()
	limiter := NewGlobalRateLimiter(cacheStore, 1, 1000)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := limiter.Middleware(next)

	// 1st request -> 200 OK
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req1.Header.Set("X-Caller-Id", "web-app")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", w1.Code)
	}

	// 2nd request -> 429 Too Many Requests
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req2.Header.Set("X-Caller-Id", "web-app")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected HTTP 429, got %d", w2.Code)
	}
}
