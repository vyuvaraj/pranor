package schema

import (
	"context"
	"fmt"
	"os"
)

type stdoutEmitter struct{}

func (e *stdoutEmitter) EmitSpan(ctx context.Context, span SpanEvent) error {
	spanStr := fmt.Sprintf(`{"span":"%s","module":"%s","outcome":"%s"}`, span.Name, span.Attrs[AttrModule], span.Attrs[AttrOutcome])
	_, err := fmt.Fprintln(os.Stderr, spanStr)
	return err
}

func (e *stdoutEmitter) EmitAgentExecution(ctx context.Context, sc SpanContext, fn func() error) error {
	if fn != nil {
		return fn()
	}
	return nil
}

func init() {
	if DefaultEmitter == nil || (fmt.Sprintf("%T", DefaultEmitter) == "*schema.noopEmitter") {
		DefaultEmitter = &stdoutEmitter{}
	}
}
