package branch

import (
	"context"
	"testing"
)

func TestCoWBranchEngine(t *testing.T) {
	engine := NewCoWBranchEngine()

	branch, err := engine.CreateBranch(context.Background(), "prod-data", "dev-test-branch")
	if err != nil || branch == nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	err = engine.WriteToBranch(context.Background(), "prod-data", "dev-test-branch", "delta.json", []byte(`{"mod":true}`))
	if err != nil {
		t.Fatalf("failed to write to branch overlay: %v", err)
	}

	merged, err := engine.MergeBranch(context.Background(), "prod-data", "dev-test-branch")
	if err != nil || merged != 1 {
		t.Errorf("merge failed: got %d merged keys", merged)
	}
}
