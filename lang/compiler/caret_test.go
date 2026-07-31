package compiler

import (
	"strings"
	"testing"
)

func TestCompilerDiagnosticsAndCaretReporter(t *testing.T) {
	source := "fn hello() {\n  let x = \n}"
	errs := []string{"[Line 2, Col 11] expected next token to be ="}

	formatted := FormatDiagnostics(errs, source)
	if !strings.Contains(formatted, "SRV-E001") || !strings.Contains(formatted, "^") {
		t.Fatalf("Expected diagnostic code and source caret pointer '^', got:\n%s", formatted)
	}
}
