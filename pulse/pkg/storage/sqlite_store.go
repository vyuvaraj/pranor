package storage

// SQ.D5: SQLite-Backed Persistent Storage Mode
//
// Provides a single-file SQLite WAL-compatible storage backend for Pranor Pulse.
// For single-node deployments where ACID durability, SQL queryability, and
// zero-dependency binary distribution matter more than raw throughput.
//
// The SQLiteStore implements the same Append/Recover API as the binary WAL
// so the broker engine can swap backends via an environment variable:
//   PRANOR_PULSE_STORAGE_BACKEND=sqlite  (default: binary WAL)
//
// Usage from BrokerEngine:
//   store, _ := storage.OpenSQLiteStore("Pranor Pulse.db")
//   // store.Append(topic, payload) / store.Recover() / store.QuerySQL(...)

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite" // pure-Go, CGO-free SQLite driver
)

// SQLiteStore is a single-file SQLite backend with the same interface as WAL.
type SQLiteStore struct {
	mu   sync.Mutex
	db   *sql.DB
	path string
}

// OpenSQLiteStore opens (or creates) a SQLite database at path and
// initialises the messages table with WAL journal mode for durability.
func OpenSQLiteStore(path string) (*SQLiteStore, error) {
	// Use modernc.org/sqlite pure-Go driver; no CGO required.
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open failed: %w", err)
	}

	// Enable WAL mode for concurrent reads and crash-safe writes.
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("sqlite: enable WAL journal mode: %w", err)
	}
	if _, err := db.Exec(`PRAGMA synchronous=NORMAL`); err != nil {
		return nil, fmt.Errorf("sqlite: set synchronous=NORMAL: %w", err)
	}

	// Create messages table if it does not exist.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		topic     TEXT    NOT NULL,
		payload   TEXT    NOT NULL,
		timestamp INTEGER NOT NULL
	)`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: create messages table: %w", err)
	}

	// Index for fast per-topic queries.
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_topic ON messages(topic)`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: create topic index: %w", err)
	}

	return &SQLiteStore{db: db, path: path}, nil
}

// Append inserts a message into the SQLite store — equivalent to WAL.Append.
func (s *SQLiteStore) Append(topic, payload string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO messages (topic, payload, timestamp) VALUES (?, ?, ?)`,
		topic, payload, time.Now().UnixNano(),
	)
	return err
}

// Recover returns all stored messages — equivalent to WAL.Recover.
func (s *SQLiteStore) Recover() ([]LogEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT topic, payload, timestamp FROM messages ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: recover query: %w", err)
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		if err := rows.Scan(&e.Topic, &e.Payload, &e.Timestamp); err != nil {
			return nil, fmt.Errorf("sqlite: recover scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// QuerySQL allows raw SQL SELECT queries against the messages table.
// Only SELECT statements are permitted for safety.
//   Example: SELECT topic, payload FROM messages WHERE topic = 'orders' LIMIT 50
func (s *SQLiteStore) QuerySQL(query string) ([]map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trimmed := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(trimmed, "SELECT") {
		return nil, fmt.Errorf("sqlite: only SELECT queries are permitted")
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("sqlite: query failed: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = vals[i]
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// TrimTopic deletes all messages for a topic — used by retention / compaction.
func (s *SQLiteStore) TrimTopic(topic string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`DELETE FROM messages WHERE topic = ?`, topic)
	return err
}

// TrimBefore deletes messages older than the given nanosecond timestamp.
func (s *SQLiteStore) TrimBefore(topic string, beforeNs int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`DELETE FROM messages WHERE topic = ? AND timestamp < ?`,
		topic, beforeNs,
	)
	return err
}

// MessageCount returns the total number of stored messages (across all topics).
func (s *SQLiteStore) MessageCount() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&count)
	return count, err
}

// TopicStats returns per-topic message counts.
func (s *SQLiteStore) TopicStats() (map[string]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`SELECT topic, COUNT(*) FROM messages GROUP BY topic`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int64)
	for rows.Next() {
		var topic string
		var count int64
		if err := rows.Scan(&topic, &count); err != nil {
			continue
		}
		stats[topic] = count
	}
	return stats, rows.Err()
}

// Close closes the underlying SQLite database.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// IsSQLiteAvailable returns true if the pure-Go SQLite driver binary is
// loadable — used to guard backend selection at startup.
func IsSQLiteAvailable() bool {
	tmpPath := os.TempDir() + "/pranorPulse_probe.db"
	db, err := sql.Open("sqlite", tmpPath)
	if err != nil {
		return false
	}
	db.Close()
	os.Remove(tmpPath)
	return true
}
