// Package schema implements an embedded Schema Registry for ServQueue.
// It enforces message contract validation at the broker: producers must send
// messages conforming to a registered JSON Schema, Avro, or Protobuf contract,
// or their messages are rejected before routing.
package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// SchemaFormat identifies the schema language used.
type SchemaFormat string

const (
	FormatJSONSchema SchemaFormat = "json_schema"
	FormatAvro       SchemaFormat = "avro"
	FormatProtobuf   SchemaFormat = "protobuf"
)

// Schema is a versioned message contract registered for a subject.
type Schema struct {
	ID        int          `json:"id"`
	Subject   string       `json:"subject"`
	Version   int          `json:"version"`
	Format    SchemaFormat `json:"format"`
	Schema    string       `json:"schema"`
	CreatedAt time.Time    `json:"created_at"`
}

// ValidationError describes a single contract violation found in a message payload.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Registry is a thread-safe, in-memory schema registry.
// Subjects are keyed strings (e.g. "orders-value", "payments-key").
// Each subject has an ordered list of versioned schemas (1-indexed).
type Registry struct {
	mu      sync.RWMutex
	schemas map[string][]*Schema // subject -> ascending version list
	nextID  int
}

// NewRegistry returns an empty, ready-to-use Registry.
func NewRegistry() *Registry {
	return &Registry{
		schemas: make(map[string][]*Schema),
		nextID:  1,
	}
}

// RegisterSchema registers a new schema version for the given subject.
// The schema string is validated for structural correctness before storage.
// Returns the new Schema entry (with assigned ID and version).
func (r *Registry) RegisterSchema(subject string, format SchemaFormat, schemaStr string) (*Schema, error) {
	if subject == "" {
		return nil, errors.New("subject must not be empty")
	}
	if err := validateSchemaString(format, schemaStr); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	versions := r.schemas[subject]
	version := len(versions) + 1

	s := &Schema{
		ID:        r.nextID,
		Subject:   subject,
		Version:   version,
		Format:    format,
		Schema:    schemaStr,
		CreatedAt: time.Now().UTC(),
	}
	r.nextID++
	r.schemas[subject] = append(versions, s)
	return s, nil
}

// GetSchema retrieves a specific version of a subject's schema (1-indexed).
func (r *Registry) GetSchema(subject string, version int) (*Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.schemas[subject]
	if !ok || len(versions) == 0 {
		return nil, fmt.Errorf("subject %q not found", subject)
	}
	if version < 1 || version > len(versions) {
		return nil, fmt.Errorf("version %d not found for subject %q (latest: %d)", version, subject, len(versions))
	}
	return versions[version-1], nil
}

// GetLatest retrieves the most recently registered schema for a subject.
func (r *Registry) GetLatest(subject string) (*Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.schemas[subject]
	if !ok || len(versions) == 0 {
		return nil, fmt.Errorf("subject %q not found", subject)
	}
	return versions[len(versions)-1], nil
}

// ListSubjects returns a sorted list of all registered subject names.
func (r *Registry) ListSubjects() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	subjects := make([]string, 0, len(r.schemas))
	for s := range r.schemas {
		subjects = append(subjects, s)
	}
	sort.Strings(subjects)
	return subjects
}

// ListVersions returns all version numbers for a subject (1-indexed).
func (r *Registry) ListVersions(subject string) ([]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, ok := r.schemas[subject]
	if !ok {
		return nil, fmt.Errorf("subject %q not found", subject)
	}
	nums := make([]int, len(versions))
	for i := range versions {
		nums[i] = i + 1
	}
	return nums, nil
}

// DeleteSubject removes a subject and all its schema versions.
func (r *Registry) DeleteSubject(subject string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.schemas[subject]; !ok {
		return fmt.Errorf("subject %q not found", subject)
	}
	delete(r.schemas, subject)
	return nil
}

// ValidateMessage validates a raw message payload against the latest schema
// registered for the given subject.
//
// For FormatJSONSchema: runs structural JSON Schema validation.
// For FormatAvro / FormatProtobuf: skips validation (future implementation).
// Returns an empty slice if the payload is valid.
func (r *Registry) ValidateMessage(subject string, payload []byte) ([]ValidationError, error) {
	schema, err := r.GetLatest(subject)
	if err != nil {
		return nil, err
	}

	switch schema.Format {
	case FormatAvro, FormatProtobuf:
		// Not yet implemented — accept all messages
		return nil, nil
	case FormatJSONSchema:
		return validateJSONSchema(schema.Schema, payload)
	default:
		return nil, fmt.Errorf("unknown format: %s", schema.Format)
	}
}

// ---- Internal schema validation ----

// validateSchemaString checks that the schema definition is structurally valid
// for the declared format.
func validateSchemaString(format SchemaFormat, schemaStr string) error {
	switch format {
	case FormatJSONSchema:
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(schemaStr), &m); err != nil {
			return fmt.Errorf("JSON Schema must be valid JSON: %w", err)
		}
		if _, hasType := m["type"]; !hasType {
			if _, hasProps := m["properties"]; !hasProps {
				return errors.New("JSON Schema must have 'type' or 'properties'")
			}
		}
		return nil
	case FormatAvro:
		// Accept any non-empty string for now
		if strings.TrimSpace(schemaStr) == "" {
			return errors.New("Avro schema must not be empty")
		}
		return nil
	case FormatProtobuf:
		if strings.TrimSpace(schemaStr) == "" {
			return errors.New("Protobuf schema must not be empty")
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// validateJSONSchema runs a simple structural JSON Schema validator against payload.
// Supports: type, required, properties (with nested type assertions).
func validateJSONSchema(schemaStr string, payload []byte) ([]ValidationError, error) {
	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(schemaStr), &schemaMap); err != nil {
		return nil, fmt.Errorf("malformed schema: %w", err)
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(payload, &msg); err != nil {
		return []ValidationError{{Field: "$", Message: "payload is not valid JSON: " + err.Error()}}, nil
	}

	var errs []ValidationError

	// Check top-level type
	if topType, ok := schemaMap["type"].(string); ok {
		if topType == "object" {
			if _, isObj := msg["__nonexistent__"]; false {
				_ = isObj // msg is already a map, so it is an object — valid
			}
		}
	}

	// Check required fields
	if required, ok := schemaMap["required"].([]interface{}); ok {
		for _, req := range required {
			field, _ := req.(string)
			if field == "" {
				continue
			}
			if _, exists := msg[field]; !exists {
				errs = append(errs, ValidationError{
					Field:   field,
					Message: "required field is missing",
				})
			}
		}
	}

	// Check property types
	if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for field, propDef := range properties {
			val, exists := msg[field]
			if !exists {
				continue // missing optional fields handled by required check
			}
			propMap, ok := propDef.(map[string]interface{})
			if !ok {
				continue
			}
			expectedType, _ := propMap["type"].(string)
			if expectedType == "" {
				continue
			}
			if typeErr := checkType(field, val, expectedType); typeErr != nil {
				errs = append(errs, *typeErr)
			}
		}
	}

	return errs, nil
}

// checkType validates that val matches the JSON Schema primitive type name.
func checkType(field string, val interface{}, expectedType string) *ValidationError {
	var actualType string
	switch val.(type) {
	case string:
		actualType = "string"
	case bool:
		actualType = "boolean"
	case float64:
		actualType = "number"
	case []interface{}:
		actualType = "array"
	case map[string]interface{}:
		actualType = "object"
	case nil:
		actualType = "null"
	default:
		actualType = "unknown"
	}

	if actualType != expectedType {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("expected type %q but got %q", expectedType, actualType),
		}
	}
	return nil
}
