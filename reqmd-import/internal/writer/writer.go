// Package writer renders reqmd markdown from resolved requirements and
// writes generated files idempotently. It is the single renderer for every
// importer: producers emit the reqmd-import/v1 exchange format, reqmd-import
// turns it into .md + schema.yaml here.
package writer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// GeneratedMarker is the HTML-comment sentinel embedded in the header of
// every generated file. The idempotent write path uses it to distinguish
// generated content from hand-authored content in the same tree.
const GeneratedMarker = "<!-- reqmd: generated -->"

// legacyMarker is the pre-rename marker emitted by reqmd-code before the
// renderer moved to reqmd-import. It is still recognized so a re-run
// overwrites an existing generated tree instead of mistaking it for
// hand-authored content.
const legacyMarker = "<!-- reqmd-code: generated -->"

// Req is one resolved requirement to render: its final reqmd ID, the trace
// refs, the remaining attr-block entries, and its heading nesting depth.
type Req struct {
	Attrs  map[string]any
	ID     string
	Title  string
	Status string
	Body   string
	Trace  []string
	Depth  int
}

// File is one rendered output file, relative to the output root.
type File struct {
	Path string
	Data []byte
}

// IsGenerated reports whether content carries the generated marker.
func IsGenerated(content []byte) bool {
	return bytes.Contains(content, []byte(GeneratedMarker)) || bytes.Contains(content, []byte(legacyMarker))
}

// Write writes every file under outDir, creating directories as needed. A
// file is only rewritten when its content changed; existing content without
// the generated marker is left untouched so hand-authored files are never
// clobbered.
func Write(outDir string, files []File) error {
	for _, f := range files {
		if err := writeIfChanged(filepath.Join(outDir, f.Path), f.Data); err != nil {
			return fmt.Errorf("write %s: %w", f.Path, err)
		}
	}
	return nil
}

func writeIfChanged(path string, data []byte) error {
	existing, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, data, 0o644)
	}
	if err != nil {
		return err
	}
	if !IsGenerated(existing) {
		return nil
	}
	if bytes.Equal(existing, data) {
		return nil
	}
	return os.WriteFile(path, data, 0o644)
}

// RenderDocument renders a document: title, generated marker, optional notice
// line, then one section per requirement in the order given.
func RenderDocument(title, notice string, reqs []Req) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", title)
	b.WriteString(GeneratedMarker)
	b.WriteString("\n\n")
	if notice != "" {
		b.WriteString(notice)
		b.WriteString("\n\n")
	}
	for _, r := range reqs {
		s, err := RenderRequirement(r)
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	return b.String(), nil
}

// RenderRequirement renders one requirement: heading, attr block, body. Depth
// is the nesting level: 0 renders "##", 1 renders "###", and so on (reqmd
// accepts any heading level from h2 down).
func RenderRequirement(r Req) (string, error) {
	var b strings.Builder
	heading := strings.Repeat("#", 2+r.Depth)
	if r.Title != "" {
		fmt.Fprintf(&b, "%s %s: %s\n", heading, r.ID, r.Title)
	} else {
		fmt.Fprintf(&b, "%s %s\n", heading, r.ID)
	}
	b.WriteString("```attr\n")
	if err := writeAttr(&b, r); err != nil {
		return "", fmt.Errorf("requirement %s: %w", r.ID, err)
	}
	b.WriteString("```\n")
	if r.Body != "" {
		b.WriteString("\n")
		b.WriteString(r.Body)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String(), nil
}

// writeAttr writes the attr-block body: status first, then the remaining
// attributes in sorted key order (deterministic output), then the trace list.
func writeAttr(b *strings.Builder, r Req) error {
	if r.Status != "" {
		fmt.Fprintf(b, "status: %s\n", quoteIfNeeded(r.Status))
	}
	keys := make([]string, 0, len(r.Attrs))
	for k := range r.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := writeAttrValue(b, k, r.Attrs[k]); err != nil {
			return err
		}
	}
	writeRefList(b, "trace", r.Trace)
	return nil
}

func writeAttrValue(b *strings.Builder, key string, v any) error {
	switch val := v.(type) {
	case nil:
		return nil
	case string:
		fmt.Fprintf(b, "%s: %s\n", key, quoteIfNeeded(val))
		return nil
	case bool:
		fmt.Fprintf(b, "%s: %t\n", key, val)
		return nil
	case json.Number:
		fmt.Fprintf(b, "%s: %s\n", key, val.String())
		return nil
	case int, int64, float64:
		fmt.Fprintf(b, "%s: %v\n", key, val)
		return nil
	case []any:
		return writeList(b, key, val)
	case []string:
		items := make([]any, len(val))
		for i, s := range val {
			items[i] = s
		}
		return writeList(b, key, items)
	default:
		return fmt.Errorf("attribute %q: unsupported value type %T", key, v)
	}
}

func writeList(b *strings.Builder, key string, items []any) error {
	if len(items) == 0 {
		fmt.Fprintf(b, "%s: []\n", key)
		return nil
	}
	fmt.Fprintf(b, "%s:\n", key)
	for _, it := range items {
		s, err := scalar(it)
		if err != nil {
			return fmt.Errorf("attribute %q: %w", key, err)
		}
		fmt.Fprintf(b, "  - %s\n", s)
	}
	return nil
}

func scalar(v any) (string, error) {
	switch val := v.(type) {
	case string:
		return quoteIfNeeded(val), nil
	case bool:
		return strconv.FormatBool(val), nil
	case json.Number:
		return val.String(), nil
	case int, int64, float64:
		return fmt.Sprintf("%v", val), nil
	default:
		return "", fmt.Errorf("unsupported list item type %T", v)
	}
}

// writeRefList writes a YAML block list, or `key: []` when empty.
func writeRefList(b *strings.Builder, key string, refs []string) {
	if len(refs) == 0 {
		fmt.Fprintf(b, "%s: []\n", key)
		return
	}
	fmt.Fprintf(b, "%s:\n", key)
	for _, r := range refs {
		fmt.Fprintf(b, "  - %s\n", quoteIfNeeded(r))
	}
}

// quoteIfNeeded formats a value as a YAML scalar that reqmd reads back
// unchanged. A plain scalar is unsafe when it contains YAML punctuation, when
// it has leading or trailing whitespace, or when reqmd's attr scanner
// (internal/parser/attrscan.go, parseScalar) would retype it as null, a bool,
// or an integer — so all of those are double-quoted with the escapes that
// scanner understands. Over-quoting is harmless; under-quoting silently
// changes the value's type.
func quoteIfNeeded(s string) string {
	if plainScalarSafe(s) {
		return s
	}
	// Escape backslash first, then double-quote, then the control characters.
	// Backslash must be first so the backslashes introduced by the later passes
	// are not themselves re-escaped.
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	escaped = strings.ReplaceAll(escaped, "\t", `\t`)
	escaped = strings.ReplaceAll(escaped, "\r", `\r`)
	return `"` + escaped + `"`
}

// plainScalarSafe reports whether s can be written as a plain (unquoted) YAML
// scalar without changing its meaning for reqmd's attr scanner. The retype
// list must stay a superset of that scanner's parseScalar: if the scanner
// starts resolving something new to a non-string, a value that this function
// calls safe becomes corrupt.
func plainScalarSafe(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch r {
		case ':', '#', '"', '\'', '\n', '\t', '\r',
			'[', ']', '{', '}', ',', '&', '!', '%', '@', '`', '>', '|', '?':
			return false
		}
	}
	if s[0] == ' ' || s[0] == '\t' || s[0] == '-' || s[0] == '*' {
		return false
	}
	if s[len(s)-1] == ' ' || s[len(s)-1] == '\t' {
		return false
	}
	// Plain scalars the scanner resolves to a non-string type. Anything else
	// (including floats, which the scanner leaves as strings) stays plain.
	switch s {
	case "null", "Null", "NULL", "~",
		"true", "True", "TRUE", "false", "False", "FALSE":
		return false
	}
	if _, err := strconv.Atoi(s); err == nil {
		return false
	}
	return true
}
