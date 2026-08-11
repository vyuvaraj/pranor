package engine

import (
	"context"
	"fmt"
	"math"
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

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}
	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (m *ossEpisodicMemory) RecallSemantic(ctx context.Context, ec *execctx.ExecutionContext, queryVector []float32, topK int) ([]api.MemoryEntry, error) {
	tenantID, agentID := "", ""
	if ec != nil {
		tenantID = ec.TenantID
		agentID = ec.AgentID
	}
	
	var scoredEntries []api.MemoryEntry
	
	m.mu.RLock()
	for _, entry := range m.entries {
		if entry.TenantID != tenantID || entry.AgentID != agentID {
			continue
		}
		
		if len(entry.Vector) > 0 {
			sim := cosineSimilarity(queryVector, entry.Vector)
			// Need a local copy to modify the score safely without race conditions
			scoredEntry := entry
			scoredEntry.Score = sim
			scoredEntries = append(scoredEntries, scoredEntry)
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

// SetVectorForTest is a helper method used for testing vector recall.
func (m *ossEpisodicMemory) SetVectorForTest(id string, vec []float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.entries {
		if m.entries[i].ID == id {
			m.entries[i].Vector = vec
			break
		}
	}
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
