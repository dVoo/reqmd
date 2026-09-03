package exporter

import (
	"encoding/json"
	"reqmd/internal/model"
)

func formatAttr(v any) string {
	switch v.(type) {
	case string, bool, int, int64, float64:
		return model.FormatScalar(v)
	case []any, []string, map[string]any:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return ""
	}
}
