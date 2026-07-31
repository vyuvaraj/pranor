package import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalyticsTracker_MetricsCalculation(t *testing.T) {
	tracker := NewAnalyticsTracker()

	for i := 0; i < 100; i++ {
		tracker.RecordSent()
		if i < 90 {
			tracker.RecordDelivered()
		} else {
			tracker.RecordBounce()
		}
	}
	for i := 0; i < 45; i++ {
		tracker.RecordOpen()
	}
	for i := 0; i < 15; i++ {
		tracker.RecordClick()
	}

	metrics := tracker.GetMetrics()

	if metrics.Sent != 100 || metrics.Delivered != 90 || metrics.Bounced != 10 {
		t.Errorf("unexpected counts: %+v", metrics)
	}

	if metrics.OpenRate != 0.5 { // 45 / 90
		t.Errorf("expected OpenRate 0.5, got %f", metrics.OpenRate)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mail/analytics", nil)
	w := httptest.NewRecorder()
	tracker.HTTPHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", w.Code)
	}

	var resp DeliveryMetrics
	_ = json.NewDecoder(w.Body).Decode(&resp)

	if resp.Sent != 100 || resp.OpenRate != 0.5 {
		t.Errorf("unexpected JSON response: %+v", resp)
	}
}
