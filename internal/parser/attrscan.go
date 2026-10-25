package parser

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// errUnsupported is returned by the scanner for input outside the subset it
// can parse *exactly*: nested maps, multi-line scalars, duplicate keys, and
// scalars yaml.v3 would type differently (floats, timestamps, hex, big
// integers). decodeAttrs then re-parses the whole block with yaml.v3, so the
// result never depends on which path produced it.
var errUnsupported = errors.New("attr scanner: unsupported syntax")

// decodeAttrs parses the contents of an ```attr block into a map. It runs the
// fast line scanner first and falls back to yaml.v3 for anything the scanner
// cannot reproduce byte-for-byte (see errUnsupported).
func decodeAttrs(src string) (map[string]any, error) {
	attrs, err := decodeAttrYAML(src)
	if err == nil {
		return attrs, nil
	}
	if !errors.Is(err, errUnsupported) {
		return nil, err
	}
	// yaml.v3 merges into an existing map, so starting from an empty one
	// keeps the result identical to a fresh nil map and allocates once.
	m := make(map[string]any, 8)
	if yamlErr := yaml.Unmarshal([]byte(src), &m); yamlErr != nil {
		return nil, fmt.Errorf("parsing attr block as YAML: %w", yamlErr)
	}
	return m, nil
}

// decodeAttrYAML is the fast path: it parses a small flat YAML map (the
// contents of an ```attr block) into map[string]any without building the
// intermediate YAML document AST that yaml.v3 allocates.
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
// Anything it cannot reproduce exactly — malformed input, nested maps, flow
// mappings, duplicate keys, or a scalar yaml.v3 would retype (float,
// timestamp, hex, big integer) — returns errUnsupported so decodeAttrs
// re-parses the block with yaml.v3. Callers should use decodeAttrs, not this
// function directly.
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
		if _, dup := m[key]; dup {
			return nil, fmt.Errorf("duplicate key %q", key)
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
					scalar, err := parseScalar(valuePart)
					if err != nil {
						return nil, err
					}
					m[key] = scalar
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
					scalar, err := parseScalar(itemText)
					if err != nil {
						return nil, err
					}
					items = append(items, scalar)
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
