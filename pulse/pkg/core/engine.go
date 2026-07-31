package import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

type LogEntry struct {
	Offset      uint64 `json:"offset"`
	Topic       string `json:"topic"`
	Payload     string `json:"payload"`
	Timestamp   int64  `json:"timestamp"`
	Synced      bool   `json:"synced"`
	Traceparent string `json:"traceparent,omitempty"`
}

type StorageDriver interface {
	Append(topic, payload string, metadata ...map[string]string) (LogEntry, error)
	ReadRange(topic string, startOffset, limit uint64) ([]LogEntry, error)
	SeekToTime(topic string, targetTimestamp int64) (uint64, error)
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

func (m *MemoryDriver) Append(topic, payload string, metadata ...map[string]string) (LogEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentOffset := m.offsets[topic] + 1
	m.offsets[topic] = currentOffset

	var tp string
	if len(metadata) > 0 && metadata[0] != nil {
		for k, v := range metadata[0] {
			if strings.EqualFold(k, "traceparent") {
				tp = v
				break
			}
		}
	}

	entry := LogEntry{
		Offset:      currentOffset,
		Topic:       topic,
		Payload:     payload,
		Timestamp:   time.Now().UnixNano(),
		Synced:      false,
		Traceparent: tp,
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

func (m *MemoryDriver) SeekToTime(topic string, targetTimestamp int64) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.entries {
		if entry.Topic == topic && entry.Timestamp >= targetTimestamp {
			return entry.Offset, nil
		}
	}
	if offset, ok := m.offsets[topic]; ok {
		return offset, nil
	}
	return 0, nil
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

// Engine represents the Pranor Pulse embedded log engine.
type Engine struct {
	driver        StorageDriver
	encryptionKey []byte
	mu            sync.RWMutex
}

func NewEngine(driver StorageDriver) *Engine {
	if driver == nil {
		driver = NewMemoryDriver()
	}
	return &Engine{
		driver: driver,
	}
}

// SetEncryptionKey configures 256-bit AES-GCM encryption key for payload at rest.
func (e *Engine) SetEncryptionKey(key []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(key) != 32 {
		return fmt.Errorf("engine: key must be 32 bytes for AES-256")
	}
	e.encryptionKey = make([]byte, 32)
	copy(e.encryptionKey, key)
	return nil
}

// Append appends a log entry to the topic with optional metadata (e.g., traceparent header).
func (e *Engine) Append(topic, payload string, metadata ...map[string]string) (LogEntry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	writePayload := payload
	if len(e.encryptionKey) == 32 {
		enc, err := EncryptPayload(payload, e.encryptionKey)
		if err != nil {
			return LogEntry{}, fmt.Errorf("engine: encryption failed: %w", err)
		}
		writePayload = "ENC:" + enc
	}

	return e.driver.Append(topic, writePayload, metadata...)
}

func (e *Engine) Enqueue(topic, payload string) (LogEntry, error) {
	return e.Append(topic, payload)
}

func (e *Engine) Dequeue(topic string, startOffset, limit uint64) ([]LogEntry, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	entries, err := e.driver.ReadRange(topic, startOffset, limit)
	if err != nil {
		return nil, err
	}

	if len(e.encryptionKey) == 32 {
		for i := range entries {
			if len(entries[i].Payload) > 4 && entries[i].Payload[:4] == "ENC:" {
				dec, err := DecryptPayload(entries[i].Payload[4:], e.encryptionKey)
				if err == nil {
					entries[i].Payload = dec
				}
			}
		}
	}

	return entries, nil
}

func (e *Engine) SeekToTime(topic string, targetTimestamp int64) (uint64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.driver.SeekToTime(topic, targetTimestamp)
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
