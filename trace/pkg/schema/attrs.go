package schema

const (
	AttrAgentID   = "pranor.agent_id"
	AttrUserID    = "pranor.user_id"
	AttrTenantID  = "pranor.tenant_id"
	AttrRequestID = "pranor.request_id"
	AttrModule    = "pranor.module"
	AttrOutcome   = "pranor.outcome"
)

const (
	OutcomeAllow     = "ALLOW"
	OutcomeDeny      = "DENY"
	OutcomeApprove   = "APPROVE"
	OutcomeTransform = "TRANSFORM"
	OutcomeError     = "ERROR"
)

const (
	ModuleGate     = "gate"
	ModuleGraph    = "graph"
	ModuleDecision = "decision"
	ModuleFlow     = "flow"
	ModuleLearn    = "learn"
	ModuleTrace    = "trace"
)
