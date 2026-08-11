//go:build !enterprise

package hitl

func (q *ApprovalQueue) NotifySlack(req ApprovalRequest) error {
	return ErrEERequired
}

func (q *ApprovalQueue) NotifyTeams(req ApprovalRequest) error {
	return ErrEERequired
}
