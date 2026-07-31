package import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAlertDispatcher_SlackNotification(t *testing.T) {
	var receivedSlack map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedSlack)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := AlertConfig{
		ID:          "slack-target",
		Type:        AlertTypeSlack,
		TargetURL:   ts.URL,
		MinFailures: 2,
	}

	ad := NewAlertDispatcher([]AlertConfig{cfg})

	// 1 failure - below MinFailures threshold (2), should not dispatch
	errs := ad.DispatchEvent(AlertPayload{
		JobID:        "job-1",
		TargetURL:    "http://api/run",
		FailureCount: 1,
		LastOutcome:  "failed",
		Timestamp:    time.Now(),
	})
	if len(errs) != 0 || len(receivedSlack) != 0 {
		t.Errorf("expected no alert for 1 failure (threshold 2)")
	}

	// 2 failures - threshold reached, should dispatch Slack attachment
	errs = ad.DispatchEvent(AlertPayload{
		JobID:        "job-1",
		TargetURL:    "http://api/run",
		FailureCount: 2,
		LastOutcome:  "failed",
		ErrorMessage: "connection timeout",
		Timestamp:    time.Now(),
	})
	if len(errs) != 0 {
		t.Fatalf("unexpected dispatch error: %v", errs)
	}

	attachments, ok := receivedSlack["attachments"].([]interface{})
	if !ok || len(attachments) == 0 {
		t.Fatalf("expected Slack attachments payload, got: %+v", receivedSlack)
	}
}

func TestAlertDispatcher_PagerDutyNotification(t *testing.T) {
	var receivedPD map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedPD)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := AlertConfig{
		ID:             "pd-target",
		Type:           AlertTypePagerDuty,
		TargetURL:      ts.URL,
		RoutingKey:     "pd-routing-key-xyz",
		MinFailures:    1,
		NotifyOnRescue: true,
	}

	ad := NewAlertDispatcher([]AlertConfig{cfg})

	// Trigger critical event
	_ = ad.DispatchEvent(AlertPayload{
		JobID:        "etl-job",
		TargetURL:    "http://etl/start",
		FailureCount: 3,
		LastOutcome:  "failed",
		ErrorMessage: "database lock timeout",
		Timestamp:    time.Now(),
	})

	if receivedPD["event_action"] != "trigger" {
		t.Errorf("expected event_action trigger, got %v", receivedPD["event_action"])
	}
	if receivedPD["routing_key"] != "pd-routing-key-xyz" {
		t.Errorf("expected routing_key pd-routing-key-xyz, got %v", receivedPD["routing_key"])
	}

	// Resolve rescue event
	_ = ad.DispatchEvent(AlertPayload{
		JobID:        "etl-job",
		TargetURL:    "http://etl/start",
		FailureCount: 0,
		LastOutcome:  "success",
		Timestamp:    time.Now(),
	})

	if receivedPD["event_action"] != "resolve" {
		t.Errorf("expected event_action resolve, got %v", receivedPD["event_action"])
	}
}
