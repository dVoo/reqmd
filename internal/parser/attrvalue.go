package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// parseScalar converts an unquoted YAML plain scalar to the Go type yaml.v3
// would produce: true/false → bool, null/~ → nil, integer → int, anything
// else → string. Scalars yaml.v3 would retype as something this function
// cannot produce (float, timestamp, hex, big integer) report errUnsupported
// so decodeAttrs re-parses the block with yaml.v3 instead of silently
// turning them into strings.
func parseScalar(s string) (any, error) {
	// Strip inline comments: value # comment
	if idx := strings.Index(s, " #"); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}
	// bool
	switch s {
	case "true", "True", "TRUE":
		return true, nil
	case "false", "False", "FALSE":
		return false, nil
	case "null", "Null", "NULL", "~":
		return nil, nil
	}
	// int — only call strconv.Atoi when the string is actually an integer
	// literal. Atoi allocates an error object on non-numeric input, which
	// is the common case for attribute values (IDs, statuses, priorities).
	if isIntegerString(s) {
		if n, err := strconv.Atoi(s); err == nil {
			return n, nil
		}
	}
	// Every scalar yaml.v3 retypes beyond the cases above starts with a
	// digit, sign, or dot (floats, timestamps, 0x/0o prefixes, .inf/.nan).
	// A plain scalar starting the same way cannot be assumed to be a string.
	if s == "" || !isPlainStringStart(s[0]) {
		return nil, errUnsupported
	}
	return s, nil
}

// isPlainStringStart reports whether c can begin a plain YAML scalar that
// yaml.v3 resolves as a string.
func isPlainStringStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
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
			scalar, err := parseScalar(part)
			if err != nil {
				return nil, err
			}
			items = append(items, scalar)
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
