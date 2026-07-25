package core

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type LogEntry struct {
	Offset    uint64 `json:"offset"`
	Topic     string `json:"topic"`
	Payload   string `json:"payload"`
	Timestamp int64  `json:"timestamp"`
	Synced    bool   `json:"synced"`
}

type StorageDriver interface {
	Append(topic, payload string) (LogEntry, error)
	ReadRange(topic string, startOffset, limit uint64) ([]LogEntry, error)
	GetUnsynced(limit uint64) ([]LogEntry, error)
	MarkSynced(offsets []uint64) error
	Recover() ([]LogEntry, error)
	Flush() error
	Close() error
}

// MemoryDriver is an in-memory implementation of StorageDriver for testing and transient usage.
type MemoryDriver struct {
	mu      sync.RWMutex
	entries []LogEntry
	offsets map[string]uint64
}

func NewMemoryDriver() *MemoryDriver {
	return &MemoryDriver{
		offsets: make(map[string]uint64),
	}
}

func (m *MemoryDriver) Append(topic, payload string) (LogEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentOffset := m.offsets[topic] + 1
	m.offsets[topic] = currentOffset

	entry := LogEntry{
		Offset:    currentOffset,
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
		Synced:    false,
	}
	m.entries = append(m.entries, entry)
	return entry, nil
}

func (m *MemoryDriver) ReadRange(topic string, startOffset, limit uint64) ([]LogEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []LogEntry
	for _, entry := range m.entries {
		if entry.Topic == topic && entry.Offset >= startOffset {
			result = append(result, entry)
			if uint64(len(result)) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MemoryDriver) GetUnsynced(limit uint64) ([]LogEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []LogEntry
	for _, entry := range m.entries {
		if !entry.Synced {
			result = append(result, entry)
			if uint64(len(result)) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MemoryDriver) MarkSynced(offsets []uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	offsetMap := make(map[uint64]bool)
	for _, o := range offsets {
		offsetMap[o] = true
	}

	for i := range m.entries {
		if offsetMap[m.entries[i].Offset] {
			m.entries[i].Synced = true
		}
	}
	return nil
}

func (m *MemoryDriver) Recover() ([]LogEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]LogEntry, len(m.entries))
	copy(result, m.entries)
	return result, nil
}

func (m *MemoryDriver) Flush() error {
	return nil
}

func (m *MemoryDriver) Close() error {
	return nil
}

// Engine represents the ServQueue embedded log engine.
type Engine struct {
	driver StorageDriver
	mu     sync.RWMutex
}

func NewEngine(driver StorageDriver) *Engine {
	if driver == nil {
		driver = NewMemoryDriver()
	}
	return &Engine{
		driver: driver,
	}
}

func (e *Engine) Enqueue(topic, payload string) (LogEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.driver.Append(topic, payload)
}

func (e *Engine) Dequeue(topic string, startOffset, limit uint64) ([]LogEntry, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.driver.ReadRange(topic, startOffset, limit)
}

func (e *Engine) GetPendingSync(limit uint64) ([]LogEntry, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.driver.GetUnsynced(limit)
}

func (e *Engine) AcknowledgeSync(offsets []uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.driver.MarkSynced(offsets)
}

func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.driver.Close()
}

// HashChecksum computes SHA256 checksum for record payload verification
func HashChecksum(data string) string {
	h := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", h)
}
