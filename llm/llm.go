package llm

import (
    "context"
    "github.com/vyuvaraj/pranor/llm/api"
    "github.com/vyuvaraj/pranor/llm/pkg/router"
)

// DefaultRouter is the package-level router. Populated by init() in OSS/EE files.
var DefaultRouter api.Router = router.NewOSSRouter()

// Register adds a provider to DefaultRouter.
func Register(p api.ChatProvider) { DefaultRouter.Register(p) }

// Route routes a chat request through DefaultRouter.
func Route(ctx context.Context, req api.ChatRequest) (api.ChatResponse, error) {
    return DefaultRouter.Route(ctx, req)
}

// HealthCheck checks all providers in DefaultRouter.
func HealthCheck(ctx context.Context) map[string]error {
    return DefaultRouter.HealthCheck(ctx)
}
