package tooling

import (
	"fmt"
	"strings"

	"github.com/vyuvaraj/serv/packages/Serv-lang/pkg/codegen"
)

// BreakingChangeSeverity indicates the impact level of a schema change.
type BreakingChangeSeverity string

const (
	SeverityBreaking    BreakingChangeSeverity = "BREAKING"
	SeverityNonBreaking BreakingChangeSeverity = "NON_BREAKING"
)

// SchemaChangeDetail describes a detected schema modification between two versions.
type SchemaChangeDetail struct {
	StructName string                 `json:"struct_name"`
	FieldName  string                 `json:"field_name,omitempty"`
	ChangeType string                 `json:"change_type"` // "removed_struct", "removed_field", "type_changed", "added_required_field"
	Severity   BreakingChangeSeverity `json:"severity"`
	Message    string                 `json:"message"`
}

// BreakingChangeDetector compares two schema definitions and identifies backwards-incompatible breaking changes.
type BreakingChangeDetector struct{}

// NewBreakingChangeDetector creates a BreakingChangeDetector instance.
func NewBreakingChangeDetector() *BreakingChangeDetector {
	return &BreakingChangeDetector{}
}

// DetectChanges compares old vs new struct definitions and returns detailed diff analysis.
func (bcd *BreakingChangeDetector) DetectChanges(oldStructs, newStructs []codegen.StructDef) []SchemaChangeDetail {
	var changes []SchemaChangeDetail

	oldMap := make(map[string]codegen.StructDef)
	for _, s := range oldStructs {
		oldMap[s.Name] = s
	}

	newMap := make(map[string]codegen.StructDef)
	for _, s := range newStructs {
		newMap[s.Name] = s
	}

	// 1. Check for removed structs
	for name, oldS := range oldMap {
		newS, exists := newMap[name]
		if !exists {
			changes = append(changes, SchemaChangeDetail{
				StructName: name,
				ChangeType: "removed_struct",
				Severity:   SeverityBreaking,
				Message:    fmt.Sprintf("Struct '%s' was completely removed", name),
			})
			continue
		}

		// 2. Compare fields within existing struct
		changes = append(changes, compareStructFields(oldS, newS)...)
	}

	return changes
}

func compareStructFields(oldS, newS codegen.StructDef) []SchemaChangeDetail {
	var changes []SchemaChangeDetail

	oldFields := make(map[string]codegen.FieldDef)
	for _, f := range oldS.Fields {
		oldFields[f.Name] = f
	}

	newFields := make(map[string]codegen.FieldDef)
	for _, f := range newS.Fields {
		newFields[f.Name] = f
	}

	// Check removed fields or mutated types
	for name, oldF := range oldFields {
		newF, exists := newFields[name]
		if !exists {
			changes = append(changes, SchemaChangeDetail{
				StructName: oldS.Name,
				FieldName:  name,
				ChangeType: "removed_field",
				Severity:   SeverityBreaking,
				Message:    fmt.Sprintf("Field '%s.%s' was removed", oldS.Name, name),
			})
			continue
		}

		if !strings.EqualFold(oldF.Type, newF.Type) {
			changes = append(changes, SchemaChangeDetail{
				StructName: oldS.Name,
				FieldName:  name,
				ChangeType: "type_changed",
				Severity:   SeverityBreaking,
				Message:    fmt.Sprintf("Field '%s.%s' type changed from '%s' to '%s'", oldS.Name, name, oldF.Type, newF.Type),
			})
		}
	}

	// Check newly added required fields
	for name, newF := range newFields {
		_, exists := oldFields[name]
		if !exists && !newF.Optional {
			changes = append(changes, SchemaChangeDetail{
				StructName: oldS.Name,
				FieldName:  name,
				ChangeType: "added_required_field",
				Severity:   SeverityBreaking,
				Message:    fmt.Sprintf("Added new required field '%s.%s'", oldS.Name, name),
			})
		}
	}

	return changes
}
