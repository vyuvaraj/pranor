package import (
	"testing"
)

func TestWebAuthnRegistrationAndAssertion(t *testing.T) {
	engine := NewWebAuthnEngine("Pranor Auth.dev", "Pranor Auth Identity Provider")

	user := WebAuthnUser{
		ID:          "user-123",
		Name:        "alice@example.com",
		DisplayName: "Alice Developer",
	}

	// 1. Generate Creation Options
	opts, err := engine.GenerateCreationOptions(user)
	if err != nil {
		t.Fatalf("GenerateCreationOptions failed: %v", err)
	}

	if opts.Challenge == "" {
		t.Fatal("expected non-empty challenge")
	}
	if opts.RP.ID != "Pranor Auth.dev" {
		t.Errorf("expected RP ID Pranor Auth.dev, got %s", opts.RP.ID)
	}

	// 2. Verify Registration
	credID := []byte("cred-id-abc-123")
	pubKey := []byte("fake-public-key-bytes")

	err = engine.VerifyRegistration(user.ID, opts.Challenge, credID, pubKey)
	if err != nil {
		t.Fatalf("VerifyRegistration failed: %v", err)
	}

	// 3. Challenge consumption check (reusing consumed challenge must fail)
	err = engine.VerifyRegistration(user.ID, opts.Challenge, credID, pubKey)
	if err == nil {
		t.Error("expected failure when reusing consumed challenge")
	}

	// 4. Generate Request Options for login
	reqOpts, err := engine.GenerateRequestOptions(user.ID)
	if err != nil {
		t.Fatalf("GenerateRequestOptions failed: %v", err)
	}

	if len(reqOpts.AllowCredentials) != 1 {
		t.Fatalf("expected 1 registered credential, got %d", len(reqOpts.AllowCredentials))
	}

	// 5. Verify Assertion
	authData := []byte("authenticator-data-payload")
	sig := []byte("simulated-signature-bytes")

	err = engine.VerifyAssertion(user.ID, reqOpts.Challenge, credID, authData, sig)
	if err != nil {
		t.Fatalf("VerifyAssertion failed: %v", err)
	}
}

func TestWebAuthnExpiredOrInvalidChallenge(t *testing.T) {
	engine := NewWebAuthnEngine("Pranor Auth.dev", "Pranor Auth")

	err := engine.VerifyRegistration("user-1", "invalid-challenge", []byte("cid"), []byte("pk"))
	if err == nil {
		t.Error("expected error for invalid challenge")
	}

	err = engine.VerifyAssertion("user-1", "invalid-challenge", []byte("cid"), []byte("auth"), []byte("sig"))
	if err == nil {
		t.Error("expected error for invalid assertion challenge")
	}
}
