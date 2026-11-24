package exporter

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/internal/model"
)

// testDoc builds a document with a container, requirement children, and an
// info item so the render data covers every node kind.
func testDoc() model.Document {
	child := &model.Node{
		Kind:     model.KindRequirement,
		ID:       "SYS-002",
		Title:    "Child requirement",
		Level:    3,
		Body:     "Child body.",
		ParentID: "SYS-001",
		Attrs:    map[string]any{"status": "draft", "priority": "High"},
	}
	root := &model.Node{
		Kind:     model.KindRequirement,
		ID:       "SYS-001",
		Title:    "Root requirement",
		Level:    2,
		Body:     "Root body.",
		Attrs:    map[string]any{"status": "approved", "priority": "Low", "owner": "alice"},
		Children: []*model.Node{child},
	}
	container := &model.Node{
		Kind:     model.KindSection,
		Title:    "Section",
		Level:    2,
		Body:     "Container body.",
		Children: []*model.Node{root},
	}
	info := &model.Node{Kind: model.KindSection, Title: "Note", Level: 2, Body: "Info body."}
	return model.Document{
		Path:   "/test/doc",
		Title:  "Test Document",
		Schema: map[string]any{"title": "Schema Title"},
		Nodes:  []*model.Node{container, info},
		XReqmd: &model.XReqmd{DocumentID: "test", Level: "system"},
	}
}

// TestBuildRenderData_SlugAnchors covers the section anchor rules: a slug from
// the title, a deterministic suffix on collisions, and a fixed anchor for the
// untitled preamble.
func TestBuildRenderData_SlugAnchors(t *testing.T) {
	mk := func(title string) *model.Node {
		return &model.Node{Kind: model.KindSection, Title: title, Level: 2}
	}
	first := mk("General")
	second := mk("General")
	preamble := &model.Node{Kind: model.KindSection, Level: 0, Body: "Intro."}
	weird := mk("  Brakes / Hydraulics (v2)  ")
	req := &model.Node{Kind: model.KindRequirement, ID: "SYS-001", Title: "Real", Level: 2}

	rd := buildRenderData(model.Document{
		Path:  "/test/doc",
		Nodes: []*model.Node{first, second, preamble, weird, req},
	})

	want := map[*model.Node]string{
		first:    "sec-general",
		second:   "sec-general-2",
		preamble: "sec-preamble",
		weird:    "sec-brakes-hydraulics-v2",
		req:      "SYS-001",
	}
	for node, expected := range want {
		if got := rd.Anchors[node]; got != expected {
			t.Errorf("anchor for %q = %q, want %q", node.Title, got, expected)
		}
	}
}

// TestBuildRenderData_SlugNeverShadowsRequirementID guards the collision rule:
// a requirement whose ID happens to look like a section slug keeps its own
// anchor, and the section takes the next free suffix.
func TestBuildRenderData_SlugNeverShadowsRequirementID(t *testing.T) {
	req := &model.Node{Kind: model.KindRequirement, ID: "sec-general", Title: "Real", Level: 2}
	section := &model.Node{Kind: model.KindSection, Title: "General", Level: 2}

	rd := buildRenderData(model.Document{Path: "/test/doc", Nodes: []*model.Node{req, section}})

	if got := rd.Anchors[req]; got != "sec-general" {
		t.Errorf("requirement anchor = %q, want its own ID", got)
	}
	if got := rd.Anchors[section]; got != "sec-general-2" {
		t.Errorf("section anchor = %q, want sec-general-2", got)
	}
}

func TestBuildRenderData(t *testing.T) {
	doc := testDoc()
	rd := buildRenderData(doc)

	if got, want := len(rd.Nodes), len(doc.Nodes); got != want {
		t.Errorf("Nodes = %d, want %d", got, want)
	}
	// Requirements anchor on their ID; items get stable generated anchors.
	container := doc.Nodes[0]
	if got := rd.Anchors[container]; got != "sec-section" {
		t.Errorf("section anchor = %q, want sec-section", got)
	}
	if got := rd.Anchors[container.Children[0]]; got != "SYS-001" {
		t.Errorf("requirement anchor = %q, want SYS-001", got)
	}
	if got := rd.Anchors[container.Children[0].Children[0]]; got != "SYS-002" {
		t.Errorf("child anchor = %q, want SYS-002", got)
	}
	if got := rd.Anchors[doc.Nodes[1]]; got != "sec-note" {
		t.Errorf("leaf section anchor = %q, want sec-note", got)
	}

	// Filters collect values from every requirement; status is ordered
	// built-ins first.
	if got, want := rd.Filters["status"], []string{"draft", "approved"}; !reflect.DeepEqual(got, want) {
		t.Errorf("status filter = %v, want %v", got, want)
	}
	if got, want := rd.Filters["priority"], []string{"High", "Low"}; !reflect.DeepEqual(got, want) {
		t.Errorf("priority filter = %v, want %v", got, want)
	}
	if _, ok := rd.Filters["trace"]; ok {
		t.Error("trace must never become a filter")
	}

	// Status breakdown counts the implicit approved default.
	if len(rd.StatusBreakdown) != 2 {
		t.Fatalf("status breakdown = %+v, want 2 entries", rd.StatusBreakdown)
	}
	if rd.StatusBreakdown[0].Value != "draft" || rd.StatusBreakdown[0].Count != 1 {
		t.Errorf("breakdown[0] = %+v", rd.StatusBreakdown[0])
	}
	if rd.StatusBreakdown[1].Value != "approved" || rd.StatusBreakdown[1].Count != 1 {
		t.Errorf("breakdown[1] = %+v", rd.StatusBreakdown[1])
	}
}

func TestBuildRenderDataIgnoreStatus(t *testing.T) {
	doc := testDoc()
	doc.XReqmd.IgnoreStatus = true
	rd := buildRenderData(doc)
	if _, ok := rd.Filters["status"]; ok {
		t.Error("status filter must be hidden for ignore-status documents")
	}
	if len(rd.StatusBreakdown) != 0 {
		t.Error("status breakdown must be hidden for ignore-status documents")
	}
}

func TestBuildFiltersOmitsSingleValueAttrs(t *testing.T) {
	reqs := []*model.Node{
		{ID: "A", Attrs: map[string]any{"status": "approved", "owner": "alice"}},
		{ID: "B", Attrs: map[string]any{"status": "approved", "owner": "alice"}},
	}
	doc := model.Document{Path: "/d"}
	if got := buildFilters(reqs, doc); len(got) != 0 {
		t.Errorf("filters = %v, want empty (single distinct values)", got)
	}
}

func TestBuildFiltersStatusOrdering(t *testing.T) {
	reqs := []*model.Node{
		{ID: "A", Attrs: map[string]any{"status": "review"}},
		{ID: "B", Attrs: map[string]any{"status": "approved"}},
		{ID: "C", Attrs: map[string]any{"status": "draft"}},
		{ID: "D", Attrs: map[string]any{"status": "zzz"}},
	}
	doc := model.Document{XReqmd: &model.XReqmd{AdditionalStatusValues: []string{"review"}}}
	got := buildFilters(reqs, doc)["status"]
	want := []string{"draft", "approved", "review", "zzz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("status filter = %v, want %v", got, want)
	}
}

func TestNormalizeAttrValue(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want []string
	}{
		{"nil", nil, nil},
		{"string", "x", []string{"x"}},
		{"int", 3, []string{"3"}},
		{"bool", true, []string{"true"}},
		{"string slice", []string{"a", "b"}, []string{"a", "b"}},
		{"any slice", []any{"a", 1, nil, map[string]any{"x": 1}}, []string{"a", "1"}},
		{"map", map[string]any{"x": 1}, nil},
		{"empty any slice", []any{}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAttrValue(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeAttrValue(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSortedStatusValues(t *testing.T) {
	got := sortedStatusValues(
		[]string{"zzz", "b-review", "approved", "draft"},
		[]string{"b-review"},
	)
	want := []string{"draft", "approved", "b-review", "zzz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sortedStatusValues = %v, want %v", got, want)
	}
}

func TestItemHeadingTagNativeLevels(t *testing.T) {
	// Top-level items never render as h1 (that is the document title); nested
	// items keep their native depth, still capped at h6.
	tests := []struct {
		level   int
		isChild bool
		want    string
	}{
		{0, false, "h2"},
		{1, false, "h2"},
		{2, false, "h2"},
		{3, false, "h3"},
		{6, false, "h6"},
		{7, false, "h6"},
		{3, true, "h3"},
		{4, true, "h4"},
		{7, true, "h6"},
		// A nested node with no recorded level still renders below the title.
		{0, true, "h3"},
		{1, true, "h3"},
	}
	for _, tt := range tests {
		if got := itemHeadingTag(tt.level, tt.isChild); got != tt.want {
			t.Errorf("itemHeadingTag(%d, %v) = %q, want %q", tt.level, tt.isChild, got, tt.want)
		}
	}
}

func TestDocTitleFallback(t *testing.T) {
	if got := docTitle(model.Document{Path: "/p"}); got != "" {
		t.Errorf("docTitle without title = %q, want empty", got)
	}
	if got := docTitle(model.Document{Title: "H1"}); got != "H1" {
		t.Errorf("docTitle = %q, want H1", got)
	}
	if got := docTitle(model.Document{Schema: map[string]any{"title": "Schema"}}); got != "Schema" {
		t.Errorf("docTitle schema fallback = %q, want Schema", got)
	}
}

func TestAttributeColumns(t *testing.T) {
	doc := testDoc()
	doc.Properties = []string{"status", "trace", "priority", "owner"}
	got := attributeColumns(doc, doc.Properties)
	want := []string{"priority", "owner"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("attributeColumns = %v, want %v", got, want)
	}
}

func TestBuildPageModel(t *testing.T) {
	doc := testDoc()
	rd := buildRenderData(doc)
	m := buildPageModel(doc, rd, nil, nil, doc.Properties)

	if m.Doc.Title != "Test Document" {
		t.Errorf("doc title = %q", m.Doc.Title)
	}
	if m.Doc.Level != "system" || m.Doc.DocumentID != "test" {
		t.Errorf("doc meta = %+v", m.Doc)
	}
	if m.DefaultStatus != model.StatusApproved {
		t.Errorf("default status = %q", m.DefaultStatus)
	}
	if len(m.Nodes) != 4 {
		t.Fatalf("nodes = %d, want 4", len(m.Nodes))
	}
	// Node order is document order, depth-first: container, SYS-001,
	// SYS-002, info.
	wantIDs := []string{"sec-section", "SYS-001", "SYS-002", "sec-note"}
	for i, id := range wantIDs {
		if m.Nodes[i].ID != id {
			t.Errorf("node[%d].ID = %q, want %q", i, m.Nodes[i].ID, id)
		}
	}
	var container *pageNode
	for i := range m.Nodes {
		if m.Nodes[i].ID == "sec-section" {
			container = &m.Nodes[i]
		}
	}
	if container == nil || container.Type != "section" {
		t.Errorf("section node = %+v, want type section", container)
	}
	// The model deliberately carries no body: the Document view is the single
	// representation of requirement prose.
	data, _ := json.Marshal(m)
	if strings.Contains(string(data), "Root body.") {
		t.Error("page model must not duplicate requirement bodies")
	}
	if strings.Contains(string(data), `"trace"`) {
		t.Error("page model must not carry the raw trace attribute")
	}
}

func TestBuildPageModelResolveLink(t *testing.T) {
	doc := testDoc()
	rd := buildRenderData(doc)
	m := buildPageModel(doc, rd, nil, func(id string) string { return "other.html#" + id }, doc.Properties)
	var sys *pageNode
	for i := range m.Nodes {
		if m.Nodes[i].ID == "SYS-001" {
			sys = &m.Nodes[i]
		}
	}
	if sys == nil {
		t.Fatal("SYS-001 not in model")
	}
	if sys.Attrs["status"] != "approved" {
		t.Errorf("attrs = %v", sys.Attrs)
	}
}

func TestDocTitleFallsBackToPathInExport(t *testing.T) {
	var h HTML
	var buf strings.Builder
	doc := model.Document{Path: "/no/title", Nodes: []*model.Node{{ID: "R-1"}}}
	if err := h.Export(&buf, doc, nil); err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(buf.String(), "<title>/no/title</title>") {
		t.Error("title should fall back to the document path")
	}
}
