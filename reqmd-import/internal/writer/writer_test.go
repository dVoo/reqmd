package writer

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func TestRenderDocument(t *testing.T) {
	reqs := []Req{
		{
			ID:     "IMP-foo-StructA-abc1234",
			Title:  "StructA",
			Status: "approved",
			Body:   "StructA description.",
			Attrs: map[string]any{
				"x-reqmd.imported":    true,
				"x-reqmd.source-file": "internal/foo/a.go",
				"x-reqmd.source-line": 12,
			},
		},
		{
			ID:     "IMP-foo-StructA-Method1-xyz5678",
			Title:  "Method1",
			Status: "approved",
			Depth:  1,
			Body:   "Method1 description.",
		},
	}
	out, err := RenderDocument("Package: foo", "This file is auto-generated. Do not edit by hand.", reqs)
	if err != nil {
		t.Fatalf("RenderDocument: %v", err)
	}

	checks := []string{
		"# Package: foo",
		GeneratedMarker,
		"This file is auto-generated. Do not edit by hand.",
		"## IMP-foo-StructA-abc1234: StructA",
		"### IMP-foo-StructA-Method1-xyz5678: Method1",
		"x-reqmd.imported: true",
		"x-reqmd.source-file: internal/foo/a.go",
		"x-reqmd.source-line: 12",
		"status: approved",
		"trace: []",
		"```attr",
		"StructA description.",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in rendered output:\n%s", want, out)
		}
	}

	// Parent heading must come before the child heading.
	if strings.Index(out, "## IMP-foo-StructA-abc1234") > strings.Index(out, "### IMP-foo-StructA-Method1") {
		t.Error("parent heading should come before child heading")
	}
}

func TestRenderRequirementHeadingDepth(t *testing.T) {
	top, err := RenderRequirement(Req{ID: "IMP-bar-Const-c0nst", Title: "Const", Status: "approved"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(top, "## IMP-bar-Const-c0nst") {
		t.Errorf("top-level requirement should use ##, got: %q", firstLine(top))
	}
	child, err := RenderRequirement(Req{ID: "IMP-bar-Sub-sub0", Title: "Sub", Status: "approved", Depth: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(child, "### IMP-bar-Sub-sub0") {
		t.Errorf("child requirement should use ###, got: %q", firstLine(child))
	}
}

func TestRenderRequirementAttrsDeterministic(t *testing.T) {
	// Attrs render in sorted key order regardless of map iteration order.
	r := Req{ID: "IMP-x-Y-0000000", Title: "Y", Status: "approved", Attrs: map[string]any{
		"zeta":  1,
		"alpha": "a",
		"mid":   []any{"p", "q"},
	}}
	first, err := RenderRequirement(r)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RenderRequirement(r)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Error("RenderRequirement must be deterministic across calls")
	}
	alpha, mid, zeta := strings.Index(first, "alpha:"), strings.Index(first, "mid:"), strings.Index(first, "zeta:")
	if alpha >= mid || mid >= zeta {
		t.Errorf("attrs not sorted; got:\n%s", first)
	}
	if !strings.Contains(first, "mid:\n  - p\n  - q\n") {
		t.Errorf("list attr not rendered as a YAML block list:\n%s", first)
	}
	if !strings.Contains(first, "zeta: 1\n") {
		t.Errorf("numeric attr not rendered verbatim:\n%s", first)
	}
}

func TestRenderRequirementUnsupportedAttrType(t *testing.T) {
	_, err := RenderRequirement(Req{ID: "IMP-x-Y-0000000", Attrs: map[string]any{
		"nested": map[string]any{"a": "b"},
	}})
	if err == nil {
		t.Fatal("expected an error for a nested-map attribute")
	}
	if !strings.Contains(err.Error(), "unsupported value type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsGenerated(t *testing.T) {
	if !IsGenerated([]byte("# hi\n" + GeneratedMarker + "\nbody")) {
		t.Error("expected IsGenerated=true for content with marker")
	}
	// Trees generated before the renderer moved to reqmd-import carry the
	// pre-rename marker; a re-run must still recognize them as generated
	// instead of treating them as hand-authored and refusing to update them.
	if !IsGenerated([]byte("# hi\n" + legacyMarker + "\nbody")) {
		t.Error("expected IsGenerated=true for content with the legacy marker")
	}
	if IsGenerated([]byte("# hi\nno marker here")) {
		t.Error("expected IsGenerated=false for content without marker")
	}
	if IsGenerated(nil) {
		t.Error("expected IsGenerated=false for nil content")
	}
}

// TestWriteIdempotent verifies a second Write produces byte-identical files.
func TestWriteIdempotent(t *testing.T) {
	dir := t.TempDir()
	files := []File{
		{Path: "schema.yaml", Data: []byte(GeneratedMarker + "\nschema: 1\n")},
		{Path: "p/package.md", Data: []byte("# Package: p\n" + GeneratedMarker + "\n")},
	}
	if err := Write(dir, files); err != nil {
		t.Fatalf("first Write: %v", err)
	}
	first := readAll(t, dir)
	if err := Write(dir, files); err != nil {
		t.Fatalf("second Write: %v", err)
	}
	if second := readAll(t, dir); !bytes.Equal(first, second) {
		t.Errorf("Write not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// TestWriteDoesNotClobberHandAuthored verifies a hand-authored file (no
// generated marker) is left untouched.
func TestWriteDoesNotClobberHandAuthored(t *testing.T) {
	dir := t.TempDir()
	handAuthored := "# hand-authored\nstatus: draft\n"
	if err := os.WriteFile(filepath.Join(dir, "schema.yaml"), []byte(handAuthored), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(dir, []File{{Path: "schema.yaml", Data: []byte(GeneratedMarker + "\ngenerated\n")}}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "schema.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != handAuthored {
		t.Errorf("hand-authored file was clobbered:\nwant:\n%s\ngot:\n%s", handAuthored, got)
	}
}

func readAll(t *testing.T, dir string) []byte {
	t.Helper()
	var paths []string
	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	var out []byte
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, data...)
	}
	return out
}

func TestQuoteIfNeeded(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", `""`},
		{"foo", "foo"},
		{"/tmp/source.go", "/tmp/source.go"},
		{"IMP-foo-Bar-abc1234", "IMP-foo-Bar-abc1234"},
		{"weird:name", `"weird:name"`},
		{"weird # comment", `"weird # comment"`},
		{`say "hi"`, `"say \"hi\""`},
		{"it's fine", `"it's fine"`},
		{"line1\nline2", `"line1\nline2"`},
		{"col1\tcol2", `"col1\tcol2"`},
		{" foo", `" foo"`},
		{"foo ", `"foo "`},
		{"\tfoo", `"\tfoo"`},
		{"-foo", `"-foo"`},
		{"*foo", `"*foo"`},
		{`C:\Users\foo`, `"C:\\Users\\foo"`},
		{"[a, b]", `"[a, b]"`},
		{"{a: b}", `"{a: b}"`},
		{"a,b", `"a,b"`},
		{"50%", `"50%"`},
		// Values reqmd's attr scanner would retype if left plain (null, bool,
		// integer) must be quoted so they stay strings.
		{"null", `"null"`},
		{"Null", `"Null"`},
		{"NULL", `"NULL"`},
		{"~", `"~"`},
		{"true", `"true"`},
		{"True", `"True"`},
		{"false", `"false"`},
		{"123", `"123"`},
		{"007", `"007"`},
		{"-5", `"-5"`},
		// The scanner leaves floats and other non-integer numbers as strings,
		// so these stay plain — quoting them would be churn, not safety.
		{"yes", "yes"},
		{"no", "no"},
		{"1.5", "1.5"},
		{"2e3", "2e3"},
	}
	for _, c := range cases {
		if got := quoteIfNeeded(c.in); got != c.want {
			t.Errorf("quoteIfNeeded(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestQuoteIfNeededRoundTrip(t *testing.T) {
	for _, in := range []string{
		"line1\nline2",
		"col1\tcol2",
		`say "hi"`,
		`C:\Users\foo`,
		"normal",
		"",
		`weird: "value" with \backslash`,
		"[not, a, list]",
		"50%",
	} {
		quoted := quoteIfNeeded(in)
		var out any
		if err := yaml.Unmarshal([]byte(quoted), &out); err != nil {
			t.Fatalf("yaml.Unmarshal(%q): %v", quoted, err)
		}
		if out != in {
			t.Errorf("round-trip lost data: in=%q out=%v", in, out)
		}
	}
}
