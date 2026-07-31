package import (
	"testing"
)

func TestVHostManagerQuotasAndACLs(t *testing.T) {
	mgr := NewVHostManager()
	quota := VHostQuota{MaxBytesPerSec: 100}

	mgr.CreateVHost("tenant_a", quota)
	_ = mgr.GrantPermission("tenant_a", "alice", PermissionWrite)

	// Test ACL
	if !mgr.Authorize("tenant_a", "alice", PermissionWrite) {
		t.Errorf("Expected alice to have WRITE permission on tenant_a")
	}

	if mgr.Authorize("tenant_a", "bob", PermissionWrite) {
		t.Errorf("Expected bob to NOT have WRITE permission on tenant_a")
	}

	// Test Rate Limiting
	err := mgr.CheckRateLimit("tenant_a", 50)
	if err != nil {
		t.Errorf("Expected payload of 50B to pass 100B rate limit, got %v", err)
	}

	err = mgr.CheckRateLimit("tenant_a", 60)
	if err == nil {
		t.Errorf("Expected rate limit error for 50B + 60B > 100B, got nil")
	}
}
