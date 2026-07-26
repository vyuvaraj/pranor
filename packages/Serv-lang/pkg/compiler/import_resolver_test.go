package compiler

import (
	"testing"
)

func TestMultiFileImportResolver_RegisterAndResolve(t *testing.T) {
	resolver := NewMultiFileImportResolver()

	modelsContent := `
import "./base.srv"

struct UserProfile {
    id: int64
}

struct OrderItem {
    id: string
}
`

	imports, err := resolver.ParseAndRegisterFile("models.srv", modelsContent)
	if err != nil {
		t.Fatalf("ParseAndRegisterFile failed: %v", err)
	}

	if len(imports) != 1 || imports[0] != "./base.srv" {
		t.Errorf("unexpected imports extracted: %v", imports)
	}

	sym, found := resolver.ResolveSymbol("UserProfile")
	if !found || sym.SourceFile != "models.srv" {
		t.Errorf("expected UserProfile symbol resolved from models.srv, got %+v", sym)
	}

	unresolved := resolver.CheckUnresolvedSymbols([]string{"UserProfile", "MissingType"})
	if len(unresolved) != 1 || unresolved[0] != "MissingType" {
		t.Errorf("unexpected unresolved symbols: %v", unresolved)
	}
}
