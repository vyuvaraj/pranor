package core

import (
	"fmt"
	"sync"
	"time"
)

type TxState string

const (
	TxActive    TxState = "ACTIVE"
	TxCommitted TxState = "COMMITTED"
	TxAborted   TxState = "ABORTED"
)

type TransactionItem struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

type Transaction struct {
	TxID      string            `json:"tx_id"`
	State     TxState           `json:"state"`
	Items     []TransactionItem `json:"items"`
	CreatedAt time.Time         `json:"created_at"`
	engine    *Engine
	mu        sync.Mutex
}

func (e *Engine) BeginTx(txID string) *Transaction {
	if txID == "" {
		txID = fmt.Sprintf("tx_%d", time.Now().UnixNano())
	}
	return &Transaction{
		TxID:      txID,
		State:     TxActive,
		CreatedAt: time.Now(),
		engine:    e,
	}
}

func (tx *Transaction) Enqueue(topic, payload string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.State != TxActive {
		return fmt.Errorf("transaction %s is no longer active (state: %s)", tx.TxID, tx.State)
	}

	tx.Items = append(tx.Items, TransactionItem{
		Topic:   topic,
		Payload: payload,
	})
	return nil
}

func (tx *Transaction) Commit() ([]LogEntry, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.State != TxActive {
		return nil, fmt.Errorf("transaction %s is not active (state: %s)", tx.TxID, tx.State)
	}

	var committedEntries []LogEntry
	for _, item := range tx.Items {
		entry, err := tx.engine.Enqueue(item.Topic, item.Payload)
		if err != nil {
			tx.State = TxAborted
			return nil, fmt.Errorf("transaction commit failed on topic '%s': %w", item.Topic, err)
		}
		committedEntries = append(committedEntries, entry)
	}

	tx.State = TxCommitted
	return committedEntries, nil
}

func (tx *Transaction) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.State != TxActive {
		return fmt.Errorf("transaction %s is not active", tx.TxID)
	}

	tx.Items = nil
	tx.State = TxAborted
	return nil
}
