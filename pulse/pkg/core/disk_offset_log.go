package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConsumerOffsetRecord tracks a consumer group's committed offset and timestamp (SQ.H1).
type ConsumerOffsetRecord struct {
	Group     string `json:"group"`
	Topic     string `json:"topic"`
	Offset    uint64 `json:"offset"`
	Timestamp int64  `json:"timestamp"`
}

// DiskOffsetLog represents an append-only disk-persisted consumer offset log & ack engine (SQ.H1).
type DiskOffsetLog struct {
	mu         sync.RWMutex
	filePath   string
	file       *os.File
	offsets    map[string]map[string]uint64 // group -> topic -> offset
	uncommitted map[string]map[string]uint64 // unacked offsets pending commit
}

// NewDiskOffsetLog creates or opens an append-only consumer offset log on disk.
func NewDiskOffsetLog(dir string) (*DiskOffsetLog, error) {
	if dir == "" {
		dir = ".pranorPulse_offsets"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("disk_offset_log: failed to create dir: %w", err)
	}

	filePath := filepath.Join(dir, "consumer_offsets.log")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("disk_offset_log: failed to open log file: %w", err)
	}

	dol := &DiskOffsetLog{
		filePath:    filePath,
		file:        file,
		offsets:     make(map[string]map[string]uint64),
		uncommitted: make(map[string]map[string]uint64),
	}

	if err := dol.recover(); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("disk_offset_log: recovery failed: %w", err)
	}

	return dol, nil
}

func (d *DiskOffsetLog) recover() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := os.ReadFile(d.filePath)
	if err != nil || len(data) == 0 {
		return nil
	}

	decoder := json.NewDecoder(d.file)
	// Seek to start of file for recovery
	if _, err := d.file.Seek(0, 0); err != nil {
		return err
	}

	for decoder.More() {
		var rec ConsumerOffsetRecord
		if err := decoder.Decode(&rec); err == nil {
			if d.offsets[rec.Group] == nil {
				d.offsets[rec.Group] = make(map[string]uint64)
			}
			d.offsets[rec.Group][rec.Topic] = rec.Offset
		}
	}

	return nil
}

// AcknowledgeOffset persists a consumer group offset to disk synchronously (SQ.H1).
func (d *DiskOffsetLog) AcknowledgeOffset(group, topic string, offset uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	rec := ConsumerOffsetRecord{
		Group:     group,
		Topic:     topic,
		Offset:    offset,
		Timestamp: time.Now().UnixNano(),
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("disk_offset_log: marshal failed: %w", err)
	}

	data = append(data, '\n')
	if _, err := d.file.Write(data); err != nil {
		return fmt.Errorf("disk_offset_log: append failed: %w", err)
	}

	if err := d.file.Sync(); err != nil {
		return fmt.Errorf("disk_offset_log: fsync failed: %w", err)
	}

	if d.offsets[group] == nil {
		d.offsets[group] = make(map[string]uint64)
	}
	d.offsets[group][topic] = offset

	return nil
}

// GetCommittedOffset returns the latest disk-persisted offset for a group & topic (SQ.H1).
func (d *DiskOffsetLog) GetCommittedOffset(group, topic string) (uint64, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if topics, ok := d.offsets[group]; ok {
		offset, found := topics[topic]
		return offset, found
	}
	return 0, false
}

// Close closes the underlying offset log file handle.
func (d *DiskOffsetLog) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.file.Close()
}
