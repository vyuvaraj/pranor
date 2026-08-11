package schema

import (
	"context"
)

type SpanContext struct {
	AgentID   string
	UserID    string
	TenantID  string
	RequestID string
	Module    string
	Outcome   string
}

type SpanEvent struct {
	Name        string
	Attrs       map[string]string
	StartTimeNs int64
	DurationNs  int64
	Error       error
}

type Emitter interface {
	EmitSpan(ctx context.Context, span SpanEvent) error
	EmitAgentExecution(ctx context.Context, sc SpanContext, fn func() error) error
}

var DefaultEmitter Emitter = &noopEmitter{}

type noopEmitter struct{}

func (e *noopEmitter) EmitSpan(ctx context.Context, span SpanEvent) error {
	return nil
}

func (e *noopEmitter) EmitAgentExecution(ctx context.Context, sc SpanContext, fn func() error) error {
	if fn != nil {
		return fn()
	}
	return nil
}

func Emit(ctx context.Context, span SpanEvent) error {
	return DefaultEmitter.EmitSpan(ctx, span)
}

func EmitAgentExecution(ctx context.Context, sc SpanContext, fn func() error) error {
	return DefaultEmitter.EmitAgentExecution(ctx, sc, fn)
}

func TruncateAttr(v string) string {
	if len(v) > 256 {
		return v[:256]
	}
	return v
}
