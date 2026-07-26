package cron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AlertType identifies the target notification channel type.
type AlertType string

const (
	AlertTypeSlack     AlertType = "slack"
	AlertTypePagerDuty AlertType = "pagerduty"
	AlertTypeWebhook   AlertType = "webhook"
)

// AlertConfig defines notification target credentials and trigger conditions.
type AlertConfig struct {
	ID             string    `json:"id"`
	Type           AlertType `json:"type"`
	TargetURL      string    `json:"target_url"`
	RoutingKey     string    `json:"routing_key,omitempty"` // For PagerDuty
	MinFailures    int       `json:"min_failures"`          // Trigger alert after N consecutive failures
	NotifyOnRescue bool      `json:"notify_on_rescue"`     // Alert when job recovers to success
}

// AlertPayload represents a job failure or recovery event.
type AlertPayload struct {
	JobID        string    `json:"job_id"`
	TargetURL    string    `json:"target_url"`
	FailureCount int       `json:"failure_count"`
	LastOutcome  string    `json:"last_outcome"` // "failed" or "success"
	ErrorMessage string    `json:"error_message,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// AlertDispatcher handles dispatching notifications to Slack, PagerDuty, or Webhooks.
type AlertDispatcher struct {
	mu      sync.RWMutex
	configs []AlertConfig
	client  *http.Client
}

// NewAlertDispatcher creates an AlertDispatcher.
func NewAlertDispatcher(configs []AlertConfig) *AlertDispatcher {
	return &AlertDispatcher{
		configs: configs,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// DispatchEvent checks configs and fires matching webhook/Slack/PagerDuty notifications.
func (ad *AlertDispatcher) DispatchEvent(payload AlertPayload) []error {
	ad.mu.RLock()
	configs := make([]AlertConfig, len(ad.configs))
	copy(configs, ad.configs)
	ad.mu.RUnlock()

	var errs []error
	for _, cfg := range configs {
		if payload.LastOutcome == "failed" {
			if payload.FailureCount < cfg.MinFailures {
				continue
			}
		} else if payload.LastOutcome == "success" {
			if !cfg.NotifyOnRescue {
				continue
			}
		}

		if err := ad.sendAlert(cfg, payload); err != nil {
			errs = append(errs, fmt.Errorf("alert config %s failed: %w", cfg.ID, err))
		}
	}
	return errs
}

func (ad *AlertDispatcher) sendAlert(cfg AlertConfig, payload AlertPayload) error {
	var reqBody []byte
	var err error

	switch cfg.Type {
	case AlertTypeSlack:
		reqBody, err = buildSlackPayload(payload)
	case AlertTypePagerDuty:
		reqBody, err = buildPagerDutyPayload(cfg.RoutingKey, payload)
	case AlertTypeWebhook:
		reqBody, err = json.Marshal(payload)
	default:
		return fmt.Errorf("unsupported alert type %s", cfg.Type)
	}

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.TargetURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ad.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification endpoint returned HTTP status %d", resp.StatusCode)
	}
	return nil
}

func buildSlackPayload(payload AlertPayload) ([]byte, error) {
	color := "#danger"
	title := fmt.Sprintf("🚨 ServCron Alert: Job %s Failed", payload.JobID)
	if payload.LastOutcome == "success" {
		color = "#good"
		title = fmt.Sprintf("✅ ServCron Recovered: Job %s Succeeded", payload.JobID)
	}

	slackMsg := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"title": title,
				"fields": []map[string]interface{}{
					{"title": "Job ID", "value": payload.JobID, "short": true},
					{"title": "Failures", "value": fmt.Sprintf("%d", payload.FailureCount), "short": true},
					{"title": "Target URL", "value": payload.TargetURL, "short": false},
					{"title": "Error", "value": payload.ErrorMessage, "short": false},
				},
				"ts": payload.Timestamp.Unix(),
			},
		},
	}
	return json.Marshal(slackMsg)
}

func buildPagerDutyPayload(routingKey string, payload AlertPayload) ([]byte, error) {
	action := "trigger"
	if payload.LastOutcome == "success" {
		action = "resolve"
	}

	pdMsg := map[string]interface{}{
		"routing_key":  routingKey,
		"event_action": action,
		"dedup_key":    "servcron-" + payload.JobID,
		"payload": map[string]interface{}{
			"summary":   fmt.Sprintf("ServCron Job %s: %s (Failures: %d)", payload.JobID, payload.LastOutcome, payload.FailureCount),
			"source":    "ServCron",
			"severity":  "critical",
			"timestamp": payload.Timestamp.Format(time.RFC3339),
			"custom_details": map[string]interface{}{
				"job_id":     payload.JobID,
				"target_url": payload.TargetURL,
				"error":      payload.ErrorMessage,
			},
		},
	}
	return json.Marshal(pdMsg)
}

// FormatSlackAlertHelper returns formatted markdown text for Slack preview.
func FormatSlackAlertHelper(jobID, targetURL, errStr string) string {
	return fmt.Sprintf("*ServCron Failure*: `%s` (%s) - Error: %s", jobID, targetURL, strings.TrimSpace(errStr))
}
