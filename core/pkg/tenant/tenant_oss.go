//go:build !enterprise

package tenant

import (
	"sync"
	"time"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

type ossEnforcer struct {
	mu     sync.RWMutex
	quotas map[string]Quota
	stats  map[string]*UsageStats
}

func NewEnforcer() Enforcer {
	return &ossEnforcer{
		quotas: make(map[string]Quota),
		stats:  make(map[string]*UsageStats),
	}
}

func (e *ossEnforcer) SetQuota(tenantID string, q Quota) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.quotas[tenantID] = q
	if _, ok := e.stats[tenantID]; !ok {
		e.stats[tenantID] = &UsageStats{
			LastResetMinute: time.Now(),
			LastResetDay:    time.Now(),
		}
	}
}

func (e *ossEnforcer) GetQuota(tenantID string) (Quota, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	q, ok := e.quotas[tenantID]
	return q, ok
}

func (e *ossEnforcer) resetStats(now time.Time, s *UsageStats) {
	if now.Sub(s.LastResetMinute) >= time.Minute {
		s.RequestsThisMin = 0
		s.LastResetMinute = now
	}
	if now.Sub(s.LastResetDay) >= 24*time.Hour { // simple 24h reset
		s.TokensToday = 0
		s.CostUSDToday = 0
		s.LastResetDay = now
	}
}

func (e *ossEnforcer) CheckRateLimit(ec *execctx.ExecutionContext) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	q, ok := e.quotas[ec.TenantID]
	if !ok {
		return ErrTenantNotFound
	}

	s, ok := e.stats[ec.TenantID]
	if !ok {
		return ErrTenantNotFound
	}

	now := time.Now()
	e.resetStats(now, s)

	if q.MaxRequestsPerMin > 0 && s.RequestsThisMin >= q.MaxRequestsPerMin {
		return ErrTenantRateLimited
	}

	s.RequestsThisMin++
	return nil
}

func (e *ossEnforcer) Enforce(ec *execctx.ExecutionContext) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	q, ok := e.quotas[ec.TenantID]
	if !ok {
		return ErrTenantNotFound
	}

	s, ok := e.stats[ec.TenantID]
	if !ok {
		return ErrTenantNotFound
	}

	now := time.Now()
	e.resetStats(now, s)

	if q.MaxRequestsPerMin > 0 && s.RequestsThisMin >= q.MaxRequestsPerMin {
		return ErrTenantRateLimited
	}

	if q.MaxConcurrentAgents > 0 && s.ActiveAgents >= q.MaxConcurrentAgents {
		return ErrTenantRateLimited
	}

	s.RequestsThisMin++
	s.ActiveAgents++
	return nil
}

func (e *ossEnforcer) RecordUsage(ec *execctx.ExecutionContext, tokens int, costUSD float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	q, ok := e.quotas[ec.TenantID]
	if !ok {
		return ErrTenantNotFound
	}

	s, ok := e.stats[ec.TenantID]
	if !ok {
		return ErrTenantNotFound
	}

	now := time.Now()
	e.resetStats(now, s)

	if q.MaxTokensPerDay > 0 && s.TokensToday+tokens > q.MaxTokensPerDay {
		return ErrTenantQuotaExceeded
	}

	if q.MaxCostUSDPerDay > 0 && s.CostUSDToday+costUSD > q.MaxCostUSDPerDay {
		return ErrTenantQuotaExceeded
	}

	s.TokensToday += tokens
	s.CostUSDToday += costUSD
	return nil
}

func (e *ossEnforcer) ReleaseAgent(ec *execctx.ExecutionContext) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if s, ok := e.stats[ec.TenantID]; ok {
		if s.ActiveAgents > 0 {
			s.ActiveAgents--
		}
	}
}
