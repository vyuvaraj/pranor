//go:build !enterprise

package import (
	"context"
)

type GatewayController struct {
	Namespace string
}

func NewGatewayController(ns string) *GatewayController {
	return &GatewayController{
		Namespace: ns,
	}
}

func (c *GatewayController) ReconcileHTTPRoutes(ctx context.Context) error {
	return nil
}
