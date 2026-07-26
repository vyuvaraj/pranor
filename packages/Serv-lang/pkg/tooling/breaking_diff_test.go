package tooling

import (
	"testing"

	"github.com/vyuvaraj/serv/packages/Serv-lang/pkg/codegen"
)

func TestBreakingChangeDetector_DetectChanges(t *testing.T) {
	detector := NewBreakingChangeDetector()

	oldStructs := []codegen.StructDef{
		{
			Name: "Account",
			Fields: []codegen.FieldDef{
				{Name: "ID", Type: "int64", Optional: false},
				{Name: "Balance", Type: "float64", Optional: false},
				{Name: "LegacyCode", Type: "string", Optional: true},
			},
		},
	}

	// Remove LegacyCode, change Balance to int64, add new required Field Email
	newStructs := []codegen.StructDef{
		{
			Name: "Account",
			Fields: []codegen.FieldDef{
				{Name: "ID", Type: "int64", Optional: false},
				{Name: "Balance", Type: "int64", Optional: false}, // Mutated
				{Name: "Email", Type: "string", Optional: false},   // Added required
			},
		},
	}

	changes := detector.DetectChanges(oldStructs, newStructs)

	if len(changes) != 3 {
		t.Fatalf("expected 3 breaking changes detected, got %d: %+v", len(changes), changes)
	}

	for _, c := range changes {
		if c.Severity != SeverityBreaking {
			t.Errorf("expected SeverityBreaking, got %s", c.Severity)
		}
	}
}
