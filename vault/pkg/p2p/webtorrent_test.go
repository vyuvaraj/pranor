package import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestBrowserP2PMesh(t *testing.T) {
	mesh := NewBrowserP2PMesh()
	mesh.RegisterPeer("peer-browser-1")

	data := []byte("p2p-chunk-data-payload")
	h := sha256.Sum256(data)
	hashHex := hex.EncodeToString(h[:])

	chunk := P2PChunk{
		ChunkID: "chunk-101",
		Data:    data,
		Hash:    hashHex,
		PeerID:  "peer-browser-1",
	}

	valid, err := mesh.VerifyAndStoreChunk(context.Background(), chunk)
	if err != nil || !valid {
		t.Fatalf("chunk verification failed: %v", err)
	}

	offload, _, ratio := mesh.GetOffloadStats()
	if offload == 0 || ratio != 100.0 {
		t.Errorf("expected 100%% offload ratio, got %.2f%%", ratio)
	}
}
