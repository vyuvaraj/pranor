package daemon

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/vyuvaraj/pranor/vault/pkg/lifecycle"
	"github.com/vyuvaraj/pranor/vault/pkg/s3select"
	"github.com/vyuvaraj/pranor/vault/pkg/tiering"
	"github.com/vyuvaraj/pranor/vault/pkg/vectorclock"
)

func TestPhase53_S3SelectStreamingQueryEngine(t *testing.T) {
	engine := s3select.NewS3SelectEngine()
	req := s3select.S3SelectRequest{
		Query:              "SELECT * FROM s3object WHERE status = 200",
		InputSerialization: "JSON",
	}

	reader := strings.NewReader(`[{"id":1,"status":200}]`)
	var out bytes.Buffer

	n, err := engine.ExecuteStreamingQuery(context.Background(), req, reader, &out)
	if err != nil || n == 0 {
		t.Fatalf("s3 select streaming query failed: %v", err)
	}
	if !strings.Contains(out.String(), "SUCCESS") {
		t.Errorf("expected SUCCESS status in S3 select output")
	}
}

func TestPhase53_MultiCloudStorageTiering(t *testing.T) {
	tierEngine := tiering.NewCloudTierEngine()
	err := tierEngine.SetPolicy(tiering.MultiCloudTierPolicy{
		Bucket:              "analytics-hot",
		DaysBeforeArchiving: 30,
		TargetProvider:      tiering.AWSGlacierDeepArchive,
		TargetBucket:        "analytics-glacier",
	})
	if err != nil {
		t.Fatalf("failed to set tiering policy: %v", err)
	}

	migrated, err := tierEngine.EvaluateAndMigrate(context.Background(), "analytics-hot", "log_2025.tar", 45)
	if err != nil || !migrated {
		t.Errorf("expected migration to Glacier for 45-day-old object")
	}
}

func TestPhase53_AutoPurgeAndLifecycleEngine(t *testing.T) {
	purgeEngine := lifecycle.NewAutoPurgeEngine()
	err := purgeEngine.AddRule("temp-bucket", lifecycle.ExpirationRule{
		ID:             "rule-expire-7d",
		Prefix:         "tmp/",
		ExpirationDays: 7,
	})
	if err != nil {
		t.Fatalf("failed to add expiration rule: %v", err)
	}

	purged, err := purgeEngine.EvaluatePurge(context.Background(), "temp-bucket", "tmp/session.json", 10)
	if err != nil || !purged {
		t.Errorf("expected object purge for 10-day-old object")
	}
}

func TestPhase53_CRDTVectorClockReplication(t *testing.T) {
	nodeA := vectorclock.NewVectorClockReplicationNode("us-east-1")
	nodeA.IncrementClock()

	remoteClock := map[string]uint64{
		"eu-west-1": 5,
	}

	updated := nodeA.MergeVectorClocks("eu-west-1", remoteClock)
	if !updated || nodeA.SyncedItems != 1 {
		t.Errorf("vector clock merge failed: updated=%v, synced=%d", updated, nodeA.SyncedItems)
	}
}
