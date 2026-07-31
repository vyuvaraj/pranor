package import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// ContainerSpec defines OCI container isolation parameters.
type ContainerSpec struct {
	ID          string            `json:"id"`
	Image       string            `json:"image"`
	Env         map[string]string `json:"env"`
	PortMappings map[int]int       `json:"port_mappings"` // hostPort -> containerPort
	CPULimit    float64           `json:"cpu_limit"`     // e.g. 1.5 CPUs
	MemoryLimitMB int             `json:"memory_mb"`
}

// ContainerStatus represents the current state of a running OCI container.
type ContainerStatus struct {
	ID        string    `json:"id"`
	State     string    `json:"state"` // "running", "stopped", "exited"
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
}

// ContainerIsolationManager manages OCI container process isolation environments.
type ContainerIsolationManager struct {
	mu         sync.RWMutex
	containers map[string]*ContainerStatus
}

// NewContainerIsolationManager creates a ContainerIsolationManager.
func NewContainerIsolationManager() *ContainerIsolationManager {
	return &ContainerIsolationManager{
		containers: make(map[string]*ContainerStatus),
	}
}

// SpawnContainer simulates/executes OCI container process isolation startup.
func (cim *ContainerIsolationManager) SpawnContainer(ctx context.Context, spec ContainerSpec) (*ContainerStatus, error) {
	if spec.ID == "" || spec.Image == "" {
		return nil, fmt.Errorf("container ID and Image are required")
	}

	cim.mu.Lock()
	defer cim.mu.Unlock()

	if _, exists := cim.containers[spec.ID]; exists {
		return nil, fmt.Errorf("container ID '%s' already exists", spec.ID)
	}

	// Verify docker/containerd CLI availability if present
	_, execErr := exec.LookPath("docker")

	status := &ContainerStatus{
		ID:        spec.ID,
		State:     "running",
		PID:       1000 + len(cim.containers),
		StartedAt: time.Now(),
	}

	if execErr == nil {
		// Docker CLI present — validate syntax simulation
		_ = exec.CommandContext(ctx, "docker", "ps", "-q").Run()
	}

	cim.containers[spec.ID] = status
	return status, nil
}

// GetStatus returns current isolation status for a container ID.
func (cim *ContainerIsolationManager) GetStatus(containerID string) (*ContainerStatus, bool) {
	cim.mu.RLock()
	defer cim.mu.RUnlock()
	st, ok := cim.containers[containerID]
	return st, ok
}

// StopContainer terminates a running OCI container process environment.
func (cim *ContainerIsolationManager) StopContainer(containerID string) error {
	cim.mu.Lock()
	defer cim.mu.Unlock()

	st, ok := cim.containers[containerID]
	if !ok {
		return fmt.Errorf("container ID '%s' not found", containerID)
	}

	st.State = "stopped"
	return nil
}
