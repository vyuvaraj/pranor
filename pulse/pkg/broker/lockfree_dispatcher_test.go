package broker

import (
	"sync"
	"testing"
)

func TestLockFreeTopicDispatcherAndBinaryFrame(t *testing.T) {
	dispatcher := NewLockFreeTopicDispatcher()

	subCh1 := make(chan BinaryFrame, 10)
	subCh2 := make(chan BinaryFrame, 10)

	topic := "events.orders.created"
	dispatcher.Subscribe(topic, subCh1)
	dispatcher.Subscribe(topic, subCh2)

	frame := BinaryFrame{
		Topic:       topic,
		Partition:   1,
		Offset:      42,
		Payload:     []byte{0x08, 0x96, 0x01, 0x12, 0x07, 0x54, 0x65, 0x73, 0x74, 0x4d, 0x73, 0x67}, // Protobuf frame
		ContentType: "application/x-protobuf",
		Headers:     map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		msg := <-subCh1
		if msg.ContentType != "application/x-protobuf" || len(msg.Payload) == 0 {
			t.Errorf("Sub1 received invalid binary frame: %+v", msg)
		}
	}()

	go func() {
		defer wg.Done()
		msg := <-subCh2
		if msg.Offset != 42 {
			t.Errorf("Sub2 received invalid offset: %d", msg.Offset)
		}
	}()

	dispatched := dispatcher.DispatchBroadcast(frame)
	if dispatched != 2 {
		t.Fatalf("Expected 2 dispatched messages, got %d", dispatched)
	}

	wg.Wait()

	// Unsubscribe subCh1 and verify single dispatch
	dispatcher.Unsubscribe(topic, subCh1)
	dispatched2 := dispatcher.DispatchBroadcast(frame)
	if dispatched2 != 1 {
		t.Fatalf("Expected 1 dispatched message after unsubscribe, got %d", dispatched2)
	}
}
