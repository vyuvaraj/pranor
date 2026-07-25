package cdc

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/vyuvaraj/serv/packages/ServQueue/pkg/core"
)

type CDCOperation string

const (
	OpInsert CDCOperation = "INSERT"
	OpUpdate CDCOperation = "UPDATE"
	OpDelete CDCOperation = "DELETE"
)

type CDCEvent struct {
	Table     string                 `json:"table"`
	Operation CDCOperation           `json:"operation"`
	Before    map[string]interface{} `json:"before,omitempty"`
	After     map[string]interface{} `json:"after,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

type CDCEngine struct {
	engine *core.Engine
	mu     sync.RWMutex
}

func NewCDCEngine(engine *core.Engine) *CDCEngine {
	return &CDCEngine{
		engine: engine,
	}
}

// StreamMutation converts database table mutations into structured ServQueue topic events.
func (c *CDCEngine) StreamMutation(table string, op CDCOperation, before, after map[string]interface{}) (core.LogEntry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	event := CDCEvent{
		Table:     table,
		Operation: op,
		Before:    before,
		After:     after,
		Timestamp: time.Now().UnixNano(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return core.LogEntry{}, fmt.Errorf("cdc: marshal error: %w", err)
	}

	topic := fmt.Sprintf("%s.cdc", table)
	return c.engine.Enqueue(topic, string(data))
}
