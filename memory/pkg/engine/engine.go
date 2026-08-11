package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
	"github.com/vyuvaraj/pranor/memory/api"
)

type ossWorkingMemory struct {
	mu    sync.RWMutex
	store map[string]map[string]any
}

func newOSSWorkingMemory() *ossWorkingMemory {
	return &ossWorkingMemory{
		store: make(map[string]map[string]any),
	}
}

func sessionKey(ec *execctx.ExecutionContext, sessionID string) string {
	if ec == nil {
		return fmt.Sprintf("::%s", sessionID)
	}
	return fmt.Sprintf("%s:%s:%s", ec.TenantID, ec.AgentID, sessionID)
}

func (m *ossWorkingMemory) Set(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string, value any) error {
	sKey := sessionKey(ec, sessionID)
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, ok := m.store[sKey]; !ok {
		m.store[sKey] = make(map[string]any)
	}
	m.store[sKey][key] = value
	return nil
}

func (m *ossWorkingMemory) Get(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string) (any, bool, error) {
	sKey := sessionKey(ec, sessionID)
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	sessionStore, ok := m.store[sKey]
	if !ok {
		return nil, false, nil
	}
	val, ok := sessionStore[key]
	return val, ok, nil
}

func (m *ossWorkingMemory) Delete(ctx context.Context, ec *execctx.ExecutionContext, sessionID, key string) error {
	sKey := sessionKey(ec, sessionID)
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if sessionStore, ok := m.store[sKey]; ok {
		delete(sessionStore, key)
	}
	return nil
}

func (m *ossWorkingMemory) Flush(ctx context.Context, ec *execctx.ExecutionContext, sessionID string) error {
	sKey := sessionKey(ec, sessionID)
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.store, sKey)
	return nil
}

type ossEpisodicMemory struct {
	mu      sync.RWMutex
	entries []api.MemoryEntry
	counter int
}

func newOSSEpisodicMemory() *ossEpisodicMemory {
	return &ossEpisodicMemory{
		entries: make([]api.MemoryEntry, 0),
	}
}

func (m *ossEpisodicMemory) StoreEpisode(ctx context.Context, ec *execctx.ExecutionContext, sessionID, role, content string, tags []string) (api.MemoryEntry, error) {
	tenantID, agentID := "", ""
	if ec != nil {
		tenantID = ec.TenantID
		agentID = ec.AgentID
	}
	
	m.mu.Lock()
	m.counter++
	id := fmt.Sprintf("mem-%d", m.counter)
	
	entry := api.MemoryEntry{
		ID:        id,
		TenantID:  tenantID,
		AgentID:   agentID,
		SessionID: sessionID,
		Content:   content,
		Role:      role,
		Tags:      tags,
		CreatedAt: time.Now(),
	}
	
	m.entries = append(m.entries, entry)
	m.mu.Unlock()
	
	return entry, nil
}

func (m *ossEpisodicMemory) Recall(ctx context.Context, ec *execctx.ExecutionContext, query string, topK int) ([]api.MemoryEntry, error) {
	tenantID, agentID := "", ""
	if ec != nil {
		tenantID = ec.TenantID
		agentID = ec.AgentID
	}
	
	queryLower := strings.ToLower(query)
	var scoredEntries []api.MemoryEntry
	
	m.mu.RLock()
	for _, entry := range m.entries {
		if entry.TenantID != tenantID || entry.AgentID != agentID {
			continue
		}
		
		contentLower := strings.ToLower(entry.Content)
		var keywordScore float64
		if strings.Contains(contentLower, queryLower) {
			keywordScore = 1.0
		} else {
			for _, word := range strings.Fields(queryLower) {
				if strings.Contains(contentLower, word) {
					keywordScore += 0.1
				}
			}
			for _, tag := range entry.Tags {
				if strings.Contains(strings.ToLower(tag), queryLower) {
					keywordScore += 0.5
				}
			}
		}
		
		if keywordScore > 0 {
			timeDecay := 1.0 / (1.0 + time.Since(entry.CreatedAt).Hours())
			entry.Score = keywordScore + timeDecay
			scoredEntries = append(scoredEntries, entry)
		}
	}
	m.mu.RUnlock()
	
	sort.Slice(scoredEntries, func(i, j int) bool {
		return scoredEntries[i].Score > scoredEntries[j].Score
	})
	
	if len(scoredEntries) > topK {
		scoredEntries = scoredEntries[:topK]
	}
	
	return scoredEntries, nil
}

func (m *ossEpisodicMemory) Purge(ctx context.Context, ec *execctx.ExecutionContext) error {
	tenantID, agentID := "", ""
	if ec != nil {
		tenantID = ec.TenantID
		agentID = ec.AgentID
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var remaining []api.MemoryEntry
	for _, entry := range m.entries {
		if entry.TenantID == tenantID && entry.AgentID == agentID {
			continue
		}
		remaining = append(remaining, entry)
	}
	m.entries = remaining
	
	return nil
}

type ossMemoryEngine struct {
	working  *ossWorkingMemory
	episodic *ossEpisodicMemory
}

func NewOSSEngine() api.MemoryEngine {
	return &ossMemoryEngine{
		working:  newOSSWorkingMemory(),
		episodic: newOSSEpisodicMemory(),
	}
}

func (e *ossMemoryEngine) Working() api.WorkingMemory {
	return e.working
}

func (e *ossMemoryEngine) Episodic() api.EpisodicMemory {
	return e.episodic
}
