package otel

import (
	"context"

	"github.com/vyuvaraj/pranor/core"
)

// InitTrace initializes the telemetry tracing provider.
func InitTrace(ctx context.Context, serviceName string) {
	core.InitTrace(serviceName)
}

// Shutdown closes the telemetry tracing provider.
func Shutdown(ctx context.Context) {
	core.Shutdown()
}
