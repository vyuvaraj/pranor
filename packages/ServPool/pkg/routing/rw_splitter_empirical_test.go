package routing_test

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/vyuvaraj/serv/packages/ServPool/pkg/pool"
	"github.com/vyuvaraj/serv/packages/ServPool/pkg/routing"
)

// TestEmpirical_SQLClassification verifies SQL classification with leading comments, whitespace, and mixed casing.
func TestEmpirical_SQLClassification(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected routing.QueryType
	}{
		// Read queries - SELECT and WITH
		{"SELECT uppercase", "SELECT * FROM users", routing.QueryTypeRead},
		{"SELECT lowercase", "select * from users", routing.QueryTypeRead},
		{"SELECT mixed case 1", "sElEcT id, name FROM table1", routing.QueryTypeRead},
		{"SELECT mixed case 2", "SeLeCt 1", routing.QueryTypeRead},
		{"WITH uppercase", "WITH summary AS (SELECT 1) SELECT * FROM summary", routing.QueryTypeRead},
		{"WITH lowercase", "with cte as (select * from logs) select * from cte", routing.QueryTypeRead},
		{"WITH mixed case", "WiTh cte AS (SELECT 1) SELECT 1", routing.QueryTypeRead},

		// Read queries with leading whitespace
		{"Leading spaces", "   SELECT * FROM users", routing.QueryTypeRead},
		{"Leading tabs", "\t\tSELECT * FROM users", routing.QueryTypeRead},
		{"Leading newlines", "\n\nSELECT * FROM users", routing.QueryTypeRead},
		{"Leading CRLF", "\r\n\r\nSELECT * FROM users", routing.QueryTypeRead},
		{"Mixed leading whitespace", " \t \n \r SELECT * FROM users", routing.QueryTypeRead},

		// Read queries with leading comments
		{"Single-line comment", "-- fetch users\nSELECT * FROM users", routing.QueryTypeRead},
		{"Multiple single-line comments", "-- comment 1\n-- comment 2\nSELECT 1", routing.QueryTypeRead},
		{"Block comment single line", "/* inline comment */ SELECT * FROM users", routing.QueryTypeRead},
		{"Block comment multiline", "/*\n * Multiline SQL comment\n */ SELECT * FROM users", routing.QueryTypeRead},
		{"Sequential block and line comments", "/* block 1 */ -- line comment\n /* block 2 */ WITH cte AS (SELECT 1) SELECT 1", routing.QueryTypeRead},

		// Write queries - INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, etc.
		{"INSERT uppercase", "INSERT INTO users (name) VALUES ('alice')", routing.QueryTypeWrite},
		{"INSERT lowercase", "insert into users (name) values ('bob')", routing.QueryTypeWrite},
		{"INSERT mixed case", "iNsErT INTO users VALUES (1)", routing.QueryTypeWrite},
		{"UPDATE uppercase", "UPDATE users SET active = true WHERE id = 1", routing.QueryTypeWrite},
		{"UPDATE lowercase", "update users set active = false", routing.QueryTypeWrite},
		{"UPDATE mixed case", "UpDaTe users SET val = 2", routing.QueryTypeWrite},
		{"DELETE uppercase", "DELETE FROM users WHERE id = 1", routing.QueryTypeWrite},
		{"DELETE lowercase", "delete from tokens where expired = true", routing.QueryTypeWrite},
		{"DELETE mixed case", "DeLeTe FROM session", routing.QueryTypeWrite},
		{"CREATE uppercase", "CREATE TABLE test (id INT PRIMARY KEY)", routing.QueryTypeWrite},
		{"CREATE lowercase", "create table test (id int)", routing.QueryTypeWrite},
		{"CREATE mixed case", "CrEaTe INDEX idx ON test(id)", routing.QueryTypeWrite},
		{"DROP uppercase", "DROP TABLE test", routing.QueryTypeWrite},
		{"DROP lowercase", "drop table if exists test", routing.QueryTypeWrite},
		{"DROP mixed case", "DrOp DATABASE temp", routing.QueryTypeWrite},
		{"ALTER uppercase", "ALTER TABLE test ADD COLUMN age INT", routing.QueryTypeWrite},
		{"ALTER lowercase", "alter table test drop column age", routing.QueryTypeWrite},
		{"ALTER mixed case", "AlTeR TABLE test RENAME TO test_old", routing.QueryTypeWrite},
		{"TRUNCATE", "TRUNCATE TABLE logs", routing.QueryTypeWrite},
		{"EXPLAIN", "EXPLAIN SELECT * FROM users", routing.QueryTypeWrite},

		// Write queries with comments & whitespace
		{"INSERT with single-line comment", "-- insert query\nINSERT INTO users VALUES (1)", routing.QueryTypeWrite},
		{"UPDATE with block comment", "/* update status */ UPDATE users SET status = 'done'", routing.QueryTypeWrite},
		{"DELETE with multiline comments", "/* comment 1 */\n-- line comment\nDELETE FROM users", routing.QueryTypeWrite},

		// Boundary & edge cases
		{"Empty string", "", routing.QueryTypeWrite},
		{"Spaces only", "     ", routing.QueryTypeWrite},
		{"Tabs and newlines only", "\t\n\r  ", routing.QueryTypeWrite},
		{"Comment without trailing newline", "-- comment only", routing.QueryTypeWrite},
		{"Unclosed block comment", "/* unclosed comment SELECT 1", routing.QueryTypeWrite},
		{"SELECT followed immediately by paren", "SELECT(1)", routing.QueryTypeRead},
		{"SELECT followed immediately by semicolon", "SELECT;1", routing.QueryTypeRead},
	}

	splitter := routing.NewRWSplitter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPkg := routing.ClassifyQuery(tt.sql)
			if gotPkg != tt.expected {
				t.Errorf("routing.ClassifyQuery(%q) = %v, expected %v", tt.sql, gotPkg, tt.expected)
			}

			gotInst := splitter.ClassifyQuery(tt.sql)
			if gotInst != tt.expected {
				t.Errorf("splitter.ClassifyQuery(%q) = %v, expected %v", tt.sql, gotInst, tt.expected)
			}
		})
	}
}

// TestEmpirical_ReplicaDistributionFairness verifies exact round-robin fairness across replicas.
func TestEmpirical_ReplicaDistributionFairness(t *testing.T) {
	primary := pool.NewConnectionPool(10, "postgres")
	numReplicas := 5
	replicas := make([]pool.Manager, numReplicas)
	for i := 0; i < numReplicas; i++ {
		replicas[i] = pool.NewConnectionPool(10, fmt.Sprintf("replica_%d", i))
	}

	splitter := routing.NewRWSplitter()
	const totalQueries = 10000

	counts := make(map[pool.Manager]int)
	for i := 0; i < totalQueries; i++ {
		selected := splitter.Route("SELECT * FROM items WHERE id = 1", primary, replicas)
		if selected == primary {
			t.Fatalf("Read query routed to primary instead of replica at query %d", i)
		}
		counts[selected]++
	}

	expectedPerReplica := totalQueries / numReplicas
	for i, replica := range replicas {
		count := counts[replica]
		if count != expectedPerReplica {
			t.Errorf("Replica %d received %d queries, expected exact round-robin count %d", i, count, expectedPerReplica)
		}
	}
}

// TestEmpirical_ReplicaDistributionFairness_Concurrent verifies round-robin fairness under concurrent load.
func TestEmpirical_ReplicaDistributionFairness_Concurrent(t *testing.T) {
	primary := pool.NewConnectionPool(10, "postgres")
	numReplicas := 4
	replicas := make([]pool.Manager, numReplicas)
	for i := 0; i < numReplicas; i++ {
		replicas[i] = pool.NewConnectionPool(10, fmt.Sprintf("replica_%d", i))
	}

	splitter := routing.NewRWSplitter()
	const numGoroutines = 100
	const queriesPerGoroutine = 100
	const totalQueries = numGoroutines * queriesPerGoroutine // 10,000

	replicaCounts := make([]int64, numReplicas)
	var primaryCount int64

	var wg sync.WaitGroup
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for q := 0; q < queriesPerGoroutine; q++ {
				selected := splitter.Route("SELECT 1", primary, replicas)
				if selected == primary {
					atomic.AddInt64(&primaryCount, 1)
					continue
				}
				for rIdx, r := range replicas {
					if selected == r {
						atomic.AddInt64(&replicaCounts[rIdx], 1)
						break
					}
				}
			}
		}()
	}
	wg.Wait()

	if primaryCount != 0 {
		t.Errorf("Primary received %d read queries during concurrent test, expected 0", primaryCount)
	}

	expectedPerReplica := float64(totalQueries) / float64(numReplicas)
	for i := 0; i < numReplicas; i++ {
		c := atomic.LoadInt64(&replicaCounts[i])
		diff := math.Abs(float64(c) - expectedPerReplica)
		// Atomic round-robin index ensures exact or near-exact distribution even concurrently
		if diff > 5 {
			t.Errorf("Replica %d received %d queries, expected close to %f (diff: %f)", i, c, expectedPerReplica, diff)
		}
	}
}

// TestEmpirical_RouteWriteQueriesAndFallback verifies primary fallback logic.
func TestEmpirical_RouteWriteQueriesAndFallback(t *testing.T) {
	primary := pool.NewConnectionPool(10, "postgres")
	replica := pool.NewConnectionPool(10, "replica")

	splitter := routing.NewRWSplitter()

	// Write query with replicas available must route to primary
	writeSQL := "/* write query */ INSERT INTO users (name) VALUES ('charlie')"
	got := splitter.Route(writeSQL, primary, []pool.Manager{replica})
	if got != primary {
		t.Errorf("Write query routed to replica instead of primary")
	}

	// Read query with nil replicas must fallback to primary
	gotNil := splitter.Route("SELECT 1", primary, nil)
	if gotNil != primary {
		t.Errorf("Read query with nil replicas did not fallback to primary")
	}

	// Read query with empty replicas slice must fallback to primary
	gotEmpty := splitter.Route("SELECT 1", primary, []pool.Manager{})
	if gotEmpty != primary {
		t.Errorf("Read query with empty replicas slice did not fallback to primary")
	}
}
