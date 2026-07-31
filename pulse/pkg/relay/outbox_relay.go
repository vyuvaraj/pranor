package import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/pulse/pkg/core"
)

type OutboxRelay struct {
	engine     *core.Engine
	serverURL  string
	syncBatch  uint64
	client     *http.Client
	mu         sync.Mutex
	running    bool
	stopChan   chan struct{}
	OnSyncAck  func(syncedCount int)
}

func NewOutboxRelay(engine *core.Engine, serverURL string) *OutboxRelay {
	if serverURL == "" {
		serverURL = "http://localhost:8080/api/v1/queue/stream"
	}
	return &OutboxRelay{
		engine:    engine,
		serverURL: serverURL,
		syncBatch: 50,
		client:    &http.Client{Timeout: 5 * time.Second},
		stopChan:  make(chan struct{}),
	}
}

func (r *OutboxRelay) Start(interval time.Duration) {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	r.stopChan = make(chan struct{})
	r.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = r.SyncNow()
			case <-r.stopChan:
				return
			}
		}
	}()
}

func (r *OutboxRelay) SyncNow() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	pending, err := r.engine.GetPendingSync(r.syncBatch)
	if err != nil || len(pending) == 0 {
		return err
	}

	payloadData, err := json.Marshal(pending)
	if err != nil {
		return fmt.Errorf("relay: marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", r.serverURL, bytes.NewBuffer(payloadData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pranor-Pulse-Outbox-Sync", "true")
	req.Header.Set("X-WebTransport-Mode", "quic-multiplex")
	req.Header.Set("Sec-WebTransport-Http3-Draft", "draft02")

	resp, err := r.client.Do(req)
	if err != nil {
		// Server offline, return silently for local-first retention
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var syncedOffsets []uint64
		for _, entry := range pending {
			syncedOffsets = append(syncedOffsets, entry.Offset)
		}
		if err := r.engine.AcknowledgeSync(syncedOffsets); err == nil {
			if r.OnSyncAck != nil {
				r.OnSyncAck(len(syncedOffsets))
			}
		}
	}
	return nil
}

func (r *OutboxRelay) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.running {
		r.running = false
		close(r.stopChan)
	}
}
