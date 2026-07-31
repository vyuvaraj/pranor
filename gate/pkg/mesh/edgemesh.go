//go:build !enterprise

package import (
	"context"
)

type EdgeMeshNode struct {
	NodeID string
	Region string
}

func NewEdgeMeshNode(nodeID, region string) *EdgeMeshNode {
	return &EdgeMeshNode{
		NodeID: nodeID,
		Region: region,
	}
}

func (m *EdgeMeshNode) SyncRoutes(ctx context.Context) error {
	return nil
}
