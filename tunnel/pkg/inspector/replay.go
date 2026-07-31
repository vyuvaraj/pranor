package import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// CapturedTransaction represents a full request/response pair flowing through a tunnel pipe.
type CapturedTransaction struct {
	ID             string              `json:"id"`
	TunnelID       string              `json:"tunnel_id"`
	Method         string              `json:"method"`
	URL            string              `json:"url"`
	RequestHeaders map[string][]string `json:"request_headers"`
	RequestBody    string              `json:"request_body"`
	ResponseStatus int                 `json:"response_status"`
	ResponseBody   string              `json:"response_body"`
	DurationMs     int64               `json:"duration_ms"`
	Timestamp      time.Time           `json:"timestamp"`
}

// RequestInspectorBuffer captures and stores HTTP transactions in a ring buffer for Pranor Console replay.
type RequestInspectorBuffer struct {
	mu           sync.RWMutex
	transactions []CapturedTransaction
	maxCapacity  int
	client       *http.Client
}

// NewRequestInspectorBuffer creates a RequestInspectorBuffer instance.
func NewRequestInspectorBuffer(maxCapacity int) *RequestInspectorBuffer {
	if maxCapacity <= 0 {
		maxCapacity = 200
	}
	return &RequestInspectorBuffer{
		transactions: make([]CapturedTransaction, 0),
		maxCapacity:  maxCapacity,
		client:       &http.Client{Timeout: 10 * time.Second},
	}
}

// CaptureTransaction records a full HTTP request/response transaction.
func (rib *RequestInspectorBuffer) CaptureTransaction(tx CapturedTransaction) {
	rib.mu.Lock()
	defer rib.mu.Unlock()

	if len(rib.transactions) >= rib.maxCapacity {
		rib.transactions = rib.transactions[1:]
	}
	rib.transactions = append(rib.transactions, tx)
}

// GetTransactions returns captured HTTP request/response history.
func (rib *RequestInspectorBuffer) GetTransactions() []CapturedTransaction {
	rib.mu.RLock()
	defer rib.mu.RUnlock()

	res := make([]CapturedTransaction, len(rib.transactions))
	copy(res, rib.transactions)
	return res
}

// ReplayTransaction re-executes a captured transaction against target URL.
func (rib *RequestInspectorBuffer) ReplayTransaction(ctx context.Context, txID string, overrideTargetURL string) (*CapturedTransaction, error) {
	rib.mu.RLock()
	var targetTx *CapturedTransaction
	for _, tx := range rib.transactions {
		if tx.ID == txID {
			targetTx = &tx
			break
		}
	}
	rib.mu.RUnlock()

	if targetTx == nil {
		return nil, fmt.Errorf("transaction ID '%s' not found", txID)
	}

	targetURL := targetTx.URL
	if overrideTargetURL != "" {
		targetURL = overrideTargetURL
	}

	req, err := http.NewRequestWithContext(ctx, targetTx.Method, targetURL, bytes.NewBufferString(targetTx.RequestBody))
	if err != nil {
		return nil, err
	}

	for k, v := range targetTx.RequestHeaders {
		req.Header[k] = v
	}

	start := time.Now()
	resp, err := rib.client.Do(req)
	dur := time.Since(start).Milliseconds()

	if err != nil {
		return nil, fmt.Errorf("replay request failed: %w", err)
	}
	defer resp.Body.Close()

	respBodyBytes, _ := io.ReadAll(resp.Body)

	replayed := &CapturedTransaction{
		ID:             fmt.Sprintf("replay-%s", targetTx.ID),
		TunnelID:       targetTx.TunnelID,
		Method:         targetTx.Method,
		URL:            targetURL,
		RequestHeaders: targetTx.RequestHeaders,
		RequestBody:    targetTx.RequestBody,
		ResponseStatus: resp.StatusCode,
		ResponseBody:   string(respBodyBytes),
		DurationMs:     dur,
		Timestamp:      time.Now(),
	}

	return replayed, nil
}

// HTTPHandler exposes `/api/v1/tunnel/inspector` for Pranor Console visual request replay.
func (rib *RequestInspectorBuffer) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		txs := rib.GetTransactions()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":        len(txs),
			"transactions": txs,
		})
	})
}
