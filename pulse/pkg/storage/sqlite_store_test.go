package storage

import (
	"os"
	"testing"
)

func TestSQLiteStoreAppendAndRecover(t *testing.T) {
	path := os.TempDir() + "/pranor-pulse_test.db"
	defer os.Remove(path)

	store, err := OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// Append some messages
	for i := 0; i < 5; i++ {
		if err := store.Append("orders", `{"id":` + string(rune('1'+i)) + `}`); err != nil {
			t.Fatalf("Append failed at i=%d: %v", i, err)
		}
	}
	store.Append("payments", `{"amount":100}`)

	// Recover all messages
	entries, err := store.Recover()
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}
	if len(entries) != 6 {
		t.Errorf("Expected 6 entries, got %d", len(entries))
	}
}

func TestSQLiteStoreQuerySQL(t *testing.T) {
	path := os.TempDir() + "/pranor-pulse_query_test.db"
	defer os.Remove(path)

	store, err := OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore failed: %v", err)
	}
	defer store.Close()

	store.Append("topicA", "msg1")
	store.Append("topicA", "msg2")
	store.Append("topicB", "msg3")

	results, err := store.QuerySQL("SELECT topic, payload FROM messages WHERE topic = 'topicA'")
	if err != nil {
		t.Fatalf("QuerySQL failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results for topicA, got %d", len(results))
	}
}

func TestSQLiteStoreQuerySQLRejectNonSelect(t *testing.T) {
	path := os.TempDir() + "/pranor-pulse_nsel_test.db"
	defer os.Remove(path)

	store, err := OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore failed: %v", err)
	}
	defer store.Close()

	_, err = store.QuerySQL("DELETE FROM messages")
	if err == nil {
		t.Error("Expected error for non-SELECT query, got nil")
	}
}

func TestSQLiteTopicStats(t *testing.T) {
	path := os.TempDir() + "/pranor-pulse_stats_test.db"
	defer os.Remove(path)

	store, err := OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore failed: %v", err)
	}
	defer store.Close()

	store.Append("alpha", "a1")
	store.Append("alpha", "a2")
	store.Append("beta", "b1")

	stats, err := store.TopicStats()
	if err != nil {
		t.Fatalf("TopicStats failed: %v", err)
	}
	if stats["alpha"] != 2 {
		t.Errorf("Expected 2 for alpha, got %d", stats["alpha"])
	}
	if stats["beta"] != 1 {
		t.Errorf("Expected 1 for beta, got %d", stats["beta"])
	}
}

func TestSQLiteMessageCount(t *testing.T) {
	path := os.TempDir() + "/pranor-pulse_count_test.db"
	defer os.Remove(path)

	store, err := OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore failed: %v", err)
	}
	defer store.Close()

	store.Append("t1", "m1")
	store.Append("t2", "m2")
	store.Append("t3", "m3")

	count, err := store.MessageCount()
	if err != nil {
		t.Fatalf("MessageCount failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3, got %d", count)
	}
}

func TestSQLiteTrimTopic(t *testing.T) {
	path := os.TempDir() + "/pranor-pulse_trim_test.db"
	defer os.Remove(path)

	store, err := OpenSQLiteStore(path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore failed: %v", err)
	}
	defer store.Close()

	store.Append("keep", "k1")
	store.Append("delete", "d1")
	store.Append("delete", "d2")

	if err := store.TrimTopic("delete"); err != nil {
		t.Fatalf("TrimTopic failed: %v", err)
	}

	count, _ := store.MessageCount()
	if count != 1 {
		t.Errorf("Expected 1 message after trim, got %d", count)
	}
}
