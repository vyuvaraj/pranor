package import (
	"fmt"
	"sync"
)

// MetricType specifies vector distance calculation metric.
type MetricType string

const (
	MetricCosine     MetricType = "cosine"
	MetricEuclidean  MetricType = "euclidean"
	MetricDotProduct MetricType = "dot_product"
)

// VectorNamespaceConfig defines parameters for a bucket's HNSW vector index namespace.
type VectorNamespaceConfig struct {
	BucketName string     `json:"bucket_name"`
	Dimension  int        `json:"dimension"`
	Metric     MetricType `json:"metric"`
	M          int        `json:"m"`          // Max HNSW connections per node
	EfSearch   int        `json:"ef_search"`  // Search candidate list size
}

// VectorNamespaceManager manages isolated vector index namespaces per storage bucket.
type VectorNamespaceManager struct {
	mu         sync.RWMutex
	namespaces map[string]*VectorNamespaceConfig // bucketName -> config
}

// NewVectorNamespaceManager creates a VectorNamespaceManager instance.
func NewVectorNamespaceManager() *VectorNamespaceManager {
	return &VectorNamespaceManager{
		namespaces: make(map[string]*VectorNamespaceConfig),
	}
}

// CreateNamespace provisions an isolated vector index namespace for a storage bucket.
func (vnm *VectorNamespaceManager) CreateNamespace(cfg VectorNamespaceConfig) (*VectorNamespaceConfig, error) {
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("bucket_name is required")
	}
	if cfg.Dimension <= 0 {
		cfg.Dimension = 128
	}
	if cfg.Metric == "" {
		cfg.Metric = MetricCosine
	}
	if cfg.M <= 0 {
		cfg.M = 16
	}
	if cfg.EfSearch <= 0 {
		cfg.EfSearch = 64
	}

	vnm.mu.Lock()
	defer vnm.mu.Unlock()

	vnm.namespaces[cfg.BucketName] = &cfg
	return &cfg, nil
}

// GetNamespace retrieves active vector namespace config for a bucket.
func (vnm *VectorNamespaceManager) GetNamespace(bucketName string) (*VectorNamespaceConfig, bool) {
	vnm.mu.RLock()
	defer vnm.mu.RUnlock()
	cfg, ok := vnm.namespaces[bucketName]
	return cfg, ok
}

// DeleteNamespace removes a bucket's vector index namespace.
func (vnm *VectorNamespaceManager) DeleteNamespace(bucketName string) bool {
	vnm.mu.Lock()
	defer vnm.mu.Unlock()
	_, exists := vnm.namespaces[bucketName]
	i