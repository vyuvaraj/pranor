package security

import (
	"fmt"
	"regexp"
	"strings"
)

// PIIType denotes the category of sensitive information detected.
type PIIType string

const (
	PIIEmail        PIIType = "EMAIL"
	PIIPhone        PIIType = "PHONE"
	PIISNN          PIIType = "SSN"
	PIICreditCard   PIIType = "CREDIT_CARD"
	PIIName         PIIType = "PERSON_NAME"
	PIIAddress      PIIType = "STREET_ADDRESS"
	PIIOrganization PIIType = "ORGANIZATION"
)

// PIIEntity details a detected entity and its location.
type PIIEntity struct {
	Type  PIIType `json:"type"`
	Value string  `json:"value"`
	Start int     `json:"start"`
	End   int     `json:"end"`
}

// RedactionResult contains the redacted string and entities found.
type RedactionResult struct {
	OriginalText string      `json:"original_text"`
	RedactedText string      `json:"redacted_text"`
	Entities     []PIIEntity `json:"entities"`
}

// NLPPIRedactor implements regex + NLP named-entity recognition (NER) for PII scrubbing (SG.H3).
type NLPPIRedactor struct {
	regexes       map[PIIType]*regexp.Regexp
	titlePrefixes []string
	streetKeywords []string
	orgSuffixes   []string
}

// NewNLPPIRedactor creates a new entity-aware PII redactor.
func NewNLPPIRedactor() *NLPPIRedactor {
	return &NLPPIRedactor{
		regexes: map[PIIType]*regexp.Regexp{
			PIIEmail:      regexp.MustCompile(`(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`),
			PIIPhone:      regexp.MustCompile(`(?:\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`),
			PIISNN:        regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
			PIICreditCard: regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
		},
		titlePrefixes:  []string{"Mr.", "Mrs.", "Ms.", "Dr.", "Prof.", "CEO", "Director", "President"},
		streetKeywords: []string{"Street", "St.", "Avenue", "Ave.", "Road", "Rd.", "Boulevard", "Blvd.", "Lane", "Ln.", "Drive", "Dr."},
		orgSuffixes:    []string{"Inc.", "Corp.", "LLC", "Ltd.", "Technologies", "Group", "Bank", "Foundation"},
	}
}

// RedactScrub scans input text and replaces contextual PII with redaction tokens (SG.H3).
func (nr *NLPPIRedactor) RedactScrub(text string) RedactionResult {
	var entities []PIIEntity
	redacted := text

	// 1. Standard Regex PII Scrubber
	for piiType, re := range nr.regexes {
		matches := re.FindAllStringIndex(redacted, -1)
		for i := len(matches) - 1; i >= 0; i-- {
			loc := matches[i]
			val := redacted[loc[0]:loc[1]]
			entities = append(entities, PIIEntity{
				Type:  piiType,
				Value: val,
				Start: loc[0],
				End:   loc[1],
			})
			replacement := fmt.Sprintf("[%s_REDACTED]", piiType)
			redacted = redacted[:loc[0]] + replacement + redacted[loc[1]:]
		}
	}

	// 2. Contextual NLP Named-Entity Recognition (NER) for Names
	words := strings.Fields(text)
	for i, w := range words {
		for _, prefix := range nr.titlePrefixes {
			if strings.EqualFold(w, prefix) && i+1 < len(words) {
				candidateName := words[i+1]
				if len(candidateName) > 1 && candidateName[0] >= 'A' && candidateName[0] <= 'Z' {
					fullVal := w + " " + candidateName
					entities = append(entities, PIIEntity{
						Type:  PIIName,
						Value: fullVal,
					})
					redacted = strings.Replace(redacted, fullVal, "[PERSON_NAME_REDACTED]", 1)
				}
			}
		}
	}

	// 3. Contextual NER for Street Addresses
	for i, w := range words {
		for _, streetKw := range nr.streetKeywords {
			if strings.EqualFold(w, streetKw) && i > 1 {
				numCandidate := words[i-2]
				streetNameCandidate := words[i-1]
				if regexp.MustCompile(`^\d+`).MatchString(numCandidate) {
					fullAddr := numCandidate + " " + streetNameCandidate + " " + w
					entities = append(entities, PIIEntity{
						Type:  PIIAddress,
						Value: fullAddr,
					})
					redacted = strings.Replace(redacted, fullAddr, "[STREET_ADDRESS_REDACTED]", 1)
				}
			}
		}
	}

	// 4. Contextual NER for Sensitive Organizations
	for i, w := range words {
		for _, orgSuff := range nr.orgSuffixes {
			if strings.EqualFold(w, orgSuff) && i > 0 {
				orgName := words[i-1] + " " + w
				entities = append(entities, PIIEntity{
					Type:  PIIOrganization,
					Value: orgName,
				})
				redacted = strings.Replace(redacted, orgName, "[ORGANIZATION_REDACTED]", 1)
			}
		}
	}

	return RedactionResult{
		OriginalText: text,
		RedactedText: redacted,
		Entities:     entities,
	}
}
