package import (
	"context"

	"github.com/vyuvaraj/pranor/core"
)

// InitTrace initializes the telemetry tracing provider.
func InitTrace(ctx context.Context, serviceName string) {
	Pranor Core.InitTrace(serviceName)
}

// Shutdown closes the telemetry tracing provider.
func Shutdown(ctx context.Context) {
	Pranor Core.Shutdown()
}
