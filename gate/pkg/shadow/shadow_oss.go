package shadow

import (
	"context"

	"github.com/vyuvaraj/pranor/core/pkg/capability"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

type ossInterceptor struct {
	registry capability.Registry
}

// NewOSSInterceptor returns an Interceptor instance.
func NewOSSInterceptor(reg capability.Registry) Interceptor {
	if reg == nil {
		reg = capability.DefaultRegistry
	}
	return &ossInterceptor{registry: reg}
}

func (i *ossInterceptor) IsShadowMode(ec *execctx.ExecutionContext) bool {
	if ec == nil {
		return false
	}
	if ec.PolicyContext != nil && ec.PolicyContext["mode"] == "SIMULATION" {
		return true
	}
	return false
}

func (i *ossInterceptor) InterceptCapability(ctx context.Context, ec *execctx.ExecutionContext, capID string, input map[string]any) (map[string]any, bool, error) {
	if !i.IsShadowMode(ec) {
		return nil, false, nil
	}

	capDef, err := i.registry.Lookup(capID)
	if err != nil {
		// If capability lookup fails in shadow mode, default to safe noop
		return map[string]any{
			"shadow_status": "[SHADOW_MODE_NOOP]",
			"simulated":     true,
			"capability":    capID,
		}, true, nil
	}

	// Side-effect detection: writes DB, sends notification, or external API calls
	isSideEffect := capDef.BlastRadius.WritesDB ||
		capDef.BlastRadius.SendsNotification ||
		capDef.BlastRadius.ExternalAPICalls ||
		capDef.Risk == capability.RiskHigh ||
		capDef.Risk == capability.RiskCritical

	if isSideEffect {
		return map[string]any{
			"shadow_status": "[SHADOW_MODE_NOOP]",
			"simulated":     true,
			"capability":    capID,
			"risk_level":    capDef.Risk.String(),
		}, true, nil
	}

	// Read operations are allowed to proceed through
	return nil, false, nil
}
