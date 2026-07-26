package storage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type TieredStorageConfig struct {
	MaxLocalSegmentAge time.Duration `json:"max_local_segment_age"`
	S3Endpoint         string        `json:"s3_endpoint"`
	S3Bucket           string        `json:"s3_bucket"`
	AccessToken        string        `json:"access_token"`
}

type TieredStorageManager struct {
	offloader *Offloader
	config    TieredStorageConfig
	mu        sync.Mutex
	offloaded map[string]string // filename -> s3URL
}

func NewTieredStorageManager(config TieredStorageConfig) *TieredStorageManager {
	off := NewOffloader(config.S3Endpoint, config.S3Bucket, config.AccessToken)
	return &TieredStorageManager{
		offloader: off,
		config:    config,
		offloaded: make(map[string]string),
	}
}

// OffloadOldSegment archives closed local WAL segments to S3 / ServStore cloud storage.
func (m *TieredStorageManager) OffloadOldSegment(localPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.offloader.OffloadSegment(localPath)
	if err != nil {
		return fmt.Errorf("tiered_storage: offload failed: %w", err)
	}

	m.offloaded[localPath] = fmt.Sprintf("%s/%s/wal/%s", m.config.S3Endpoint, m.config.S3Bucket, localPath)
	return nil
}

type AutoCompactionWorker struct {
	TTL          time.Duration
	MaxSegmentMB int64
	mu           sync.Mutex
	PurgedCount  uint64
}

func NewAutoCompactionWorker(ttl time.Duration, maxSegmentMB int64) *AutoCompactionWorker {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	if maxSegmentMB <= 0 {
		maxSegmentMB = 100
	}
	return &AutoCompactionWorker{
		TTL:          ttl,
		MaxSegmentMB: maxSegmentMB,
	}
}

func (w *AutoCompactionWorker) CompactAndPurgeExpired(entries []LogEntry) ([]LogEntry, int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	cutoff := time.Now().Add(-w.TTL).UnixNano()
	var retained []LogEntry
	purged := 0

	for _, e := range entries {
		if e.Timestamp > 0 && e.Timestamp < cutoff {
			purged++
		} else {
			retained = append(retained, e)
		}
	}

	w.PurgedCount += uint64(purged)
	return retained, purged
}

// FetchRemoteSegment retrieves cold WAL segment data from S3 / ServStore.
func (m *TieredStorageManager) FetchRemoteSegment(remoteURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", remoteURL, nil)
	if err != nil {
		return nil, err
	}
	if m.config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+m.config.AccessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tiered_storage: download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tiered_storage: HTTP status %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	return buf.Bytes(), err
}
