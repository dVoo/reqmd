package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// decodeAttrYAML parses a small flat YAML map (the contents of an ```attr
// block) directly into map[string]any, without building the intermediate
// YAML document AST that yaml.v3 allocates.
//
// Supported syntax (covers all real-world attr blocks):
//   - key: value          (scalar: string, int, bool)
//   - key: "quoted value" (quoted string — quotes stripped)
//   - key:                (list — following lines are "  - item")
//   - item1
//   - item2
//   - # comment           (ignored)
//   - blank lines          (ignored)
//
// Values are typed as yaml.v3 would type them:
//   - unquoted → string (unless it's an int or bool literal)
//   - "true"/"false" → bool
//   - integer → int
//   - quoted → string (always, even if contents look like a number)
//   - list items → []any of strings
//
// Returns an error on malformed input (no key:value, bad indentation,
// non-list content under a list key). Falls back to yaml.v3 if the
// input contains syntax this scanner doesn't handle (nested maps,
// flow sequences, multi-line scalars).
func decodeAttrYAML(src string) (map[string]any, error) {
	m := make(map[string]any, 8)
	lines := strings.Split(src, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip blank lines and comments.
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Lines must start at indent 0 (top-level keys).
		if line[0] == ' ' || line[0] == '\t' {
			return nil, fmt.Errorf("unexpected indentation: %q", line)
		}

		// Split key: value.
		keyRaw, valueRaw, found := strings.Cut(trimmed, ":")
		if !found {
			return nil, fmt.Errorf("missing key:value on line: %q", trimmed)
		}
		key := strings.TrimSpace(keyRaw)
		valuePart := strings.TrimSpace(valueRaw)

		if key == "" {
			return nil, fmt.Errorf("empty key on line: %q", trimmed)
		}

		if valuePart != "" {
			// Scalar value on the same line.
			switch valuePart[0] {
			case '"', '\'':
				s, err := unquoteYAML(valuePart)
				if err != nil {
					return nil, fmt.Errorf("line %d: %w", i+1, err)
				}
				m[key] = s
			case '[':
				// Flow sequence: [item1, item2]
				items, err := parseFlowSeq(valuePart)
				if err != nil {
					return nil, fmt.Errorf("line %d: %w", i+1, err)
				}
				m[key] = items
			default:
				if strings.HasPrefix(valuePart, "#") {
					// key: #comment → empty value (treat as null/nil)
					m[key] = nil
				} else {
					m[key] = parseScalar(valuePart)
				}
			}
			continue
		}

		// Empty value after colon — expect a list (  - item lines).
		var items []any
		listIndent := -1
		for i+1 < len(lines) {
			next := lines[i+1]
			nextTrim := strings.TrimSpace(next)
			if nextTrim == "" || strings.HasPrefix(nextTrim, "#") {
				i++
				continue
			}
			indent := leadingSpaces(next)
			// Non-indented line → end of list (next top-level key).
			if indent == 0 {
				break
			}
			// Indented but not a list item → unsupported nested structure.
			if nextTrim[0] != '-' {
				return nil, fmt.Errorf("unsupported nested structure under key %q", key)
			}
			if listIndent < 0 {
				listIndent = indent
			} else if indent != listIndent {
				return nil, fmt.Errorf("inconsistent list indentation under key %q", key)
			}
			itemText := strings.TrimSpace(nextTrim[1:])
			if itemText == "" {
				items = append(items, nil)
			} else {
				switch itemText[0] {
				case '"', '\'':
					s, err := unquoteYAML(itemText)
					if err != nil {
						return nil, fmt.Errorf("line %d: %w", i+2, err)
					}
					items = append(items, s)
				default:
					items = append(items, parseScalar(itemText))
				}
			}
			i++
		}
		if items != nil {
			m[key] = items
		} else {
			m[key] = nil
		}
	}

	return m, nil
}

// parseScalar converts an unquoted YAML scalar to the appropriate Go type.
// Integers → int, true/false → bool, everything else → string.
// This matches yaml.v3's behavior for plain scalars.
func parseScalar(s string) any {
	// Strip inline comments: value # comment
	if idx := strings.Index(s, " #"); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}
	// bool
	switch s {
	case "true", "True", "TRUE":
		return true
	case "false", "False", "FALSE":
		return false
	case "null", "Null", "NULL", "~":
		return nil
	}
	// int — only call strconv.Atoi when the string is actually an integer
	// literal. Atoi allocates an error object on non-numeric input, which
	// is the common case for attribute values (IDs, statuses, priorities).
	if isIntegerString(s) {
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
	}
	// string
	return s
}

// isIntegerString reports whether s is a plain integer literal (optional
// sign followed by digits). Used as the gate before strconv.Atoi.
func isIntegerString(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	if s[0] == '-' || s[0] == '+' {
		if len(s) == 1 {
			return false
		}
		i = 1
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// unquoteYAML strips single or double quotes from a YAML scalar.
func unquoteYAML(s string) (string, error) {
	if len(s) < 2 {
		return s, nil
	}
	q := s[0]
	if q != '"' && q != '\'' {
		return s, nil
	}
	if s[len(s)-1] != q {
		return "", fmt.Errorf("unterminated quote: %s", s)
	}
	inner := s[1 : len(s)-1]
	if q == '\'' {
		// YAML single quotes: '' is an escaped literal quote.
		return strings.ReplaceAll(inner, "''", "'"), nil
	}
	// YAML double quotes: handle common escapes.
	return unescapeYAMLDouble(inner), nil
}

func unescapeYAMLDouble(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			default:
				b.WriteByte(s[i+1])
			}
			i++
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func leadingSpaces(s string) int {
	n := 0
	for n < len(s) && s[n] == ' ' {
		n++
	}
	return n
}

// parseFlowSeq parses a YAML flow sequence [item1, item2, item3] into []any.
// Items are parsed as scalars (same rules as parseScalar). Quoted items
// are supported.
func parseFlowSeq(s string) ([]any, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil, fmt.Errorf("invalid flow sequence: %s", s)
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return []any{}, nil
	}
	parts := splitFlowItems(inner)
	items := make([]any, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part[0] == '"' || part[0] == '\'' {
			s, err := unquoteYAML(part)
			if err != nil {
				return nil, err
			}
			items = append(items, s)
		} else {
			items = append(items, parseScalar(part))
		}
	}
	return items, nil
}

// splitFlowItems splits a flow sequence body on commas, respecting quotes.
func splitFlowItems(s string) []string {
	var parts []string
	depth := 0
	inQuote := byte(0)
	start := 0
	for i := range len(s) {
		c := s[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			inQuote = c
		case '[', '{':
			depth++
		case ']', '}':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}
