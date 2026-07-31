package import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// NewTraceID generates a random 32-character hexadecimal trace ID (16 bytes).
func NewTraceID() string {
	b := make([]byte, 16)
	for {
		if _, err := rand.Read(b); err != nil {
			continue
		}
		allZero := true
		for _, v := range b {
			if v != 0 {
				allZero = false
				break
			}
		}
		if !allZero {
			break
		}
	}
	return hex.EncodeToString(b)
}

// NewSpanID generates a random 16-character hexadecimal span ID (8 bytes).
func NewSpanID() string {
	b := make([]byte, 8)
	for {
		if _, err := rand.Read(b); err != nil {
			continue
		}
		allZero := true
		for _, v := range b {
			if v != 0 {
				allZero = false
				break
			}
		}
		if !allZero {
			break
		}
	}
	return hex.EncodeToString(b)
}

// Inject writes a W3C traceparent header into the given headers map.
// The format is 00-{traceID}-{spanID}-01.
// If traceID or spanID is empty, a new random ID is generated.
func Inject(headers map[string]string, traceID, spanID string) {
	if headers == nil {
		return
	}
	if traceID == "" {
		traceID = NewTraceID()
	}
	if spanID == "" {
		spanID = NewSpanID()
	}
	headers["traceparent"] = fmt.Sprintf("00-%s-%s-01", traceID, spanID)
}

// Extract parses a W3C traceparent header from the given headers map (case-insensitive key search).
// Returns (traceID, spanID, sampled, ok).
func Extract(headers map[string]string) (traceID string, spanID string, sampled bool, ok bool) {
	if headers == nil {
		return "", "", false, false
	}

	var headerVal string
	found := false
	for k, v := range headers {
		if strings.EqualFold(k, "traceparent") {
			headerVal = strings.TrimSpace(v)
			found = true
			break
		}
	}

	if !found || headerVal == "" {
		return "", "", false, false
	}

	parts := strings.Split(headerVal, "-")
	if len(parts) != 4 {
		return "", "", false, false
	}

	// 1. Version must be "00"
	if parts[0] != "00" {
		return "", "", false, false
	}

	// 2. TraceID must be 32 hex chars, valid hex, not all zeros
	traceIDPart := parts[1]
	if len(traceIDPart) != 32 {
		return "", "", false, false
	}
	traceBytes, err := hex.DecodeString(traceIDPart)
	if err != nil {
		return "", "", false, false
	}
	allZeroTrace := true
	for _, b := range traceBytes {
		if b != 0 {
			allZeroTrace = false
			break
		}
	}
	if allZeroTrace {
		return "", "", false, false
	}

	// 3. SpanID must be 16 hex chars, valid hex, not all zeros
	spanIDPart := parts[2]
	if len(spanIDPart) != 16 {
		return "", "", false, false
	}
	spanBytes, err := hex.DecodeString(spanIDPart)
	if err != nil {
		return "", "", false, false
	}
	allZeroSpan := true
	for _, b := range spanBytes {
		if b != 0 {
			allZeroSpan = false
			break
		}
	}
	if allZeroSpan {
		return "", "", false, false
	}

	// 4. Trace flags must be 2 hex chars
	flagsPart := parts[3]
	if len(flagsPart) != 2 {
		return "", "", false, false
	}
	flagBytes, err := hex.DecodeString(flagsPart)
	if err != nil || len(flagBytes) != 1 {
		return "", "", false, false
	}

	sampled = (flagBytes[0] & 0x01) != 0
	return traceIDPart, spanIDPart, sampled, true
}
