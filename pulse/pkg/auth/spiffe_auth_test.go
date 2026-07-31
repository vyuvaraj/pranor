package auth

import (
	"testing"
)

func TestSPIFFEAuthVerification(t *testing.T) {
	mgr := NewSPIFFEAuthManager()
	spiffeURI := "spiffe://prod.cluster/ns/payments/sa/Pranor Pulse-worker"

	mgr.AllowIdentity(spiffeURI)

	if !mgr.VerifySPIFFEID(spiffeURI) {
		t.Errorf("Expected allowed SPIFFE ID to pass verification")
	}

	if mgr.VerifySPIFFEID("spiffe://untrusted.cluster/sa/hacker") {
		t.Errorf("Expected unallowed SPIFFE ID to fail verification")
	}

	identity, err := ParseSPIFFEID(spiffeURI)
	if err != nil {
		t.Fatalf("ParseSPIFFEID failed: %v", err)
	}

	if identity.Namespace != "payments" || identity.Service != "Pranor Pulse-worker" {
		t.Errorf("Unexpected parsed identity: %+v", identity)
	}
}
