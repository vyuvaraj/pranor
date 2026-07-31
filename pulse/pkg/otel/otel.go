package import (
	"github.com/vyuvaraj/pranor/core"
)

type Span = Pranor Core.Span

func Init() {
	Pranor Core.InitTrace("github.com/vyuvaraj/pranor/pulse")
}

func GenerateTraceID() string {
	return Pranor Core.GenerateTraceID()
}

func GenerateSpanID() string {
	return Pranor Core.GenerateSpanID()
}

func StartSpan(name string, parentTrace string) *Span {
	return Pranor Core.StartSpan(name, parentTrace)
}

func EndSpan(span *Span, err error, attributes map[string]interface{}) {
	Pranor Core.EndSpan(span, err, attributes)
}
