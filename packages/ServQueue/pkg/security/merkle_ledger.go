//go:build !enterprise

package security

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type AuditRecord struct {
	Index        uint64    `json:"index"`
	Operation    string    `json:"operation"`
	Actor        string    `json:"actor"`
	Detail       string    `json:"detail"`
	Timestamp    time.Time `json:"timestamp"`
	PrevHash     string    `json:"prev_hash"`
	CurrentHash  string    `json:"current_hash"`
}

type MerkleAuditLedger struct {
	mu      sync.RWMutex
	records []AuditRecord
}

func NewMerkleAuditLedger() *MerkleAuditLedger {
	ledger := &MerkleAuditLedger{}
	genesis := AuditRecord{
		Index:       0,
		Operation:   "GENESIS",
		Actor:       "SYSTEM",
		Detail:      "ServQueue Sovereign Merkle Audit Ledger Initialized",
		Timestamp:   time.Now(),
		PrevHash:    "0000000000000000000000000000000000000000000000000000000000000000",
		CurrentHash: "",
	}
	genesis.CurrentHash = computeHash(genesis)
	ledger.records = append(ledger.records, genesis)
	return ledger
}

func computeHash(record AuditRecord) string {
	raw := fmt.Sprintf("%d|%s|%s|%s|%d|%s",
		record.Index, record.Operation, record.Actor, record.Detail, record.Timestamp.UnixNano(), record.PrevHash)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

func (m *MerkleAuditLedger) AppendAuditRecord(operation, actor, detail string) AuditRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	last := m.records[len(m.records)-1]
	record := AuditRecord{
		Index:     uint64(len(m.records)),
		Operation: operation,
		Actor:     actor,
		Detail:    detail,
		Timestamp: time.Now(),
		PrevHash:  last.CurrentHash,
	}
	record.CurrentHash = computeHash(record)
	m.records = append(m.records, record)
	return record
}

func (m *MerkleAuditLedger) VerifyLedgerIntegrity() (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := 1; i < len(m.records); i++ {
		prev := m.records[i-1]
		curr := m.records[i]

		if curr.PrevHash != prev.CurrentHash {
			return false, fmt.Errorf("ledger integrity violation at index %d: prev_hash mismatch", curr.Index)
		}

		expectedHash := computeHash(curr)
		if curr.CurrentHash != expectedHash {
			return false, fmt.Errorf("ledger integrity violation at index %d: current_hash mismatch", curr.Index)
		}
	}

	return true, nil
}

func (m *MerkleAuditLedger) GetRecords() []AuditRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]AuditRecord, len(m.records))
	copy(res, m.records)
	return res
}
