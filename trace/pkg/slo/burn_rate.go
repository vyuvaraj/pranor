package import (
	"sync"
	"time"
)

// SLOBurnRateConfig defines multi-window multi-burn-rate parameters (SRE best practices).
type SLOBurnRateConfig struct {
	SLOName             string  `json:"slo_name"`
	TargetSLO           float64 `json:"target_slo"`            // e.g. 0.999 (99.9%)
	FastBurnWindow      time.Duration `json:"fast_burn_window"` // e.g., 1h
	FastBurnThreshold   float64 `json:"fast_burn_threshold"`   // e.g., 14.4x
	SlowBurnWindow      time.Duration `json:"slow_burn_window"` // e.g., 6h
	SlowBurnThreshold   float64 `json:"slow_burn_threshold"`   // e.g., 6.0x
}

// BurnRateAlert represents a predicted SLO budget breach alert.
type BurnRateAlert struct {
	SLOName       string    `json:"slo_name"`
	AlertType     string    `json:"alert_type"`     // "fast_burn" or "slow_burn"
	BurnRate      float64   `json:"burn_rate"`      // Actual burn rate multiplier
	BudgetConsumed float64  `json:"budget_consumed"` // e.g., 0.02 (2% of monthly budget consumed in 1 hour)
	TimeToExhaust time.Duration `json:"time_to_exhaustion"`
	Timestamp     time.Time `json:"timestamp"`
}

// SloBreachPredictor evaluates error rates over multiple time windows to predict SLO exhaustion.
type SloBreachPredictor struct {
	mu  sync.RWMutex
	cfg SLOBurnRateConfig
}

// NewSloBreachPredictor creates an SloBreachPredictor instance.
func NewSloBreachPredictor(cfg SLOBurnRateConfig) *SloBreachPredictor {
	if cfg.TargetSLO <= 0 {
		cfg.TargetSLO = 0.999
	}
	if cfg.FastBurnThreshold <= 0 {
		cfg.FastBurnThreshold = 14.4
	}
	if cfg.SlowBurnThreshold <= 0 {
		cfg.SlowBurnThreshold = 6.0
	}
	return &SloBreachPredictor{cfg: cfg}
}

// EvaluateBurnRate evaluates recent window error rate against SLO budget and predicts exhaustion.
func (sbp *SloBreachPredictor) EvaluateBurnRate(windowErrorRate float64, isFastWindow bool) (*BurnRateAlert, bool) {
	sbp.mu.RLock()
	defer sbp.mu.RUnlock()

	errorBudget := 1.0 - sbp.cfg.TargetSLO // e.g., 0.001 for 99.9%
	if errorBudget <= 0 {
		errorBudget = 0.001
	}

	burnRate := windowErrorRate / errorBudget

	threshold := sbp.cfg.SlowBurnThreshold
	alertType := "slow_burn"
	if isFastWindow {
		threshold = sbp.cfg.FastBurnThreshold
		alertType = "fast_burn"
	}

	if burnRate >= threshold {
		// Time to 100% budget exhaustion = errorBudget / (windowErrorRate / time)
		hoursToExhaust := 720.0 / burnRate // 720 hours in 30 days
		timeToExhaust := time.Duration(hoursToExhaust * float64(time.Hour))

		alert := &BurnRateAlert{
			SLOName:        sbp.cfg.SLOName,
			AlertType:      alertType,
			BurnRate:       burnRate,
			BudgetConsumed: windowErrorRate,
			TimeToExhaust:  timeToExhaust,
			Timestamp:      time.Now(),
		}
		return alert, true
	}

	return nil, false
}
