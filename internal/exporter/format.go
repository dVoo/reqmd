package exporter

import (
	"encoding/json"
	"strconv"
)

func formatAttr(v any) string {
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
	case []any:
		b, _ := json.Marshal(val)
		return string(b)
	case []string:
		b, _ := json.Marshal(val)
		return string(b)
	case map[string]any:
		b, _ := json.Marshal(val)
		return string(b)
	default:
		return ""
	}
}
