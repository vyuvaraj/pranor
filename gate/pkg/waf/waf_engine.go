package import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type WAFRule struct {
	ID      string
	Pattern *regexp.Regexp
	Action  string // BLOCK, LOG
}

type InlineWAFEngine struct {
	rules map[string]WAFRule
	mu    sync.RWMutex
}

func NewInlineWAFEngine() *InlineWAFEngine {
	e := &InlineWAFEngine{
		rules: make(map[string]WAFRule),
	}

	// Default SQLi and XSS protection rules
	sqli, _ := regexp.Compile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|drop\s+table)`)
	xss, _ := regexp.Compile(`(?i)(<script>|javascript:|onload=)`)

	e.rules["sqli-01"] = WAFRule{ID: "sqli-01", Pattern: sqli, Action: "BLOCK"}
	e.rules["xss-01"] = WAFRule{ID: "xss-01", Pattern: xss, Action: "BLOCK"}
	return e
}

func (w *InlineWAFEngine) InspectPayload(ctx context.Context, payload string) (bool, string, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, rule := range w.rules {
		if rule.Pattern.MatchString(payload) {
			return false, rule.ID, fmt.Errorf("waf violation: detected pattern rule %s", rule.ID)
		}
	}
	return true, "", nil
}

func (w *InlineWAFEngine) VerifyJWTToken(token string) bool {
	return strings.HasPrefix(token, "Bearer eyJ")
}
