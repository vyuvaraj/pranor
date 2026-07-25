package storage

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTieredStorageOffloadAndFetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusOK)
		} else if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("mock_wal_segment_data"))
		}
	}))
	defer ts.Close()

	cfg := TieredStorageConfig{
		S3Endpoint:  ts.URL,
		S3Bucket:    "servqueue-cold-wal",
		AccessToken: "test-token-99",
	}

	mgr := NewTieredStorageManager(cfg)

	// Create dummy local segment file
	tmpFile := filepath.Join(t.TempDir(), "segment_1001.log")
	_ = os.WriteFile(tmpFile, []byte("test_data"), 0644)

	// Offload segment to S3 mock
	if err := mgr.OffloadOldSegment(tmpFile); err != nil {
		t.Fatalf("OffloadOldSegment failed: %v", err)
	}

	// Fetch remote segment from S3 mock
	data, err := mgr.FetchRemoteSegment(ts.URL + "/servqueue-cold-wal/wal/segment_1001.log")
	if err != nil {
		t.Fatalf("FetchRemoteSegment failed: %v", err)
	}

	if string(data) != "mock_wal_segment_data" {
		t.Errorf("Unexpected remote segment data: %s", string(data))
	}
}
