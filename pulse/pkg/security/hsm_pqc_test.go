package import (
	"testing"
)

func TestHSMKeyUnsealing(t *testing.T) {
	unsealer := NewHSMKeyUnsealer(ProviderCloudHSM, 1, "hsm-pin-7712")

	key, err := unsealer.UnsealMasterKey()
	if err != nil {
		t.Fatalf("UnsealMasterKey failed: %v", err)
	}

	if len(key) != 32 {
		t.Fatalf("Expected 32-byte master key, got %d", len(key))
	}

	fingerprint, err := unsealer.GetMasterKeyFingerprint()
	if err != nil || fingerprint == "" {
		t.Errorf("GetMasterKeyFingerprint failed: %v", err)
	}

	if !unsealer.IsFIPSCompliant() {
		t.Errorf("Expected HSM key unsealer to be FIPS 140-3 compliant")
	}
}

func TestPQCHybridCrypto(t *testing.T) {
	pqc := NewPQCHybridEngine()

	keyPair, err := pqc.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	cipher, sharedSecret, err := pqc.EncapsulateSecret(keyPair.PublicKey)
	if err != nil || len(cipher) == 0 || len(sharedSecret) != 32 {
		t.Fatalf("EncapsulateSecret failed: %v", err)
	}

	sig, err := pqc.SignToken(keyPair.PrivateKey, []byte("quantum-test-message"))
	if err != nil || len(sig) == 0 {
		t.Fatalf("SignToken failed: %v", err)
	}
}
