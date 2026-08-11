//go:build !enterprise

package llm

import (
    "github.com/vyuvaraj/pranor/llm/pkg/providers"
)

func init() {
    // Register EchoProvider as the default OSS provider.
    DefaultRouter.Register(providers.NewEchoProvider())
    DefaultRouter.SetFallbackChain([]string{"echo"})
}
