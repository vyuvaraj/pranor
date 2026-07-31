//go:build !enterprise

package enterprise

import (
	"context"
)

type MultiCloudGlacierConnector struct {
	Enabled bool
}

func NewMultiCloudGlacierConnector() *MultiCloudGlacierConnector {
	return &MultiCloudGlacierConnector{
		Enabled: false,
	}
}

func (m *MultiCloudGlacierConnector) ArchiveObject(ctx context.Context, bucket, key string) error {
	return nil
}

type TagBasedAccessControl struct {
	Enabled bool
}

func NewTagBasedAccessControl() *TagBasedAccessControl {
	return &TagBasedAccessControl{
		Enabled: false,
	}
}

func (t *TagBasedAccessControl) AuthorizeByTag(ctx context.Context, userRole string, tagKey, tagVal string) bool {
	return true
}
