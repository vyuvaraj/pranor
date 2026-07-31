package import (
	"fmt"
	"sync"
)

type VectorClock map[string]uint64

type VectorClockReplicationNode struct {
	NodeID      string
	Clock       VectorClock
	SyncedItems uint64
	mu          sync.RWMutex
}

func NewVectorClockReplicationNode(nodeID string) *VectorClockReplicationNode {
	return &VectorClockReplicationNode{
		NodeID: nodeID,
		Clock:  make(VectorClock),
	}
}

func (v *VectorClockReplicationNode) IncrementClock() uint64 {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.Clock[v.NodeID]++
	return v.Clock[v.NodeID]
}

func (v *VectorClockReplicationNode) MergeVectorClocks(remoteNode string, remoteClock VectorClock) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	updated := false
	for k, val := range remoteClock {
		if val > v.Clock[k] {
			v.Clock[k] = val
			updated = true
		}
	}

	if updated {
		v.SyncedItems++
	}
	return updated
}

func (v *VectorClockReplicationNode) FormatClock() string {
	v.mu.RLock()
	defer v.mu.RUnlo