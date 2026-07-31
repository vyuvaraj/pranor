package import (
	"context"
	"fmt"
	"sync"
	"time"
)

type BucketBranch struct {
	Name        string
	ParentBucket string
	CreatedAt   time.Time
	Overlays    map[string][]byte
}

type CoWBranchEngine struct {
	branches map[string]*BucketBranch
	mu       sync.RWMutex
}

func NewCoWBranchEngine() *CoWBranchEngine {
	return &CoWBranchEngine{
		branches: make(map[string]*BucketBranch),
	}
}

func (e *CoWBranchEngine) CreateBranch(ctx context.Context, parentBucket, branchName string) (*BucketBranch, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if parentBucket == "" || branchName == "" {
		return nil, fmt.Errorf("cow branch: missing parent bucket or branch name")
	}

	key := fmt.Sprintf("%s:%s", parentBucket, branchName)
	if _, exists := e.branches[key]; exists {
		return nil, fmt.Errorf("branch %s already exists for bucket %s", branchName, parentBucket)
	}

	b := &BucketBranch{
		Name:         branchName,
		ParentBucket: parentBucket,
		CreatedAt:    time.Now(),
		Overlays:     make(map[string][]byte),
	}
	e.branches[key] = b
	return b, nil
}

func (e *CoWBranchEngine) WriteToBranch(ctx context.Context, parentBucket, branchName, key string, data []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	bKey := fmt.Sprintf("%s:%s", parentBucket, branchName)
	b, exists := e.branches[bKey]
	if !exists {
		return fmt.Errorf("branch %s not found", branchName)
	}

	b.Overlays[key] = data
	return nil
}

func (e *CoWBranchEngine) MergeBranch(ctx context.Context, parentBucket, branchName string) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	bKey := fmt.Sprintf("%s:%s", parentBucket, branchName)
	b, exists := e.branches[bKey]
	if !exists {
		return 0, fmt.Errorf("branch %s not found for merge", branchName)
	}

	mergedCount := len(b.Overlays)
	delete(e.branches, bKey)
	return mergedCount, nil
}
