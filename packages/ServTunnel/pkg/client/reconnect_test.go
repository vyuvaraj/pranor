package client

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestTunnelReconnectClient_SuccessfulReconnect(t *testing.T) {
	attempts := 0
	mockConnect := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("connection refused")
		}
		return nil
	}

	cfg := ReconnectConfig{
		InitialInterval: 5 * time.Millisecond,
		MaxInterval:     20 * time.Millisecond,
		Multiplier:      1.5,
		Jitter:          true,
		MaxRetries:      5,
	}

	client := NewTunnelReconnectClient(cfg, mockConnect)

	err := client.StartLoop(context.Background())
	if err != nil {
		t.Fatalf("StartLoop failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("expected client to be connected after successful 3rd attempt")
	}
	if attempts != 3 {
		t.Errorf("expected 3 connect attempts, got %d", attempts)
	}
}
