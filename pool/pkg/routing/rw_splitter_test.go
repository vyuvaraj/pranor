package routing

import (
	"sync"
	"testing"

	"github.com/vyuvaraj/pranor/pool/pkg/pool"
	)

func TestClassifyQuery_SQLVerbsAndCasing(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected QueryType
	}{
		// Read queries
		{"uppercase SELECT", "SELECT * FROM users", QueryTypeRead},
		{"lowercase select", "select id, name from products", QueryTypeRead},
		{"mixedcase sElEcT", "sElEcT 1", QueryTypeRead},
		{"uppercase WITH", "WITH cte AS (SELECT 1) SELECT * FROM cte", QueryTypeRead},
		{"lowercase with", "with summary as (select * from logs) select * from summary", QueryTypeRead},
		{"mixedcase wItH", "wItH cte AS (SELECT 1) SELECT 1", QueryTypeRead},

		// Write queries
		{"uppercase INSERT", "INSERT INTO users (name) VALUES ('alice')", QueryTypeWrite},
		{"lowercase insert", "insert into logs (msg) values ('hello')", QueryTypeWrite},
		{"mixedcase InSeRt", "InSeRt INTO tbl VALUES (1)", QueryTypeWrite},
		{"uppercase UPDATE", "UPDATE users SET status = 'active' WHERE id = 1", QueryTypeWrite},
		{"lowercase update", "update users set status = 'idle'", QueryTypeWrite},
		{"uppercase DELETE", "DELETE FROM users WHERE id = 1", QueryTypeWrite},
		{"lowercase delete", "delete from tokens", QueryTypeWrite},
		{"uppercase CREATE", "CREATE TABLE test (id INT)", QueryTypeWrite},
		{"lowercase create", "create table foo (bar text)", QueryTypeWrite},
		{"uppercase DROP", "DROP TABLE test", QueryTypeWrite},
		{"lowercase drop", "drop index idx_user", QueryTypeWrite},
		{"uppercase ALTER", "ALTER TABLE test ADD COLUMN age INT", QueryTypeWrite},
		{"lowercase alter", "alter table test drop column age", QueryTypeWrite},
		{"other verb TRUNCATE", "TRUNCATE TABLE users", QueryTypeWrite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyQuery(tt.sql)
			if got != tt.expected {
				t.Errorf("ClassifyQuery(%q) = %v, expected %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestClassifyQuery_WhitespaceAndComments(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected QueryType
	}{
		{"leading spaces", "   SELECT * FROM users", QueryTypeRead},
		{"leading tabs and newlines", "\n\t\r  SELECT 1", QueryTypeRead},
		{"single-line comment", "-- query comment\nSELECT * FROM users", QueryTypeRead},
		{"multi-line comment", "/* block comment */ SELECT * FROM users", QueryTypeRead},
		{"multiline comment with newlines", "/* line 1\n line 2 */\nSELECT 1", QueryTypeRead},
		{"sequential comments", "/* comment 1 */ -- comment 2\n  \t  WITH cte AS (SELECT 1) SELECT 1", QueryTypeRead},
		{"write with leading comment", "-- insert comment\nINSERT INTO users VALUES (1)", QueryTypeWrite},
		{"write with block comment", "/* write block */ UPDATE users SET status = 'ok'", QueryTypeWrite},
		{"empty query", "", QueryTypeWrite},
		{"whitespace only", "   \n\t  ", QueryTypeWrite},
		{"unclosed comment", "/* unclosed comment SELECT 1", QueryTypeWrite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyQuery(tt.sql)
			if got != tt.expected {
				t.Errorf("ClassifyQuery(%q) = %v, expected %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestRoute_RoundRobinAndFallback(t *testing.T) {
	primary := pool.NewConnectionPool(5, "postgres")
	replica1 := pool.NewConnectionPool(5, "postgres")
	replica2 := pool.NewConnectionPool(5, "postgres")
	replica3 := pool.NewConnectionPool(5, "postgres")
	replicas := []pool.Manager{replica1, replica2, replica3}

	splitter := NewRWSplitter()

	// 1. Write queries should always route to primary
	writeSQLs := []string{
		"INSERT INTO users VALUES (1)",
		"UPDATE users SET name = 'x'",
		"DELETE FROM users",
		"CREATE TABLE foo (id INT)",
		"DROP TABLE foo",
		"ALTER TABLE foo ADD c INT",
	}
	for _, sql := range writeSQLs {
		got := splitter.Route(sql, primary, replicas)
		if got != primary {
			t.Errorf("Route(%q) should route to primary, got different pool", sql)
		}
	}

	// 2. Read query with empty replicas should fallback to primary
	got := splitter.Route("SELECT 1", primary, nil)
	if got != primary {
		t.Errorf("Route(SELECT) with nil replicas should fallback to primary")
	}
	got = splitter.Route("SELECT 1", primary, []pool.Manager{})
	if got != primary {
		t.Errorf("Route(SELECT) with empty replicas should fallback to primary")
	}

	// 3. Read queries with replicas should round-robin: replica1, replica2, replica3, replica1, ...
	readSQL := "SELECT * FROM users WHERE id = 1"
	expectedOrder := []pool.Manager{replica1, replica2, replica3, replica1, replica2}

	for i, expected := range expectedOrder {
		selected := splitter.Route(readSQL, primary, replicas)
		if selected != expected {
			t.Errorf("Round %d: Route(%q) = %v, expected replica at index %d", i, readSQL, selected, i%3)
		}
	}
}

func TestClassifyQuery_InstanceMethod(t *testing.T) {
	splitter := NewRWSplitter()
	if splitter.ClassifyQuery("SELECT 1") != QueryTypeRead {
		t.Errorf("expected QueryTypeRead from instance method")
	}
	if splitter.ClassifyQuery("UPDATE tbl SET a=1") != QueryTypeWrite {
		t.Errorf("expected QueryTypeWrite from instance method")
	}
}

func TestRoute_PackageLevelFunc(t *testing.T) {
	primary := pool.NewConnectionPool(5, "postgres")
	replica := pool.NewConnectionPool(5, "postgres")
	replicas := []pool.Manager{replica}

	got := Route("SELECT 1", primary, replicas)
	if got != replica {
		t.Errorf("expected package-level Route to return replica pool")
	}

	gotWrite := Route("INSERT INTO tbl VALUES (1)", primary, replicas)
	if gotWrite != primary {
		t.Errorf("expected package-level Route to return primary pool for write")
	}
}

func TestRoute_Concurrent(t *testing.T) {
	primary := pool.NewConnectionPool(5, "postgres")
	replica1 := pool.NewConnectionPool(5, "postgres")
	replica2 := pool.NewConnectionPool(5, "postgres")
	replicas := []pool.Manager{replica1, replica2}
	splitter := NewRWSplitter()

	const numGoroutines = 20
	const queriesPerGoroutine = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < queriesPerGoroutine; j++ {
				res := splitter.Route("SELECT * FROM items", primary, replicas)
				if res != replica1 && res != replica2 {
					t.Errorf("unexpected pool returned during concurrent routing")
				}
			}
		}()
	}
	wg.Wait()
}

