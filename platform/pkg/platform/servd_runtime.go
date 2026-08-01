package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// ComponentStatus represents the runtime status of an embedded Pranor component.
type ComponentStatus struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
}

// ServdRuntime manages the single-binary unified runtime embedding all Pranor components.
type ServdRuntime struct {
	mu         sync.RWMutex
	components map[string]*ComponentStatus
	httpServer *http.Server
}

// NewServdRuntime creates a ServdRuntime instance.
func NewServdRuntime() *ServdRuntime {
	return &ServdRuntime{
		components: make(map[string]*ComponentStatus),
	}
}

// RegisterComponent registers an embedded component into the unified runtime.
func (sr *ServdRuntime) RegisterComponent(name string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.components[name] = &ComponentStatus{Name: name, Running: false}
}

// StartComponent marks a component as running.
func (sr *ServdRuntime) StartComponent(name string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	c, ok := sr.components[name]
	if !ok {
		return fmt.Errorf("component '%s' not registered", name)
	}
	c.Running = true
	return nil
}

// StopComponent marks a component as stopped.
func (sr *ServdRuntime) StopComponent(name string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	c, ok := sr.components[name]
	if !ok {
		return fmt.Errorf("component '%s' not registered", name)
	}
	c.Running = false
	return nil
}

// ListComponents returns runtime state of all embedded components.
func (sr *ServdRuntime) ListComponents() []ComponentStatus {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	res := make([]ComponentStatus, 0, len(sr.components))
	for _, c := range sr.components {
		res = append(res, *c)
	}
	return res
}

// Shutdown gracefully shuts down all running components.
func (sr *ServdRuntime) Shutdown(ctx context.Context) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for _, c := range sr.components {
		c.Running = false
	}
}

// HTTPHandler exposes `/api/v1/pranord/components` runtime status.
func (sr *ServdRuntime) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"components": sr.ListComponents(),
		})
	})
}
