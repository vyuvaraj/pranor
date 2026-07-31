package tracing

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestTraceparentInjectExtractRoundTrip(t *testing.T) {
	traceID := NewTraceID()
	spanID := NewSpanID()

	headers := make(map[string]string)
	Inject(headers, traceID, spanID)

	expectedHeader := fmt.Sprintf("00-%s-%s-01", traceID, spanID)
	if headers["traceparent"] != expectedHeader {
		t.Fatalf("Expected header %s, got %s", expectedHeader, headers["traceparent"])
	}

	extTraceID, extSpanID, sampled, ok := Extract(headers)
	if !ok {
		t.Fatalf("Extract failed for valid traceparent header")
	}
	if extTraceID != traceID {
		t.Errorf("Expected traceID %s, got %s", traceID, extTraceID)
	}
	if extSpanID != spanID {
		t.Errorf("Expected spanID %s, got %s", spanID, extSpanID)
	}
	if !sampled {
		t.Errorf("Expected sampled true, got false")
	}
}

func TestTraceparentCaseInsensitivity(t *testing.T) {
	traceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	spanID := "00f067aa0ba902b7"
	headerVal := fmt.Sprintf("00-%s-%s-01", traceID, spanID)

	testCases := []string{"traceparent", "Traceparent", "TRACEPARENT", "TraceParent"}

	for _, key := range testCases {
		headers := map[string]string{key: headerVal}
		extTraceID, extSpanID, sampled, ok := Extract(headers)
		if !ok {
			t.Errorf("Extract failed for key %s", key)
		}
		if extTraceID != traceID || extSpanID != spanID || !sampled {
			t.Errorf("Mismatch for key %s: got traceID=%s, spanID=%s, sampled=%v", key, extTraceID, extSpanID, sampled)
		}
	}
}

func TestInjectWithEmptyIDs(t *testing.T) {
	headers := make(map[string]string)
	Inject(headers, "", "")

	extTraceID, extSpanID, sampled, ok := Extract(headers)
	if !ok {
		t.Fatalf("Extract failed on auto-generated traceparent")
	}
	if len(extTraceID) != 32 {
		t.Errorf("Expected traceID length 32, got %d", len(extTraceID))
	}
	if len(extSpanID) != 16 {
		t.Errorf("Expected spanID length 16, got %d", len(extSpanID))
	}
	if !sampled {
		t.Errorf("Expected sampled true")
	}
}

func TestExtractInvalidHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
	}{
		{"nil headers", nil},
		{"missing key", map[string]string{"other-header": "value"}},
		{"empty string", map[string]string{"traceparent": ""}},
		{"too few parts", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7"}},
		{"too many parts", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-extra"}},
		{"invalid version", map[string]string{"traceparent": "01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"}},
		{"traceID too short", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e473-00f067aa0ba902b7-01"}},
		{"traceID non-hex", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e473g-00f067aa0ba902b7-01"}},
		{"traceID all zeros", map[string]string{"traceparent": "00-00000000000000000000000000000000-00f067aa0ba902b7-01"}},
		{"spanID too short", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b-01"}},
		{"spanID non-hex", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902bg-01"}},
		{"spanID all zeros", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000000000-01"}},
		{"flags invalid length", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-1"}},
		{"flags non-hex", map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-zz"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, ok := Extract(tt.headers)
			if ok {
				t.Errorf("Expected Extract to return ok=false for case %q", tt.name)
			}
		})
	}
}

func TestSampledFlagParsing(t *testing.T) {
	traceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	spanID := "00f067aa0ba902b7"

	tests := []struct {
		flags           string
		expectedSampled bool
	}{
		{"00", false},
		{"01", true},
		{"02", false},
		{"03", true},
		{"ff", true},
		{"fe", false},
	}

	for _, tt := range tests {
		headers := map[string]string{
			"traceparent": fmt.Sprintf("00-%s-%s-%s", traceID, spanID, tt.flags),
		}
		_, _, sampled, ok := Extract(headers)
		if !ok {
			t.Errorf("Extract failed for flags %s", tt.flags)
		}
		if sampled != tt.expectedSampled {
			t.Errorf("For flags %s, expected sampled=%v, got %v", tt.flags, tt.expectedSampled, sampled)
		}
	}
}

func TestIDGenerationUniqueness(t *testing.T) {
	const count = 1000
	traceIDs := make(map[string]bool)
	spanIDs := make(map[string]bool)

	for i := 0; i < count; i++ {
		tid := NewTraceID()
		sid := NewSpanID()

		if len(tid) != 32 {
			t.Fatalf("NewTraceID length should be 32, got %d", len(tid))
		}
		if len(sid) != 16 {
			t.Fatalf("NewSpanID length should be 16, got %d", len(sid))
		}

		if _, err := hex.DecodeString(tid); err != nil {
			t.Fatalf("NewTraceID produces invalid hex: %v", err)
		}
		if _, err := hex.DecodeString(sid); err != nil {
			t.Fatalf("NewSpanID produces invalid hex: %v", err)
		}

		if traceIDs[tid] {
			t.Fatalf("Duplicate TraceID generated: %s", tid)
		}
		if spanIDs[sid] {
			t.Fatalf("Duplicate SpanID generated: %s", sid)
		}

		traceIDs[tid] = true
		spanIDs[sid] = true
	}
}
