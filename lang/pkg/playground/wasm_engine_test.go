package import (
	"context"
	"strings"
	"testing"
)

func TestPlaygroundWASMEngine_ExecuteSnippet(t *testing.T) {
	pwe := NewPlaygroundWASMEngine()

	source := "fn main() { print('Hello Pranor!'); }"

	res := pwe.ExecuteSnippet(context.Background(), source)
	if !res.Success {
		t.Fatalf("expected successful execution, got error: %s", res.ErrorMsg)
	}

	if !strings.Contains(res.Stdout, "[playground] Executed successfully") {
		t.Errorf("unexpected stdout: %s", res.Stdout)
	}

	// Empty source test
	emptyRes := pwe.ExecuteSnippet(context.Background(), "")
	if emptyRes.Success {
		t.Error("expected error for empty source code")
	}
}
