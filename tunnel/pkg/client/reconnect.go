package import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ReconnectConfig configures exponential backoff parameters.
type ReconnectConfig struct {
	InitialInterval time.Duration `json:"initial_interval"` // e.g. 100ms
	MaxInterval     time.Duration `json:"max_interval"`     // e.g. 30s
	Multiplier      float64       `json:"multiplier"`       // e.g. 2.0
	Jitter          bool          `json:"jitter"`
	MaxRetries      int           `json:"max_retries"` // 0 for infinite
}

// TunnelReconnectClient manages persistent tunnel pipe reconnect attempts with randomized exponential backoff.
type TunnelReconnectClient struct {
	mu           sync.RWMutex
	cfg          ReconnectConfig
	connectFn    func(ctx context.Context) error
	attempts     int
	isConnected  bool
	rng          *rand.Rand
}

// NewTunnelReconnectClient creates a TunnelReconnectClient instance.
func NewTunnelReconnectClient(cfg ReconnectConfig, connectFn func(ctx context.Context) error) *TunnelReconnectClient {
	if cfg.InitialInterval <= 0 {
		cfg.InitialInterval = 10 * time.Millisecond
	}
	if cfg.MaxInterval <= 0 {
		cfg.MaxInterval = 1 * time.Second
	}
	if cfg.Multiplier <= 1.0 {
		cfg.Multiplier = 2.0
	}
	return &TunnelReconnectClient{
		cfg:       cfg,
		connectFn: connectFn,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// StartLoop continuously maintains tunnel pipe connection, reconnecting with exponential backoff on drop.
func (trc *TunnelReconnectClient) StartLoop(ctx context.Context) error {
	currentInterval := trc.cfg.InitialInterval
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		attempts++
		trc.mu.Lock()
		trc.attempts = attempts
		trc.mu.Unlock()

		err := trc.connectFn(ctx)
		if err == nil {
			trc.mu.Lock()
			trc.isConnected = true
			trc.attempts = 0
			trc.mu.Unlock()
			return nil
		}

		trc.mu.Lock()
		trc.isConnected = false
		trc.mu.Unlock()

		if trc.cfg.MaxRetries > 0 && attempts >= trc.cfg.MaxRetries {
			return fmt.Errorf("max reconnect attempts (%d) reached: %w", trc.cfg.MaxRetries, err)
		}

		// Calculate backoff interval
		backoff := time.Duration(float64(currentInterval) * trc.cfg.Multiplier)
		if backoff > trc.cfg.MaxInterval {
			backoff = trc.cfg.MaxInterval
		}

		if trc.cfg.Jitter {
			trc.mu.Lock()
			jitterFactor := 0.5 + trc.rng.Float64() // 0.5 to 1.5
			trc.mu.Unlock()
			backoff = time.Duration(float64(backoff) * jitterFactor)
		}

		currentInterval = backoff

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentInterval):
		}
	}
}

// IsConnected returns current connection status.
func (trc *TunnelReconnectClient) IsConnected() bool {
	trc.mu.RLock()
	defer trc.mu.RUnlock()
	return trc.isConnected
}
