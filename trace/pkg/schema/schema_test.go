package schema

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestConstants(t *testing.T) {
	spans := []string{
		SpanAgentExecution, SpanGateInspect, SpanGraphContext, SpanGraphCache,
		SpanGraphSQL, SpanDecisionEvaluate, SpanDecisionAuth, SpanDecisionBudget,
		SpanDecisionRisk, SpanDecisionRules, SpanDecisionLearn, SpanFlowSaga,
		SpanFlowStep, SpanLearnPredict,
	}
	for _, s := range spans {
		if s == "" || !strings.HasPrefix(s, "pranor.") {
			t.Errorf("Invalid span constant: %q", s)
		}
	}

	attrs := []string{
		AttrAgentID, AttrUserID, AttrTenantID, AttrRequestID, AttrModule, AttrOutcome,
	}
	for _, a := range attrs {
		if a == "" || !strings.HasPrefix(a, "pranor.") {
			t.Errorf("Invalid attr constant: %q", a)
		}
	}
}

func TestTruncateAttr(t *testing.T) {
	s := strings.Repeat("a", 300)
	trunc := TruncateAttr(s)
	if len(trunc) != 256 {
		t.Errorf("Expected length 256, got %d", len(trunc))
	}

	s2 := "short"
	if TruncateAttr(s2) != s2 {
		t.Errorf("Expected %q, got %q", s2, TruncateAttr(s2))
	}
}

func TestNoopEmitter(t *testing.T) {
	e := &noopEmitter{}
	if err := e.EmitSpan(context.Background(), SpanEvent{}); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	called := false
	err := e.EmitAgentExecution(context.Background(), SpanContext{}, func() error {
		called = true
		return errors.New("test error")
	})
	if !called {
		t.Error("Function was not called")
	}
	if err == nil || err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", err)
	}
}

func TestStdoutEmitterNonBlocking(t *testing.T) {
	e := &stdoutEmitter{}

	done := make(chan bool)
	go func() {
		_ = e.EmitSpan(context.Background(), SpanEvent{Name: "test"})
		_ = e.EmitAgentExecution(context.Background(), SpanContext{}, func() error { return nil })
		done <- true
	}()

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("Stdout emitter blocked")
	}
}
