package model

import (
	"encoding/json"
	"strconv"
)

// FormatScalar converts a primitive scalar value to its string
// representation. It handles the types that YAML decoding produces for
// attribute values: string, bool, int, int64, float64. Non-scalar types
// (slices, maps) return "" — callers should handle those with
// encoding/json or a domain-specific renderer.
//
// This is the shared scalar formatter used by the exporter and reporter
// packages to avoid duplicating the same type switch in four places.
func FormatScalar(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return ""
	}
}

// Attribute accessors. Attr values come from the YAML attr block and are
// typed by the parser (string, bool, int, []any of those). These helpers
// are the single place that type-switches on them; every consumer (graph
// checks, exporters, reporters) reads attrs through them instead of
// reimplementing the same switch.

// AttrString returns the string value of attrs[key], or "" when the key
// is absent or holds a non-string value.
func AttrString(attrs map[string]any, key string) string {
	s, _ := attrs[key].(string)
	return s
}

// AttrStrings reads an array attr as []string. Returns nil when the key is
// absent (so callers can tell "not set" from "explicitly empty") and
// []string{} when the key is present but the array is empty. Non-string
// items are dropped.
func AttrStrings(attrs map[string]any, key string) []string {
	switch v := attrs[key].(type) {
	case []any:
		if v == nil {
			return []string{}
		}
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		if v == nil {
			return []string{}
		}
		out := make([]string, len(v))
		copy(out, v)
		return out
	}
	return nil
}

// AttrList returns an array attr in its native []any form, converting a
// []string. Returns nil when the key is absent or not an array.
func AttrList(attrs map[string]any, key string) []any {
	switch v := attrs[key].(type) {
	case []any:
		return v
	case []string:
		out := make([]any, len(v))
		for i, s := range v {
			out[i] = s
		}
		return out
	}
	return nil
}

// AttrInt reads an integer attr, handling the types that appear in
// practice: int and int64 (YAML decode) and float64 (JSON round-trip).
// Returns 0 when the key is absent or undecodable.
func AttrInt(attrs map[string]any, key string) int {
	switch n := attrs[key].(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

// FormatAttr renders an attribute value for display: scalars through
// FormatScalar, slices and maps as compact JSON. Unknown types render as
// "" so a table cell never shows Go syntax.
func FormatAttr(v any) string {
	switch v.(type) {
	case string, bool, int, int64, float64:
		return FormatScalar(v)
	case []any, []string, map[string]any:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	default:
		return ""
	}
}
