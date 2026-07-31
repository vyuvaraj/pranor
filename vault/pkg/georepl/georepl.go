//go:build !enterprise

package georepl

import (
	"context"
)

type GeoReplicator struct {
	LocalRegion string
}

func NewGeoReplicator(region string) *GeoReplicator {
	return &GeoReplicator{
		LocalRegion: region,
	}
}

func (g *GeoReplicator) SyncObject(ctx context.Context, bucket, key string) error {
	return nil
}
