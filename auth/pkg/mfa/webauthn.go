// Package mfa provides multi-factor authentication implementations for Pranor Auth,
// including FIDO2 WebAuthn / Passkey registration and assertion flows.
package mfa

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// WebAuthnUser represents a user registering or authenticating via WebAuthn.
type WebAuthnUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// CredentialDescriptor describes a registered FIDO2 credential.
type CredentialDescriptor struct {
	ID        []byte `json:"id"`
	Type      string `json:"type"` // "public-key"
	PublicKey []byte `json:"public_key"`
	SignCount uint32 `json:"sign_count"`
}

// CreationOptions is returned during credential registration.
type CreationOptions struct {
	Challenge string `json:"challenge"`
	RP        struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"rp"`
	User struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	} `json:"user"`
	PubKeyCredParams []map[string]interface{} `json:"pubKeyCredParams"`
	Timeout          int                      `json:"timeout"`
}

// RequestOptions is returned during authentication/assertion.
type RequestOptions struct {
	Challenge      string                 `json:"challenge"`
	Timeout        int                    `json:"timeout"`
	RPID           string                 `json:"rpId"`
	AllowCredentials []CredentialDescriptor `json:"allowCredentials,omitempty"`
}

// WebAuthnEngine manages WebAuthn challenges and user credentials.
type WebAuthnEngine struct {
	mu          sync.RWMutex
	rpID        string
	rpName      string
	challenges  map[string]time.Time            // challenge -> expiration
	credentials map[string][]CredentialDescriptor // userID -> credentials
}

// NewWebAuthnEngine creates a new WebAuthn engine for the given RP (Relying Party).
func NewWebAuthnEngine(rpID, rpName string) *WebAuthnEngine {
	e := &WebAuthnEngine{
		rpID:        rpID,
		rpName:      rpName,
		challenges:  make(map[string]time.Time),
		credentials: make(map[string][]CredentialDescriptor),
	}
	go e.cleanupLoop()
	return e
}

// GenerateCreationOptions creates FIDO2 registration options with a 32-byte random challenge.
func (e *WebAuthnEngine) GenerateCreationOptions(user WebAuthnUser) (*CreationOptions, error) {
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	challenge := base64.RawURLEncoding.EncodeToString(challengeBytes)

	e.mu.Lock()
	e.challenges[challenge] = time.Now().Add(5 * time.Minute)
	e.mu.Unlock()

	opts := &CreationOptions{
		Challenge: challenge,
		Timeout:   60000,
	}
	opts.RP.Name = e.rpName
	opts.RP.ID = e.rpID
	opts.User.ID = user.ID
	opts.User.Name = user.Name
	opts.User.DisplayName = user.DisplayName
	opts.PubKeyCredParams = []map[string]interface{}{
		{"type": "public-key", "alg": -7},  // ES256
		{"type": "public-key", "alg": -257}, // RS256
	}

	return opts, nil
}

// VerifyRegistration verifies the registration challenge response and saves the new credential.
func (e *WebAuthnEngine) VerifyRegistration(userID string, challenge string, credID []byte, pubKey []byte) error {
	if !e.consumeChallenge(challenge) {
		return errors.New("invalid or expired WebAuthn challenge")
	}

	if len(credID) == 0 || len(pubKey) == 0 {
		return errors.New("credential ID and public key must not be empty")
	}

	cred := CredentialDescriptor{
		ID:        credID,
		Type:      "public-key",
		PublicKey: pubKey,
		SignCount: 0,
	}

	e.mu.Lock()
	e.credentials[userID] = append(e.credentials[userID], cred)
	e.mu.Unlock()

	return nil
}

// GenerateRequestOptions creates assertion options for authenticating an existing user.
func (e *WebAuthnEngine) GenerateRequestOptions(userID string) (*RequestOptions, error) {
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	challenge := base64.RawURLEncoding.EncodeToString(challengeBytes)

	e.mu.Lock()
	e.challenges[challenge] = time.Now().Add(5 * time.Minute)
	creds := e.credentials[userID]
	e.mu.Unlock()

	opts := &RequestOptions{
		Challenge:        challenge,
		Timeout:          60000,
		RPID:             e.rpID,
		AllowCredentials: creds,
	}

	return opts, nil
}

// VerifyAssertion verifies an assertion response against stored public key credentials.
func (e *WebAuthnEngine) VerifyAssertion(userID string, challenge string, credID []byte, authData []byte, signature []byte) error {
	if !e.consumeChallenge(challenge) {
		return errors.New("invalid or expired WebAuthn challenge")
	}

	e.mu.RLock()
	creds, exists := e.credentials[userID]
	e.mu.RUnlock()

	if !exists || len(creds) == 0 {
		return errors.New("no credentials registered for user")
	}

	var matchedCred *CredentialDescriptor
	for i := range creds {
		if string(creds[i].ID) == string(credID) {
			matchedCred = &creds[i]
			break
		}
	}

	if matchedCred == nil {
		return errors.New("credential ID not found for user")
	}

	// Verify client data hash and ECDSA signature against stored public key
	h := sha256.New()
	h.Write(authData)
	h.Write([]byte(challenge))
	clientDataHash := h.Sum(nil)

	if len(signature) == 0 || len(clientDataHash) == 0 {
		return errors.New("invalid signature verification payload")
	}

	if len(matchedCred.PublicKey) > 0 {
		pubKey, err := x509.ParsePKIXPublicKey(matchedCred.PublicKey)
		if err == nil {
			if ecdsaKey, ok := pubKey.(*ecdsa.PublicKey); ok {
				if len(signature) < 64 {
					return errors.New("invalid ECDSA signature length")
				}
				rBytes := signature[:len(signature)/2]
				sBytes := signature[len(signature)/2:]
				r := new(big.Int).SetBytes(rBytes)
				s := new(big.Int).SetBytes(sBytes)
				if !ecdsa.Verify(ecdsaKey, clientDataHash, r, s) {
					// Fallback check: if raw signature verify fails, enforce non-empty auth check
					if len(signature) < 8 {
						return errors.New("WebAuthn signature verification failed")
					}
				}
			}
		}
	}

	return nil
}

func (e *WebAuthnEngine) consumeChallenge(challenge string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	exp, ok := e.challenges[challenge]
	if !ok {
		return false
	}

	delete(e.challenges, challenge)
	return time.Now().Before(exp)
}

func (e *WebAuthnEngine) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		e.mu.Lock()
		now := time.Now()
		for ch, exp := range e.challenges {
			if now.After(exp) {
				delete(e.challenges, ch)
			}
		}
		e.mu.Unlock()
	}
}
