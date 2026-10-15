// Package schema loads the canonical reqmd-code document schema template
// (standard_schema.yaml) and projects it into the exchange document metadata
// that every generated document carries.
//
// standard_schema.yaml is the single source of truth for the attr-block
// properties reqmd-code emits; reqmd-import re-renders it as each document's
// schema.yaml, so there is no second copy to drift.
package schema

import (
	_ "embed"
	"fmt"
	"reqmd-import/exchange"

	"gopkg.in/yaml.v3"
)

//go:embed standard_schema.yaml
var standardSchema string

// Document returns the document metadata — schema header, attr properties,
// level and upstream — read from standard_schema.yaml. Callers override
// Dir, File, Title and IDPrefix per document.
func Document() (exchange.Document, error) {
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(standardSchema), &raw); err != nil {
		return exchange.Document{}, fmt.Errorf("parse standard schema: %w", err)
	}

	d := exchange.Document{
		SchemaID:          asString(raw["$id"]),
		SchemaTitle:       asString(raw["title"]),
		SchemaDescription: asString(raw["description"]),
	}
	if props, ok := raw["properties"].(map[string]any); ok {
		d.Properties = props
	}
	if required, ok := raw["required"].([]any); ok {
		for _, v := range required {
			if s, ok := v.(string); ok {
				d.Required = append(d.Required, s)
			}
		}
	}
	if x, ok := raw["x-reqmd"].(map[string]any); ok {
		d.Level = asString(x["level"])
		d.IDPrefix = asString(x["id-prefix"])
		if up, ok := x["upstream"].(map[string]any); ok {
			d.Upstream = &exchange.Upstream{
				Level:   asString(up["level"]),
				Sources: asStrings(up["sources"]),
			}
		}
	}
	return d, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asStrings(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		if s, ok := it.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
