package raft

import (
	"testing"
)

func TestBFTConsensusQuorum(t *testing.T) {
	cluster := []string{"node_1", "node_2", "node_3", "node_4"} // 4 nodes -> f = 1 -> 2f+1 = 3 quorum
	bft := NewBFTConsensusManager("node_1", cluster)

	prop := bft.ProposeBlock(1, 1, "ENQUEUE orders payload_99")

	if prop.QuorumPassed {
		t.Errorf("Expected quorum to NOT pass with 1 signature")
	}

	// Add second node signature
	bft.AddSignature(prop, "node_2", "sig_node_2")
	if prop.QuorumPassed {
		t.Errorf("Expected quorum to NOT pass with 2 signatures (need 3)")
	}

	// Add third node signature (quorum met!)
	passed := bft.AddSignature(prop, "node_3", "sig_node_3")
	if !passed || !prop.QuorumPassed {
		t.Errorf("Expected 2f+1 BFT quorum to pass with 3 signatures")
	}
}
