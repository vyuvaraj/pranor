package opfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vyuvaraj/serv/packages/ServQueue/pkg/core"
)

// OPFSDriver implements core.StorageDriver targeting Origin Private File System (OPFS) / file handles.
type OPFSDriver struct {
	mu         sync.RWMutex
	basePath   string
	walPath    string
	entries    []core.LogEntry
	offsets    map[string]uint64
	fileHandle *os.File
}

func NewOPFSDriver(basePath string) (*OPFSDriver, error) {
	if basePath == "" {
		basePath = filepath.Join(os.TempDir(), "servqueue_opfs")
	}
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("opfs: failed to create base path: %w", err)
	}

	walPath := filepath.Join(basePath, "opfs_wal.log")
	file, err := os.OpenFile(walPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("opfs: open WAL failed: %w", err)
	}

	driver := &OPFSDriver{
		basePath:   basePath,
		walPath:    walPath,
		offsets:    make(map[string]uint64),
		fileHandle: file,
	}

	if err := driver.recoverFromDisk(); err != nil {
		// Log recovery warning
		_ = err
	}

	return driver, nil
}

func (o *OPFSDriver) recoverFromDisk() error {
	data, err := os.ReadFile(o.walPath)
	if err != nil || len(data) == 0 {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry core.LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			o.entries = append(o.entries, entry)
			if entry.Offset > o.offsets[entry.Topic] {
				o.offsets[entry.Topic] = entry.Offset
			}
		}
	}
	return nil
}

func (o *OPFSDriver) Append(topic, payload string) (core.LogEntry, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	currentOffset := o.offsets[topic] + 1
	o.offsets[topic] = currentOffset

	entry := core.LogEntry{
		Offset:    currentOffset,
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
		Synced:    false,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return entry, err
	}

	line := append(data, '\n')
	if _, err := o.fileHandle.Write(line); err != nil {
		return entry, fmt.Errorf("opfs: write failed: %w", err)
	}
	_ = o.fileHandle.Sync()

	o.entries = append(o.entries, entry)
	return entry, nil
}

func (o *OPFSDriver) ReadRange(topic string, startOffset, limit uint64) ([]core.LogEntry, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var result []core.LogEntry
	for _, entry := range o.entries {
		if entry.Topic == topic && entry.Offset >= startOffset {
			result = append(result, entry)
			if uint64(len(result)) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (o *OPFSDriver) GetUnsynced(limit uint64) ([]core.LogEntry, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var result []core.LogEntry
	for _, entry := range o.entries {
		if !entry.Synced {
			result = append(result, entry)
			if uint64(len(result)) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (o *OPFSDriver) MarkSynced(offsets []uint64) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	offsetMap := make(map[uint64]bool)
	for _, off := range offsets {
		offsetMap[off] = true
	}

	for i := range o.entries {
		if offsetMap[o.entries[i].Offset] {
			o.entries[i].Synced = true
		}
	}

	// Rewrite WAL with updated sync state
	_ = o.fileHandle.Close()
	file, err := os.OpenFile(o.walPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	o.fileHandle = file

	for _, entry := range o.entries {
		data, _ := json.Marshal(entry)
		_, _ = o.fileHandle.Write(append(data, '\n'))
	}
	_ = o.fileHandle.Sync()
	return nil
}

func (o *OPFSDriver) Recover() ([]core.LogEntry, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	result := make([]core.LogEntry, len(o.entries))
	copy(result, o.entries)
	return result, nil
}

func (o *OPFSDriver) Flush() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.fileHandle.Sync()
}

func (o *OPFSDriver) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.fileHandle.Close()
}
