//go:build !enterprise

package import (
	"context"
	"time"
)

type WORMLockManager struct {
	Enabled bool
}

func NewWORMLockManager() *WORMLockManager {
	return &WORMLockManager{
		Enabled: false,
	}
}

func (w *WORMLockManager) CanDeleteObject(ctx context.Context, bucket, key string) (bool, error) {
	// OSS Fallback: allow deletion
	return true, nil
}

func (w *WORMLockManager) LockObject(ctx context.Context, bucket, key string, retainUntil time.Time) error {
	return nil
}
