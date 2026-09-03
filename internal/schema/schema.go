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
	"reqmd/internal/model"
	"sort"
	"strings"

	jsonschema "github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
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

	// Validate x-reqmd.requires-trace-from (the document-wide default
	// coverage expectation) so a malformed default fails fast instead of
	// being silently ignored by graph coverage checks.
	if _, err := collectRequiresTraceFromDefault(normMap, schemaPath); err != nil {
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
	if err = json.Unmarshal(schemaBytes, &s); err != nil {
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

// requiresTraceFromTokenPattern mirrors the built-in requires-trace-from
// item pattern: lowercase letters, digits, hyphen, or underscore. Document
// defaults name document-ids or levels, which follow the same shape.
var requiresTraceFromTokenPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// collectRequiresTraceFromDefault reads x-reqmd.requires-trace-from (the
// document-wide default coverage expectation) from a normalized schema map,
// validates each token, and returns the cleaned slice. A bare string is
// accepted and treated as a one-token list. The empty list is allowed (it
// means every requirement in the document opts out of downstream coverage).
// Returns nil (no error) when the key is absent.
func collectRequiresTraceFromDefault(normMap map[string]any, schemaPath string) ([]string, error) {
	xr, ok := normMap["x-reqmd"].(map[string]any)
	if !ok {
		return nil, nil
	}
	raw, ok := xr["requires-trace-from"]
	if !ok {
		return nil, nil
	}

	var tokens []string
	switch v := raw.(type) {
	case string:
		tokens = []string{v}
	case []any:
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.requires-trace-from: every entry must be a string, got %T", schemaPath, item)
			}
			tokens = append(tokens, s)
		}
	case []string:
		tokens = append(tokens, v...)
	default:
		return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.requires-trace-from: must be an array of strings (or a single string), got %T", schemaPath, raw)
	}

	seen := make(map[string]bool, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, s := range tokens {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.requires-trace-from: empty/whitespace-only token rejected", schemaPath)
		}
		if !requiresTraceFromTokenPattern.MatchString(s) {
			return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.requires-trace-from: invalid token %q: must be lowercase alphanumeric+hyphen+underscore, matching ^[a-z0-9_-]+$", schemaPath, s)
		}
		if seen[s] {
			return nil, fmt.Errorf("%s/schema.yaml: x-reqmd.requires-trace-from: duplicate token %q", schemaPath, s)
		}
		seen[s] = true
		out = append(out, s)
	}
	return out, nil
}

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
	return orderedPropertyNames(s.Required, propertyKeySet(s.Properties))
}

// orderedPropertyNames is the single implementation of the property-order
// convention: required fields in schema order, then remaining property
// keys sorted alphabetically, then built-in attribute names not already
// listed. Shared by PropertyNames (compiled schema) and ExtractProperties
// (raw schema map) so the ordering rule lives in one place.
func orderedPropertyNames(required []string, optional map[string]bool) []string {
	seen := make(map[string]bool, len(required)+len(optional)+len(builtinDefs))
	result := make([]string, 0, len(required)+len(optional)+len(builtinDefs))

	for _, r := range required {
		if !seen[r] {
			result = append(result, r)
			seen[r] = true
		}
	}

	var optKeys []string
	for k := range optional {
		if !seen[k] {
			optKeys = append(optKeys, k)
		}
	}
	sort.Strings(optKeys)
	result = append(result, optKeys...)
	for _, k := range optKeys {
		seen[k] = true
	}

	for _, a := range builtinDefs {
		if !seen[a.Name] {
			result = append(result, a.Name)
		}
	}
	return result
}

// propertyKeySet converts a property map (either form) into a key set for
// orderedPropertyNames. The compiled schema uses *jsonschema.Schema
// values, the raw map uses any.
func propertyKeySet[V any](props map[string]V) map[string]bool {
	set := make(map[string]bool, len(props))
	for k := range props {
		set[k] = true
	}
	return set
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
	return fmt.Errorf("validating attributes: %w", err)
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

	var required []string
	if req, ok := m["required"].([]any); ok {
		required = make([]string, 0, len(req))
		for _, r := range req {
			if s, ok := r.(string); ok {
				required = append(required, s)
			}
		}
	}

	optional := make(map[string]bool)
	if props, ok := m["properties"].(map[string]any); ok {
		for k := range props {
			optional[k] = true
		}
	}

	return orderedPropertyNames(required, optional)
}

// Title extracts a human-readable title from a parsed schema.
// Uses "$id — title", "$id", or "title" from the JSON Schema.
func Title(schema any) string {
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

	// Normalize disjoint-check: accept both string and array forms.
	// A bare string is converted to a single-element array so the
	// yaml.Unmarshal into []string succeeds. Normalization happens on a
	// shallow copy so the caller's schema map (used later for Compile
	// validation) is left pristine.
	if xrSrc, ok := raw.(map[string]any); ok {
		xrMap := make(map[string]any, len(xrSrc))
		for k, v := range xrSrc {
			xrMap[k] = v
		}
		if dc, ok := xrMap["disjoint-check"]; ok {
			switch v := dc.(type) {
			case string:
				if v != "" {
					xrMap["disjoint-check"] = []string{v}
				} else {
					delete(xrMap, "disjoint-check")
				}
			case []any:
				// Already array form — keep as-is.
			}
		}

		// Normalize requires-trace-from the same way: accept a bare
		// string (wrapped into a one-element array), an array of
		// strings, or an empty array (document-wide opt-out, kept
		// distinct from an absent key). Malformed values are dropped
		// from the parsed metadata only; the original value stays in the
		// caller's schema so Compile validation can report it.
		if rtf, ok := xrMap["requires-trace-from"]; ok {
			switch v := rtf.(type) {
			case string:
				if v != "" {
					xrMap["requires-trace-from"] = []string{v}
				} else {
					delete(xrMap, "requires-trace-from")
				}
			case []any:
				strs := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok {
						strs = append(strs, s)
					}
				}
				xrMap["requires-trace-from"] = strs
			default:
				delete(xrMap, "requires-trace-from")
			}
		}

		// Re-marshal the nested map to YAML bytes, then unmarshal into
		// model.XReqmd. This approach handles all field types cleanly
		// via yaml struct tags.
		b, err := yaml.Marshal(xrMap)
		if err != nil {
			return nil
		}

		var xr model.XReqmd
		if err := yaml.Unmarshal(b, &xr); err != nil {
			return nil
		}

		// An explicitly-empty document default must survive the
		// round-trip as a non-nil empty slice (opt-out), never as nil.
		if _, present := xrMap["requires-trace-from"]; present && xr.RequiresTraceFrom == nil {
			xr.RequiresTraceFrom = []string{}
		}

		return &xr
	}

	return nil
}
