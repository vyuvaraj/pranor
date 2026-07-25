package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptPayload encrypts plaintext using AES-256-GCM with key.
func EncryptPayload(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("crypto: key must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("crypto: cipher creation failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: gcm creation failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: nonce generation failed: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPayload decrypts base64 ciphertext using AES-256-GCM with key.
func DecryptPayload(encodedCiphertext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("crypto: key must be 32 bytes for AES-256")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("crypto: base64 decode failed: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("crypto: cipher creation failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: gcm creation failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("crypto: ciphertext too short")
	}

	nonce, ciphertextBytes := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: decryption failed: %w", err)
	}

	return string(plaintext), nil
}
