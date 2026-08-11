package shadow

import (
	"context"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

// Interceptor intercepts capability calls for Gate-level side-effect isolation during shadow/simulation execution.
type Interceptor interface {
	IsShadowMode(ec *execctx.ExecutionContext) bool
	InterceptCapability(ctx context.Context, ec *execctx.ExecutionContext, capID string, input map[string]any) (map[string]any, bool, error)
}

// ResultAnnotation provides audit metadata for intercepted shadow operations.
type ResultAnnotation struct {
	Intercepted  bool
	Reason       string
	CapabilityID string
}
