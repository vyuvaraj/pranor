package guardrails

import (
	"context"
	"errors"
	"time"

	"github.com/vyuvaraj/pranor/core/pkg/capability"
	"github.com/vyuvaraj/pranor/core/pkg/execctx"
)

// Action represents the guardrail enforcement decision.
type Action int

const (
	ActionAllow Action = iota
	ActionMask
	ActionBlock
)

func (a Action) String() string {
	switch a {
	case ActionAllow:
		return "ALLOW"
	case ActionMask:
		return "MASK"
	case ActionBlock:
		return "BLOCK"
	default:
		return "UNKNOWN"
	}
}

// PIIType defines categories of detected PII.
type PIIType string

const (
	PIIEmail      PIIType = "EMAIL"
	PIIPhone      PIIType = "PHONE"
	PIISSN        PIIType = "SSN"
	PIICreditCard PIIType = "CREDIT_CARD"
)

// PIISpan describes a single detected PII fragment.
type PIISpan struct {
	Type  PIIType
	Start int
	End   int
	Value string
}

// InputInspectionResult holds findings from input prompt inspection.
type InputInspectionResult struct {
	Action         Action
	Prompt         string    // original or masked prompt
	PIISpans       []PIISpan
	InjectionRisk  float64   // 0.0 to 1.0
	BlockedReason  string
	InspectedAt    time.Time
}

// OutputValidationResult holds findings from output inspection.
type OutputValidationResult struct {
	Action        Action
	Output        string
	SecretLeaks   []string
	BlockedReason string
	InspectedAt   time.Time
}

// Inspector evaluates prompt inputs and LLM outputs against security guardrails.
type Inspector interface {
	InspectInput(ctx context.Context, ec *execctx.ExecutionContext, prompt string) (InputInspectionResult, error)
	ValidateOutput(ctx context.Context, ec *execctx.ExecutionContext, output string, schema *capability.CapabilitySchema) (OutputValidationResult, error)
}

// Sentinel errors
var (
	ErrPromptInjectionBlocked = errors.New("pranor/gate/guardrails: prompt injection detected")
	ErrPIIViolationBlocked   = errors.New("pranor/gate/guardrails: PII policy violation")
	ErrSecretLeakBlocked      = errors.New("pranor/gate/guardrails: output secret leak detected")
	ErrOutputSchemaMismatch   = errors.New("pranor/gate/guardrails: output schema validation failed")
)
