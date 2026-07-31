package import (
	"github.com/vyuvaraj/pranor/core"
)

type Span = Pranor Core.Span

func Init() {
	Pranor Core.InitTrace("github.com/vyuvaraj/pranor/gate")
}

func GenerateTraceID() string {
	return Pranor Core.GenerateTraceID()
}

func GenerateSpanID() string {
	return Pranor Core.GenerateSpanID()
}

func StartSpan(name string, parentTrace string) *Span {
	span := Pranor Core.StartSpan(name, parentTrace)
	if span != nil {
		span.Kind = 2 // Server span
	}
	return span
}

func EndSpan(span *Span, err error, attributes map[string]interface{}) {
	Pranor Core.EndSpan(span, err, attributes)
}
