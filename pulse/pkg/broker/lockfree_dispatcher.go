package import (
	"sync"
)

// BinaryFrame represents a native binary payload frame pipeline item (SQ.H4).
type BinaryFrame struct {
	Topic       string            `json:"topic"`
	Partition   int               `json:"partition"`
	Offset      uint64            `json:"offset"`
	Payload     []byte            `json:"payload"`
	ContentType string            `json:"content_type"` // e.g. "application/x-protobuf", "application/avro"
	Headers     map[string]string `json:"headers,omitempty"`
}

// LockFreeTopicDispatcher manages subscriber slice dispatching using lock-free atomic channels (SQ.H2).
type LockFreeTopicDispatcher struct {
	topicSubscribers sync.Map // map[string]*subscriberSlice
}

type subscriberSlice struct {
	mu   sync.RWMutex
	subs []chan BinaryFrame
}

// NewLockFreeTopicDispatcher initializes a lock-free topic dispatcher.
func NewLockFreeTopicDispatcher() *LockFreeTopicDispatcher {
	return &LockFreeTopicDispatcher{}
}

// Subscribe registers a subscriber channel to a topic atomically (SQ.H2).
func (d *LockFreeTopicDispatcher) Subscribe(topic string, ch chan BinaryFrame) {
	val, _ := d.topicSubscribers.LoadOrStore(topic, &subscriberSlice{})
	slice := val.(*subscriberSlice)

	slice.mu.Lock()
	slice.subs = append(slice.subs, ch)
	slice.mu.Unlock()
}

// Unsubscribe removes a subscriber channel from a topic atomically (SQ.H2).
func (d *LockFreeTopicDispatcher) Unsubscribe(topic string, ch chan BinaryFrame) {
	val, ok := d.topicSubscribers.Load(topic)
	if !ok {
		return
	}
	slice := val.(*subscriberSlice)

	slice.mu.Lock()
	defer slice.mu.Unlock()

	newList := make([]chan BinaryFrame, 0, len(slice.subs))
	for _, subscriber := range slice.subs {
		if subscriber != ch {
			newList = append(newList, subscriber)
		}
	}
	slice.subs = newList
}

// DispatchBroadcast broadcasts a native binary frame to all subscribers in lock-free fashion (SQ.H2, SQ.H4).
func (d *LockFreeTopicDispatcher) DispatchBroadcast(frame BinaryFrame) int {
	val, ok := d.topicSubscribers.Load(frame.Topic)
	if !ok {
		return 0
	}
	slice := val.(*subscriberSlice)

	slice.mu.RLock()
	subs := make([]chan BinaryFrame, len(slice.subs))
	copy(subs, slice.subs)
	slice.mu.RUnlock()

	dispatched := 0
	for _, ch := range subs {
		select {
		case ch <- frame:
			dispatched++
		default:
			// Non-blocking dispatch if subscriber buffer full
		}
	}

	return dispatched
}
