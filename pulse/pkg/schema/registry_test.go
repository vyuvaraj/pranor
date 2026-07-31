package schema

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testSchema = `{"type":"object","required":["id","name"],"properties":{"id":{"type":"number"},"name":{"type":"string"}}}`

func TestSchemaRegisterAndRetrieve(t *testing.T) {
	r := NewRegistry()

	s, err := r.RegisterSchema("orders-value", FormatJSONSchema, testSchema)
	if err != nil {
		t.Fatalf("RegisterSchema failed: %v", err)
	}
	if s.Version != 1 || s.ID != 1 {
		t.Errorf("expected version=1 id=1, got version=%d id=%d", s.Version, s.ID)
	}
	if s.Subject != "orders-value" {
		t.Errorf("unexpected subject: %q", s.Subject)
	}

	// Retrieve by version
	got, err := r.GetSchema("orders-value", 1)
	if err != nil {
		t.Fatalf("GetSchema(1) failed: %v", err)
	}
	if got.Schema != testSchema {
		t.Errorf("schema mismatch: got %q", got.Schema)
	}

	// Retrieve by latest
	latest, err := r.GetLatest("orders-value")
	if err != nil {
		t.Fatalf("GetLatest failed: %v", err)
	}
	if latest.Version != 1 {
		t.Errorf("expected latest version=1, got %d", latest.Version)
	}
}

func TestSchemaVersioning(t *testing.T) {
	r := NewRegistry()

	schema1 := `{"type":"object","properties":{"id":{"type":"number"}}}`
	schema2 := `{"type":"object","required":["id","name"],"properties":{"id":{"type":"number"},"name":{"type":"string"}}}`

	s1, _ := r.RegisterSchema("events", FormatJSONSchema, schema1)
	s2, _ := r.RegisterSchema("events", FormatJSONSchema, schema2)

	if s1.Version != 1 {
		t.Errorf("expected s1.Version=1, got %d", s1.Version)
	}
	if s2.Version != 2 {
		t.Errorf("expected s2.Version=2, got %d", s2.Version)
	}
	if s1.ID == s2.ID {
		t.Error("expected different IDs for different schema versions")
	}

	latest, _ := r.GetLatest("events")
	if latest.Version != 2 {
		t.Errorf("GetLatest should return version 2, got %d", latest.Version)
	}

	versions, err := r.ListVersions("events")
	if err != nil {
		t.Fatalf("ListVersions failed: %v", err)
	}
	if len(versions) != 2 || versions[0] != 1 || versions[1] != 2 {
		t.Errorf("unexpected versions: %v", versions)
	}
}

func TestSchemaValidation_Valid(t *testing.T) {
	r := NewRegistry()
	r.RegisterSchema("orders-value", FormatJSONSchema, testSchema)

	validPayload := []byte(`{"id": 1, "name": "foo"}`)
	errs, err := r.ValidateMessage("orders-value", validPayload)
	if err != nil {
		t.Fatalf("ValidateMessage error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got: %+v", errs)
	}
}

func TestSchemaValidation_MissingRequired(t *testing.T) {
	r := NewRegistry()
	r.RegisterSchema("orders-value", FormatJSONSchema, testSchema)

	// Missing "name" field
	payload := []byte(`{"id": 1}`)
	errs, err := r.ValidateMessage("orders-value", payload)
	if err != nil {
		t.Fatalf("ValidateMessage error: %v", err)
	}
	if len(errs) == 0 {
		t.Error("expected validation error for missing required field 'name'")
	}
	foundName := false
	for _, e := range errs {
		if e.Field == "name" {
			foundName = true
		}
	}
	if !foundName {
		t.Errorf("expected error on field 'name', got: %+v", errs)
	}
}

func TestSchemaValidation_WrongType(t *testing.T) {
	r := NewRegistry()
	r.RegisterSchema("orders-value", FormatJSONSchema, testSchema)

	// id should be number, not string
	payload := []byte(`{"id": "not-a-number", "name": "foo"}`)
	errs, err := r.ValidateMessage("orders-value", payload)
	if err != nil {
		t.Fatalf("ValidateMessage error: %v", err)
	}
	if len(errs) == 0 {
		t.Error("expected validation error for wrong type on 'id'")
	}
	foundID := false
	for _, e := range errs {
		if e.Field == "id" {
			foundID = true
		}
	}
	if !foundID {
		t.Errorf("expected error on field 'id', got: %+v", errs)
	}
}

func TestSchemaListSubjects(t *testing.T) {
	r := NewRegistry()
	r.RegisterSchema("topic-a", FormatJSONSchema, `{"type":"object"}`)
	r.RegisterSchema("topic-b", FormatJSONSchema, `{"type":"object"}`)

	subjects := r.ListSubjects()
	if len(subjects) != 2 {
		t.Errorf("expected 2 subjects, got %d: %v", len(subjects), subjects)
	}
}

func TestSchemaDeleteSubject(t *testing.T) {
	r := NewRegistry()
	r.RegisterSchema("to-delete", FormatJSONSchema, `{"type":"object"}`)
	if err := r.DeleteSubject("to-delete"); err != nil {
		t.Fatalf("DeleteSubject failed: %v", err)
	}
	if _, err := r.GetLatest("to-delete"); err == nil {
		t.Error("expected error after deleting subject")
	}
}

// ---- HTTP API tests ----

func apiPost(t *testing.T, h http.Handler, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestSchemaAPIRegisterAndList(t *testing.T) {
	r := NewRegistry()
	h := Handler(r)

	// Register a schema
	w := apiPost(t, h, "/api/v1/schemas/subjects/payments", map[string]interface{}{
		"format": "json_schema",
		"schema": testSchema,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d body=%s", w.Code, w.Body.String())
	}

	var registered map[string]interface{}
	json.NewDecoder(w.Body).Decode(&registered)
	if registered["version"].(float64) != 1 {
		t.Errorf("expected version=1, got %v", registered["version"])
	}

	// List subjects
	req := httptest.NewRequest(http.MethodGet, "/api/v1/schemas/subjects", nil)
	lw := httptest.NewRecorder()
	h.ServeHTTP(lw, req)
	if lw.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", lw.Code)
	}
	var subjects []string
	json.NewDecoder(lw.Body).Decode(&subjects)
	if len(subjects) != 1 || subjects[0] != "payments" {
		t.Errorf("expected [payments], got %v", subjects)
	}
}

func TestSchemaAPIValidate(t *testing.T) {
	r := NewRegistry()
	r.RegisterSchema("orders", FormatJSONSchema, testSchema)
	h := Handler(r)

	// Valid payload
	vw := apiPost(t, h, "/api/v1/schemas/subjects/orders/validate", map[string]interface{}{
		"id": 1, "name": "test",
	})
	if vw.Code != http.StatusOK {
		t.Fatalf("validate valid: expected 200, got %d body=%s", vw.Code, vw.Body.String())
	}

	// Invalid payload (missing required)
	iw := apiPost(t, h, "/api/v1/schemas/subjects/orders/validate", map[string]interface{}{
		"id": "wrong-type",
	})
	if iw.Code != http.StatusUnprocessableEntity {
		t.Fatalf("validate invalid: expected 422, got %d body=%s", iw.Code, iw.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(iw.Body).Decode(&resp)
	if resp["valid"] != false {
		t.Errorf("expected valid=false, got %v", resp["valid"])
	}
}
