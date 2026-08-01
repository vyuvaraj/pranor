package otel

import (
	"github.com/vyuvaraj/pranor/core"
)

type Span = core.Span

func Init() {
	core.InitTrace("github.com/vyuvaraj/pranor/gate")
}

func GenerateTraceID() string {
	return core.GenerateTraceID()
}

func GenerateSpanID() string {
	return core.GenerateSpanID()
}

func StartSpan(name string, parentTrace string) *Span {
	span := core.StartSpan(name, parentTrace)
	if span != nil {
		span.Kind = 2 // Server span
	}
	return span
}

func EndSpan(span *Span, err error, attributes map[string]interface{}) {
	core.EndSpan(span, err, attributes)
}
