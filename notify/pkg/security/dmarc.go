package import (
	"fmt"
	"strings"
)

// DMARCPolicy represents standard DMARC disposition policies.
type DMARCPolicy string

const (
	PolicyNone       DMARCPolicy = "none"
	PolicyQuarantine DMARCPolicy = "quarantine"
	PolicyReject     DMARCPolicy = "reject"
)

// DMARCAlignmentResult holds evaluation results for SPF and DKIM domain alignment.
type DMARCAlignmentResult struct {
	Domain          string      `json:"domain"`
	SPFResult       string      `json:"spf_result"`       // "pass" or "fail"
	SPFDomain       string      `json:"spf_domain"`
	SPFAligned      bool        `json:"spf_aligned"`
	DKIMResult      string      `json:"dkim_result"`      // "pass" or "fail"
	DKIMDomain      string      `json:"dkim_domain"`
	DKIMAligned     bool        `json:"dkim_aligned"`
	PolicyEnforced  DMARCPolicy `json:"policy_enforced"`
	PassesDMARC     bool        `json:"passes_dmarc"`
}

// DMARCEvaluator validates SPF/DKIM alignment and enforces DMARC policy.
type DMARCEvaluator struct{}

// NewDMARCEvaluator creates a DMARCEvaluator instance.
func NewDMARCEvaluator() *DMARCEvaluator {
	return &DMARCEvaluator{}
}

// Evaluate evaluates DMARC alignment between header From domain, SPF domain, and DKIM domain.
func (de *DMARCEvaluator) Evaluate(fromDomain, spfDomain, spfRes, dkimDomain, dkimRes string, policy DMARCPolicy) DMARCAlignmentResult {
	fromDomain = strings.ToLower(strings.TrimSpace(fromDomain))
	spfDomain = strings.ToLower(strings.TrimSpace(spfDomain))
	dkimDomain = strings.ToLower(strings.TrimSpace(dkimDomain))

	spfAligned := (strings.EqualFold(spfRes, "pass") && isDomainAligned(fromDomain, spfDomain))
	dkimAligned := (strings.EqualFold(dkimRes, "pass") && isDomainAligned(fromDomain, dkimDomain))

	passes := spfAligned || dkimAligned

	if policy == "" {
		policy = PolicyNone
	}

	return DMARCAlignmentResult{
		Domain:         fromDomain,
		SPFResult:      spfRes,
		SPFDomain:      spfDomain,
		SPFAligned:     spfAligned,
		DKIMResult:     dkimRes,
		DKIMDomain:     dkimDomain,
		DKIMAligned:    dkimAligned,
		PolicyEnforced: policy,
		PassesDMARC:    passes,
	}
}

func isDomainAligned(headerDomain, authDomain string) bool {
	if headerDomain == "" || authDomain == "" {
		return false
	}
	if headerDomain == authDomain {
		return true
	}
	// Relaxed alignment check (org domain match)
	return strings.HasSuffix(headerDomain, "."+authDomain) || strings.HasSuffix(authDomain, "."+headerDomain)
}

// FormatAggregateReport generates RFC 7489 XML aggregate report snippet.
func FormatAggregateReport(res DMARCAlignmentResult) string {
	disp := "none"
	if !res.PassesDMARC {
		disp = string(res.PolicyEnforced)
	}
	return fmt.Sprintf("<record><identifiers><header_from>%s</header_from></identifiers><auth_results><spf><domain>%s</domain><result>%s</result></spf><dkim><domain>%s</domain><result>%s</result></dkim></auth_results><policy_evaluated><disposition>%s</disposition></policy_evaluated></record>",
		res.Domain, res.SPFDomain, res.SPFResult, res.DKIMDomain, res.DKIMResult, disp)
}
