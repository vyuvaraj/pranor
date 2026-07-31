package import (
	"strings"
	"testing"
)

func TestPythonCodeGenerator_GeneratePython(t *testing.T) {
	generator := NewPythonCodeGenerator()

	structs := []StructDef{
		{
			Name: "OrderData",
			Fields: []FieldDef{
				{Name: "OrderID", Type: "string", Optional: false},
				{Name: "Amount", Type: "float64", Optional: false},
				{Name: "Discount", Type: "float64", Optional: true},
			},
		},
	}

	code := generator.GeneratePython(structs)

	if !strings.Contains(code, "from dataclasses import dataclass") {
		t.Errorf("missing dataclass import: %s", code)
	}
	if !strings.Contains(code, "class OrderData:") {
		t.Errorf("missing class header: %s", code)
	}
	if !strings.Contains(code, "order_id: str") {
		t.Errorf("missing order_id str field: %s", code)
	}
	if !strings.Contains(code, "discount: Optional[float] = None") {
		t.Errorf("missing discount Optional[float] field: %s", code)
	}
}
