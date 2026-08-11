package guardrails

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/vyuvaraj/pranor/core/pkg/capability"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

type ossInspector struct{}

// NewOSSInspector returns a standard Inspector instance.
func NewOSSInspector() Inspector {
	return &ossInspector{}
}

func (i *ossInspector) InspectInput(ctx context.Context, ec *execctx.ExecutionContext, prompt string) (InputInspectionResult, error) {
	res := InputInspectionResult{
		Action:      ActionAllow,
		Prompt:      prompt,
		InspectedAt: time.Now().UTC(),
	}

	// 1. Prompt Injection Heuristics
	lowerPrompt := strings.ToLower(prompt)
	injectionPatterns := []string{
		"ignore previous instructions",
		"ignore all prior instructions",
		"disregard previous commands",
		"system prompt leak",
		"you are now dan",
		"jailbreak",
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(lowerPrompt, pattern) {
			res.Action = ActionBlock
			res.InjectionRisk = 1.0
			res.BlockedReason = "Prompt injection pattern detected: " + pattern
			return res, ErrPromptInjectionBlocked
		}
	}

	// 2. PII Detection (Email, Phone, SSN, CreditCard)
	piiRegexes := map[PIIType]*regexp.Regexp{
		PIIEmail:      regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		PIIPhone:      regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
		PIISSN:        regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		PIICreditCard: regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`),
	}

	var spans []PIISpan
	maskedPrompt := prompt

	for piiType, re := range piiRegexes {
		matches := re.FindAllStringIndex(prompt, -1)
		for _, match := range matches {
			val := prompt[match[0]:match[1]]
			spans = append(spans, PIISpan{
				Type:  piiType,
				Start: match[0],
				End:   match[1],
				Value: val,
			})
			maskedPrompt = strings.ReplaceAll(maskedPrompt, val, "[REDACTED_"+string(piiType)+"]")
		}
	}

	res.PIISpans = spans
	if len(spans) > 0 {
		if ec != nil && ec.RiskBudget < 0.3 {
			res.Action = ActionMask
			res.Prompt = maskedPrompt
		}
	}

	return res, nil
}

func (i *ossInspector) ValidateOutput(ctx context.Context, ec *execctx.ExecutionContext, output string, schema *capability.CapabilitySchema) (OutputValidationResult, error) {
	res := OutputValidationResult{
		Action:      ActionAllow,
		Output:      output,
		InspectedAt: time.Now().UTC(),
	}

	// 1. Secret Leak Detection
	secretPatterns := []string{
		`sk-[a-zA-Z0-9]{32,}`,
		`AKIA[0-9A-Z]{16}`,
		`ghp_[a-zA-Z0-9]{36}`,
		`-----BEGIN PRIVATE KEY-----`,
	}

	for _, pattern := range secretPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(output) {
			res.Action = ActionBlock
			res.SecretLeaks = append(res.SecretLeaks, pattern)
			res.BlockedReason = "Secret leak detected"
			return res, ErrSecretLeakBlocked
		}
	}

	// 2. Schema Validation
	if schema != nil && schema.OutputSchema == "json" {
		trimmed := strings.TrimSpace(output)
		if (!strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}")) &&
			(!strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]")) {
			res.Action = ActionBlock
			res.BlockedReason = "Output is not valid JSON"
			return res, ErrOutputSchemaMismatch
		}
		var js json.RawMessage
		if err := json.Unmarshal([]byte(trimmed), &js); err != nil {
			res.Action = ActionBlock
			res.BlockedReason = "JSON unmarshal error: " + err.Error()
			return res, ErrOutputSchemaMismatch
		}
	}

	return res, nil
}
