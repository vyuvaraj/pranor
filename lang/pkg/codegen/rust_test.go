package codegen

import (
	"strings"
	"testing"
)

func TestRustCodeGenerator_GenerateRust(t *testing.T) {
	generator := NewRustCodeGenerator()

	structs := []StructDef{
		{
			Name: "UserProfile",
			Fields: []FieldDef{
				{Name: "UserID", Type: "int64", Optional: false},
				{Name: "Email", Type: "string", Optional: false},
				{Name: "Bio", Type: "string", Optional: true},
			},
		},
	}

	code := generator.GenerateRust(structs)

	if !strings.Contains(code, "#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]") {
		t.Errorf("missing Serde derive macro: %s", code)
	}
	if !strings.Contains(code, "pub struct UserProfile {") {
		t.Errorf("missing struct header: %s", code)
	}
	if !strings.Contains(code, "pub user_id: i64,") {
		t.Errorf("missing user_id i64 field: %s", code)
	}
	if !strings.Contains(code, "pub bio: Option<String>,") {
		t.Errorf("missing bio Option<String> field: %s", code)
	}
}
