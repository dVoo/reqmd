// Package schema compiles and validates requirement attribute schemas.
//
// Schema format (schema.yaml) — standard JSON Schema 2020-12 in YAML:
//
//	$schema: "https://json-schema.org/draft/2020-12/schema"
//	$id: "ivi-requirements"
//	title: "IVI Requirements"
//	type: object
//	required: [status, asil, maturity, verify]
//	properties:
//	  status:
//	    type: string
//	    enum: [Draft, "In Review", Approved]
//	  asil:
//	    type: string
//	    enum: [QM, A, B, C, D]
//	  owner:
//	    type: string
//	additionalProperties: false
//
// reqmd built-in attributes (trace, disposition, disposition-reason,
// requires-trace-from, version) are injected automatically and must not
// appear in the schema.
package schema

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	jsonschema "github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"

	"reqmd/internal/model"
)

// Compiled holds a compiled schema ready for validation and property introspection.
type Compiled struct {
	resolved *jsonschema.Resolved
}

// Compile validates the schema, injects built-in attrs, and compiles it
// against JSON Schema 2020-12. schemaPath is the directory containing the
// schema file.
func Compile(schemaRaw any, schemaPath string) (*Compiled, error) {
	norm, err := Normalize(schemaRaw)
	if err != nil {
		return nil, fmt.Errorf("normalizing schema: %w", err)
	}

	// Inject built-in attribute definitions into every schema
	normMap, ok := norm.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema must be a map")
	}

	// Validate and harvest x-reqmd.additional-status-values (extensions to
	// the built-in status enum). Must run BEFORE InjectBuiltins so a
	// user-defined `properties.status` is still rejected by Normalize
	// (built-in redefinition), but the additions themselves are
	// appended to the built-in's enum AFTER injection.
	additions, err := collectStatusAdditions(normMap, schemaPath)
	if err != nil {
		return nil, err
	}

	InjectBuiltins(normMap)
	applyStatusAdditions(normMap, additions)
	norm = normMap

	schemaBytes, err := json.Marshal(norm)
	if err != nil {
		return nil, fmt.Errorf("marshaling schema to JSON: %w", err)
	}

	var s jsonschema.Schema
	if err := json.Unmarshal(schemaBytes, &s); err != nil {
		return nil, fmt.Errorf("unmarshaling schema: %w", err)
	}

	absPath, err := filepath.Abs(schemaPath)
	if err != nil {
		absPath = schemaPath
	}
	baseURI := "file://" + filepath.ToSlash(absPath) + "/schema.yaml"

	resolved, err := s.Resolve(&jsonschema.ResolveOptions{
		BaseURI: baseURI,
	})
	if err != nil {
		return nil, fmt.Errorf("resolving schema: %w", err)
	}

	return &Compiled{resolved: resolved}, nil
}

// statusAdditionPattern matches the lowercase-only requirement for
// x-reqmd.additional-status-values entries: starts with a lowercase letter,
// then lowercase letters, digits, hyphen, or underscore.
var statusAdditionPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// collectStatusAdditions reads x-reqmd.additional-status-values from a
// normalized schema map, validates each entry (lowercase only, no
// duplication, no built-in collision), and returns the cleaned slice.
// Returns nil (no error) when the field is absent or empty — empty
// additions are a no-op.
func collectStatusAdditions(normMap map[string]any, schemaPath string) ([]string, error) {
	xr, ok := normMap["x-reqmd"].(map[string]any)
	if !ok {
		return nil, nil
	}
	raw, ok := xr["additional-status-values"]
	if !ok {
		return nil, nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.additional-status-values: must be an array of strings", schemaPath)
	}
	if len(arr) == 0 {
		return nil, nil
	}

	seen := make(map[string]bool, len(arr))
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.additional-status-values: every entry must be a string, got %T", schemaPath, item)
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.additional-status-values: empty/whitespace-only value rejected", schemaPath)
		}
		if strings.EqualFold(s, model.StatusDraft) || strings.EqualFold(s, model.StatusApproved) {
			return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.additional-status-values: invalid value %q: reserved (use the built-in 'status' enum)", schemaPath, s)
		}
		if !statusAdditionPattern.MatchString(s) {
			return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.additional-status-values: invalid value %q: must be lowercase alphanumeric+hyphen+underscore, matching ^[a-z][a-z0-9_-]*$", schemaPath, s)
		}
		key := strings.ToLower(s)
		if seen[key] {
			return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.additional-status-values: duplicate value %q", schemaPath, s)
		}
		seen[key] = true
		out = append(out, s)
	}
	return out, nil
}

// applyStatusAdditions appends the validated extension values to the
// built-in `status` enum in the schema's properties. Called after
// InjectBuiltins so the built-in definition is already present.
func applyStatusAdditions(normMap map[string]any, additions []string) {
	if len(additions) == 0 {
		return
	}
	props, ok := normMap["properties"].(map[string]any)
	if !ok {
		return
	}
	statusDef, ok := props[model.AttrStatus].(map[string]any)
	if !ok {
		return
	}
	enum, ok := statusDef["enum"].([]any)
	if !ok {
		return
	}
	for _, v := range additions {
		enum = append(enum, v)
	}
	statusDef["enum"] = enum
}

// PropertyNames returns the ordered list of property names from the schema:
// required fields first (in schema order), then optional fields
// alphabetically, then built-in attributes at the end.
func (c *Compiled) PropertyNames() []string {
	s := c.resolved.Schema()
	if s == nil {
		return nil
	}

	var result []string
	seen := make(map[string]bool)

	// Required fields in schema order
	for _, r := range s.Required {
		result = append(result, r)
		seen[r] = true
	}

	// Optional fields sorted alphabetically
	var optKeys []string
	for k := range s.Properties {
		if !seen[k] {
			optKeys = append(optKeys, k)
		}
	}
	sort.Strings(optKeys)
	result = append(result, optKeys...)
	for _, k := range optKeys {
		seen[k] = true
	}

	// Built-in attr names in canonical order
	for _, a := range builtinDefs {
		if !seen[a.Name] {
			result = append(result, a.Name)
		}
	}
	return result
}

// Validate checks a requirement's attrs against the compiled schema.
func (c *Compiled) Validate(attrs map[string]any) error {
	if attrs == nil {
		attrs = map[string]any{}
	}
	err := c.resolved.Validate(attrs)
	if err == nil {
		return nil
	}
	// Augment status-enum violations with a hint so authors immediately
	// see the correct built-in values. Catches "Draft", "Approved",
	// "Verified" — anything not in [draft, approved, ...additions].
	// We gate on both: caller supplied a `status` attr AND the error
	// message references `status` (so we don't pollute other failures).
	if _, hasStatus := attrs[model.AttrStatus]; hasStatus && strings.Contains(err.Error(), "status") {
		return fmt.Errorf("%w\n  → hint: did you mean \"draft\" or \"approved\"?", err)
	}
	return err
}

// Normalize validates a schema and checks for built-in redefinition.
// The schema must be valid JSON Schema 2020-12 (in YAML map form).
// It is returned as-is; built-in attrs are injected later by Compile.
func Normalize(raw any) (any, error) {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema must be a YAML map with '$schema' key")
	}

	// Must have $schema to be valid JSON Schema
	if _, has := m["$schema"]; !has {
		return nil, fmt.Errorf("schema must have a '$schema' key (e.g. \"https://json-schema.org/draft/2020-12/schema\")")
	}

	// Check for built-in redefinition in properties
	if props, ok := m["properties"].(map[string]any); ok {
		if err := RejectBuiltinRedefinition(props); err != nil {
			return nil, err
		}
	}

	return raw, nil
}

// ExtractProperties extracts property order from a raw JSON Schema map.
// Uses the google/jsonschema-go library to parse the schema struct.
// Required fields first (in schema's required order), then optional
// fields alphabetically, then built-in attributes at the end.
func ExtractProperties(schemaRaw any) (propNames []string) {
	m, ok := schemaRaw.(map[string]any)
	if !ok {
		return nil
	}

	var result []string
	seen := make(map[string]bool)

	// Required fields in schema order
	if req, ok := m["required"].([]any); ok {
		for _, r := range req {
			if s, ok := r.(string); ok && !seen[s] {
				result = append(result, s)
				seen[s] = true
			}
		}
	}

	// Optional fields sorted alphabetically
	if props, ok := m["properties"].(map[string]any); ok {
		var optKeys []string
		for k := range props {
			if !seen[k] {
				optKeys = append(optKeys, k)
			}
		}
		sort.Strings(optKeys)
		result = append(result, optKeys...)
		for _, k := range optKeys {
			seen[k] = true
		}
	}

	// Built-in attr names in canonical order
	for _, a := range builtinDefs {
		if !seen[a.Name] {
			result = append(result, a.Name)
		}
	}
	return result
}

// SchemaTitle extracts a human-readable title from a parsed schema.
// Uses "$id — title", "$id", or "title" from the JSON Schema.
func SchemaTitle(schema any) string {
	if m, ok := schema.(map[string]any); ok {
		id, _ := m["$id"].(string)
		title, _ := m["title"].(string)
		switch {
		case id != "" && title != "":
			return id + " — " + title
		case id != "":
			return id
		case title != "":
			return title
		}
	}
	return ""
}

// ExtractXReqmd reads the x-reqmd extension block from a parsed schema.
// Returns nil if the key is absent. All fields are optional.
func ExtractXReqmd(schema any) *model.XReqmd {
	m, ok := schema.(map[string]any)
	if !ok {
		return nil
	}

	raw, ok := m["x-reqmd"]
	if !ok {
		return nil
	}

	// Marshal the nested map to YAML bytes, then unmarshal into model.XReqmd.
	// This approach handles all field types cleanly via yaml struct tags.
	b, err := yaml.Marshal(raw)
	if err != nil {
		return nil
	}

	var xr model.XReqmd
	if err := yaml.Unmarshal(b, &xr); err != nil {
		return nil
	}

	return &xr
}
