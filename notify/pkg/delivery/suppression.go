package delivery

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// SuppressionReason indicates why an address was suppressed.
type SuppressionReason string

const (
	ReasonHardBounce SuppressionReason = "hard_bounce"
	ReasonComplaint  SuppressionReason = "complaint"
	ReasonManual     SuppressionReason = "manual"
)

// SuppressedAddress holds suppression record metadata.
type SuppressedAddress struct {
	Email      string            `json:"email"`
	Reason     SuppressionReason `json:"reason"`
	Details    string            `json:"details,omitempty"`
	Created    time.Time         `json:"created_at"`
}

// SuppressionList prevents sending emails to bounced or complaining recipients.
type SuppressionList struct {
	mu   sync.RWMutex
	list map[string]*SuppressedAddress // normalized email -> record
}

// NewSuppressionList creates a SuppressionList instance.
func NewSuppressionList() *SuppressionList {
	return &SuppressionList{
		list: make(map[string]*SuppressedAddress),
	}
}

// AddSuppression adds an email address to the suppression list.
func (sl *SuppressionList) AddSuppression(email string, reason SuppressionReason, details string) {
	norm := normalizeEmail(email)
	if norm == "" {
		return
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.list[norm] = &SuppressedAddress{
		Email:   norm,
		Reason:  reason,
		Details: details,
		Created: time.Now(),
	}
}

// IsSuppressed checks if an email address is present in the suppression list.
func (sl *SuppressionList) IsSuppressed(email string) (bool, *SuppressedAddress) {
	norm := normalizeEmail(email)
	if norm == "" {
		return false, nil
	}

	sl.mu.RLock()
	defer sl.mu.RUnlock()

	rec, ok := sl.list[norm]
	return ok, rec
}

// RemoveSuppression removes an address from the suppression list.
func (sl *SuppressionList) RemoveSuppression(email string) bool {
	norm := normalizeEmail(email)
	sl.mu.Lock()
	defer sl.mu.Unlock()

	_, ok := sl.list[norm]
	if ok {
		delete(sl.list, norm)
	}
	return ok
}

// ValidateRecipient returns an error if the recipient is suppressed.
func (sl *SuppressionList) ValidateRecipient(email string) error {
	if suppressed, rec := sl.IsSuppressed(email); suppressed {
		return fmt.Errorf("recipient '%s' is suppressed due to %s (%s)", email, rec.Reason, rec.Details)
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
