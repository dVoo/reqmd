package schema

import (
	"strings"
	"testing"
)

// validSchema is a valid JSON Schema 2020-12 in YAML map form.
func validSchema() map[string]any {
	return map[string]any{
		"$schema":  "https://json-schema.org/draft/2020-12/schema",
		"$id":      "test-requirements",
		"title":    "Test Requirement Attributes",
		"type":     "object",
		"required": []any{"asil", "maturity", "verify"},
		"properties": map[string]any{
			"asil":            map[string]any{"type": "string", "enum": []any{"QM", "A", "B", "C", "D"}},
			"maturity":        map[string]any{"type": "string", "enum": []any{"Concept", "Prototype", "Production"}},
			"verify":          map[string]any{"type": "string", "enum": []any{"Test", "Analysis", "Inspection"}},
			"owner":           map[string]any{"type": "string"},
			"priority":        map[string]any{"type": "string", "enum": []any{"Low", "Medium", "High"}},
			"safety_relevant": map[string]any{"type": "boolean"},
		},
		"additionalProperties": false,
	}
}

func TestCompile_HappyPath(t *testing.T) {
	schemaRaw := validSchema()
	c, err := Compile(schemaRaw, ".")
	if err != nil {
		t.Fatalf("Compile returned unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("Compile returned nil Compiled")
	}
}

func TestCompile_BadSchema(t *testing.T) {
	// Passing a scalar string instead of a schema map should fail unmarshaling.
	_, err := Compile("not a schema", ".")
	if err == nil {
		t.Fatal("Compile should have returned an error for a bad schema")
	}
}

func TestValidate_ValidAttrs(t *testing.T) {
	schemaRaw := validSchema()
	c, err := Compile(schemaRaw, ".")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	attrs := map[string]any{
		"status":   "approved",
		"asil":     "QM",
		"maturity": "Concept",
		"verify":   "Test",
	}
	if err := c.Validate(attrs); err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
}

func TestValidate_MissingRequired(t *testing.T) {
	schemaRaw := validSchema()
	c, err := Compile(schemaRaw, ".")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	attrs := map[string]any{
		"status":   "approved",
		"maturity": "Concept",
		"verify":   "Test",
		// "asil" is missing — required by schema
	}
	err = c.Validate(attrs)
	if err == nil {
		t.Fatal("Validate should have returned an error for missing required field")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "required") {
		t.Errorf("Validate error should mention 'required'; got: %v", err)
	}
}

func TestValidate_WrongEnum(t *testing.T) {
	schemaRaw := validSchema()
	c, err := Compile(schemaRaw, ".")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	attrs := map[string]any{
		"status":   "approved",
		"asil":     "ZZZ", // not in the enum
		"maturity": "Concept",
		"verify":   "Test",
	}
	err = c.Validate(attrs)
	if err == nil {
		t.Fatal("Validate should have returned an error for invalid enum value")
	}
}

func TestValidate_NilAttrs(t *testing.T) {
	schemaRaw := validSchema()
	c, err := Compile(schemaRaw, ".")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Validate(nil) should not panic; the method converts nil to empty map.
	// With required fields missing, it should return a non-nil error.
	err = c.Validate(nil)
	if err == nil {
		t.Fatal("Validate(nil) should return an error (missing required fields)")
	}
}

func TestProperties_NilSchema(t *testing.T) {
	result := ExtractProperties(nil)
	if result != nil {
		t.Fatalf("ExtractProperties(nil) should return nil, got %v", result)
	}
}

func TestProperties_EmptySchema(t *testing.T) {
	result := ExtractProperties(map[string]any{})
	// Even an empty schema should include built-in attr names
	expected := []string{"trace", "disposition", "disposition-reason", "requires-trace-from", "version", "status"}
	if len(result) != len(expected) {
		t.Fatalf("ExtractProperties(empty map) returned %d names, want %d: got %v", len(result), len(expected), result)
	}
	for i, want := range expected {
		if result[i] != want {
			t.Errorf("Properties[%d] = %q, want %q", i, result[i], want)
		}
	}
}

func TestProperties_Ordering(t *testing.T) {
	schemaRaw := validSchema()
	props := ExtractProperties(schemaRaw)

	if props == nil {
		t.Fatal("Properties returned nil for valid schema")
	}

	// Expected order: required fields first (asil, maturity, verify) in
	// schema's required order, then optional fields sorted alphabetically,
	// then built-ins.
	expected := []string{"asil", "maturity", "verify", "owner", "priority", "safety_relevant", "trace", "disposition", "disposition-reason", "requires-trace-from", "version", "status"}

	if len(props) != len(expected) {
		t.Fatalf("Properties returned %d names, want %d\ngot:  %v\nwant: %v",
			len(props), len(expected), props, expected)
	}

	for i, want := range expected {
		if props[i] != want {
			t.Errorf("Properties[%d] = %q, want %q\nfull: %v", i, props[i], want, props)
		}
	}
}

func TestProperties_OnlyRequired(t *testing.T) {
	// Schema with only required fields (no optional properties)
	schemaRaw := map[string]any{
		"type":     "object",
		"required": []any{"a", "b"},
		"properties": map[string]any{
			"a": map[string]any{"type": "string"},
			"b": map[string]any{"type": "integer"},
		},
	}
	props := ExtractProperties(schemaRaw)
	// Required attrs first (schema's required order): a, b, then built-ins appended
	expected := []string{"a", "b", "trace", "disposition", "disposition-reason", "requires-trace-from", "version", "status"}
	if len(props) != len(expected) {
		t.Fatalf("got %v, want %v", props, expected)
	}
	for i, want := range expected {
		if props[i] != want {
			t.Errorf("Properties[%d] = %q, want %q", i, props[i], want)
		}
	}
}

func TestProperties_NoRequired(t *testing.T) {
	// Schema with no required field — all are optional
	schemaRaw := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"z": map[string]any{"type": "string"},
			"a": map[string]any{"type": "string"},
		},
	}
	props := ExtractProperties(schemaRaw)
	// Optional attrs sorted alphabetically (a, z), then built-ins appended
	expected := []string{"a", "z", "trace", "disposition", "disposition-reason", "requires-trace-from", "version", "status"}
	if len(props) != len(expected) {
		t.Fatalf("got %v, want %v", props, expected)
	}
	for i, want := range expected {
		if props[i] != want {
			t.Errorf("Properties[%d] = %q, want %q", i, props[i], want)
		}
	}
}

// ---------------------------------------------------------------------------
// Status built-in: extension mechanism (additional-status-values)
// ---------------------------------------------------------------------------

// T8 / T17: additional-status-values accepted and applied to the built-in
// `status` enum. Empty array is a no-op (T17), values extend the enum
// (T8) so the validator accepts them.
func TestCompile_StatusAdditionsAccepted(t *testing.T) {
	schemaRaw := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"x-reqmd": map[string]any{
			"additional-status-values": []any{"review"},
		},
	}
	c, err := Compile(schemaRaw, ".")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if err := c.Validate(map[string]any{"status": "review"}); err != nil {
		t.Errorf("Validate with extension value: %v", err)
	}
	if err := c.Validate(map[string]any{"status": "approved"}); err != nil {
		t.Errorf("Validate with built-in value: %v", err)
	}
	if err := c.Validate(map[string]any{"status": "nope"}); err == nil {
		t.Error("Validate with unknown value: want error, got nil")
	}
}

func TestCompile_StatusAdditionsEmptyIsNoop(t *testing.T) {
	// T17: empty array must not produce a config-error.
	schemaRaw := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"x-reqmd": map[string]any{
			"additional-status-values": []any{},
		},
	}
	if _, err := Compile(schemaRaw, "."); err != nil {
		t.Fatalf("empty additions should be a no-op, got: %v", err)
	}
}

// T9: built-in collision rejected as config-error.
func TestCompile_StatusAdditionsRejectsBuiltins(t *testing.T) {
	for _, val := range []string{"draft", "approved", "Draft", "APPROVED"} {
		schemaRaw := map[string]any{
			"$schema": "https://json-schema.org/draft/2020-12/schema",
			"type":    "object",
			"x-reqmd": map[string]any{
				"additional-status-values": []any{val},
			},
		}
		_, err := Compile(schemaRaw, ".")
		if err == nil {
			t.Errorf("addition %q should be rejected as built-in collision", val)
		}
	}
}

// T10: capitalized additions rejected (lowercase-only is config-error).
func TestCompile_StatusAdditionsRejectsCapitalized(t *testing.T) {
	schemaRaw := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"x-reqmd": map[string]any{
			"additional-status-values": []any{"Review"},
		},
	}
	_, err := Compile(schemaRaw, ".")
	if err == nil {
		t.Fatal("capitalized addition must be rejected")
	}
	if !strings.Contains(err.Error(), "lowercase") {
		t.Errorf("error must mention 'lowercase', got: %v", err)
	}
	if !strings.Contains(err.Error(), "Review") {
		t.Errorf("error must mention offending value %q, got: %v", "Review", err)
	}
}

// T18: duplicate additions rejected.
func TestCompile_StatusAdditionsRejectsDuplicates(t *testing.T) {
	schemaRaw := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"x-reqmd": map[string]any{
			"additional-status-values": []any{"review", "review"},
		},
	}
	_, err := Compile(schemaRaw, ".")
	if err == nil {
		t.Fatal("duplicate additions must be rejected")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error must mention 'duplicate', got: %v", err)
	}
}

// T12: user schema with `properties: {status: ...}` triggers the
// built-in redefinition check.
func TestCompile_StatusBuiltInIsInjectedAndRedefinitionRejected(t *testing.T) {
	raw := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"properties": map[string]any{
			"status": map[string]any{"type": "string", "enum": []any{"X"}},
		},
	}
	_, err := Compile(raw, ".")
	if err == nil {
		t.Fatal("Compile must reject user-defined properties.status")
	}
	if !strings.Contains(err.Error(), "status") {
		t.Errorf("error must mention 'status', got: %v", err)
	}
}

// T15: schemas with `additionalProperties: false` (e.g. 00-aspice,
// 01-stakeholder) must still accept the new built-in `status` attribute.
// The built-in is injected before the additionalProperties check runs,
// so it's not flagged as an "additional" property.
func TestCompile_AdditionalPropertiesFalseAcceptsBuiltInStatus(t *testing.T) {
	raw := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"asil": map[string]any{"type": "string", "enum": []any{"QM", "A", "B"}},
		},
		"required": []any{"asil"},
	}
	c, err := Compile(raw, ".")
	if err != nil {
		t.Fatalf("Compile with additionalProperties: false: %v", err)
	}
	// Built-in status must validate.
	attrs := map[string]any{
		"asil":   "QM",
		"status": "approved",
	}
	if err := c.Validate(attrs); err != nil {
		t.Errorf("Validate with built-in status: %v", err)
	}
	// Capitalized status (old-style) must be rejected by the enum.
	bad := map[string]any{
		"asil":   "QM",
		"status": "Draft",
	}
	if err := c.Validate(bad); err == nil {
		t.Error("Validate with capitalized status: want error, got nil")
	}
}
