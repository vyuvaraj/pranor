package cache

import (
	"fmt"
	"strings"
	"sync"
)

// SQLDialect represents target SQL database dialect for parameter placeholder normalization.
type SQLDialect string

const (
	DialectPostgres SQLDialect = "postgres"
	DialectSQLite   SQLDialect = "sqlite"
	DialectMySQL    SQLDialect = "mysql"
)

// PreparedStatement represents a cached prepared SQL statement handle.
type PreparedStatement struct {
	ID            string     `json:"id"`
	RawQuery      string     `json:"raw_query"`
	NormalizedQuery string   `json:"normalized_query"`
	Dialect       SQLDialect `json:"dialect"`
	HitCount      int64      `json:"hit_count"`
}

// StatementCache caches normalized prepared statements to optimize query execution.
type StatementCache struct {
	mu       sync.RWMutex
	cache    map[string]*PreparedStatement // normalizedQuery -> PreparedStatement
	capacity int
}

// NewStatementCache creates a StatementCache instance.
func NewStatementCache(capacity int) *StatementCache {
	if capacity <= 0 {
		capacity = 500
	}
	return &StatementCache{
		cache:    make(map[string]*PreparedStatement),
		capacity: capacity,
	}
}

// GetOrPrepare returns cached prepared statement or normalizes and caches a new one.
func (sc *StatementCache) GetOrPrepare(query string, dialect SQLDialect) (*PreparedStatement, bool) {
	norm := NormalizeSQL(query, dialect)

	sc.mu.Lock()
	defer sc.mu.Unlock()

	stmt, exists := sc.cache[norm]
	if exists {
		stmt.HitCount++
		return stmt, true
	}

	if len(sc.cache) >= sc.capacity {
		// Evict first key
		for k := range sc.cache {
			delete(sc.cache, k)
			break
		}
	}

	stmt = &PreparedStatement{
		ID:              fmt.Sprintf("stmt-%d", len(sc.cache)+1),
		RawQuery:        query,
		NormalizedQuery: norm,
		Dialect:         dialect,
		HitCount:        1,
	}
	sc.cache[norm] = stmt

	return stmt, false
}

// NormalizeSQL normalizes parameter placeholder syntax across Postgres ($1, $2), MySQL (?), and SQLite (?).
func NormalizeSQL(query string, dialect SQLDialect) string {
	q := strings.TrimSpace(query)
	switch dialect {
	case DialectPostgres:
		// Normalize ? to $1, $2 for Postgres if needed
		var sb strings.Builder
		paramIdx := 1
		for i := 0; i < len(q); i++ {
			if q[i] == '?' {
				sb.WriteString(fmt.Sprintf("$%d", paramIdx))
				paramIdx++
			} else {
				sb.WriteByte(q[i])
			}
		}
		return sb.String()
	case DialectSQLite, DialectMySQL:
		return q
	default:
		return q
	}
}
