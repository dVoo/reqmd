package exchange

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// This file holds the structural assertions the rendering tests use instead of
// substring matching. A rendered document is asserted through its outline —
// heading level, heading text, and the parsed reqmd block — so a change in
// formatting (spacing, quoting style, key order) does not break a test that is
// really about structure, while a change in nesting or identity does.

// outlineHeading is one rendered heading.
type outlineHeading struct {
	Text  string
	Attrs map[string]any
	Level int
}

// outline parses a rendered document into its headings. It only understands the
// shape the renderer produces (ATX headings, opaque ```yaml reqmd blocks), which is
// exactly what the assertions need.
func outline(t *testing.T, md string) []outlineHeading {
	t.Helper()
	lines := strings.Split(md, "\n")
	var out []outlineHeading
	for i := range lines {
		level, text, ok := atxLevel(lines[i])
		if !ok || level == 1 {
			// Level 1 is the document title, not a content heading.
			continue
		}
		h := outlineHeading{Level: level, Text: text}
		// The reqmd block, when present, is the next non-blank line.
		for j := i + 1; j < len(lines); j++ {
			next := strings.TrimSpace(lines[j])
			if next == "" {
				continue
			}
			if next != "```yaml reqmd" {
				break
			}
			var block []string
			for k := j + 1; k < len(lines) && strings.TrimSpace(lines[k]) != "```"; k++ {
				block = append(block, lines[k])
			}
			attrs := map[string]any{}
			if err := yaml.Unmarshal([]byte(strings.Join(block, "\n")), &attrs); err != nil {
				t.Fatalf("rendered reqmd block for %q is not valid YAML: %v", text, err)
			}
			h.Attrs = attrs
			break
		}
		out = append(out, h)
	}
	return out
}

// atxLevel parses "### Text" into its level and text.
func atxLevel(line string) (int, string, bool) {
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n == 0 || n > 6 || n == len(line) || line[n] != ' ' {
		return 0, "", false
	}
	return n, strings.TrimSpace(line[n+1:]), true
}

// heading returns the outline entry whose text contains want, failing when it
// is absent or ambiguous.
func heading(t *testing.T, md, want string) outlineHeading {
	t.Helper()
	var hits []outlineHeading
	for _, h := range outline(t, md) {
		if strings.Contains(h.Text, want) {
			hits = append(hits, h)
		}
	}
	switch len(hits) {
	case 0:
		t.Fatalf("no heading matching %q in rendered document:\n%s", want, md)
	case 1:
		return hits[0]
	}
	t.Fatalf("heading %q is ambiguous (%d matches)", want, len(hits))
	return outlineHeading{}
}

// attrString reads a string attribute from a parsed heading.
func attrString(t *testing.T, h outlineHeading, key string) string {
	t.Helper()
	v, ok := h.Attrs[key]
	if !ok {
		t.Fatalf("heading %q has no attribute %q; got %v", h.Text, key, h.Attrs)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("attribute %q of %q = %T, want string", key, h.Text, v)
	}
	return s
}

// attrList reads a list attribute from a parsed heading.
func attrList(t *testing.T, h outlineHeading, key string) []string {
	t.Helper()
	v, ok := h.Attrs[key]
	if !ok {
		t.Fatalf("heading %q has no attribute %q; got %v", h.Text, key, h.Attrs)
	}
	raw, ok := v.([]any)
	if !ok {
		t.Fatalf("attribute %q of %q = %T, want a list", key, h.Text, v)
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			t.Fatalf("attribute %q of %q has a non-string item %T", key, h.Text, item)
		}
		out = append(out, s)
	}
	return out
}
