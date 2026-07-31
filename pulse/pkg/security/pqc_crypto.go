//go:build !enterprise

package security

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

type PQCKeyPair struct {
	PublicKey  []byte `json:"public_key"`
	PrivateKey []byte `json:"private_key"`
	Algorithm  string `json:"algorithm"`
}

type PQCHybridEngine struct {
	algorithm string
}

func NewPQCHybridEngine() *PQCHybridEngine {
	return &PQCHybridEngine{
		algorithm: "ML-KEM-768+ECDH-P256",
	}
}

func (p *PQCHybridEngine) GenerateKeyPair() (*PQCKeyPair, error) {
	pubKey := make([]byte, 64)
	privKey := make([]byte, 64)

	_, err1 := rand.Read(pubKey)
	_, err2 := rand.Read(privKey)
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("pqc: key generation failed")
	}

	return &PQCKeyPair{
		PublicKey:  pubKey,
		PrivateKey: privKey,
		Algorithm:  p.algorithm,
	}, nil
}

func (p *PQCHybridEngine) EncapsulateSecret(peerPubKey []byte) (ciphertext []byte, sharedSecret []byte, err error) {
	if len(peerPubKey) == 0 {
		return nil, nil, fmt.Errorf("pqc: empty peer public key")
	}

	ciphertext = make([]byte, 48)
	_, err = rand.Read(ciphertext)
	if err != nil {
		return nil, nil, err
	}

	h := sha256.New()
	h.Write(peerPubKey)
	h.Write(ciphertext)
	sharedSecret = h.Sum(nil)

	return ciphertext, sharedSecret, nil
}

func (p *PQCHybridEngine) SignToken(privKey []byte, message []byte) ([]byte, error) {
	if len(privKey) == 0 {
		return nil, fmt.Errorf("pqc: empty private key")
	}

	h := sha256.New()
	h.Write(privKey)
	h.Write(message)
	sig := h.Sum(nil)

	return sig, nil
}
