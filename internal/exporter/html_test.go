package exporter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/internal/graph"
	"github.com/dVoo/reqmd/internal/model"
)

// analysisFor builds the presentation analysis from a document set.
func analysisFor(t *testing.T, docs ...model.Document) *Analysis {
	t.Helper()
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("graph.New: %v", err)
	}
	return NewAnalysis(g, BuildTitleMap(docs))
}

// exportPage renders one document and returns the HTML.
func exportPage(t *testing.T, h *HTML, doc model.Document, props []string) string {
	t.Helper()
	var buf strings.Builder
	if err := h.Export(&buf, doc, props); err != nil {
		t.Fatalf("HTML export: %v", err)
	}
	return buf.String()
}

// pageModelOf extracts and decodes the embedded page model.
func pageModelOf(t *testing.T, output string) pageModel {
	t.Helper()
	re := regexp.MustCompile(`(?s)<script type="application/json" id="req-model">(.*?)</script>`)
	m := re.FindStringSubmatch(output)
	if m == nil {
		t.Fatal("req-model script not found")
	}
	var model pageModel
	if err := json.Unmarshal([]byte(m[1]), &model); err != nil {
		t.Fatalf("decoding req-model: %v", err)
	}
	return model
}

func basicDoc() model.Document {
	return model.Document{
		Path:       "/test/doc",
		Schema:     schemaWithTitle("Test Export"),
		Properties: []string{"status", "asil", "maturity", "verify"},
		Nodes: []*model.Node{
			{
				ID: "REQ-001",
				Attrs: map[string]any{
					"status":   "approved",
					"asil":     "B",
					"maturity": "released",
					"verify":   "Test",
				},
				Body:      "The system shall do X.",
				Rationale: "Required for safety.",
			},
			{
				ID:    "REQ-002",
				Attrs: map[string]any{"status": "draft"},
				Body:  "The system shall do Y.",
			},
		},
	}
}

func TestHTMLExport(t *testing.T) {
	doc := basicDoc()
	out := exportPage(t, &HTML{}, doc, doc.Properties)

	if !strings.Contains(out, "<!DOCTYPE html>") || !strings.Contains(out, `<html lang="en">`) {
		t.Error("missing doctype or lang")
	}
	if !strings.Contains(out, "<title>Test Export</title>") || !strings.Contains(out, "<h1>Test Export</h1>") {
		t.Error("missing document title")
	}

	// One quiet card per requirement, content visible by default.
	if !strings.Contains(out, `<article class="req" id="REQ-001"`) {
		t.Error("missing REQ-001 card")
	}
	if !strings.Contains(out, `class="req__id">REQ-001</code>`) {
		t.Error("missing REQ-001 id in the card header")
	}
	if !strings.Contains(out, "The system shall do X.") {
		t.Error("missing body text")
	}
	if !strings.Contains(out, "Required for safety.") {
		t.Error("missing rationale")
	}
	// Plain counts only: parent/child is the internal content model and the
	// nested suffix appears only when a document actually nests requirements.
	if !strings.Contains(out, ">2 requirements<") {
		t.Error("missing requirement count")
	}
	if strings.Contains(out, "parents,") {
		t.Error("doc header exposes the internal parent/child vocabulary")
	}

	// Progressive disclosure: attributes live in a collapsed detail block.
	if !strings.Contains(out, `<details class="req__detail">`) {
		t.Error("missing progressive-disclosure detail block")
	}
	if !strings.Contains(out, "<dt>Status</dt>") || !strings.Contains(out, "<dd>approved</dd>") {
		t.Error("missing attribute grid entries")
	}

	// Server-rendered outline (works without JavaScript).
	if !strings.Contains(out, `class="toc"`) || !strings.Contains(out, `data-toc-for="REQ-001"`) {
		t.Error("missing server-rendered outline")
	}

	// Exactly one reading view is visible; analytical views are empty shells
	// that render on demand.
	for _, view := range []string{"document", "table", "trace", "coverage"} {
		if !strings.Contains(out, `data-view-panel="`+view+`"`) {
			t.Errorf("missing view shell %q", view)
		}
	}
	if !strings.Contains(out, `<section class="view view--document" data-view-panel="document">`) {
		t.Error("document view must be visible by default")
	}
	if !strings.Contains(out, `<section class="view" data-view-panel="table" hidden>`) {
		t.Error("table view must start hidden and empty")
	}
	if !strings.Contains(out, `<section class="view" data-view-panel="trace" hidden></section>`) {
		t.Error("trace view must start as an empty shell")
	}

	// Shared model + palette/shortcuts overlays.
	if !strings.Contains(out, `<script type="application/json" id="req-model">`) {
		t.Error("missing shared view model")
	}
	if !strings.Contains(out, `id="palette"`) || !strings.Contains(out, `id="shortcuts"`) {
		t.Error("missing palette or shortcuts overlay")
	}
}

// TestHTMLExportOfflineContent is the Phase 0/1 regression guard: primary
// content is visible by default, and nothing loads from a CDN.
func TestHTMLExportOfflineContent(t *testing.T) {
	doc := basicDoc()
	out := exportPage(t, &HTML{}, doc, doc.Properties)

	if strings.Contains(out, "x-cloak") {
		t.Error("primary content must not use x-cloak")
	}
	if strings.Contains(out, "cdn.jsdelivr.net") || strings.Contains(out, "unpkg.com") {
		t.Error("export must not load assets from a CDN")
	}
	if strings.Contains(out, "alpinejs") || strings.Contains(out, "x-data") {
		t.Error("export must not depend on Alpine.js")
	}
	// The card markup contains no hidden/suppression hooks.
	if strings.Contains(out, `hidden id="REQ-001"`) {
		t.Error("requirement cards must not be hidden by default")
	}
	// The stylesheet and controller are inlined.
	if !strings.Contains(out, "reqmd HTML export") {
		t.Error("embedded stylesheet missing")
	}
	if !strings.Contains(out, "reqmd HTML export client") {
		t.Error("embedded client script missing")
	}
}

func TestHTMLExportHTMLEscaping(t *testing.T) {
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test"),
		Nodes: []*model.Node{
			{
				ID:        "REQ-&",
				Attrs:     map[string]any{"status": "a < b & c > d"},
				Body:      "Body with <tag> & stuff",
				Rationale: "Rationale with \"quotes\" & <angle>",
			},
		},
	}
	out := exportPage(t, &HTML{}, doc, []string{"status"})

	if strings.Contains(out, ">REQ-&<") || !strings.Contains(out, "REQ-&amp;") {
		t.Error("requirement ID must be HTML-escaped")
	}
	if strings.Contains(out, "a < b") {
		t.Error("attribute value must be HTML-escaped")
	}
	if strings.Contains(out, "<tag>") {
		t.Error("raw HTML in the body must be stripped")
	}
	if strings.Contains(out, `"quotes"`) {
		t.Error("rationale quotes must be escaped")
	}
}

// TestHTMLExportModelIsHTMLSafe pins that author-controlled values in the
// embedded JSON cannot break out of the script element.
func TestHTMLExportModelIsHTMLSafe(t *testing.T) {
	doc := model.Document{
		Path:       "/test/esc",
		Schema:     schemaWithTitle("Esc"),
		Properties: []string{"status", "note"},
		Nodes: []*model.Node{{
			ID:    "SYS-001",
			Title: `Quote " and <tag>`,
			Attrs: map[string]any{
				"status": "approved",
				"note":   `</script><script>alert(1)</script>`,
			},
			Body: "Body.",
		}},
	}
	out := exportPage(t, &HTML{}, doc, doc.Properties)
	if strings.Contains(out, "</script><script>alert(1)</script>") {
		t.Error("attribute value broke out of the model script element")
	}
	if !strings.Contains(out, `\u003c/script\u003e`) {
		t.Error("JSON should use HTML-safe escaping for < and >")
	}
}

// TestHTMLExportModelOmitsBodies verifies the duplicate-data fix: the reading
// view is the only representation of requirement prose.
func TestHTMLExportModelOmitsBodies(t *testing.T) {
	doc := basicDoc()
	doc.Nodes[0].Rationale = "A very distinctive rationale."
	out := exportPage(t, &HTML{}, doc, doc.Properties)
	m := pageModelOf(t, out)
	data, _ := json.Marshal(m)
	if strings.Contains(string(data), "The system shall do X.") {
		t.Error("model must not duplicate requirement body")
	}
	if strings.Contains(string(data), "A very distinctive rationale.") {
		t.Error("model must not duplicate requirement rationale")
	}
	if !strings.Contains(out, "A very distinctive rationale.") {
		t.Error("rationale should still be in the server-rendered Document view")
	}
}

func TestHTMLExportTraceAttributeNotInGrid(t *testing.T) {
	doc := basicDoc()
	doc.Nodes[0].Attrs[model.AttrTrace] = []any{"REQ-002"}
	doc.Properties = append([]string{model.AttrTrace}, doc.Properties...)
	out := exportPage(t, &HTML{}, doc, doc.Properties)
	if strings.Contains(out, `data-attr="trace"`) {
		t.Error("trace must not appear in the attribute grid")
	}
	if strings.Contains(out, "&lt;dt&gt;Trace&lt;/dt&gt;") {
		t.Error("trace label must not render")
	}
}

func TestHTMLExportArrayPillsAndNestedObjects(t *testing.T) {
	doc := model.Document{
		Path:       "/test/doc",
		Schema:     schemaWithTitle("Array Test"),
		Properties: []string{"aspice-bp", "status", "metadata", "flags"},
		Nodes: []*model.Node{
			{
				ID: "REQ-001",
				Attrs: map[string]any{
					"aspice-bp": []string{"SYS2-BP2", "SWE1-BP2"},
					"status":    "Draft",
					"metadata":  map[string]any{"source": "legacy", "priority": 2},
					"flags":     []any{true, 42, nil, "fast"},
				},
				Body: "Body.",
			},
			{
				ID: "REQ-002",
				Attrs: map[string]any{
					"aspice-bp": []any{},
					"status":    "Draft",
				},
				Body: "Body.",
			},
		},
	}
	out := exportPage(t, &HTML{}, doc, doc.Properties)

	for _, want := range []string{
		`class="attr__pills"`, `class="attr__pill"`, "SYS2-BP2", "SWE1-BP2",
		"true", "42", `class="attr__nested"`, "<dd>—</dd>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
	if !strings.Contains(out, `&#34;source&#34;: &#34;legacy&#34;`) {
		t.Error("nested map should be JSON-escaped inside <pre>")
	}
}

func TestHTMLExportItems(t *testing.T) {
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Export"),
		Nodes: []*model.Node{
			{
				Kind:  model.KindContainer,
				Title: "Section One",
				Level: 1,
				Body:  "Container prose.",
				Children: []*model.Node{
					{ID: "REQ-001", Attrs: map[string]any{"status": "approved"}, Body: "Req body."},
				},
			},
			{Kind: model.KindInfo, Title: "Note", Level: 2, Body: "Some note."},
		},
	}
	out := exportPage(t, &HTML{}, doc, []string{"status"})

	if !strings.Contains(out, `<section class="item item--container" id="item-1"`) {
		t.Error("container section missing")
	}
	if !strings.Contains(out, "Section One") || !strings.Contains(out, "Container prose.") {
		t.Error("container content missing")
	}
	if !strings.Contains(out, `<article class="req req--child" id="REQ-001"`) {
		t.Error("child card missing")
	}
	if !strings.Contains(out, `<section class="item" id="item-2"`) {
		t.Error("info section missing")
	}
	if !strings.Contains(out, "1 requirement · 2 items") {
		t.Error("doc count line missing")
	}
	m := pageModelOf(t, out)
	var types []string
	for _, n := range m.Nodes {
		types = append(types, n.Type)
	}
	if types[0] != "container" || types[2] != "info" {
		t.Errorf("model types = %v, want container ... info", types)
	}
}

// TestHTMLExport_UndeclaredAttributeRendered verifies undeclared attributes
// still render (flagged) and feed the filter menu.
func TestHTMLExport_UndeclaredAttributeRendered(t *testing.T) {
	doc := model.Document{
		Path:       "/test/doc",
		Schema:     schemaWithTitle("Undeclared Attrs"),
		Properties: []string{"status", "trace"},
		Nodes: []*model.Node{{
			ID: "REQ-001",
			Attrs: map[string]any{
				"status": "approved",
				"owner":  "alice",
				"team":   "platform",
			},
			Body: "Uses undeclared attributes.",
		}},
	}
	out := exportPage(t, &HTML{}, doc, doc.Properties)

	if !strings.Contains(out, `<div class="attrs__row" data-attr="status">`) {
		t.Error("declared attribute should render without the undeclared modifier")
	}
	for _, name := range []string{"owner", "team"} {
		want := `<div class="attrs__row attrs__row--undeclared" data-attr="` + name + `">`
		if !strings.Contains(out, want) {
			t.Errorf("undeclared attribute %q missing; want %s", name, want)
		}
	}
	if !strings.Contains(out, "<dt>Owner</dt>") || !strings.Contains(out, "<dd>alice</dd>") {
		t.Error("undeclared attribute should render its label and value")
	}
	if strings.Index(out, `data-attr="owner"`) > strings.Index(out, `data-attr="team"`) {
		t.Error("undeclared attributes should be alphabetical")
	}
	if strings.Contains(out, `data-attr="trace"`) {
		t.Error("trace must not appear in the attribute grid")
	}
}

// TestHTMLExportTraceLinkState checks pins, versions, and link health in the
// shared model and the server-rendered links.
func TestHTMLExportTraceLinkState(t *testing.T) {
	up := model.Document{
		Path:   "/test/up",
		Schema: schemaWithTitle("Up"),
		XReqmd: &model.XReqmd{DocumentID: "up", Level: "system"},
		Nodes: []*model.Node{{
			ID: "UP-001", Source: "/test/up/up.md", Title: "Upstream requirement",
			Attrs: map[string]any{"status": "approved", "version": 3},
			Body:  "up",
		}},
	}
	down := model.Document{
		Path:       "/test/down",
		Schema:     schemaWithTitle("Down"),
		Properties: []string{"status", "trace"},
		XReqmd:     &model.XReqmd{DocumentID: "down", Level: "software"},
		Nodes: []*model.Node{{
			ID: "SW-001", Source: "/test/down/down.md", Title: "Downstream requirement",
			Attrs: map[string]any{
				"status": "approved",
				"trace":  []any{"up/UP-001~2", "up/MISSING"},
			},
			Body: "down",
		}},
	}
	an := analysisFor(t, up, down)
	h := &HTML{}
	h.SetAnalysis(an)
	out := exportPage(t, h, down, down.Properties)
	m := pageModelOf(t, out)

	var sw *pageNode
	for i := range m.Nodes {
		if m.Nodes[i].ID == "SW-001" {
			sw = &m.Nodes[i]
		}
	}
	if sw == nil {
		t.Fatal("SW-001 missing from model")
	}
	if len(sw.Up) != 2 {
		t.Fatalf("up links = %d, want 2", len(sw.Up))
	}
	if sw.Up[0].ID != "UP-001" || !sw.Up[0].Resolved || sw.Up[0].Direction != graph.DirOutdated || sw.Up[0].Version != 3 {
		t.Errorf("resolved link = %+v, want outdated UP-001 at v3", sw.Up[0])
	}
	if sw.Up[1].Resolved {
		t.Errorf("dangling link = %+v, want unresolved", sw.Up[1])
	}
	if !strings.Contains(out, "link__pin--outdated") {
		t.Error("outdated pin should render with its stale modifier")
	}
	if !strings.Contains(out, "link--broken") {
		t.Error("unresolved reference should render as broken")
	}
}

// TestHTMLExportDownstreamLinkDirection pins that a requirement's inbound
// links appear as Down (who traces to it), never as Up.
func TestHTMLExportDownstreamLinkDirection(t *testing.T) {
	doc := model.Document{
		Path:   "/test/up",
		Schema: schemaWithTitle("Up"),
		XReqmd: &model.XReqmd{DocumentID: "up", Level: "system"},
		Nodes: []*model.Node{{
			ID: "UP-001", Source: "/test/up/up.md", Title: "Upstream requirement",
			Attrs: map[string]any{"status": "approved"},
			Body:  "up",
		}},
	}
	child := model.Document{
		Path:   "/test/down",
		Schema: schemaWithTitle("Down"),
		XReqmd: &model.XReqmd{DocumentID: "down", Level: "software"},
		Nodes: []*model.Node{{
			ID: "SW-001", Source: "/test/down/down.md", Title: "Downstream requirement",
			Attrs: map[string]any{"trace": []any{"up/UP-001"}},
			Body:  "down",
		}},
	}
	h := &HTML{}
	h.SetAnalysis(analysisFor(t, doc, child))
	m := pageModelOf(t, exportPage(t, h, doc, nil))
	var up *pageNode
	for i := range m.Nodes {
		if m.Nodes[i].ID == "UP-001" {
			up = &m.Nodes[i]
		}
	}
	if up == nil {
		t.Fatal("UP-001 missing")
	}
	if len(up.Up) != 0 {
		t.Errorf("UP-001 up links = %+v, want none", up.Up)
	}
	if len(up.Down) != 1 || up.Down[0].ID != "SW-001" {
		t.Errorf("UP-001 down links = %+v, want SW-001", up.Down)
	}
}

// TestMatrixTooltipDirection pins the client's direction-specific matrix
// tooltip, the regression from the old downstream matrix.
func TestMatrixTooltipDirection(t *testing.T) {
	// Upstream: row requirement traces to the column.
	if !strings.Contains(js, "dir === 'up' ? n.id + ' → ' + (col.id || col.ref)") {
		t.Error("matrix tooltip must render row → column for the upstream direction")
	}
	// Downstream: the column requirement traces to the row.
	if !strings.Contains(js, "(col.id || col.ref) + ' → ' + n.id") {
		t.Error("matrix tooltip must render column → row for the downstream direction")
	}
}

// TestHTMLExportCoverageStates checks that coverage comes from graph
// semantics, not link presence: a draft downstream does not cover.
func TestHTMLExportCoverageStates(t *testing.T) {
	system := model.Document{
		Path:   "/test/system",
		Schema: schemaWithTitle("System"),
		XReqmd: &model.XReqmd{DocumentID: "system", Level: "system", RequiresTraceFrom: []string{"software"}},
		Nodes: []*model.Node{
			{ID: "SYS-COVERED", Source: "/test/system/a.md", Attrs: map[string]any{"status": "approved"}},
			{ID: "SYS-MISSING", Source: "/test/system/b.md", Attrs: map[string]any{"status": "approved", "requires-trace-from": []any{"software"}}},
		},
	}
	software := model.Document{
		Path:   "/test/software",
		Schema: schemaWithTitle("Software"),
		XReqmd: &model.XReqmd{DocumentID: "software", Level: "software"},
		Nodes: []*model.Node{
			{ID: "SW-APPROVED", Source: "/test/software/a.md", Attrs: map[string]any{"status": "approved", "trace": []any{"system/SYS-COVERED"}}},
			{ID: "SW-DRAFT", Source: "/test/software/b.md", Attrs: map[string]any{"status": "draft", "trace": []any{"system/SYS-MISSING"}}},
		},
	}
	h := &HTML{}
	h.SetAnalysis(analysisFor(t, system, software))
	out := exportPage(t, h, system, nil)

	if !strings.Contains(out, `id="SYS-COVERED" data-status="approved" data-coverage="covered"`) {
		t.Error("approved downstream should cover SYS-COVERED")
	}
	if !strings.Contains(out, `id="SYS-MISSING" data-status="approved" data-coverage="missing-coverage"`) {
		t.Error("draft downstream must not cover SYS-MISSING")
	}
	if !strings.Contains(out, "chip--coverage-missing") {
		t.Error("missing coverage should be surfaced as a chip")
	}
}

// TestHTMLExportNoVerdictIsFirstClass verifies unverified eligible measures
// show a no-verdict chip and are marked eligible in the model, so the client
// denominator includes them.
func TestHTMLExportNoVerdictIsFirstClass(t *testing.T) {
	doc := model.Document{
		Path:   "/test/measures",
		Schema: schemaWithTitle("Measures"),
		Nodes: []*model.Node{
			{ID: "TST-NONE", Source: "/test/measures/a.md", Attrs: map[string]any{"status": "approved", "verify": "Test"}},
			{ID: "TST-PASS", Source: "/test/measures/b.md", Attrs: map[string]any{"status": "approved", "verify": "Test"}},
		},
	}
	results := model.Document{
		Path:  "<verify-results>",
		Nodes: []*model.Node{{ID: "RESULT:TST-PASS", Source: "ci/run.ctrf.json", Attrs: map[string]any{model.AttrResultOutcome: "pass", model.AttrTrace: []any{"TST-PASS"}}}},
	}
	h := &HTML{}
	h.SetAnalysis(analysisFor(t, doc, results))
	out := exportPage(t, h, doc, nil)

	if !strings.Contains(out, "chip--no-verdict") || !strings.Contains(out, "no verdict") {
		t.Error("eligible measure without a result should show a no-verdict chip")
	}
	m := pageModelOf(t, out)
	for _, n := range m.Nodes {
		if strings.HasPrefix(n.ID, "RESULT:") {
			t.Errorf("synthetic result node %q must not appear in the page model", n.ID)
		}
		for _, l := range append(n.Up, n.Down...) {
			if strings.HasPrefix(l.ID, "RESULT:") || strings.HasPrefix(l.ID, "TC:") {
				t.Errorf("node %q links to synthetic %q; results are evidence, not trace links", n.ID, l.ID)
			}
		}
		switch n.ID {
		case "TST-NONE":
			if !n.Eligible || n.Verdict != "" {
				t.Errorf("TST-NONE = %+v, want eligible with no verdict", n)
			}
		case "TST-PASS":
			if !n.Eligible || n.Verdict != "pass" {
				t.Errorf("TST-PASS = %+v, want eligible pass", n)
			}
		}
	}
}

// TestHTMLExportEvidenceList keeps the rolled-up evidence rendering.
func TestHTMLExportEvidenceList(t *testing.T) {
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Export"),
		Nodes: []*model.Node{{
			ID: "REQ-001", Source: "/test/doc/req.md",
			Attrs: map[string]any{"status": "approved", "verify": "Test"},
			Body:  "The system shall boot.",
		}},
	}
	// A synthesized test case traces to the measure; its result traces to
	// the case, which is how the graph derives the evidence description.
	results := model.Document{
		Path:      "<verify-results>",
		Synthetic: true,
		Nodes: []*model.Node{
			{ID: "TC:BOOT-TIME", Source: "<verify-results>", Body: "Measures boot time.\n\n```go\nboot()\n```", Attrs: map[string]any{model.AttrTrace: []any{"REQ-001"}}},
			{ID: "RESULT:REQ-001#BOOT-TIME", Source: "ci/run.ctrf.json", Attrs: map[string]any{model.AttrResultOutcome: "pass", model.AttrTrace: []any{"TC:BOOT-TIME"}}},
		},
	}
	h := &HTML{}
	h.SetAnalysis(analysisFor(t, doc, results))
	out := exportPage(t, h, doc, nil)

	for _, want := range []string{"evidence", "Evidence (1)", "BOOT-TIME", "evidence__desc", "Measures boot time."} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

// TestHTMLExportLargeDocumentBound guards against quadratic hidden markup:
// a large document's output grows linearly and emits no matrix/table DOM.
func TestHTMLExportLargeDocumentBound(t *testing.T) {
	const n = 300
	nodes := make([]*model.Node, 0, n)
	for i := range n {
		id := fmt.Sprintf("REQ-%04d", i)
		attrs := map[string]any{"status": "approved"}
		if i > 0 {
			attrs[model.AttrTrace] = []any{fmt.Sprintf("REQ-%04d", i-1)}
		}
		nodes = append(nodes, &model.Node{
			ID: id, Source: fmt.Sprintf("/test/large/%s.md", id), Attrs: attrs,
			Body: "A requirement body of moderate length.",
		})
	}
	doc := model.Document{
		Path:       "/test/large",
		Schema:     schemaWithTitle("Large"),
		Properties: []string{"status", "trace"},
		Nodes:      nodes,
	}
	h := &HTML{}
	h.SetAnalysis(analysisFor(t, doc))
	out := exportPage(t, h, doc, doc.Properties)

	if len(out) > n*8000 {
		t.Errorf("output = %d bytes for %d requirements; expected linear, bounded markup", len(out), n)
	}
	// Analytical views are generated client-side: no server-rendered table
	// or matrix markup exists before a view is selected.
	if strings.Contains(out, `<table class="grid" id="req-table">`) {
		t.Error("table view markup must not be emitted before it is selected")
	}
	if strings.Contains(out, `class="matrix__cell matrix__cell--on"`) {
		t.Error("matrix cells must not be emitted before the matrix is selected")
	}
	// Every requirement is still present in the reading view.
	for _, id := range []string{"REQ-0000", "REQ-0150", "REQ-0299"} {
		if !strings.Contains(out, `id="`+id+`"`) {
			t.Errorf("requirement %s missing from the reading view", id)
		}
	}
}

// TestProjectIndex checks the project home page content and queues.
func TestProjectIndex(t *testing.T) {
	system := model.Document{
		Path:   "/spec/system",
		Schema: schemaWithTitle("System"),
		XReqmd: &model.XReqmd{DocumentID: "system", Level: "system", RequiresTraceFrom: []string{"software"}},
		Nodes: []*model.Node{
			{ID: "SYS-001", Source: "/spec/system/a.md", Title: "Covered item", Attrs: map[string]any{"status": "approved"}},
			{ID: "SYS-002", Source: "/spec/system/b.md", Title: "Uncovered item", Attrs: map[string]any{"status": "approved"}},
			{ID: "TST-001", Source: "/spec/system/c.md", Title: "Unverified measure", Attrs: map[string]any{"status": "approved", "verify": "Test"}},
		},
	}
	software := model.Document{
		Path:   "/spec/software",
		Schema: schemaWithTitle("Software"),
		XReqmd: &model.XReqmd{DocumentID: "software", Level: "software"},
		Nodes: []*model.Node{
			{ID: "SW-001", Source: "/spec/software/a.md", Attrs: map[string]any{"status": "approved", "trace": []any{"system/SYS-001"}}},
		},
	}
	docs := []model.Document{system, software}
	an := analysisFor(t, docs...)
	p := NewProjectIndex(docs, an, BuildTitleMap(docs), func(path string) string { return "page-" + strings.TrimPrefix(path, "/spec/") + ".html" })

	var buf strings.Builder
	if err := p.Export(&buf); err != nil {
		t.Fatalf("project export: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		`data-page="project"`, "Documents", "System", "Software",
		"Missing coverage", "Missing verdicts", "Unverified measure",
		"page-system.html#SYS-002", `id="project-reqs"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("project page missing %q", want)
		}
	}
	if strings.Contains(out, "cdn.jsdelivr.net") {
		t.Error("project page must not load assets from a CDN")
	}
}
