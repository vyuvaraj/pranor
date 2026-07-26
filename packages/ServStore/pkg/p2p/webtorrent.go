package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

type P2PChunk struct {
	ChunkID   string
	Data      []byte
	Hash      string
	PeerID    string
}

type BrowserP2PMesh struct {
	peers          map[string]bool
	offloadBytes   uint64
	originBytes    uint64
	mu             sync.RWMutex
}

func NewBrowserP2PMesh() *BrowserP2PMesh {
	return &BrowserP2PMesh{
		peers: make(map[string]bool),
	}
}

func (p *BrowserP2PMesh) RegisterPeer(peerID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers[peerID] = true
}

func (p *BrowserP2PMesh) VerifyAndStoreChunk(ctx context.Context, chunk P2PChunk) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	h := sha256.Sum256(chunk.Data)
	computedHash := hex.EncodeToString(h[:])

	if chunk.Hash != "" && computedHash != chunk.Hash {
		return false, fmt.Errorf("p2p chunk integrity verification failed")
	}

	p.offloadBytes += uint64(len(chunk.Data))
	return true, nil
}

func (p *BrowserP2PMesh) GetOffloadStats() (uint64, uint64, float64) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := p.offloadBytes + p.originBytes
	if total == 0 {
		return 0, 0, 0.0
	}
	ratio := (float64(p.offloadBytes) / float64(total)) * 100.0
	return p.offloadBytes, p.originBytes, ratio
}
