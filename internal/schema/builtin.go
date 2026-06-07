// Package schema — built-in attribute definitions.
//
// reqmd reserves a set of attribute names that have tool-level meaning:
// trace, disposition, disposition-reason, requires-trace-from, version, status.
// These attributes are always valid in any attr block — they are injected
// into every schema before JSON Schema validation runs. Users must not
// redefine them in schema.yaml.
package schema

import (
	"fmt"
	"sort"
	"strings"

	"reqmd/internal/model"
)

// BuiltinAttr describes a single reqmd built-in attribute.
type BuiltinAttr struct {
	Name       string
	Definition map[string]any // JSON Schema property definition
}

// builtinDefs is the canonical list of all reqmd built-in attributes.
//
// Order defines the display order: trace first (most common), then the
// remaining attributes in the order they were added to the tool.
var builtinDefs = []BuiltinAttr{
	{
		Name: model.AttrTrace,
		Definition: map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":    "string",
				"pattern": "^([a-z0-9_-]+/)?[A-Z][A-Z0-9]*(-[A-Z0-9]+)*-[A-Z]*[0-9]+(~[0-9]+)?$",
			},
			"description": "Upstream requirement ID references for traceability",
		},
	},
	{
		Name: model.AttrDisposition,
		Definition: map[string]any{
			"type":        "string",
			"enum":        []any{"implemented", "deferred", "rejected"},
			"description": "How this requirement's intent is addressed downstream",
		},
	},
	{
		Name: model.AttrDispositionReason,
		Definition: map[string]any{
			"type":        "string",
			"description": "Required when disposition is not 'implemented' — explains why",
		},
	},
	{
		Name: model.AttrRequiresTraceFrom,
		Definition: map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":    "string",
				"pattern": "^[a-z0-9_-]+$",
			},
			"description": "Coverage expectations — which downstream document levels or document-ids are expected to trace to this requirement",
		},
	},
	{
		Name: model.AttrVersion,
		Definition: map[string]any{
			"type":        "integer",
			"description": "Version number — bumped on substantive changes to the requirement statement or rationale",
		},
	},
	{
		Name: model.AttrStatus,
		Definition: map[string]any{
			"type":        "string",
			"enum":        []any{"draft", "approved"},
			"description": "Approval lifecycle. Default: approved. Only 'approved' satisfies traceability coverage.",
		},
	},
}

// builtinSet is a name → true map for fast lookup.
var builtinSet map[string]bool

func init() {
	builtinSet = make(map[string]bool, len(builtinDefs))
	for _, a := range builtinDefs {
		builtinSet[a.Name] = true
	}
}

// IsBuiltin returns true if name is a reserved reqmd built-in attribute.
func IsBuiltin(name string) bool {
	return builtinSet[name]
}

// RejectBuiltinRedefinition checks a user schema's attrs map for any
// attribute names that are reserved by reqmd. Returns an error listing
// all redefined names, or nil if the schema is clean.
func RejectBuiltinRedefinition(attrs map[string]any) error {
	var bad []string
	for key := range attrs {
		if IsBuiltin(key) {
			bad = append(bad, key)
		}
	}
	if len(bad) > 0 {
		sort.Strings(bad)
		return fmt.Errorf("reserved built-in attribute(s) must not appear in schema.yaml: %s; reqmd injects these automatically", strings.Join(bad, ", "))
	}
	return nil
}

// BuiltinPropertyDefs returns a JSON Schema properties map containing
// all built-in attribute definitions. The returned map can be merged
// into a schema's "properties" field.
func BuiltinPropertyDefs() map[string]any {
	m := make(map[string]any, len(builtinDefs))
	for _, a := range builtinDefs {
		m[a.Name] = a.Definition
	}
	return m
}

// BuiltinNames returns the sorted list of built-in attribute names.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtinDefs))
	for _, a := range builtinDefs {
		names = append(names, a.Name)
	}
	sort.Strings(names)
	return names
}

// InjectBuiltins adds all built-in property definitions to the given
// JSON Schema map's "properties" field, creating it if needed.
// It does not modify the schema's "required" list — built-ins are
// always optional.
//
// Must be called after Normalize and before Compile (JSON marshal).
func InjectBuiltins(schema map[string]any) {
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		props = make(map[string]any)
		schema["properties"] = props
	}
	for _, a := range builtinDefs {
		props[a.Name] = a.Definition
	}
}
