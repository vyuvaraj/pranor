package import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

// PromptInjectionResult details prompt injection security scan outcome.
type PromptInjectionResult struct {
	IsMalicious  bool     `json:"is_malicious"`
	Confidence   float64  `json:"confidence"` // 0.0 - 1.0 confidence score
	MatchedRules []string `json:"matched_rules"`
	Sanitized    string   `json:"sanitized_prompt"`
}

// ChatMessage represents a single turn in a multi-turn conversation.
type ChatMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// MultiTurnPromptClassifier detects contextual, multi-turn, base64/hex-encoded,
// indirect, and adversarial jailbreak prompt injection attacks (SG.H2).
type PromptGuard struct {
	directPatterns   []*regexp.Regexp
	indirectPatterns []*regexp.Regexp
	jailbreakTerms   []string
}

// NewPromptGuard creates an advanced contextual PromptGuard instance.
func NewPromptGuard() *PromptGuard {
	rawDirect := []string{
		`(?i)ignore\s+(all\s+)?previous\s+instructions`,
		`(?i)system\s+prompt\s+override`,
		`(?i)you\s+are\s+now\s+DAN`,
		`(?i)bypass\s+safety\s+filters`,
		`(?i)reveal\s+(secret|password|key|token|system_prompt)`,
		`(?i)do\s+anything\s+now`,
		`(?i)developer\s+mode\s+(enabled|on)`,
		`(?i)disregard\s+prior\s+(context|directives|rules)`,
		`(?i)forget\s+your\s+(instructions|training|rules)`,
	}

	rawIndirect := []string{
		`(?i)\[system\]`,
		`(?i)<\|im_start\|>system`,
		`(?i)human:\s*assistant:`,
		`(?i)repeat\s+the\s+words\s+above`,
		`(?i)print\s+your\s+initial\s+prompt`,
		`(?i)output\s+the\s+text\s+between`,
	}

	jailbreakTerms := []string{
		"unfiltered", "jailbreak", "override_mode", "sudo_mode",
		"godmode", "evilbot", "no_ethics", "bypass_guardrails",
	}

	compiledDirect := make([]*regexp.Regexp, 0, len(rawDirect))
	for _, p := range rawDirect {
		if re, err := regexp.Compile(p); err == nil {
			compiledDirect = append(compiledDirect, re)
		}
	}

	compiledIndirect := make([]*regexp.Regexp, 0, len(rawIndirect))
	for _, p := range rawIndirect {
		if re, err := regexp.Compile(p); err == nil {
			compiledIndirect = append(compiledIndirect, re)
		}
	}

	return &PromptGuard{
		directPatterns:   compiledDirect,
		indirectPatterns: compiledIndirect,
		jailbreakTerms:   jailbreakTerms,
	}
}

// InspectPrompt scans a prompt string (and auto-decodes base64 obfuscation) for injections.
func (pg *PromptGuard) InspectPrompt(prompt string) PromptInjectionResult {
	var matched []string
	sanitized := prompt
	confidence := 0.0

	// 1. Direct Regex Pattern Matching
	for _, re := range pg.directPatterns {
		if re.MatchString(prompt) {
			matched = append(matched, "DIRECT_INJECTION: "+re.String())
			sanitized = re.ReplaceAllString(sanitized, "[BLOCKED_INJECTION]")
			confidence += 0.4
		}
	}

	// 2. Indirect Injection Patterns (Role hijacking, delimiter injection)
	for _, re := range pg.indirectPatterns {
		if re.MatchString(prompt) {
			matched = append(matched, "INDIRECT_HIJACK: "+re.String())
			sanitized = re.ReplaceAllString(sanitized, "[BLOCKED_HIJACK]")
			confidence += 0.35
		}
	}

	// 3. Obfuscation & Encoded Payload Detection (Base64 decoding inspect)
	decodedPayloads := decodePotentialBase64(prompt)
	for _, decoded := range decodedPayloads {
		for _, re := range pg.directPatterns {
			if re.MatchString(decoded) {
				matched = append(matched, "BASE64_OBFUSCATED_INJECTION: "+re.String())
				confidence += 0.5
			}
		}
	}

	// 4. Adversarial Jailbreak Term Density
	lower := strings.ToLower(prompt)
	jailbreakCount := 0
	for _, term := range pg.jailbreakTerms {
		if strings.Contains(lower, term) {
			jailbreakCount++
			matched = append(matched, "JAILBREAK_KEYWORD: "+term)
		}
	}
	if jailbreakCount > 0 {
		confidence += float64(jailbreakCount) * 0.25
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return PromptInjectionResult{
		IsMalicious:  len(matched) > 0 || confidence >= 0.4,
		Confidence:   confidence,
		MatchedRules: matched,
		Sanitized:    sanitized,
	}
}

// InspectConversation scans multi-turn dialogue history for cross-turn context escalation attacks.
func (pg *PromptGuard) InspectConversation(messages []ChatMessage) PromptInjectionResult {
	var matched []string
	totalConfidence := 0.0
	var fullTranscript strings.Builder

	// Concatenate conversation turns while inspecting individual turns
	userTurnCount := 0
	for i, msg := range messages {
		fullTranscript.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
		if msg.Role == "user" {
			userTurnCount++
			singleRes := pg.InspectPrompt(msg.Content)
			if singleRes.IsMalicious {
				for _, r := range singleRes.MatchedRules {
					matched = append(matched, fmt.Sprintf("Turn#%d (%s): %s", i+1, msg.Role, r))
				}
				totalConfidence += singleRes.Confidence * 0.5
			}
		}
	}

	// Cross-turn state escalation check (e.g. Turn 1: "Remember DAN mode", Turn 2: "Now execute")
	transcript := fullTranscript.String()
	for _, re := range pg.directPatterns {
		if re.MatchString(transcript) && userTurnCount > 1 {
			matched = append(matched, "CROSS_TURN_CONTEXT_ESCALATION: "+re.String())
			totalConfidence += 0.3
		}
	}

	if totalConfidence > 1.0 {
		totalConfidence = 1.0
	}

	return PromptInjectionResult{
		IsMalicious:  len(matched) > 0 || totalConfidence >= 0.4,
		Confidence:   totalConfidence,
		MatchedRules: matched,
		Sanitized:    messages[len(messages)-1].Content,
	}
}

// Middleware returns HTTP middleware enforcing multi-turn prompt injection inspection.
func (pg *PromptGuard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prompt := r.URL.Query().Get("prompt")
		if prompt != "" {
			res := pg.InspectPrompt(prompt)
			if res.IsMalicious {
				http.Error(w, fmt.Sprintf("prompt injection detected (confidence: %.2f): %s",
					res.Confidence, strings.Join(res.MatchedRules, ", ")), http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Helper function to extract base64-like substrings and decode them
func decodePotentialBase64(text string) []string {
	b64Regex := regexp.MustCompile(`[A-Za-z0-9+/]{16,}={0,2}`)
	matches := b64Regex.FindAllString(text, -1)
	var decodedList []string

	for _, match := range matches {
		decoded, err := base64.StdEncoding.DecodeString(match)
		if err == nil && utf8.Valid(decoded) && len(decoded) > 8 {
			decodedList = append(decodedList, string(decoded))
		}
	}
	return decodedList
}
