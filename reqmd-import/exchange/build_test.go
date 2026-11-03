package exchange

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/reqmd-import/internal/writer"

	"gopkg.in/yaml.v3"
)

// bundle wraps documents in a valid envelope.
func bundle(docs ...Document) *Bundle {
	return &Bundle{Schema: SchemaURL, Documents: docs}
}

// render builds the bundle and returns rendered files keyed by path.
func render(t *testing.T, b *Bundle) map[string]string {
	t.Helper()
	files, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	out := make(map[string]string, len(files))
	for _, f := range files {
		out[f.Path] = string(f.Data)
	}
	return out
}

// renderErr builds and returns the validation error text.
func renderErr(t *testing.T, b *Bundle) string {
	t.Helper()
	_, err := b.Build()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	return err.Error()
}

// wantProblems asserts that every fragment appears in the error report, i.e.
// that violations are batched rather than reported one at a time.
func wantProblems(t *testing.T, got string, fragments ...string) {
	t.Helper()
	for _, f := range fragments {
		if !strings.Contains(got, f) {
			t.Errorf("error report is missing %q:\n%s", f, got)
		}
	}
}

func TestBuild_RendersDocument(t *testing.T) {
	got := render(t, bundle(Document{
		Key:   "docs",
		Title: "Docs",
		Body:  "Preamble prose.",
		XReqmd: &XReqmd{
			IDPrefix: "X-",
			Level:    "system-requirements",
		},
		Properties: map[string]any{"priority": map[string]any{"type": "string"}},
		Items: []Item{
			{
				Kind:   KindRequirement,
				Key:    "1",
				Title:  "Brakes",
				Body:   "The vehicle shall stop.",
				Status: "approved",
				Fields: map[string]any{"priority": "High"},
				Refs:   []string{"SYS-001"},
			},
		},
	}))

	md := got["docs/requirements.md"]
	if !strings.Contains(md, writer.GeneratedMarker) {
		t.Error("rendered document is missing the generated marker")
	}
	h := heading(t, md, "Brakes")
	if h.Level != 2 {
		t.Errorf("top-level requirement rendered at heading level %d, want 2", h.Level)
	}
	if h.Text != "X-1: Brakes" {
		t.Errorf("heading = %q, want the ID prefix and title \"X-1: Brakes\"", h.Text)
	}
	if got := attrString(t, h, "status"); got != "approved" {
		t.Errorf("status = %q, want approved", got)
	}
	if got := attrString(t, h, "priority"); got != "High" {
		t.Errorf("priority = %q, want High", got)
	}
	if got := attrList(t, h, "trace"); len(got) != 1 || got[0] != "SYS-001" {
		t.Errorf("trace = %v, want [SYS-001]", got)
	}

	// The schema is asserted as parsed YAML: field order and quoting are the
	// renderer's business, the x-reqmd mapping is the contract.
	schema := map[string]any{}
	if err := yaml.Unmarshal([]byte(got["docs/schema.yaml"]), &schema); err != nil {
		t.Fatalf("schema is not valid YAML: %v", err)
	}
	xr, ok := schema["x-reqmd"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no x-reqmd block: %v", schema)
	}
	if xr["id-prefix"] != "X-" || xr["level"] != "system-requirements" {
		t.Errorf("x-reqmd = %v, want id-prefix X- and level system-requirements", xr)
	}
}

// TestBuild_NestedSectionsAndDepth covers the tree: a section renders a heading
// with children one level deeper, and no attr block.
func TestBuild_NestedSectionsAndDepth(t *testing.T) {
	got := render(t, bundle(Document{
		Key:   "d",
		Title: "D",
		Items: []Item{
			{
				Kind:  KindSection,
				Title: "Chapter",
				Body:  "Chapter intro.",
				Children: []Item{{
					Kind:   KindRequirement,
					Key:    "R1",
					Title:  "Req",
					Status: "approved",
					Children: []Item{{
						Kind:   KindRequirement,
						Key:    "R2",
						Title:  "Child",
						Status: "approved",
					}},
				}},
			},
		},
	}))

	md := got["d/requirements.md"]
	chapter := heading(t, md, "Chapter")
	if chapter.Level != 2 {
		t.Errorf("section rendered at level %d, want 2", chapter.Level)
	}
	if chapter.Attrs != nil {
		t.Errorf("a section must render no attr block, got %v", chapter.Attrs)
	}
	if !strings.Contains(md, "Chapter intro.") {
		t.Error("section prose missing from the rendered document")
	}
	// Nesting is the contract: each level adds exactly one '#'.
	if got := heading(t, md, "R1: Req").Level; got != 3 {
		t.Errorf("R1 rendered at level %d, want 3 (one level under its section)", got)
	}
	if got := heading(t, md, "R2: Child").Level; got != 4 {
		t.Errorf("R2 rendered at level %d, want 4 (one level under R1)", got)
	}
}

// TestBuild_LinksResolve covers document-local and cross-document links: a link
// becomes a trace ref, qualified by the target document's id when it has one.
func TestBuild_LinksResolve(t *testing.T) {
	got := render(t, bundle(
		Document{
			Key:    "up",
			Title:  "Up",
			XReqmd: &XReqmd{IDPrefix: "UP-", DocumentID: "upstream"},
			Items:  []Item{{Kind: KindRequirement, Key: "1", Title: "Up", Status: "approved"}},
		},
		Document{
			Key:    "down",
			Title:  "Down",
			XReqmd: &XReqmd{IDPrefix: "DN-"},
			Items: []Item{{
				Kind:   KindRequirement,
				Key:    "2",
				Title:  "Down",
				Status: "approved",
				Links:  []string{"up/1"},
			}},
		},
	))

	trace := attrList(t, heading(t, got["down/requirements.md"], "Down"), "trace")
	if len(trace) != 1 || trace[0] != "upstream/UP-1" {
		t.Errorf("trace = %v, want [upstream/UP-1] (qualified by the target document-id)", trace)
	}
}

// TestBuild_CrossDocumentLinkFallsBackToDir covers a target document without a
// document-id: reqmd resolves "<dir>/<id>", so the ref must use the dir.
func TestBuild_CrossDocumentLinkFallsBackToDir(t *testing.T) {
	got := render(t, bundle(
		Document{
			Key:    "up-doc",
			Title:  "Up",
			XReqmd: &XReqmd{IDPrefix: "UP-", Dir: "up-dir"},
			Items:  []Item{{Kind: KindRequirement, Key: "1", Title: "Up", Status: "approved"}},
		},
		Document{
			Key:    "down",
			Title:  "Down",
			XReqmd: &XReqmd{IDPrefix: "DN-"},
			Items: []Item{{
				Kind:   KindRequirement,
				Key:    "2",
				Title:  "Down",
				Status: "approved",
				Links:  []string{"up-doc/1"},
			}},
		},
	))

	trace := attrList(t, heading(t, got["down/requirements.md"], "Down"), "trace")
	if len(trace) != 1 || trace[0] != "up-dir/UP-1" {
		t.Errorf("trace = %v, want [up-dir/UP-1] (dir fallback when the target has no document-id)", trace)
	}
}

// TestBuild_StatusMapping covers the renderer-side status map: a mapped value is
// used and extends the document enum; an unmapped one falls back to the default
// and keeps the original as provenance.
func TestBuild_StatusMapping(t *testing.T) {
	b := bundle(Document{
		Key:   "d",
		Title: "D",
		Items: []Item{
			{Kind: KindRequirement, Key: "1", Title: "Mapped", Status: "In Review"},
			{Kind: KindRequirement, Key: "2", Title: "Unmapped", Status: "Obsolete"},
		},
	})
	files, err := b.BuildWith(Options{StatusMap: map[string]string{"In Review": "review"}})
	if err != nil {
		t.Fatalf("BuildWith: %v", err)
	}
	byPath := map[string]string{}
	for _, f := range files {
		byPath[f.Path] = string(f.Data)
	}
	md := byPath["d/requirements.md"]
	if got := attrString(t, heading(t, md, "Mapped"), "status"); got != "review" {
		t.Errorf("mapped status = %q, want review", got)
	}
	unmapped := heading(t, md, "Unmapped")
	if got := attrString(t, unmapped, "status"); got != "approved" {
		t.Errorf("unmapped status = %q, want the approved fallback", got)
	}
	if got := attrString(t, unmapped, "x-reqmd.source-status"); got != "Obsolete" {
		t.Errorf("provenance = %q, want the original status Obsolete", got)
	}

	schema := map[string]any{}
	if err := yaml.Unmarshal([]byte(byPath["d/schema.yaml"]), &schema); err != nil {
		t.Fatalf("schema is not valid YAML: %v", err)
	}
	xr, _ := schema["x-reqmd"].(map[string]any)
	raw, ok := xr["additional-status-values"].([]any)
	if !ok || len(raw) != 1 || raw[0] != "review" {
		t.Errorf("additional-status-values = %v, want [review]", xr["additional-status-values"])
	}
}

// TestBuild_SchemaOnlyDocument covers a document without items.
func TestBuild_SchemaOnlyDocument(t *testing.T) {
	got := render(t, bundle(Document{Key: "root", Title: "Root"}))
	if _, ok := got["root/requirements.md"]; ok {
		t.Error("a document without items must not produce a .md file")
	}
	if _, ok := got["root/schema.yaml"]; !ok {
		t.Error("a document without items must still produce its schema.yaml")
	}
}

func TestBuild_DecodeRejectsUnknownFields(t *testing.T) {
	raw := `{"$schema":"` + SchemaURL + `","documents":[{"key":"d","title":"D","items":[{"kind":"section","parent":"x"}]}]}`
	_, err := Decode(strings.NewReader(raw))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("want unknown-field rejection, got %v", err)
	}
}

func TestBuild_ItemTreeKeepsInputOrder(t *testing.T) {
	got := render(t, bundle(Document{
		Key:   "d",
		Title: "D",
		Items: []Item{
			{Kind: KindSection, Title: "B", Children: []Item{{Kind: KindRequirement, Key: "b1", Title: "b1", Status: "approved"}}},
			{Kind: KindSection, Title: "A", Children: []Item{{Kind: KindRequirement, Key: "a1", Title: "a1", Status: "approved"}}},
		},
	}))
	md := got["d/requirements.md"]
	if strings.Index(md, "## B") > strings.Index(md, "## A") {
		t.Errorf("the renderer must not reorder items:\n%s", md)
	}
}

// TestBuild_BatchesViolations is the core producer-facing guarantee: every
// violation is reported in one run instead of one at a time.
func TestBuild_BatchesViolations(t *testing.T) {
	got := renderErr(t, bundle(Document{
		Key:   "d",
		Title: "D",
		Items: []Item{
			{Kind: "bogus", Title: "wrong kind"},
			{Kind: KindRequirement, Title: "no key"},
			{Kind: KindRequirement, Key: "1", Title: "ok"},
			{Kind: KindSection, Title: "with fields", Fields: map[string]any{"a": 1}},
			{Kind: KindRequirement, Key: "1", Title: "duplicate key"},
			{Kind: KindRequirement, Key: "2", Title: "bad link", Links: []string{"missing"}},
		},
	}))
	wantProblems(t, got,
		`kind "bogus"`,
		"requirement without a key",
		"section and must not carry fields",
		"is used by both",
		"links to \"missing\"",
	)
}

func TestBuild_ValidationRules(t *testing.T) {
	tests := []struct {
		name     string
		bundle   *Bundle
		contains string
	}{
		{
			name:     "wrong schema url",
			bundle:   &Bundle{Schema: "https://example.invalid/schema.json", Documents: []Document{{Key: "d", Title: "D"}}},
			contains: "want \"https://reqmd.dev/exchange/v1/schema.json\"",
		},
		{
			name:     "no documents",
			bundle:   &Bundle{Schema: SchemaURL},
			contains: "bundle has no documents",
		},
		{
			name:     "duplicate document key",
			bundle:   bundle(Document{Key: "d", Title: "D"}, Document{Key: "d", Title: "D2"}),
			contains: `document key "d" is used by both`,
		},
		{
			name:     "document key with slash",
			bundle:   bundle(Document{Key: "a/b", Title: "D"}),
			contains: "key must not contain",
		},
		{
			name:     "missing title",
			bundle:   bundle(Document{Key: "d"}),
			contains: "title is required",
		},
		{
			name:     "section with status",
			bundle:   bundle(Document{Key: "d", Title: "D", Items: []Item{{Kind: KindSection, Title: "S", Status: "approved"}}}),
			contains: "must not carry a status",
		},
		{
			name: "undeclared field",
			bundle: bundle(Document{
				Key: "d", Title: "D",
				Properties: map[string]any{"known": map[string]any{"type": "string"}},
				Items:      []Item{{Kind: KindRequirement, Key: "1", Title: "R", Fields: map[string]any{"other": "x"}}},
			}),
			contains: `attribute "other" is not declared`,
		},
		{
			name: "unsatisfied required",
			bundle: bundle(Document{
				Key: "d", Title: "D",
				Required: []string{"priority"},
				Items:    []Item{{Kind: KindRequirement, Key: "1", Title: "R"}},
			}),
			contains: `missing required attribute "priority"`,
		},
		{
			name: "nested field value",
			bundle: bundle(Document{
				Key: "d", Title: "D",
				Items: []Item{{Kind: KindRequirement, Key: "1", Title: "R", Fields: map[string]any{"a": map[string]any{"b": 1}}}},
			}),
			contains: "nested",
		},
		{
			name: "bad attribute name",
			bundle: bundle(Document{
				Key: "d", Title: "D",
				Items: []Item{{Kind: KindRequirement, Key: "1", Title: "R", Fields: map[string]any{"a:b": 1}}},
			}),
			contains: "name may only contain",
		},
		{
			name: "link to a section",
			bundle: bundle(Document{
				Key: "d", Title: "D",
				Items: []Item{
					{Kind: KindSection, Key: "sec", Title: "S"},
					{Kind: KindRequirement, Key: "1", Title: "R", Links: []string{"sec"}},
				},
			}),
			contains: "links may only target requirements",
		},
		{
			name: "link to unknown document",
			bundle: bundle(Document{
				Key: "d", Title: "D",
				Items: []Item{{Kind: KindRequirement, Key: "1", Title: "R", Links: []string{"nope/1"}}},
			}),
			contains: `unknown document "nope"`,
		},
		{
			name: "title with a line break",
			bundle: bundle(Document{
				Key: "d", Title: "D",
				Items: []Item{{Kind: KindRequirement, Key: "1", Title: "a\nb"}},
			}),
			contains: "must not contain a line break",
		},
		{
			name: "depth budget",
			bundle: bundle(Document{
				Key: "d", Title: "D",
				Items: nestItems(6),
			}),
			contains: "would render heading level",
		},
		{
			name: "escaping directory",
			bundle: bundle(Document{
				Key: "d", Title: "D", XReqmd: &XReqmd{Dir: "../outside"},
				Items: []Item{{Kind: KindRequirement, Key: "1", Title: "R"}},
			}),
			contains: "escapes the output root",
		},
		{
			name: "duplicate output path",
			bundle: bundle(
				Document{Key: "a", Title: "A", XReqmd: &XReqmd{Dir: "same"}, Items: []Item{{Kind: KindRequirement, Key: "1", Title: "R"}}},
				Document{Key: "b", Title: "B", XReqmd: &XReqmd{Dir: "same"}, Items: []Item{{Kind: KindRequirement, Key: "2", Title: "R"}}},
			),
			contains: "is claimed by both",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := renderErr(t, tc.bundle)
			if !strings.Contains(got, tc.contains) {
				t.Errorf("error report does not contain %q:\n%s", tc.contains, got)
			}
		})
	}
}

// nestItems builds a chain of sections deep enough to exceed the render budget.
func nestItems(depth int) []Item {
	if depth == 0 {
		return nil
	}
	return []Item{{Kind: KindSection, Title: "s", Children: nestItems(depth - 1)}}
}

func TestDecodeRejectsDeepNesting(t *testing.T) {
	var item = map[string]any{"kind": "section"}
	for i := 0; i < maxDepth+2; i++ {
		item = map[string]any{"kind": "section", "children": []any{item}}
	}
	raw, err := json.Marshal(map[string]any{
		"$schema":   SchemaURL,
		"documents": []any{map[string]any{"key": "d", "title": "D", "items": []any{item}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Decode(strings.NewReader(string(raw)))
	if err == nil || !strings.Contains(err.Error(), "nesting exceeds") {
		t.Fatalf("want nesting guard, got %v", err)
	}
}

func TestDecodeRejectsTrailingData(t *testing.T) {
	raw := `{"$schema":"` + SchemaURL + `","documents":[{"key":"d","title":"D"}]} {}`
	_, err := Decode(strings.NewReader(raw))
	if err == nil || !strings.Contains(err.Error(), "trailing data") {
		t.Fatalf("want trailing-data rejection, got %v", err)
	}
}

// TestBuild_IsIdempotent verifies byte-determinism: the same bundle renders the
// same bytes, so re-running an import is a no-op.
func TestBuild_IsIdempotent(t *testing.T) {
	b := bundle(Document{
		Key:   "d",
		Title: "D",
		Properties: map[string]any{
			"z": map[string]any{"type": "string"},
			"a": map[string]any{"type": "string"},
			"m": map[string]any{"type": "string"},
		},
		Items: []Item{{
			Kind:   KindRequirement,
			Key:    "1",
			Title:  "R",
			Status: "approved",
			Fields: map[string]any{"z": "1", "a": "2", "m": "3"},
		}},
	})
	first, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	second, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(first) != len(second) {
		t.Fatalf("file count differs between runs")
	}
	for i := range first {
		if first[i].Path != second[i].Path || string(first[i].Data) != string(second[i].Data) {
			t.Errorf("output is not byte-deterministic for %s", first[i].Path)
		}
	}
}

// TestValidationError_ListsEveryProblem pins the report shape producers rely on.
func TestValidationError_ListsEveryProblem(t *testing.T) {
	_, err := bundle(Document{Key: "d"}).Build()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *ValidationError, got %T", err)
	}
	if len(ve.Problems) == 0 || !strings.HasPrefix(ve.Error(), "invalid exchange bundle:") {
		t.Errorf("unexpected report: %v", err)
	}
}
