package cache

import (
	"testing"
)

func TestStatementCache_GetOrPrepare(t *testing.T) {
	sc := NewStatementCache(10)

	query := "SELECT * FROM users WHERE email = ?"

	// First call -> Miss
	stmt1, cached := sc.GetOrPrepare(query, DialectPostgres)
	if cached || stmt1.NormalizedQuery != "SELECT * FROM users WHERE email = $1" {
		t.Fatalf("expected miss and normalized $1 query, got cached=%v norm=%s", cached, stmt1.NormalizedQuery)
	}

	// Second call -> Hit
	stmt2, cached := sc.GetOrPrepare(query, DialectPostgres)
	if !cached || stmt2.HitCount != 2 {
		t.Errorf("expected cache hit with hitCount 2, got cached=%v hitCount=%d", cached, stmt2.HitCount)
	}
}
