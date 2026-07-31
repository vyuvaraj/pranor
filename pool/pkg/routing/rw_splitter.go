package routing

import (
	"strings"
	"sync/atomic"
	"unicode"

	"github.com/vyuvaraj/pranor/pool/pkg/pool"
)

// QueryType indicates whether a SQL query is a read or write operation.
type QueryType string

const (
	QueryTypeRead  QueryType = "READ"
	QueryTypeWrite QueryType = "WRITE"
)

// RWSplitter provides read/write split query routing with round-robin replica load balancing.
type RWSplitter struct {
	rrIndex uint64
}

// NewRWSplitter creates a new RWSplitter instance.
func NewRWSplitter() *RWSplitter {
	return &RWSplitter{}
}

// ClassifyQuery determines whether a SQL statement is a read or write operation.
// Leading whitespace and SQL comments (-- or /* ... */) are stripped before checking the SQL verb.
// SELECT and WITH queries are classified as QueryTypeRead.
// INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, and all other statements are classified as QueryTypeWrite.
func ClassifyQuery(sql string) QueryType {
	clean := stripLeadingWhitespaceAndComments(sql)
	if clean == "" {
		return QueryTypeWrite
	}

	// Extract the leading verb (up to space, semicolon, or parenthesis)
	end := strings.IndexFunc(clean, func(r rune) bool {
		return unicode.IsSpace(r) || r == ';' || r == '(' || r == ')'
	})
	var verb string
	if end == -1 {
		verb = clean
	} else {
		verb = clean[:end]
	}

	verb = strings.ToUpper(verb)
	if verb == "SELECT" || verb == "WITH" {
		return QueryTypeRead
	}
	return QueryTypeWrite
}

// ClassifyQuery classifies queries on the RWSplitter instance.
func (s *RWSplitter) ClassifyQuery(sql string) QueryType {
	return ClassifyQuery(sql)
}

// Route selects an appropriate connection pool for the given SQL query.
// Write queries always route to primary.
// Read queries route to replicas in round-robin order if replicas are available; otherwise fallback to primary.
func (s *RWSplitter) Route(sql string, primary pool.Manager, replicas []pool.Manager) pool.Manager {
	if ClassifyQuery(sql) == QueryTypeWrite || len(replicas) == 0 {
		return primary
	}
	idx := atomic.AddUint64(&s.rrIndex, 1) - 1
	return replicas[idx%uint64(len(replicas))]
}

// Route is a package-level helper that routes queries using a default RWSplitter instance.
var defaultSplitter = NewRWSplitter()

func Route(sql string, primary pool.Manager, replicas []pool.Manager) pool.Manager {
	return defaultSplitter.Route(sql, primary, replicas)
}

// stripLeadingWhitespaceAndComments removes leading whitespace, line comments (-- ...),
// and block comments (/* ... */) from the SQL string.
func stripLeadingWhitespaceAndComments(sql string) string {
	for {
		sql = strings.TrimLeftFunc(sql, unicode.IsSpace)
		if strings.HasPrefix(sql, "--") {
			idx := strings.IndexByte(sql, '\n')
			if idx == -1 {
				return ""
			}
			sql = sql[idx+1:]
			continue
		}
		if strings.HasPrefix(sql, "/*") {
			idx := strings.Index(sql, "*/")
			if idx == -1 {
				return ""
			}
			sql = sql[idx+2:]
			continue
		}
		break
	}
	return sql
}
