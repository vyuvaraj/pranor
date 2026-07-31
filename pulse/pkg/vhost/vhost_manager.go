package import (
	"fmt"
	"sync"
	"time"
)

type ACLPermission string

const (
	PermissionRead  ACLPermission = "read"
	PermissionWrite ACLPermission = "write"
	PermissionAdmin ACLPermission = "admin"
)

type VHostQuota struct {
	MaxTopicCount  int   `json:"max_topic_count"`
	MaxBytesPerSec int64 `json:"max_bytes_per_sec"`
}

type TenantVHost struct {
	Name        string                    `json:"name"`
	Quota       VHostQuota                `json:"quota"`
	UserACLs    map[string]map[string]bool `json:"user_acls"` // username -> permission -> bool
	bytesWritten int64
	lastReset    time.Time
}

type VHostManager struct {
	mu     sync.RWMutex
	vhosts map[string]*TenantVHost
}

func NewVHostManager() *VHostManager {
	return &VHostManager{
		vhosts: make(map[string]*TenantVHost),
	}
}

func (v *VHostManager) CreateVHost(name string, quota VHostQuota) *TenantVHost {
	v.mu.Lock()
	defer v.mu.Unlock()

	host := &TenantVHost{
		Name:      name,
		Quota:     quota,
		UserACLs:  make(map[string]map[string]bool),
		lastReset: time.Now(),
	}
	v.vhosts[name] = host
	return host
}

func (v *VHostManager) GrantPermission(vhostName, username string, perm ACLPermission) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	host, exists := v.vhosts[vhostName]
	if !exists {
		return fmt.Errorf("vhost: tenant %s does not exist", vhostName)
	}

	if _, ok := host.UserACLs[username]; !ok {
		host.UserACLs[username] = make(map[string]bool)
	}
	host.UserACLs[username][string(perm)] = true
	return nil
}

func (v *VHostManager) Authorize(vhostName, username string, perm ACLPermission) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	host, exists := v.vhosts[vhostName]
	if !exists {
		return false
	}

	acls, userExists := host.UserACLs[username]
	if !userExists {
		return false
	}

	return acls[string(perm)] || acls[string(PermissionAdmin)]
}

func (v *VHostManager) CheckRateLimit(vhostName string, payloadSize int) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	host, exists := v.vhosts[vhostName]
	if !exists {
		return nil
	}

	if host.Quota.MaxBytesPerSec <= 0 {
		return nil
	}

	now := time.Now()
	if now.Sub(host.lastReset) >= time.Second {
		host.bytesWritten = 0
		host.lastReset = now
	}

	if host.bytesWritten+int64(payloadSize) > host.Quota.MaxBytesPerSec {
		return fmt.Errorf("vhost: rate limit quota exceeded for tenant %s", vhostName)
	}

	host.bytesWritten += int64(payloadSize)
	return nil
}

// QualifyTopic resolves tenant isolation namespace prefix
func QualifyTopic(vhostName, topic string) string {
	if vhostName == "" || vhostName == "default" {
		return topic
	}
	return fmt.Sprintf("vhost/%s/%s", vhostName, topic)
}
