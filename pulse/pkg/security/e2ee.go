package security

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/vyuvaraj/pranor/pulse/pkg/core"
)

type BlindBrokerE2EE struct {
	mu        sync.RWMutex
	topicKeys map[string][]byte
}

func NewBlindBrokerE2EE() *BlindBrokerE2EE {
	return &BlindBrokerE2EE{
		topicKeys: make(map[string][]byte),
	}
}

// RegisterTopicKey configures client-side encryption key for a specific topic envelope.
func (b *BlindBrokerE2EE) RegisterTopicKey(topic string, key []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(key) != 32 {
		return fmt.Errorf("e2ee: key must be 32 bytes for AES-256")
	}
	keyCopy := make([]byte, 32)
	copy(keyCopy, key)
	b.topicKeys[topic] = keyCopy
	return nil
}

// ProducerEncrypt seals plaintext at the client boundary before broker dispatch.
func (b *BlindBrokerE2EE) ProducerEncrypt(topic, plaintext string) (string, error) {
	b.mu.RLock()
	key, exists := b.topicKeys[topic]
	b.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("e2ee: no encryption key registered for topic '%s'", topic)
	}

	enc, err := core.EncryptPayload(plaintext, key)
	if err != nil {
		return "", fmt.Errorf("e2ee: producer encryption failed: %w", err)
	}

	return "E2EE:" + enc, nil
}

// SubscriberDecrypt opens envelope at the client subscriber boundary.
func (b *BlindBrokerE2EE) SubscriberDecrypt(topic, ciphertext string) (string, error) {
	b.mu.RLock()
	key, exists := b.topicKeys[topic]
	b.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("e2ee: no decryption key registered for topic '%s'", topic)
	}

	if len(ciphertext) <= 5 || ciphertext[:5] != "E2EE:" {
		return ciphertext, nil // Plaintext pass-through
	}

	dec, err := core.DecryptPayload(ciphertext[5:], key)
	if err != nil {
		return "", fmt.Errorf("e2ee: subscriber decryption failed: %w", err)
	}

	return dec, nil
}

// KeyFingerprint computes SHA256 fingerprint of the topic key without revealing secret key.
func KeyFingerprint(key []byte) string {
	h := sha256.Sum256(key)
	return fmt.Sprintf("%x", h[:8])
}
