package erasure

import (
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
