package import (
	"testing"
)

func TestSuppressionList_AddAndValidate(t *testing.T) {
	sl := NewSuppressionList()

	email := "bounced-user@example.com"
	sl.AddSuppression(email, ReasonHardBounce, "550 5.1.1 User unknown")

	if suppressed, _ := sl.IsSuppressed("BOUNCED-USER@example.com"); !suppressed {
		t.Fatal("expected email to be suppressed (case-insensitive)")
	}

	err := sl.ValidateRecipient(email)
	if err == nil {
		t.Error("expected error when validating suppressed recipient")
	}

	// Remove suppression
	removed := sl.RemoveSuppression(email)
	if !removed {
		t.Error("expected successful suppression removal")
	}

	if err := sl.ValidateRecipient(email); err != nil {
		t.Errorf("expected clean validation after removal, got: %v", err)
	}
}
