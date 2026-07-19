package model

import "strconv"

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
