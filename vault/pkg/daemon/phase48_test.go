package import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vyuvaraj/pranor/vault/pkg/analytics"
	"github.com/vyuvaraj/pranor/vault/pkg/erasure"
	"github.com/vyuvaraj/pranor/vault/pkg/opfs"
	"github.com/vyuvaraj/pranor/vault/pkg/s3wire"
)

func TestPhase48_StandaloneDaemonAndCLI(t *testing.T) {
	d, err := NewPranorVaultDaemon("")
	if err != nil {
		t.Fatalf("failed to create daemon: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	d.handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestPhase48_S3WireProtocolEngine(t *testing.T) {
	s3 := s3wire.NewS3WireEngine()
	uploadID, err := s3.InitiateMultipartUpload("test-bucket", "video.mp4")
	if err != nil || uploadID == "" {
		t.Fatalf("failed to initiate multipart upload: %v", err)
	}
}

func TestPhase48_BrowserStoreOPFSSync(t *testing.T) {
	sdk := opfs.NewBrowserStoreSDK()
	err := sdk.SaveToOPFS(context.Background(), "/data.json", []byte(`{"hello":"world"}`))
	if err != nil {
		t.Fatalf("failed to save to OPFS: %v", err)
	}

	res, err := sdk.SyncToPranorVault(context.Background(), "/data.json")
	if err != nil || len(res) == 0 {
		t.Errorf("failed to sync OPFS file to Pranor Vault: %v", err)
	}
}

func TestPhase48_ErasureCoding(t *testing.T) {
	codec, err := erasure.NewErasureCodec(4, 2)
	if err != nil {
		t.Fatalf("failed to create codec: %v", err)
	}

	data := []byte("Pranor Vault Enterprise Erasure Coding Data Chunk Test")
	shards, err := codec.Encode(data)
	if err != nil || len(shards) != 6 {
		t.Fatalf("erasure encode failed: %v", err)
	}

	reconstructed, err := codec.Reconstruct(shards, len(data))
	if err != nil || string(reconstructed) != string(data) {
		t.Errorf("reconstructed data mismatch: got %s, want %s", string(reconstructed), string(data))
	}
}

func TestPhase48_InlineAnalyticsEngine(t *testing.T) {
	engine := analytics.NewInlineQueryEngine("json")
	res, err := engine.QueryToJSON(context.Background(), "SELECT * FROM dataset", []byte(`[{"id":1}]`))
	if err != nil || len(res) ==