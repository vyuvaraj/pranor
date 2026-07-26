package inbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// InboundEmailPayload represents a parsed incoming email payload.
type InboundEmailPayload struct {
	ID          string            `json:"id"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Subject     string            `json:"subject"`
	TextBody    string            `json:"text_body"`
	HTMLBody    string            `json:"html_body,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	ReceivedAt  time.Time         `json:"received_at"`
}

// InboundWebhookRoute defines routing rules for incoming emails.
type InboundWebhookRoute struct {
	ID        string `json:"id"`
	Recipient string `json:"recipient"` // Exact email address or wildcard domain (e.g. "*@support.dev")
	WebhookURL string `json:"webhook_url"`
}

// InboundRouter handles routing incoming emails to HTTP webhooks.
type InboundRouter struct {
	mu     sync.RWMutex
	routes map[string]InboundWebhookRoute // recipient -> route
	client *http.Client
}

// NewInboundRouter creates an InboundRouter instance.
func NewInboundRouter() *InboundRouter {
	return &InboundRouter{
		routes: make(map[string]InboundWebhookRoute),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// RegisterRoute adds a webhook route for incoming emails.
func (ir *InboundRouter) RegisterRoute(route InboundWebhookRoute) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	ir.routes[route.Recipient] = route
}

// RouteEmail dispatches an incoming email to matching webhook targets.
func (ir *InboundRouter) RouteEmail(ctx context.Context, email InboundEmailPayload) ([]error, int) {
	ir.mu.RLock()
	routes := make([]InboundWebhookRoute, 0, len(ir.routes))
	for _, r := range ir.routes {
		routes = append(routes, r)
	}
	ir.mu.RUnlock()

	var errs []error
	dispatchedCount := 0

	for _, recipient := range email.To {
		for _, route := range routes {
			if matchRecipient(recipient, route.Recipient) {
				if err := ir.sendWebhook(ctx, route.WebhookURL, email); err != nil {
					errs = append(errs, fmt.Errorf("route %s failed: %w", route.ID, err))
				} else {
					dispatchedCount++
				}
			}
		}
	}

	return errs, dispatchedCount
}

func (ir *InboundRouter) sendWebhook(ctx context.Context, webhookURL string, email InboundEmailPayload) error {
	bodyBytes, err := json.Marshal(email)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ServMail-InboundRouter/2.0")

	resp, err := ir.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook endpoint returned status %d", resp.StatusCode)
	}
	return nil
}

func matchRecipient(actual, rule string) bool {
	if actual == rule || rule == "*" {
		return true
	}
	if len(rule) > 2 && rule[:2] == "*@" {
		parts := splitRecipient(actual)
		if len(parts) == 2 && parts[1] == rule[2:] {
			return true
		}
	}
	return false
}

func splitRecipient(email string) []string {
	var res []string
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			res = append(res, email[:i], email[i+1:])
			break
		}
	}
	return res
}
