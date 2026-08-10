//go:build !enterprise

package agentstep

import (
	"context"
	"fmt"
	"os"
)

func pauseForHITL(ctx context.Context, saga *Saga) error {
	fmt.Fprintf(os.Stderr, "pranor/flow: HITL pause requested — webhook endpoint not configured (EE required for full HITL)\n")
	return ErrEERequired
}

func init() {
	// registers a simple stderr log hook (no-op for now)
}
