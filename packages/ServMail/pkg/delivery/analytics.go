package delivery

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
)

// DeliveryMetrics tracks aggregate mail delivery events for ServConsole visual analytics.
type DeliveryMetrics struct {
	Sent       int64   `json:"sent"`
	Delivered  int64   `json:"delivered"`
	Opened     int64   `json:"opened"`
	Clicked    int64   `json:"clicked"`
	Bounced    int64   `json:"bounced"`
	Complaints int64   `json:"complaints"`
	OpenRate   float64 `json:"open_rate"`
	ClickRate  float64 `json:"click_rate"`
	BounceRate float64 `json:"bounce_rate"`
}

// AnalyticsTracker records and aggregates email engagement events.
type AnalyticsTracker struct {
	sent       int64
	delivered  int64
	opened     int64
	clicked    int64
	bounced    int64
	complaints int64
	mu         sync.RWMutex
}

// NewAnalyticsTracker creates an AnalyticsTracker instance.
func NewAnalyticsTracker() *AnalyticsTracker {
	return &AnalyticsTracker{}
}

// RecordSent increments sent count.
func (at *AnalyticsTracker) RecordSent() { atomic.AddInt64(&at.sent, 1) }

// RecordDelivered increments delivered count.
func (at *AnalyticsTracker) RecordDelivered() { atomic.AddInt64(&at.delivered, 1) }

// RecordOpen increments open count.
func (at *AnalyticsTracker) RecordOpen() { atomic.AddInt64(&at.opened, 1) }

// RecordClick increments click count.
func (at *AnalyticsTracker) RecordClick() { atomic.AddInt64(&at.clicked, 1) }

// RecordBounce increments bounce count.
func (at *AnalyticsTracker) RecordBounce() { atomic.AddInt64(&at.bounced, 1) }

// RecordComplaint increments complaint count.
func (at *AnalyticsTracker) RecordComplaint() { atomic.AddInt64(&at.complaints, 1) }

// GetMetrics computes current rates and aggregate counts.
func (at *AnalyticsTracker) GetMetrics() DeliveryMetrics {
	sent := atomic.LoadInt64(&at.sent)
	delivered := atomic.LoadInt64(&at.delivered)
	opened := atomic.LoadInt64(&at.opened)
	clicked := atomic.LoadInt64(&at.clicked)
	bounced := atomic.LoadInt64(&at.bounced)
	complaints := atomic.LoadInt64(&at.complaints)

	openRate := 0.0
	clickRate := 0.0
	bounceRate := 0.0

	if delivered > 0 {
		openRate = float64(opened) / float64(delivered)
		clickRate = float64(clicked) / float64(delivered)
	}
	if sent > 0 {
		bounceRate = float64(bounced) / float64(sent)
	}

	return DeliveryMetrics{
		Sent:       sent,
		Delivered:  delivered,
		Opened:     opened,
		Clicked:    clicked,
		Bounced:    bounced,
		Complaints: complaints,
		OpenRate:   openRate,
		ClickRate:  clickRate,
		BounceRate: bounceRate,
	}
}

// HTTPHandler exposes `/api/v1/mail/analytics` endpoint for ServConsole dashboard graphs.
func (at *AnalyticsTracker) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		metrics := at.GetMetrics()
		_ = json.NewEncoder(w).Encode(metrics)
	})
}
