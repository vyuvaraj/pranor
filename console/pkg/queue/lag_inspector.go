package queue

import (
	"encoding/json"
	"net/http"
	"sync"
)

// PartitionOffsetLag represents partition-level offset and consumer lag metrics.
type PartitionOffsetLag struct {
	Topic         string `json:"topic"`
	Partition     int    `json:"partition"`
	ConsumerGroup string `json:"consumer_group"`
	HighWatermark int64  `json:"high_watermark"`
	CurrentOffset int64  `json:"current_offset"`
	Lag           int64  `json:"lag"`
}

// ConsumerLagInspector computes partition-level consumer group lag for Pranor Console rendering.
type ConsumerLagInspector struct {
	mu   sync.RWMutex
	lags map[string]*PartitionOffsetLag // key -> lag
}

// NewConsumerLagInspector creates a ConsumerLagInspector instance.
func NewConsumerLagInspector() *ConsumerLagInspector {
	return &ConsumerLagInspector{
		lags: make(map[string]*PartitionOffsetLag),
	}
}

// RecordLag updates partition offset metrics for a consumer group.
func (cli *ConsumerLagInspector) RecordLag(topic string, partition int, consumerGroup string, highWatermark, currentOffset int64) {
	key := consumerGroup + ":" + topic
	lag := highWatermark - currentOffset
	if lag < 0 {
		lag = 0
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()

	cli.lags[key] = &PartitionOffsetLag{
		Topic:         topic,
		Partition:     partition,
		ConsumerGroup: consumerGroup,
		HighWatermark: highWatermark,
		CurrentOffset: currentOffset,
		Lag:           lag,
	}
}

// GetGroupLag returns lag metrics for all consumer groups.
func (cli *ConsumerLagInspector) GetGroupLag() []PartitionOffsetLag {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	res := make([]PartitionOffsetLag, 0, len(cli.lags))
	for _, l := range cli.lags {
		res = append(res, *l)
	}
	return res
}

// HTTPHandler exposes `/api/v1/console/consumer-groups/lag`.
func (cli *ConsumerLagInspector) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		lags := cli.GetGroupLag()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count": len(lags),
			"lags":  lags,
		})
	})
}
