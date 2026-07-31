package import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DeploymentTarget identifies Blue or Green active environment slot.
type DeploymentTarget string

const (
	TargetBlue  DeploymentTarget = "blue"
	TargetGreen DeploymentTarget = "green"
)

// BlueGreenConfig defines parameters for atomic Blue/Green deployments.
type BlueGreenConfig struct {
	ServiceName     string `json:"service_name"`
	GatewayURL      string `json:"gateway_url"`
	BlueURL         string `json:"blue_url"`
	GreenURL        string `json:"green_url"`
	ActiveSlot      DeploymentTarget `json:"active_slot"`
	HealthCheckPath string `json:"health_check_path"`
}

// BlueGreenManager orchestrates zero-downtime Blue/Green deployments and Pranor Gate cutover.
type BlueGreenManager struct {
	mu     sync.RWMutex
	cfg    BlueGreenConfig
	client *http.Client
}

// NewBlueGreenManager creates a BlueGreenManager.
func NewBlueGreenManager(cfg BlueGreenConfig) *BlueGreenManager {
	if cfg.HealthCheckPath == "" {
		cfg.HealthCheckPath = "/healthz"
	}
	if cfg.ActiveSlot == "" {
		cfg.ActiveSlot = TargetBlue
	}
	return &BlueGreenManager{
		cfg:    cfg,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetActiveSlot returns currently active deployment target.
func (bgm *BlueGreenManager) GetActiveSlot() DeploymentTarget {
	bgm.mu.RLock()
	defer bgm.mu.RUnlock()
	return bgm.cfg.ActiveSlot
}

// DeployNewVersion deploys to inactive slot, health checks it, and performs atomic cutover via Pranor Gate.
func (bgm *BlueGreenManager) DeployNewVersion(ctx context.Context) (DeploymentTarget, error) {
	bgm.mu.Lock()
	defer bgm.mu.Unlock()

	inactiveSlot := TargetGreen
	inactiveURL := bgm.cfg.GreenURL
	if bgm.cfg.ActiveSlot == TargetGreen {
		inactiveSlot = TargetBlue
		inactiveURL = bgm.cfg.BlueURL
	}

	// 1. Health check inactive target slot
	healthURL := fmt.Sprintf("%s%s", inactiveURL, bgm.cfg.HealthCheckPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := bgm.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return "", fmt.Errorf("health check failed for inactive target slot %s at %s", inactiveSlot, healthURL)
	}
	resp.Body.Close()

	// 2. Perform atomic Pranor Gate route weight cutover
	if err := bgm.shiftGatewayTraffic(ctx, inactiveSlot, inactiveURL); err != nil {
		return "", fmt.Errorf("Pranor Gate traffic cutover failed: %w", err)
	}

	bgm.cfg.ActiveSlot = inactiveSlot
	return inactiveSlot, nil
}

func (bgm *BlueGreenManager) shiftGatewayTraffic(ctx context.Context, newSlot DeploymentTarget, targetURL string) error {
	if bgm.cfg.GatewayURL == "" {
		return nil // Soft bypass when GatewayURL unconfigured in standalone mode
	}

	cutoverURL := fmt.Sprintf("%s/api/v1/routes/weights", bgm.cfg.GatewayURL)
	payload := map[string]interface{}{
		"service":   bgm.cfg.ServiceName,
		"active":    newSlot,
		"target_url": targetURL,
		"weight":    100,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cutoverURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := bgm.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Pranor Gate returned HTTP %d on cutover", resp.StatusCode)
	}
	return nil
}
