//go:build !enterprise

package mirror

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type EventRecord struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	NodeID    string    `json:"node_id"`
}

type CRDTState struct {
	mu      sync.RWMutex
	records map[string]EventRecord
}

func NewCRDTState() *CRDTState {
	return &CRDTState{
		records: make(map[string]EventRecord),
	}
}

func (c *CRDTState) Merge(remote EventRecord) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	existing, found := c.records[remote.ID]
	if !found || remote.Timestamp.After(existing.Timestamp) {
		c.records[remote.ID] = remote
		return true
	}
	return false
}

type MirrorWorker struct {
	NodeID      string
	RemoteURLs  []string
	State       *CRDTState
	mu          sync.Mutex
	syncedCount uint64
}

func NewMirrorWorker(nodeID string, remoteURLs []string) *MirrorWorker {
	return &MirrorWorker{
		NodeID:     nodeID,
		RemoteURLs: remoteURLs,
		State:      NewCRDTState(),
	}
}

func (m *MirrorWorker) SyncEvent(ctx context.Context, topic, id, payload string) (bool, error) {
	m.mu.Lock()
	m.syncedCount++
	m.mu.Unlock()

	rec := EventRecord{
		ID:        id,
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
		NodeID:    m.NodeID,
	}

	updated := m.State.Merge(rec)
	return updated, nil
}

func (m *MirrorWorker) GetSyncedCount() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.syncedCount
}

func (m *MirrorWorker) Status() string {
	return fmt.Sprintf("Node %s active-active mirroring %d remote endpoints", m.NodeID, len(m.RemoteURLs))
}
