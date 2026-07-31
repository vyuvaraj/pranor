//go:build !enterprise

package import (
	"context"
	"fmt"
)

type FIPSTLSManager struct {
	FIPSMode bool
}

func NewFIPSTLSManager() *FIPSTLSManager {
	return &FIPSTLSManager{
		FIPSMode: false,
	}
}

func (f *FIPSTLSManager) ValidateSPIFFEID(ctx context.Context, spiffeID string) (bool, error) {
	if spiffeID == "" {
		return false, fmt.Errorf("empty spiffe ID")
	}
	// OSS Fallback: standard identity check
	return true, nil
}
