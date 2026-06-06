package schema

import (
	"testing"
)

func TestIsBuiltin(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"trace", true},
		{"version", true},
		{"disposition", true},
		{"disposition-reason", true},
		{"requires-trace-from", true},
		{"status", true},
		{"asil", false},
		{"owner", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBuiltin(tt.name); got != tt.want {
				t.Errorf("IsBuiltin(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestRejectBuiltinRedefinition_Empty(t *testing.T) {
	if err := RejectBuiltinRedefinition(nil); err != nil {
		t.Errorf("RejectBuiltinRedefinition(nil) = %v, want nil", err)
	}
	if err := RejectBuiltinRedefinition(map[string]any{}); err != nil {
		t.Errorf("RejectBuiltinRedefinition({}) = %v, want nil", err)
	}
}

func TestRejectBuiltinRedefinition_Clean(t *testing.T) {
	attrs := map[string]any{
		"asil":  map[string]any{"type": "string"},
		"owner": map[string]any{"type": "string"},
	}
	if err := RejectBuiltinRedefinition(attrs); err != nil {
		t.Errorf("RejectBuiltinRedefinition(clean) = %v, want nil", err)
	}
}

func TestRejectBuiltinRedefinition_RejectsBuiltins(t *testing.T) {
	attrs := map[string]any{
		"status":      map[string]any{"type": "string"},
		"trace":       map[string]any{"type": "array"},
		"disposition": map[string]any{"type": "string"},
	}
	err := RejectBuiltinRedefinition(attrs)
	if err == nil {
		t.Fatal("RejectBuiltinRedefinition should return error for built-in redefinition")
	}
	for _, name := range []string{"trace", "disposition"} {
		if !contains(err.Error(), name) {
			t.Errorf("error should mention %q: %v", name, err)
		}
	}
}

func TestBuiltinPropertyDefs(t *testing.T) {
	defs := BuiltinPropertyDefs()
	if len(defs) != 6 {
		t.Fatalf("BuiltinPropertyDefs() returned %d entries, want 6", len(defs))
	}
	for _, a := range builtinDefs {
		if _, ok := defs[a.Name]; !ok {
			t.Errorf("BuiltinPropertyDefs() missing %q", a.Name)
		}
	}
}

func TestBuiltinNames(t *testing.T) {
	names := BuiltinNames()
	if len(names) != 6 {
		t.Fatalf("BuiltinNames() returned %d names, want 6", len(names))
	}
	// Should be sorted
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("BuiltinNames() not sorted: %v", names)
			break
		}
	}
}

func TestInjectBuiltins_EmptySchema(t *testing.T) {
	schema := map[string]any{}
	InjectBuiltins(schema)

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("InjectBuiltins should create properties map")
	}
	for _, a := range builtinDefs {
		if _, ok := props[a.Name]; !ok {
			t.Errorf("InjectBuiltins missing %q in properties", a.Name)
		}
	}
}

func TestInjectBuiltins_ExistingProps(t *testing.T) {
	schema := map[string]any{
		"properties": map[string]any{
			"asil": map[string]any{"type": "string"},
		},
	}
	InjectBuiltins(schema)

	props := schema["properties"].(map[string]any)
	// User props preserved
	if _, ok := props["asil"]; !ok {
		t.Error("InjectBuiltins removed user-defined 'asil'")
	}
	// Built-in props injected
	for _, a := range builtinDefs {
		if _, ok := props[a.Name]; !ok {
			t.Errorf("InjectBuiltins missing %q in properties", a.Name)
		}
	}
}

func TestRejectBuiltinRedefinition_JSONSchema(t *testing.T) {
	// Simulate what happens when a user puts built-in attrs in JSON Schema properties
	raw := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"properties": map[string]any{
			"trace": map[string]any{"type": "array"},
		},
	}
	_, err := Normalize(raw)
	if err == nil {
		t.Fatal("Normalize should return error when properties contains built-in 'trace'")
	}
	if !contains(err.Error(), "trace") {
		t.Errorf("error should mention 'trace': %v", err)
	}
}

// T1: BuiltinNames must include "status".
func TestBuiltinNames_IncludesStatus(t *testing.T) {
	names := BuiltinNames()
	found := false
	for _, n := range names {
		if n == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("BuiltinNames() = %v, want to include 'status'", names)
	}
}

// T2: Compile injects the built-in `status` attribute with the canonical
// enum [draft, approved] into an empty schema. A subsequent validate
// accepts both values and rejects everything else.
func TestStatus_BuiltInInjectedWithEnum(t *testing.T) {
	raw := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
	}
	c, err := Compile(raw, ".")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if err := c.Validate(map[string]any{"status": "draft"}); err != nil {
		t.Errorf("Validate(draft): %v", err)
	}
	if err := c.Validate(map[string]any{"status": "approved"}); err != nil {
		t.Errorf("Validate(approved): %v", err)
	}
	if err := c.Validate(map[string]any{"status": "Draft"}); err == nil {
		t.Error("Validate(Draft): want error, got nil")
	}
	if err := c.Validate(map[string]any{"status": "Verified"}); err == nil {
		t.Error("Validate(Verified): want error, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
