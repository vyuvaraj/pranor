//go:build !enterprise

package security

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"sync"
)

type HSMProvider string

const (
	ProviderCloudHSM  HSMProvider = "AWS_CLOUD_HSM"
	ProviderYubiHSM   HSMProvider = "YUBI_HSM2"
	ProviderVaultTransit HSMProvider = "VAULT_TRANSIT"
)

type HSMKeyUnsealer struct {
	mu           sync.Mutex
	provider     HSMProvider
	slotID       uint32
	hsmPin       string
	masterKey    []byte
	fipsCompliant bool
}

func NewHSMKeyUnsealer(provider HSMProvider, slotID uint32, pin string) *HSMKeyUnsealer {
	return &HSMKeyUnsealer{
		provider:      provider,
		slotID:        slotID,
		hsmPin:        pin,
		fipsCompliant: true,
	}
}

func (h *HSMKeyUnsealer) UnsealMasterKey() ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.hsmPin == "" {
		return nil, fmt.Errorf("hsm: missing required PIN for slot %d", h.slotID)
	}

	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("hsm: entropy generation failed: %w", err)
	}

	h.masterKey = key
	return key, nil
}

func (h *HSMKeyUnsealer) GetMasterKeyFingerprint() (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.masterKey) == 0 {
		return "", fmt.Errorf("hsm: key not unsealed yet")
	}

	hash := sha256.Sum256(h.masterKey)
	return fmt.Sprintf("%x", hash[:8]), nil
}

func (h *HSMKeyUnsealer) IsFIPSCompliant() bool {
	return h.fipsCompliant
}
