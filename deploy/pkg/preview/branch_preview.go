package import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// PreviewEnvironment represents an isolated per-branch environment.
type PreviewEnvironment struct {
	ID        string    `json:"id"`
	Branch    string    `json:"branch"`
	Subdomain string    `json:"subdomain"`
	Status    string    `json:"status"` // "provisioned", "active", "destroyed"
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// BranchPreviewProvisioner manages ephemeral preview environments per Git branch.
type BranchPreviewProvisioner struct {
	mu           sync.RWMutex
	baseDomain   string
	environments map[string]*PreviewEnvironment // branch -> PreviewEnvironment
	defaultTTL   time.Duration
}

// NewBranchPreviewProvisioner creates a BranchPreviewProvisioner instance.
func NewBranchPreviewProvisioner(baseDomain string, defaultTTL time.Duration) *BranchPreviewProvisioner {
	if baseDomain == "" {
		baseDomain = "Pranor Deploy.dev"
	}
	if defaultTTL <= 0 {
		defaultTTL = 24 * time.Hour
	}
	return &BranchPreviewProvisioner{
		baseDomain:   baseDomain,
		environments: make(map[string]*PreviewEnvironment),
		defaultTTL:   defaultTTL,
	}
}

// ProvisionEnvironment provisions a new preview environment for a Git branch.
func (bpp *BranchPreviewProvisioner) ProvisionEnvironment(branch string) (*PreviewEnvironment, error) {
	if branch == "" {
		return nil, fmt.Errorf("branch name cannot be empty")
	}

	subdomain := sanitizeBranchName(branch)
	envID := fmt.Sprintf("prev-%s", subdomain)
	fullDomain := fmt.Sprintf("https://%s.%s", subdomain, bpp.baseDomain)

	now := time.Now()
	env := &PreviewEnvironment{
		ID:        envID,
		Branch:    branch,
		Subdomain: fullDomain,
		Status:    "active",
		CreatedAt: now,
		ExpiresAt: now.Add(bpp.defaultTTL),
	}

	bpp.mu.Lock()
	bpp.environments[branch] = env
	bpp.mu.Unlock()

	return env, nil
}

// GetEnvironment returns active preview environment for a branch if exists.
func (bpp *BranchPreviewProvisioner) GetEnvironment(branch string) (*PreviewEnvironment, bool) {
	bpp.mu.RLock()
	defer bpp.mu.RUnlock()
	env, ok := bpp.environments[branch]
	return env, ok
}

// DestroyEnvironment cleans up and destroys a branch preview environment.
func (bpp *BranchPreviewProvisioner) DestroyEnvironment(branch string) error {
	bpp.mu.Lock()
	defer bpp.mu.Unlock()

	env, ok := bpp.environments[branch]
	if !ok {
		return fmt.Errorf("preview environment for branch '%s' not found", branch)
	}

	env.Status = "destroyed"
	delete(bpp.environments, branch)
	return nil
}

func sanitizeBranchName(branch string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	sanitized := reg.ReplaceAllString(branch, "-")
	return strings.ToLower(strings.Trim(sanitized, "-"))
}
