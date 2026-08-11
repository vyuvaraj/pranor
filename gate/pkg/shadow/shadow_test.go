package shadow_test

import (
	"context"
	"testing"

	"github.com/vyuvaraj/pranor/core/pkg/capability"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
	"github.com/vyuvaraj/pranor/gate/pkg/shadow"
)

func TestShadowInterceptor(t *testing.T) {
	reg := capability.NewInMemoryRegistry()
	_ = reg.Register(capability.Capability{
		ID:       "pool.read",
		Version:  "1.0.0",
		Risk:     capability.RiskLow,
		BlastRadius: capability.BlastRadius{WritesDB: false},
	})
	_ = reg.Register(capability.Capability{
		ID:       "pool.delete",
		Version:  "1.0.0",
		Risk:     capability.RiskCritical,
		BlastRadius: capability.BlastRadius{WritesDB: true},
	})

	interceptor := shadow.NewOSSInterceptor(reg)

	ctx := context.Background()
	ecReal := execctx.New(ctx, "t1", "a1", "u1")
	ecShadow := ecReal.WithPolicy("mode", "SIMULATION")

	// Test IsShadowMode
	if interceptor.IsShadowMode(ecReal) {
		t.Errorf("expected ecReal not to be in shadow mode")
	}
	if !interceptor.IsShadowMode(ecShadow) {
		t.Errorf("expected ecShadow to be in shadow mode")
	}

	// Read call in shadow mode -> passthrough (intercepted = false)
	_, intercepted, err := interceptor.InterceptCapability(ctx, ecShadow, "pool.read", nil)
	if err != nil || intercepted {
		t.Errorf("expected read call passthrough, got intercepted=%v err=%v", intercepted, err)
	}

	// Write call in shadow mode -> noop intercepted
	res, intercepted, err := interceptor.InterceptCapability(ctx, ecShadow, "pool.delete", nil)
	if err != nil || !intercepted {
		t.Fatalf("expected write call to be intercepted, got intercepted=%v err=%v", intercepted, err)
	}
	if res["shadow_status"] != "[SHADOW_MODE_NOOP]" {
		t.Errorf("expected [SHADOW_MODE_NOOP], got %v", res["shadow_status"])
	}
}
