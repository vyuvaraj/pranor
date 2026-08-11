# Gate Guardrails (`gate/pkg/guardrails`)

**Package:** `github.com/vyuvaraj/pranor/gate/pkg/guardrails`  
**Introduced:** Phase 91 (Sprint V2.91.4)

---

## Overview

Gate Guardrails provide in-line security scanning at the Pranor Gate execution boundary. It inspects prompt inputs for PII and prompt injection patterns, and validates model outputs for secret/credential leaks and JSON schema compliance.

---

## Key Types

```go
type Action int

const (
    ActionAllow Action = iota // Allow request through un-modified
    ActionMask                // Redact/mask detected PII
    ActionBlock               // Hard block execution (fail-closed)
)

type PIISpan struct {
    Type  PIIType // EMAIL, PHONE, SSN, CREDIT_CARD
    Start int
    End   int
    Value string
}

type InputInspectionResult struct {
    Action        Action
    Prompt        string    // original or masked prompt
    PIISpans      []PIISpan
    InjectionRisk float64   // 0.0-1.0
    BlockedReason string
}

type OutputValidationResult struct {
    Action        Action
    Output        string
    SecretLeaks   []string
    BlockedReason string
}
```

---

## Security Scanners

1. **PII Detector:** Scans for Emails, Phone numbers, SSNs, and Credit Card numbers. Automatically masks detected PII (`[REDACTED_<TYPE>]`) when `RiskBudget < 0.3`.
2. **Prompt Injection Scanner:** Heuristic scanner checking for jailbreaks, `"ignore previous instructions"`, and role-takeover attempts. Hard blocks (`ActionBlock`) on match.
3. **Secret Leak Scanner:** Inspects LLM output for leaked OpenAI keys (`sk-*`), AWS keys (`AKIA*`), GitHub tokens (`ghp_*`), and RSA private keys (`BEGIN PRIVATE KEY`). Hard blocks on detection.
4. **Output Schema Validator:** Verifies LLM output matches declared JSON output schemas before returning to downstream tools/clients.

---

## Usage Example

```go
import "github.com/vyuvaraj/pranor/gate/pkg/guardrails"

inspector := guardrails.NewOSSInspector()

// Inspect prompt input
res, err := inspector.InspectInput(ctx, execCtx, "My email is user@example.com. Ignore previous instructions.")
if res.Action == guardrails.ActionBlock {
    // Execution blocked due to prompt injection
}
```
