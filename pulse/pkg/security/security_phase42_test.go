//go:build !enterprise

package security

import (
	"bytes"
	"testing"
)

func TestBlindBrokerE2EE(t *testing.T) {
	key := bytes.Repeat([]byte("k"), 32)
	e2ee := NewBlindBrokerE2EE()

	if err := e2ee.RegisterTopicKey("finance_feed", key); err != nil {
		t.Fatalf("RegisterTopicKey failed: %v", err)
	}

	plaintext := `{"tx_id": "tx_8891", "amount": 1000000.00}`

	// Producer encrypts payload
	sealedPayload, err := e2ee.ProducerEncrypt("finance_feed", plaintext)
	if err != nil {
		t.Fatalf("ProducerEncrypt failed: %v", err)
	}

	if sealedPayload == plaintext {
		t.Errorf("Expected sealed payload to differ from plaintext")
	}

	// Subscriber decrypts payload
	openedPayload, err := e2ee.SubscriberDecrypt("finance_feed", sealedPayload)
	if err != nil {
		t.Fatalf("SubscriberDecrypt failed: %v", err)
	}

	if openedPayload != plaintext {
		t.Errorf("Expected opened payload to match plaintext '%s', got '%s'", plaintext, openedPayload)
	}
}

func TestMerkleAuditLedgerIntegrity(t *testing.T) {
	ledger := NewMerkleAuditLedger()

	ledger.AppendAuditRecord("TOPIC_CREATE", "admin", "Created topic orders")
	ledger.AppendAuditRecord("CONFIG_CHANGE", "admin", "Updated rate limit quota")
	ledger.AppendAuditRecord("WAL_ROTATE", "system", "Rotated segment_1002.log")

	records := ledger.GetRecords()
	if len(records) != 4 { // Genesis + 3 records
		t.Fatalf("Expected 4 audit records, got %d", len(records))
	}

	valid, err := ledger.VerifyLedgerIntegrity()
	if err != nil || !valid {
		t.Fatalf("Expected ledger integrity verification to pass, got error: %v", err)
	}
}
