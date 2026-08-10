//go:build !enterprise

package agentgov

// IsEnterpriseAgentGovernanceEnabled indicates if enterprise cluster-wide agent governance is active.
const IsEnterpriseAgentGovernanceEnabled = false

// EnterpriseRiskScoringEngine is the open-source stub for dynamic ML risk scoring.
func EnterpriseRiskScoringEngine(toolName string) float64 {
	// OSS returns baseline static risk scores
	switch toolName {
	case "payment_refund":
		return 0.9
	case "database_drop":
		return 0.95
	case "user_delete":
		return 0.85
	case "system_reboot":
		return 0.99
	case "permissions_grant":
		return 0.8
	default:
		return 0.1
	}
}

// EnterpriseDistributedBudgetCheck is the open-source stub for local budget checks.
func EnterpriseDistributedBudgetCheck(agentID, sessionID string, currentCount, maxCount int) bool {
	if maxCount > 0 && currentCount >= maxCount {
		return false // Exceeded local session budget
	}
	return true
}
