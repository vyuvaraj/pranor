package import (
	"fmt"
	"testing"
)

func TestQuorumWriteConfirmation(t *testing.T) {
	codec, err := NewErasureCodec(4, 2)
	if err != nil {
		t.Fatalf("Failed to create codec: %v", err)
	}

	data := []byte("Hello World! Pranor Vault Erasure Quorum Testing Data")
	shards, err := codec.Encode(data)
	if err != nil {
		t.Fatalf("Failed to encode data: %v", err)
	}

	// 1. Successful Quorum write (all shards succeeded)
	status, err := codec.WriteWithQuorum(shards, 4, func(idx int, shard []byte) error {
		return nil
	})
	if err != nil || !status.QuorumSatisfied || status.VerifiedShards != 6 {
		t.Fatalf("Expected successful quorum write, got status: %+v, err: %v", status, err)
	}

	// 2. Partial failure but Quorum satisfied (1 failure out of 6)
	status, err = codec.WriteWithQuorum(shards, 4, func(idx int, shard []byte) error {
		if idx == 5 {
			return fmt.Errorf("node 5 disk failure")
		}
		return nil
	})
	if err != nil || !status.QuorumSatisfied || status.VerifiedShards != 5 {
		t.Fatalf("Expected quorum satisfied with 5/6 shards, got status: %+v, err: %v", status, err)
	}

	// 3. Quorum failed (3 failures out of 6, verified 3, required 4)
	status, err = codec.WriteWithQuorum(shards, 4, func(idx int, shard []byte) error {
		if idx >= 3 {
			return fmt.Errorf("node failure")
		}
		return nil
	})
	if err == nil || status.QuorumSatisfied || status.VerifiedShards != 3 {
		t.Fatalf("Expected quorum failure when verified shards < minQuorum, got status: %+v, err: %v", status, err)
	}
}
