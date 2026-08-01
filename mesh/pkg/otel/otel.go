package otel

import (
	"github.com/vyuvaraj/pranor/core"
)

type Span = core.Span

func Init() {
	core.InitTrace("github.com/vyuvaraj/pranor/mesh")
}

func GenerateTraceID() string {
	return core.GenerateTraceID()
}

func GenerateSpanID() string {
	return core.GenerateSpanID()
}

func StartSpan(name string, parentTrace string) *Span {
	return core.StartSpan(name, parentTrace)
}

func EndSpan(span *Span, err error, attributes map[string]interface{}) {
	core.EndSpan(span, err, attributes)
}
