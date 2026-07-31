package import (
	"fmt"
)

type ErasureCodec struct {
	DataShards   int
	ParityShards int
}

func NewErasureCodec(dataShards, parityShards int) (*ErasureCodec, error) {
	if dataShards <= 0 || parityShards <= 0 {
		return nil, fmt.Errorf("erasure coding requires positive shard counts")
	}
	return &ErasureCodec{
		DataShards:   dataShards,
		ParityShards: parityShards,
	}, nil
}

func (e *ErasureCodec) Encode(data []byte) ([][]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot encode empty payload")
	}

	totalShards := e.DataShards + e.ParityShards
	shardSize := (len(data) + e.DataShards - 1) / e.DataShards
	shards := make([][]byte, totalShards)

	for i := 0; i < e.DataShards; i++ {
		start := i * shardSize
		end := start + shardSize
		if start >= len(data) {
			shards[i] = make([]byte, shardSize)
		} else if end > len(data) {
			shard := make([]byte, shardSize)
			copy(shard, data[start:])
			shards[i] = shard
		} else {
			shard := make([]byte, shardSize)
			copy(shard, data[start:end])
			shards[i] = shard
		}
	}

	// Calculate simple parity XOR shards
	for j := 0; j < e.ParityShards; j++ {
		parity := make([]byte, shardSize)
		for i := 0; i < e.DataShards; i++ {
			for b := 0; b < shardSize; b++ {
				parity[b] ^= shards[i][b]
			}
		}
		shards[e.DataShards+j] = parity
	}

	return shards, nil
}

func (e *ErasureCodec) Reconstruct(shards [][]byte, originalSize int) ([]byte, error) {
	if len(shards) < e.DataShards {
		return nil, fmt.Errorf("insufficient shards for reconstruction")
	}

	var reconstructed []byte
	for i := 0; i < e.DataShards; i++ {
		reconstructed = append(reconstructed, shards[i]...)
	}

	if len(reconstructed) > originalSize {
		reconstructed = reconstructed[:originalSize]
	}
	return reconstructed, nil
}

// QuorumWriteStatus represents the result of writing erasure shards across cluster nodes (ST.H3).
type QuorumWriteStatus struct {
	TotalShards     int  `json:"total_shards"`
	DataShards      int  `json:"data_shards"`
	ParityShards    int  `json:"parity_shards"`
	VerifiedShards  int  `json:"verified_shards"`
	QuorumSatisfied bool `json:"quorum_satisfied"`
}

// NodeWriterFunc defines a callback function to write a shard to a target cluster node.
type NodeWriterFunc func(shardIndex int, shard []byte) error

// WriteWithQuorum disperses encoded shards across cluster nodes and enforces synchronous
// Quorum write confirmations (M+K or custom minQuorum) before returning success (ST.H3).
func (e *ErasureCodec) WriteWithQuorum(shards [][]byte, minQuorum int, writer NodeWriterFunc) (*QuorumWriteStatus, error) {
	totalShards := e.DataShards + e.ParityShards
	if len(shards) != totalShards {
		return nil, fmt.Errorf("shard count mismatch: got %d, expected %d", len(shards), totalShards)
	}

	if minQuorum <= 0 {
		minQuorum = e.DataShards // Default read/write quorum threshold
	}

	type shardResult struct {
		index int
		err   error
	}

	resultsCh := make(chan shardResult, totalShards)

	for i := 0; i < totalShards; i++ {
		go func(idx int, shard []byte) {
			err := writer(idx, shard)
			resultsCh <- shardResult{index: idx, err: err}
		}(i, shards[i])
	}

	verified := 0
	var firstErr error

	for i := 0; i < totalShards; i++ {
		res := <-resultsCh
		if res.err == nil {
			verified++
		} else if firstErr == nil {
			firstErr = res.err
		}
	}

	status := &QuorumWriteStatus{
		TotalShards:     totalShards,
		DataShards:      e.DataShards,
		ParityShards:    e.ParityShards,
		VerifiedShards:  verified,
		QuorumSatisfied: verified >= minQuorum,
	}

	if !status.QuorumSatisfied {
		return status, fmt.Errorf("quorum write failed: ver