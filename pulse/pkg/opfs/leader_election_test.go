package import (
	"testing"
	"time"
)

func TestMultiTabLeaderElection_AcquireAndExpire(t *testing.T) {
	election := NewMultiTabLeaderElection("tab-1", 100*time.Millisecond)

	now := time.Now()

	// 1. Acquire leader
	acquired := election.TryAcquireLeader(now)
	if !acquired || !election.IsLeader(now) {
		t.Fatalf("expected tab-1 to acquire leader status")
	}

	// 2. Check lease before expiration
	if !election.IsLeader(now.Add(50 * time.Millisecond)) {
		t.Error("expected tab-1 to remain leader before TTL expiration")
	}

	// 3. Lease expires after 100ms
	if election.IsLeader(now.Add(150 * time.Millisecond)) {
		t.Error("expected tab-1 leader lease to expire after TTL")
	}

	// 4. Release leader
	election.ReleaseLeader()
	if election.IsLeader(now) {
		t.Error("expected tab-1 to step down as leader after release")
	}
}
