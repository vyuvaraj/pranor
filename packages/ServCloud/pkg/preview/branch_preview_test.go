package preview

import (
	"strings"
	"testing"
	"time"
)

func TestBranchPreviewProvisioner_ProvisionAndDestroy(t *testing.T) {
	provisioner := NewBranchPreviewProvisioner("preview.servcloud.io", 1*time.Hour)

	branch := "feature/checkout-v2"
	env, err := provisioner.ProvisionEnvironment(branch)
	if err != nil {
		t.Fatalf("ProvisionEnvironment failed: %v", err)
	}

	if !strings.Contains(env.Subdomain, "feature-checkout-v2.preview.servcloud.io") {
		t.Errorf("unexpected subdomain format: %s", env.Subdomain)
	}

	retrieved, found := provisioner.GetEnvironment(branch)
	if !found || retrieved.Status != "active" {
		t.Fatalf("failed to retrieve active environment: %+v", retrieved)
	}

	// Destroy environment
	err = provisioner.DestroyEnvironment(branch)
	if err != nil {
		t.Fatalf("DestroyEnvironment failed: %v", err)
	}

	_, found = provisioner.GetEnvironment(branch)
	if found {
		t.Error("expected environment to be deleted after destroy")
	}
}
