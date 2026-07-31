package import (
	"context"
	"testing"
)

// Minimal valid WASM binary exporting function "add" that returns 42
// (wasm magic + version + export "add")
var sampleWASM = []byte{
	0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // Magic & Version
	0x01, 0x05, 0x01, 0x60, 0x00, 0x01, 0x7f,       // Type section: () -> i32
	0x03, 0x02, 0x01, 0x00,                         // Function section
	0x07, 0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00, // Export section: "add"
	0x0a, 0x06, 0x01, 0x04, 0x00, 0x41, 0x2a, 0x0b, // Code section: i32.const 42; end
}

func TestWASMStepRunner_ExecuteWASM(t *testing.T) {
	ctx := context.Background()
	runner, err := NewWASMStepRunner(ctx)
	if err != nil {
		t.Fatalf("NewWASMStepRunner failed: %v", err)
	}
	defer runner.Close(ctx)

	output, err := runner.ExecuteWASM(ctx, sampleWASM, "add")
	if err != nil {
		t.Fatalf("ExecuteWASM failed: %v", err)
	}

	if string(output) != "42" {
		t.Errorf("expected WASM function result 42, got %s", string(output))
	}
}

func TestWASMStepRunner_MissingFunction(t *testing.T) {
	ctx := context.Background()
	runner, err := NewWASMStepRunner(ctx)
	if err != nil {
		t.Fatalf("NewWASMStepRunner failed: %v", err)
	}
	defer runner.Close(ctx)

	_, err = runner.ExecuteWASM(ctx, sampleWASM, "non_existent_func")
	if err == nil {
		t.Error("expected error for non-existent exported function")
	}
}
