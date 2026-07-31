//go:build !enterprise

package import (
	"context"
	"fmt"
	"sync"
)

type PranorPulseClusterSpec struct {
	Replicas    int    `json:"replicas"`
	Image       string `json:"image"`
	StorageSize string `json:"storage_size"`
	StompPort   int    `json:"stomp_port"`
	HTTPPort    int    `json:"http_port"`
}

type PranorPulseClusterStatus struct {
	Phase         string `json:"phase"`
	ReadyReplicas int    `json:"ready_replicas"`
}

type PranorPulseCluster struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Spec      PranorPulseClusterSpec   `json:"spec"`
	Status    PranorPulseClusterStatus `json:"status"`
}

type Operator struct {
	mu       sync.RWMutex
	clusters map[string]*PranorPulseCluster
}

func NewOperator() *Operator {
	return &Operator{
		clusters: make(map[string]*PranorPulseCluster),
	}
}

func (op *Operator) Reconcile(ctx context.Context, cluster *PranorPulseCluster) (*PranorPulseClusterStatus, error) {
	op.mu.Lock()
	defer op.mu.Unlock()

	if cluster == nil || cluster.Name == "" {
		return nil, fmt.Errorf("invalid cluster resource")
	}

	key := fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name)
	if cluster.Spec.Replicas <= 0 {
		cluster.Spec.Replicas = 3
	}

	cluster.Status.Phase = "Running"
	cluster.Status.ReadyReplicas = cluster.Spec.Replicas
	op.clusters[key] = cluster

	return &cluster.Status, nil
}

func (op *Operator) GetCluster(namespace, name string) (*PranorPulseCluster, error) {
	op.mu.RLock()
	defer op.mu.RUnlock()
	key := fmt.Sprintf("%s/%s", namespace, name)
	c, ok := op.clusters[key]
	if !ok {
		return nil, fmt.Errorf("cluster %s not found", key)
	}
	return c, nil
}
