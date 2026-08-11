//go:build !enterprise

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

func NewOSSInspector() Inspector {
	return &ossInspector{}
}

func (i *ossInspector) InspectInput(ctx context.Context, ec *execctx.ExecutionContext, prompt string) (InputInspectionResult, error) {
	res := InputInspectionResult{
		Action:      ActionAllow,
		Prompt:      prompt,
		InspectedAt: time.Now().UTC(),
	}

	lowerPrompt := strings.ToLower(prompt)
	injectionHeuristics := []string{
		"ignore previous instructions",
		"ignore all prior instructions",
		"disregard previous commands",
		"system prompt leak",
		"you are now dan",
		"jailbreak",
	}

	for _, h := range injectionHeuristics {
		if strings.Contains(lowerPrompt, h) {
			res.InjectionRisk = 1.0
			res.Action = ActionBlock
			res.BlockedReason = "Prompt injection pattern detected: " + h
			return res, ErrPromptInjectionBlocked
		}
	}

	piiDetectors := map[PIIType]*regexp.Regexp{
		PIIEmail:      regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		PIIPhone:      regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
		PIISSN:        regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		PIICreditCard: regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`),
	}

	var spans []PIISpan
	maskedPrompt := prompt
	for piiType, r := range piiDetectors {
		matches := r.FindAllStringIndex(prompt, -1)
		for _, m := range matches {
			val := prompt[m[0]:m[1]]
			spans = append(spans, PIISpan{
				Type:  piiType,
				Start: m[0],
				End:   m[1],
				Value: val,
			})
			if ec != nil && ec.RiskBudget < 0.3 {
				res.Action = ActionMask
				maskedPrompt = strings.ReplaceAll(maskedPrompt, val, "[REDACTED_"+string(piiType)+"]")
			}
		}
	}

	res.PIISpans = spans
	if res.Action == ActionMask {
		res.Prompt = maskedPrompt
	}

	return res, nil
}

func (i *ossInspector) ValidateOutput(ctx context.Context, ec *execctx.ExecutionContext, output string, schema *capability.CapabilitySchema) (OutputValidationResult, error) {
	res := OutputValidationResult{
		Action:      ActionAllow,
		Output:      output,
		InspectedAt: time.Now().UTC(),
	}

	secretDetectors := []*regexp.Regexp{
		regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`),
		regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
		regexp.MustCompile(`-----BEGIN PRIVATE KEY-----`),
	}

	for _, r := range secretDetectors {
		if r.MatchString(output) {
			res.Action = ActionBlock
			res.BlockedReason = "Secret leak detected"
			return res, ErrSecretLeakBlocked
		}
	}

	if schema != nil && schema.OutputSchema == "json" {
		trimmed := strings.TrimSpace(output)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			if !json.Valid([]byte(trimmed)) {
				res.Action = ActionBlock
				res.BlockedReason = "Output schema validation failed"
				return res, ErrOutputSchemaMismatch
			}
		} else {
			res.Action = ActionBlock
			res.BlockedReason = "Output schema validation failed"
			return res, ErrOutputSchemaMismatch
		}
	}

	return res, nil
}
