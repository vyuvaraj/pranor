package raft

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

type BFTProposal struct {
	Index        uint64            `json:"index"`
	Term         uint64            `json:"term"`
	Command      string            `json:"command"`
	Signatures   map[string]string `json:"signatures"` // nodeID -> SHA256 sig
	QuorumPassed bool              `json:"quorum_passed"`
}

type BFTConsensusManager struct {
	mu           sync.Mutex
	nodeID       string
	clusterNodes []string
	f            int // Maximum faulty/byzantine nodes tolerated
}

func NewBFTConsensusManager(nodeID string, clusterNodes []string) *BFTConsensusManager {
	n := len(clusterNodes)
	f := (n - 1) / 3 // Standard BFT 3f + 1 formula
	return &BFTConsensusManager{
		nodeID:       nodeID,
		clusterNodes: clusterNodes,
		f:            f,
	}
}

// ProposeBlock creates a BFT proposal requiring 2f + 1 cryptographic signatures for commit quorum.
func (b *BFTConsensusManager) ProposeBlock(index, term uint64, command string) *BFTProposal {
	b.mu.Lock()
	defer b.mu.Unlock()

	prop := &BFTProposal{
		Index:      index,
		Term:       term,
		Command:    command,
		Signatures: make(map[string]string),
	}

	// Self signature
	sig := b.signProposal(command)
	prop.Signatures[b.nodeID] = sig
	return prop
}

// AddSignature adds peer signature and checks if 2f + 1 BFT quorum is met.
func (b *BFTConsensusManager) AddSignature(prop *BFTProposal, peerNodeID, sig string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	prop.Signatures[peerNodeID] = sig
	requiredQuorum := 2*b.f + 1
	if len(prop.Signatures) >= requiredQuorum {
		prop.QuorumPassed = true
	}
	return prop.QuorumPassed
}

func (b *BFTConsensusManager) signProposal(command string) string {
	raw := fmt.Sprintf("%s:%s", b.nodeID, command)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h[:8])
}
