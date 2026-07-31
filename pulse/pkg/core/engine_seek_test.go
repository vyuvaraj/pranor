package import (
	"testing"
	"time"
)

func TestEngineSeekToTime(t *testing.T) {
	driver := NewMemoryDriver()
	engine := NewEngine(driver)

	t1 := time.Now().UnixNano()
	e1, err := engine.Enqueue("orders", "order-1")
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	t2 := time.Now().UnixNano()
	e2, err := engine.Enqueue("orders", "order-2")
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Seek to t1
	offset1, err := engine.SeekToTime("orders", t1)
	if err != nil {
		t.Fatalf("SeekToTime failed: %v", err)
	}
	if offset1 != e1.Offset {
		t.Errorf("Expected offset %d at t1, got %d", e1.Offset, offset1)
	}

	// Seek to t2
	offset2, err := engine.SeekToTime("orders", t2)
	if err != nil {
		t.Fatalf("SeekToTime failed: %v", err)
	}
	if offset2 != e2.Offset {
		t.Errorf("Expected offset %d at t2, got %d", e2.Offset, offset2)
	}
}
