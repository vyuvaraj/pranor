package import (
	"strings"
	"testing"
)

func TestDMARCEvaluator_AlignedPass(t *testing.T) {
	evaluator := NewDMARCEvaluator()

	// From: example.com, SPF: example.com (pass), DKIM: example.com (pass) -> Passes DMARC
	res := evaluator.Evaluate("example.com", "example.com", "pass", "example.com", "pass", PolicyReject)
	if !res.PassesDMARC || !res.SPFAligned || !res.DKIMAligned {
		t.Fatalf("expected aligned pass, got %+v", res)
	}

	xmlReport := FormatAggregateReport(res)
	if !strings.Contains(xmlReport, "<header_from>example.com</header_from>") {
		t.Errorf("unexpected XML report: %s", xmlReport)
	}
}

func TestDMARCEvaluator_AlignmentFail(t *testing.T) {
	evaluator := NewDMARCEvaluator()

	// From: example.com, SPF: attacker.org (pass), DKIM: attacker.org (pass) -> Alignment fails
	res := evaluator.Evaluate("example.com", "attacker.org", "pass", "attacker.org", "pass", PolicyQuarantine)
	if res.PassesDMARC || res.SPFAligned || res.DKIMAligned {
		t.Fatalf("expected alignment failure, got %+v", res)
	}

	if res.PolicyEnforced != PolicyQuarantine {
		t.Errorf("expected policy quarantine, got %s", res.PolicyEnforced)
	}
}
