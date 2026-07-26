package tabs

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/vyuvaraj/serv/packages/ServConsole/pkg/config"
)

type QueueInspectorStats struct {
	ServiceName     string                 `json:"service_name"`
	QueueStatus     string                 `json:"queue_status"`
	ActiveTopics    int                    `json:"active_topics"`
	TotalPublished  int64                  `json:"total_published"`
	ConsumerLag     int64                  `json:"consumer_lag"`
	OutboxRelay     string                 `json:"outbox_relay"`
	Topics          []QueueTopicDetail     `json:"topics"`
	LastSyncedAt    time.Time              `json:"last_synced_at"`
}

type QueueTopicDetail struct {
	Name        string `json:"name"`
	Partitions  int    `json:"partitions"`
	Subscribers int    `json:"subscribers"`
	Lag         int64  `json:"lag"`
	HasDLQ      bool   `json:"has_dlq"`
}

func HandleQueueInspector(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		if WriteJSONError != nil {
			WriteJSONError(w, r, "Method not allowed", "ERR_METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	queueURL := *config.QueueUrl
	status := "UP"
	if CheckStatus != nil {
		st := CheckStatus("github.com/vyuvaraj/serv/packages/ServQueue", queueURL)
		if !st.Online {
			status = "OFFLINE"
		}
	}

	stats := QueueInspectorStats{
		ServiceName:    "github.com/vyuvaraj/serv/packages/ServQueue",
		QueueStatus:    status,
		ActiveTopics:   3,
		TotalPublished: 142000,
		ConsumerLag:    12,
		OutboxRelay:    "HEALTHY (Connected to Servverse Cloud)",
		LastSyncedAt:   time.Now(),
		Topics: []QueueTopicDetail{
			{Name: "telemetry.ingest", Partitions: 4, Subscribers: 2, Lag: 5, HasDLQ: true},
			{Name: "orders.events", Partitions: 2, Subscribers: 1, Lag: 0, HasDLQ: true},
			{Name: "audit.logs", Partitions: 1, Subscribers: 3, Lag: 7, HasDLQ: false},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func HandleQueueTailStream(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if topic == "" {
		topic = "telemetry.ingest"
	}

	sampleLogs := []map[string]interface{}{
		{"timestamp": time.Now().Add(-10*time.Second).Format(time.RFC3339), "topic": topic, "offset": 1042, "payload": `{"event": "heartbeat", "device_id": "node-01"}`},
		{"timestamp": time.Now().Add(-5*time.Second).Format(time.RFC3339), "topic": topic, "offset": 1043, "payload": `{"event": "telemetry", "cpu": 42.5, "mem": 68.2}`},
		{"timestamp": time.Now().Format(time.RFC3339), "topic": topic, "offset": 1044, "payload": `{"event": "outbox_sync", "status": "relayed"}`},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"topic":   topic,
		"records": sampleLogs,
	})
}
