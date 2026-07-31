//go:build !enterprise

package kms

import (
	"context"
)

type KMSManager struct {
	Provider string
}

func NewKMSManager() *KMSManager {
	return &KMSManager{
		Provider: "none",
	}
}

func (k *KMSManager) EncryptObject(ctx context.Context, payload []byte) ([]byte, error) {
	// OSS Fallback: pass-through
	return payload, nil
}
