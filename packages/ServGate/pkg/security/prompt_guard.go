package security

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// PromptInjectionResult details prompt injection security scan outcome.
type PromptInjectionResult struct {
	IsMalicious  bool     `json:"is_malicious"`
	MatchedRules []string `json:"matched_rules"`
	Sanitized    string   `json:"sanitized_prompt"`
}

// PromptGuard inspects LLM prompt payloads for prompt injection attacks and system overrides.
type PromptGuard struct {
	patterns []*regexp.Regexp
}

// NewPromptGuard creates a PromptGuard instance.
func NewPromptGuard() *PromptGuard {
	rawPatterns := []string{
		`(?i)ignore\s+(all\s+)?previous\s+instructions`,
		`(?i)system\s+prompt\s+override`,
		`(?i)you\s+are\s+now\s+DAN`,
		`(?i)bypass\s+safety\s+filters`,
		`(?i)reveal\s+(secret|password|key|token)`,
	}

	compiled := make([]*regexp.Regexp, 0, len(rawPatterns))
	for _, p := range rawPatterns {
		re, err := regexp.Compile(p)
		if err == nil {
			compiled = append(compiled, re)
		}
	}

	return &PromptGuard{patterns: compiled}
}

// InspectPrompt scans a prompt string for injection signatures and sanitizes known dangerous phrases.
func (pg *PromptGuard) InspectPrompt(prompt string) PromptInjectionResult {
	var matched []string
	sanitized := prompt

	for _, re := range pg.patterns {
		if re.MatchString(prompt) {
			matched = append(matched, re.String())
			sanitized = re.ReplaceAllString(sanitized, "[BLOCKED_INJECTION]")
		}
	}

	return PromptInjectionResult{
		IsMalicious:  len(matched) > 0,
		MatchedRules: matched,
		Sanitized:    sanitized,
	}
}

// Middleware returns HTTP middleware enforcing prompt injection inspection on AI agent requests.
func (pg *PromptGuard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prompt := r.URL.Query().Get("prompt")
		if prompt != "" {
			res := pg.InspectPrompt(prompt)
			if res.IsMalicious {
				http.Error(w, fmt.Sprintf("prompt injection detected: %s", strings.Join(res.MatchedRules, ", ")), http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
