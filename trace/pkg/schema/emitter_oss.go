//go:build !enterprise

package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func init() {
	DefaultEmitter = &stdoutEmitter{}
}

type stdoutEmitter struct{}

func (e *stdoutEmitter) EmitSpan(ctx context.Context, span SpanEvent) error {
	go func(s SpanEvent) {
		// Error interface doesn't marshal well with json by default, but this is best-effort.
		b, err := json.Marshal(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal span: %v\n", err)
			return
		}
		fmt.Fprintf(os.Stderr, "SPAN: %s\n", string(b))
	}(span)
	return nil
}

func (e *stdoutEmitter) EmitAgentExecution(ctx context.Context, sc SpanContext, fn func() error) error {
	var err error
	if fn != nil {
		err = fn()
	}
	go func(err error) {
		b, mErr := json.Marshal(sc)
		if mErr != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal span context: %v\n", mErr)
			return
		}
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		fmt.Fprintf(os.Stderr, "AGENT EXEC: %s error=%q\n", string(b), errMsg)
	}(err)
	return err
}
