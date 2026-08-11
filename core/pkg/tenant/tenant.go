package tenant

import (
	"errors"
	"time"

	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

var (
	ErrTenantQuotaExceeded = errors.New("tenant: quota exceeded")
	ErrTenantRateLimited   = errors.New("tenant: rate limited")
	ErrTenantNotFound      = errors.New("tenant: not found")
)

type Quota struct {
	MaxRequestsPerMin   int
	MaxConcurrentAgents int
	MaxTokensPerDay     int
	MaxCostUSDPerDay    float64
}

type UsageStats struct {
	RequestsThisMin int
	ActiveAgents    int
	TokensToday     int
	CostUSDToday    float64
	LastResetMinute time.Time
	LastResetDay    time.Time
}

type Enforcer interface {
	SetQuota(tenantID string, q Quota)
	GetQuota(tenantID string) (Quota, bool)
	Enforce(ec *execctx.ExecutionContext) error
	CheckRateLimit(ec *execctx.ExecutionContext) error
	RecordUsage(ec *execctx.ExecutionContext, tokens int, costUSD float64) error
	ReleaseAgent(ec *execctx.ExecutionContext)
}
