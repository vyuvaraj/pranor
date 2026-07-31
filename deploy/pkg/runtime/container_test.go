package import (
	"context"
	"testing"
)

func TestContainerIsolationManager_SpawnAndStop(t *testing.T) {
	cim := NewContainerIsolationManager()

	spec := ContainerSpec{
		ID:            "container-app-1",
		Image:         "golang:1.22-alpine",
		CPULimit:      1.0,
		MemoryLimitMB: 512,
		Env:           map[string]string{"PORT": "8080"},
	}

	ctx := context.Background()
	st, err := cim.SpawnContainer(ctx, spec)
	if err != nil {
		t.Fatalf("SpawnContainer failed: %v", err)
	}

	if st.ID != "container-app-1" || st.State != "running" {
		t.Errorf("unexpected container status: %+v", st)
	}

	// Duplicate spawn should fail
	_, err = cim.SpawnContainer(ctx, spec)
	if err == nil {
		t.Error("expected error spawning duplicate container ID")
	}

	// Stop container
	err = cim.StopContainer("container-app-1")
	if err != nil {
		t.Fatalf("StopContainer failed: %v", err)
	}

	st, found := cim.GetStatus("container-app-1")
	if !found || st.State != "stopped" {
		t.Errorf("expected state 'stopped', got %+v", st)
	}
}
