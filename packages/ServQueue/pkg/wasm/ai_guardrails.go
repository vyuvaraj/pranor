//go:build !enterprise

package wasm

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type GuardrailAction string

const (
	ActionAllow  GuardrailAction = "ALLOW"
	ActionBlock  GuardrailAction = "BLOCK"
	ActionRedact GuardrailAction = "REDACT"
)

type GuardrailResult struct {
	Action       GuardrailAction `json:"action"`
	Reason       string          `json:"reason"`
	CleanPayload string          `json:"clean_payload"`
}

type AIGuardrailEngine struct {
	mu            sync.RWMutex
	injectionRegs []*regexp.Regexp
}

func NewAIGuardrailEngine() *AIGuardrailEngine {
	patterns := []string{
		`(?i)ignore\s+previous\s+instructions`,
		`(?i)system\s+prompt\s+override`,
		`(?i)dump\s+database\s+passwords`,
		`(?i)eval\(`,
		`(?i)<script.*?>`,
	}

	var regs []*regexp.Regexp
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err == nil {
			regs = append(regs, re)
		}
	}

	return &AIGuardrailEngine{
		injectionRegs: regs,
	}
}

func (a *AIGuardrailEngine) EvaluatePayload(payload string) GuardrailResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, re := range a.injectionRegs {
		if re.MatchString(payload) {
			return GuardrailResult{
				Action:       ActionBlock,
				Reason:       fmt.Sprintf("ai_guardrail: intercepted security threat matching pattern '%s'", re.String()),
				CleanPayload: "",
			}
		}
	}

	ccReg := regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)
	cleanPayload := ccReg.ReplaceAllString(payload, "[REDACTED_PII]")

	action := ActionAllow
	if strings.Contains(cleanPayload, "[REDACTED_PII]") {
		action = ActionRedact
	}

	return GuardrailResult{
		Action:       action,
		Reason:       "payload passed ai guardrail inspection",
		CleanPayload: cleanPayload,
	}
}
