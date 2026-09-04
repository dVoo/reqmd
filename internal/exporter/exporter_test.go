package exporter

import (
	"fmt"
	"reqmd/internal/model"
	"strings"
	"testing"
)

func schemaWithTitle(title string) map[string]any {
	return map[string]any{
		"title":    title,
		"type":     "object",
		"required": []any{"status", "asil"},
		"properties": map[string]any{
			"status":   map[string]any{"type": "string"},
			"asil":     map[string]any{"type": "string"},
			"maturity": map[string]any{"type": "string"},
			"verify":   map[string]any{"type": "string"},
		},
	}
}

func TestCSVExport(t *testing.T) {
	propOrder := []string{"status", "asil", "maturity", "verify"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Export"),
		Nodes: []*model.Node{
			{
				ID: "REQ-001",
				Attrs: map[string]any{
					"status":   "approved",
					"asil":     "B",
					"maturity": "released",
					"verify":   "test",
				},
				Body:      "The system shall do X.",
				Rationale: "Required for safety.",
			},
			{
				ID: "REQ-002",
				Attrs: map[string]any{
					"status": "draft",
					// asil, maturity, verify intentionally missing
				},
				Body:      "The system shall do Y.",
				Rationale: "",
			},
		},
	}

	var buf strings.Builder
	err := (&CSV{}).Export(&buf, doc, propOrder)
	if err != nil {
		t.Fatalf("CSV export failed: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d", len(lines))
	}

	// Header
	header := lines[0]
	wantHeader := "Type,ID,Title,status,asil,maturity,verify,Body,Rationale"
	if header != wantHeader {
		t.Errorf("header = %q, want %q", header, wantHeader)
	}

	// Row 1 - all fields (Title empty)
	row1 := lines[1]
	wantRow1 := "req,REQ-001,,approved,B,released,test,The system shall do X.,Required for safety."
	if row1 != wantRow1 {
		t.Errorf("row 1 = %q, want %q", row1, wantRow1)
	}

	// Row 2 - missing attrs (Title empty)
	row2 := lines[2]
	wantRow2 := "req,REQ-002,,draft,,,,The system shall do Y.,"
	if row2 != wantRow2 {
		t.Errorf("row 2 = %q, want %q", row2, wantRow2)
	}
}

func TestCSVExportSpecialTypes(t *testing.T) {
	propOrder := []string{"status", "safety_relevant", "level", "trace"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: map[string]any{"type": "object"},
		Nodes: []*model.Node{
			{
				ID: "REQ-003",
				Attrs: map[string]any{
					"status":          "approved",
					"safety_relevant": true,
					"level":           3,
					"trace":           []string{"T1", "T2", "T3"},
				},
				Body: "Special types requirement.",
			},
		},
	}

	var buf strings.Builder
	err := (&CSV{}).Export(&buf, doc, propOrder)
	if err != nil {
		t.Fatalf("CSV export failed: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + 1 row), got %d", len(lines))
	}

	header := lines[0]
	wantHeader := "Type,ID,Title,status,safety_relevant,level,trace,Body,Rationale"
	if header != wantHeader {
		t.Errorf("header = %q, want %q", header, wantHeader)
	}

	row := lines[1]
	// boolean "true", int "3", array "[T1 T2 T3]" (Go fmt representation of []string)
	if !strings.Contains(row, "true") {
		t.Errorf("row should contain 'true' for boolean attr, got: %s", row)
	}
	if !strings.Contains(row, ",3,") {
		t.Errorf("row should contain ',3,' for int attr, got: %s", row)
	}
	if !strings.Contains(row, `[""T1"",""T2"",""T3""]`) {
		t.Errorf("row should contain escaped CSV array for []string attr, got: %s", row)
	}
}

func TestHTMLExport(t *testing.T) {
	propOrder := []string{"status", "asil", "maturity", "verify"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Export"),
		Nodes: []*model.Node{
			{
				ID: "REQ-001",
				Attrs: map[string]any{
					"status":   "approved",
					"asil":     "B",
					"maturity": "released",
					"verify":   "test",
				},
				Body:      "The system shall do X.",
				Rationale: "Required for safety.",
			},
			{
				ID: "REQ-002",
				Attrs: map[string]any{
					"status": "draft",
				},
				Body: "The system shall do Y.",
			},
		},
	}

	var h HTML
	var buf strings.Builder
	err := h.Export(&buf, doc, propOrder)
	if err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}

	output := buf.String()

	// DOCTYPE and lang
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("output missing DOCTYPE")
	}
	if !strings.Contains(output, `<html lang="en">`) {
		t.Error("output missing html lang=en")
	}

	// Title
	if !strings.Contains(output, "<title>Test Export</title>") {
		t.Error("output missing <title>Test Export</title>")
	}
	if !strings.Contains(output, "<h1>Test Export</h1>") {
		t.Error("output missing <h1>Test Export</h1>")
	}

	// Card-based layout: no table
	if strings.Contains(output, "<table>") {
		t.Error("output should not contain a table in card-based layout")
	}
	if strings.Contains(output, "water.min.css") {
		t.Error("output should not reference water.css CDN")
	}

	// Requirement IDs in card headers (h2 with req-id class, wrapped in <code>)
	if !strings.Contains(output, `class="req-id"><code>REQ-001</code></h2>`) {
		t.Error("output missing REQ-001 in card header")
	}
	if !strings.Contains(output, `class="req-id"><code>REQ-002</code></h2>`) {
		t.Error("output missing REQ-002 in card header")
	}

	// Card articles
	if !strings.Contains(output, `<article class="req-card"`) {
		t.Error("output missing req-card articles")
	}
	if !strings.Contains(output, `id="REQ-001"`) {
		t.Error("output missing article id for REQ-001")
	}
	if !strings.Contains(output, `id="REQ-002"`) {
		t.Error("output missing article id for REQ-002")
	}

	// Attributes rendered as description list
	if !strings.Contains(output, "<dl class=\"req-attrs\">") {
		t.Error("output missing attrs description list")
	}
	if !strings.Contains(output, "<dt>Status</dt>") {
		t.Error("output missing Status label in attrs")
	}
	if !strings.Contains(output, "<dd>approved</dd>") {
		t.Error("output missing approved value in attrs")
	}

	// Body and Rationale
	if !strings.Contains(output, "The system shall do X.") {
		t.Error("output missing body text for REQ-001")
	}
	if !strings.Contains(output, "Required for safety.") {
		t.Error("output missing rationale text for REQ-001")
	}
	if !strings.Contains(output, "<strong>Rationale:</strong>") {
		t.Error("output missing Rationale strong tag")
	}

	// Count: text says "2 requirements (2 parents, 0 children)"
	if !strings.Contains(output, "2 requirements (2 parents, 0 children)") {
		t.Error("output missing requirement count")
	}

	// ── Sidebar TOC ──

	// Sidebar TOC panel
	if !strings.Contains(output, `class="req-toc"`) {
		t.Error("output missing req-toc sidebar")
	}
	if !strings.Contains(output, `id="req-toc"`) {
		t.Error("output missing req-toc id")
	}

	// TOC header with title and count placeholder
	if !strings.Contains(output, `class="req-toc-title"`) {
		t.Error("output missing req-toc-title")
	}
	if !strings.Contains(output, `id="req-toc-count"`) {
		t.Error("output missing req-toc-count")
	}

	// TOC list container populated by JS
	if !strings.Contains(output, `id="req-toc-list"`) {
		t.Error("output missing req-toc-list")
	}

	// Content wrapper
	if !strings.Contains(output, `class="content-wrapper"`) {
		t.Error("output missing content-wrapper")
	}

	// Mobile toggle
	if !strings.Contains(output, `id="toc-toggle-btn"`) {
		t.Error("output missing toc-toggle-btn")
	}
	if !strings.Contains(output, `class="toc-toggle-btn"`) {
		t.Error("output missing toc-toggle-btn class")
	}

	// No browser-related elements
	if strings.Contains(output, `id="browse-btn"`) {
		t.Error("output should NOT contain browse-btn (replaced by TOC sidebar)")
	}
	if strings.Contains(output, `id="req-browser"`) {
		t.Error("output should NOT contain req-browser (replaced by TOC sidebar)")
	}
	if strings.Contains(output, `id="req-backdrop"`) {
		t.Error("output should NOT contain req-backdrop (replaced by TOC sidebar)")
	}

	// Index JSON script tag
	if !strings.Contains(output, `<script type="application/json" id="req-index">`) {
		t.Error("output missing req-index JSON script tag")
	}

	// Verify JSON contains both requirement IDs and truncated bodies
	if !strings.Contains(output, `"id":"REQ-001"`) {
		t.Error("req-index JSON missing REQ-001")
	}
	if !strings.Contains(output, `"id":"REQ-002"`) {
		t.Error("req-index JSON missing REQ-002")
	}
	if !strings.Contains(output, `"body":"The system shall do X."`) {
		t.Error("req-index JSON missing body for REQ-001")
	}
	if !strings.Contains(output, `"body":"The system shall do Y."`) {
		t.Error("req-index JSON missing body for REQ-002")
	}

	// Verify TOC sidebar JS elements are in the script
	if !strings.Contains(output, "Build Tree TOC Sidebar") {
		t.Error("output missing Build Tree TOC Sidebar JS code")
	}
	if !strings.Contains(output, "Scroll-spy") {
		t.Error("output missing scroll-spy JS code")
	}

	// Alpine.js integration
	if !strings.Contains(output, "https://cdn.jsdelivr.net/npm/alpinejs@3.14.3/dist/cdn.min.js") {
		t.Error("output missing Alpine.js CDN script tag")
	}
	if !strings.Contains(output, `x-data="reqmdApp()"`) {
		t.Error("output missing Alpine x-data on <body>")
	}
	if !strings.Contains(output, `x-init="initApp()"`) {
		t.Error("output missing x-init on <body>")
	}
	if !strings.Contains(output, "[x-cloak]") {
		t.Error("output missing x-cloak CSS rule")
	}
	if !strings.Contains(output, `id="req-filter-index"`) {
		t.Error("output missing req-filter-index JSON block")
	}
	// Toolbar uses Alpine directives
	if !strings.Contains(output, "x-model.debounce.150ms=\"searchTerm\"") {
		t.Error("output missing Alpine search input binding")
	}
	if !strings.Contains(output, "activeFiltersByAttr") {
		t.Error("output missing activeFiltersByAttr references in template")
	}
	// Theme toggle uses Alpine
	if !strings.Contains(output, "@click=\"toggleTheme()\"") {
		t.Error("output missing Alpine theme toggle binding")
	}
	// Mobile TOC toggle uses Alpine
	if !strings.Contains(output, "@click=\"tocOpen = !tocOpen\"") {
		t.Error("output missing Alpine mobile TOC toggle binding")
	}
	if !strings.Contains(output, ":class=\"{'req-toc--open': tocOpen}\"") {
		t.Error("output missing Alpine :class binding on sidebar")
	}
	// Cards use x-show
	if !strings.Contains(output, "x-show=\"isCardVisible('") {
		t.Error("output missing Alpine x-show on requirement cards")
	}
	// reqmdApp() function defined
	if !strings.Contains(output, "function reqmdApp()") {
		t.Error("output missing reqmdApp() Alpine component")
	}
}

func TestHTMLExportHTMLEscaping(t *testing.T) {
	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test"),
		Nodes: []*model.Node{
			{
				ID: "REQ-&",
				Attrs: map[string]any{
					"status": "a < b & c > d",
				},
				Body:      "Body with <tag> & stuff",
				Rationale: "Rationale with \"quotes\" & <angle>",
			},
		},
	}

	var h HTML
	var buf strings.Builder
	err := h.Export(&buf, doc, propOrder)
	if err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}

	output := buf.String()

	// ID escaped — check no unescaped & immediately before non-amp; text
	if strings.Contains(output, ">REQ-&<") {
		t.Error("ID should be HTML-escaped, found raw unescaped &")
	}
	if !strings.Contains(output, "REQ-&amp;") {
		t.Error("ID should contain &amp;")
	}

	// Attr value escaped
	if strings.Contains(output, "a < b") {
		t.Error("attr value should be HTML-escaped, found raw '<'")
	}
	if !strings.Contains(output, "a &lt; b &amp; c &gt; d") {
		t.Error("attr value not properly escaped")
	}

	// Body rendered through goldmark (raw HTML stripped, entities escaped)
	if strings.Contains(output, "<tag>") {
		t.Error("Body should have raw HTML stripped, found raw '<tag>'")
	}
	if !strings.Contains(output, "raw HTML omitted") {
		t.Error("Body should have raw HTML stripped by goldmark")
	}

	// Rationale rendered through goldmark (quotes escaped)
	if strings.Contains(output, "\"quotes\"") {
		t.Error("Rationale should have quotes escaped, found raw quotes")
	}
	if !strings.Contains(output, "&quot;quotes&quot;") {
		t.Error("Rationale quotes not properly escaped by goldmark")
	}
}

func TestHTMLExportTOCBodyTruncation(t *testing.T) {
	// Build a body that's exactly 81 characters (will be truncated to 80 + "...")
	longBody := strings.Repeat("a", 81)

	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Truncation Test"),
		Nodes: []*model.Node{
			{
				ID:   "REQ-LONG",
				Body: longBody,
			},
			{
				ID:   "REQ-SHORT",
				Body: "Short body.",
			},
		},
	}

	var h HTML
	var buf strings.Builder
	err := h.Export(&buf, doc, propOrder)
	if err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}

	output := buf.String()

	// Long body should be truncated to 80 chars + "..."
	expectedTruncated := strings.Repeat("a", 80) + "..."

	if !strings.Contains(output, `"body":"`+expectedTruncated+`"`) {
		t.Error("long body should be truncated to 80 chars + '...' in req-index")
	}

	// Short body should remain untouched
	if !strings.Contains(output, `"body":"Short body."`) {
		t.Error("short body should remain untruncated in req-index")
	}

	// Check the untruncated body is still in the HTML body (card content)
	if !strings.Contains(output, longBody) {
		t.Error("full long body should appear in card content")
	}
}

func TestHTMLExportTitleFallback(t *testing.T) {
	t.Run("no title key in schema", func(t *testing.T) {
		doc := model.Document{
			Path:   "/some/path",
			Schema: map[string]any{"type": "object"},
			Nodes: []*model.Node{
				{ID: "REQ-001", Body: "Body."},
			},
		}
		var h HTML
		var buf strings.Builder
		err := h.Export(&buf, doc, []string{"status"})
		if err != nil {
			t.Fatalf("HTML export failed: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "<title>/some/path</title>") {
			t.Errorf("expected title to fall back to doc.Path, got: %s", output)
		}
	})

	t.Run("empty title string in schema", func(t *testing.T) {
		doc := model.Document{
			Path:   "/other/path",
			Schema: schemaWithTitle(""),
			Nodes: []*model.Node{
				{ID: "REQ-001", Body: "Body."},
			},
		}
		var h HTML
		var buf strings.Builder
		err := h.Export(&buf, doc, []string{"status"})
		if err != nil {
			t.Fatalf("HTML export failed: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "<title>/other/path</title>") {
			t.Errorf("expected title to fall back to doc.Path when schema title is empty, got: %s", output)
		}
	})

	t.Run("title key present with value", func(t *testing.T) {
		doc := model.Document{
			Path:   "/some/path",
			Schema: schemaWithTitle("My Document"),
			Nodes: []*model.Node{
				{ID: "REQ-001", Body: "Body."},
			},
		}
		var h HTML
		var buf strings.Builder
		err := h.Export(&buf, doc, []string{"status"})
		if err != nil {
			t.Fatalf("HTML export failed: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "<title>My Document</title>") {
			t.Errorf("expected title to be 'My Document', got: %s", output)
		}
	})
}

func TestHTMLExportTraceFiltered(t *testing.T) {
	propOrder := []string{"status", "trace", "asil"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Trace Test"),
		Nodes: []*model.Node{
			{
				ID: "REQ-001",
				Attrs: map[string]any{
					"status": "approved",
					"trace":  []string{"REQ-002"},
					"asil":   "B",
				},
				Body: "Test.",
			},
		},
	}

	var h HTML
	var buf strings.Builder
	err := h.Export(&buf, doc, propOrder)
	if err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}

	output := buf.String()

	// Status and asil should appear in attrs
	if !strings.Contains(output, "<dt>Status</dt>") {
		t.Error("output should contain Status attr label")
	}
	if !strings.Contains(output, "<dt>Asil</dt>") {
		t.Error("output should contain Asil attr label")
	}

	// Trace should NOT appear in the attrs dl
	if strings.Contains(output, "<dt>Trace</dt>") {
		t.Error("output should NOT contain Trace in attrs list")
	}
	if strings.Contains(output, "REQ-002") && strings.Contains(output, "<dd>") {
		// Ensure trace value doesn't appear in any <dd>
		t.Error("output should NOT contain trace values in attrs list")
	}
}

func TestHTMLExportArrayPillsAndNestedObjects(t *testing.T) {
	propOrder := []string{"aspice-bp", "status", "metadata", "flags"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Array Test"),
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

	var h HTML
	var buf strings.Builder
	if err := h.Export(&buf, doc, propOrder); err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}
	output := buf.String()

	// Pills wrapper
	if !strings.Contains(output, `class="attr-pills"`) {
		t.Error("output should contain attr-pills wrapper for arrays")
	}

	// Individual pill spans
	if !strings.Contains(output, `class="attr-pill"`) {
		t.Error("output should contain attr-pill spans")
	}

	// String array values rendered as pills
	if !strings.Contains(output, "SYS2-BP2") {
		t.Error("output should contain SYS2-BP2 pill")
	}
	if !strings.Contains(output, "SWE1-BP2") {
		t.Error("output should contain SWE1-BP2 pill")
	}

	// Boolean and int inside mixed array rendered correctly
	if !strings.Contains(output, "true") {
		t.Error("output should contain 'true' pill from mixed array")
	}
	if !strings.Contains(output, "42") {
		t.Error("output should contain '42' pill from mixed array")
	}

	// null inside array renders null pill
	if !strings.Contains(output, `class="attr-pill-null"`) {
		t.Error("output should contain attr-pill-null for nil array element")
	}

	// Nested object renders as <pre>
	if !strings.Contains(output, `class="attr-nested"`) {
		t.Error("output should contain attr-nested <pre> for map attribute")
	}
	if !strings.Contains(output, `&#34;source&#34;: &#34;legacy&#34;`) {
		t.Error("output should contain escaped JSON content inside attr-nested")
	}

	// Empty array renders em-dash
	if !strings.Contains(output, "<dd>—</dd>") {
		t.Error("output should render em-dash for empty array")
	}
}

func TestHTMLExportDocChain(t *testing.T) {
	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Doc"),
		Nodes: []*model.Node{
			{ID: "REQ-001", Body: "Body."},
		},
	}

	// Confluence-Flow graph: one upstream card, current doc, one downstream card.
	graph := DocChainGraph{
		Upstream: []ChainTier{
			{
				Level: -1,
				Label: "Upstream",
				Cards: []ChainCard{
					{Title: "Upstream Doc", Path: "upstream.html", DirName: "upstream"},
				},
			},
		},
		Current: ChainCard{Title: "Test Doc", DirName: "test", IsCurrent: true},
		Downstream: []ChainTier{
			{
				Level: 1,
				Label: "Downstream",
				Cards: []ChainCard{
					{Title: "Downstream Doc", Path: "downstream.html", DirName: "downstream"},
				},
			},
		},
	}

	var h HTML
	h.SetDocChainGraph(graph)
	var buf strings.Builder
	err := h.Export(&buf, doc, propOrder)
	if err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}

	output := buf.String()

	// Nav wrapper must be present.
	if !strings.Contains(output, `<nav class="doc-chain"`) {
		t.Error("output should contain doc-chain nav element")
	}
	// Folding wrapper: shallow chain (3 tiers) is open by default.
	if !strings.Contains(output, `<details class="chain-details" open>`) {
		t.Error("output should contain open chain-details element for shallow chain")
	}
	// Summary header.
	if !strings.Contains(output, `chain-summary`) {
		t.Error("output should contain chain-summary header")
	}
	if !strings.Contains(output, `Traceability Chain`) {
		t.Error("output should contain 'Traceability Chain' summary label")
	}
	if !strings.Contains(output, `1 upstream · 1 downstream`) {
		t.Error("output should contain upstream/downstream count in summary")
	}
	// Stack wrapper is rendered.
	if !strings.Contains(output, `class="chain-stack"`) {
		t.Error("output should contain chain-stack wrapper")
	}
	// Upstream zone is rendered with the right label.
	if !strings.Contains(output, `chain-zone--upstream`) {
		t.Error("output should contain chain-zone--upstream")
	}
	// Current zone is rendered.
	if !strings.Contains(output, `chain-zone--current`) {
		t.Error("output should contain chain-zone--current")
	}
	// Downstream zone is rendered.
	if !strings.Contains(output, `chain-zone--downstream`) {
		t.Error("output should contain chain-zone--downstream")
	}
	// Inactive cards are links.
	if !strings.Contains(output, `href="upstream.html"`) {
		t.Error("output should contain href for upstream.html")
	}
	if !strings.Contains(output, `href="downstream.html"`) {
		t.Error("output should contain href for downstream.html")
	}
	if !strings.Contains(output, ">Upstream Doc<") {
		t.Error("output should contain 'Upstream Doc' link text")
	}
	if !strings.Contains(output, ">Downstream Doc<") {
		t.Error("output should contain 'Downstream Doc' link text")
	}
	// The current card is a non-link span with chain-card--current class and aria-current.
	if !strings.Contains(output, `class="chain-card chain-card--current"`) {
		t.Error("output should contain chain-card--current class on the current card")
	}
	if !strings.Contains(output, `aria-current="page"`) {
		t.Error("output should contain aria-current=page on the current card")
	}
	if !strings.Contains(output, ">Test Doc</span>") {
		t.Error("output should contain 'Test Doc' text on the current card")
	}
	// Zone separators between upstream/current and current/downstream.
	if !strings.Contains(output, `chain-separator`) {
		t.Error("output should contain chain-separator between zones")
	}
	if !strings.Contains(output, `chain-separator-chevron`) {
		t.Error("output should contain chain-separator-chevron (▾)")
	}
}

func TestHTMLExportDocChainExternal(t *testing.T) {
	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Doc"),
		Nodes: []*model.Node{
			{ID: "REQ-001", Body: "Body."},
		},
	}

	graph := DocChainGraph{
		Upstream: []ChainTier{
			{
				Level: -1,
				Label: "Upstream",
				Cards: []ChainCard{
					{Title: "External Spec", Path: "external.html", DirName: "external", IsExternal: true},
				},
			},
		},
		Current: ChainCard{Title: "Test Doc", DirName: "test", IsCurrent: true},
	}

	var h HTML
	h.SetDocChainGraph(graph)
	var buf strings.Builder
	if err := h.Export(&buf, doc, propOrder); err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}

	output := buf.String()

	// External card should be flagged with .chain-card--external.
	if !strings.Contains(output, `chain-card--external`) {
		t.Error("output should contain chain-card--external class for external docs")
	}
	// ↗ glyph (&#8599;) should appear next to the external card title.
	if !strings.Contains(output, `&#8599;`) {
		t.Error("output should contain the external ↗ glyph")
	}
	// Summary header should still appear even with only upstream (no downstream).
	if !strings.Contains(output, `chain-summary`) {
		t.Error("output should contain chain-summary header")
	}
	if !strings.Contains(output, `1 upstream`) {
		t.Error("output should contain upstream count")
	}
	// Summary count should not contain the downstream separator when none exist.
	if strings.Contains(output, `upstream ·`) {
		t.Error("output should NOT contain upstream· separator when no downstream")
	}
}

func TestHTMLExportDocChainDeep(t *testing.T) {
	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Doc"),
		Nodes: []*model.Node{
			{ID: "REQ-001", Body: "Body."},
		},
	}

	// Build a deep chain: 4 upstream tiers + 3 downstream tiers (>6).
	upstream := make([]ChainTier, 4)
	for i := range 4 {
		upstream[i] = ChainTier{
			Level: -(i + 1),
			Label: fmt.Sprintf("Upstream %d", i+1),
			Cards: []ChainCard{{Title: fmt.Sprintf("U%d", i+1), Path: fmt.Sprintf("u%d.html", i+1), DirName: fmt.Sprintf("u%d", i+1)}},
		}
	}
	downstream := make([]ChainTier, 3)
	for i := range 3 {
		downstream[i] = ChainTier{
			Level: i + 1,
			Label: fmt.Sprintf("Downstream %d", i+1),
			Cards: []ChainCard{{Title: fmt.Sprintf("D%d", i+1), Path: fmt.Sprintf("d%d.html", i+1), DirName: fmt.Sprintf("d%d", i+1)}},
		}
	}
	graph := DocChainGraph{
		Upstream:   upstream,
		Current:    ChainCard{Title: "Test Doc", DirName: "test", IsCurrent: true},
		Downstream: downstream,
	}

	var h HTML
	h.SetDocChainGraph(graph)
	var buf strings.Builder
	if err := h.Export(&buf, doc, propOrder); err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}
	output := buf.String()

	// Deep chain class on nav.
	if !strings.Contains(output, `doc-chain doc-chain--deep`) {
		t.Error("deep chain should add doc-chain--deep class")
	}
	// Details should NOT have open attribute (collapsed by default).
	if strings.Contains(output, `<details class="chain-details" open>`) {
		t.Error("deep chain details should NOT have open attribute by default")
	}
	// Deep chain compacts cards via the doc-chain--deep modifier class
	// (already asserted above) — the markup itself stays identical, so we
	// simply verify all upstream tier labels are present in the output.
	for i := 1; i <= 4; i++ {
		label := fmt.Sprintf(">Upstream %d<", i)
		if !strings.Contains(output, label) {
			t.Errorf("deep chain should render upstream tier label %q", label)
		}
	}
	// Summary count.
	if !strings.Contains(output, `upstream`) {
		t.Error("summary should contain upstream count")
	}
}

func TestHTMLExportDocChainCrowded(t *testing.T) {
	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Doc"),
		Nodes: []*model.Node{
			{ID: "REQ-001", Body: "Body."},
		},
	}

	cards := make([]ChainCard, 9)
	for i := range 9 {
		cards[i] = ChainCard{Title: fmt.Sprintf("Spec %d", i+1), Path: fmt.Sprintf("s%d.html", i+1), DirName: fmt.Sprintf("s%d", i+1)}
	}
	graph := DocChainGraph{
		Upstream: []ChainTier{{Level: -1, Label: "Upstream", Cards: cards}},
		Current:  ChainCard{Title: "Test Doc", DirName: "test", IsCurrent: true},
	}

	var h HTML
	h.SetDocChainGraph(graph)
	var buf strings.Builder
	if err := h.Export(&buf, doc, propOrder); err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}
	output := buf.String()

	// All 9 cards should be rendered as upstream cards (one per directory).
	for i := 1; i <= 9; i++ {
		title := fmt.Sprintf("Spec %d", i)
		if !strings.Contains(output, title) {
			t.Errorf("crowded tier should render upstream card %q", title)
		}
	}
	if !strings.Contains(output, `9 upstream`) {
		t.Error("summary should show 9 upstream")
	}
}

func TestHTMLExportNoDocLinks(t *testing.T) {
	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test"),
		Nodes: []*model.Node{
			{ID: "REQ-001", Body: "Body."},
		},
	}

	var h HTML
	var buf strings.Builder
	err := h.Export(&buf, doc, propOrder)
	if err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}

	output := buf.String()

	// When no graph was configured (zero value), the chain must not render.
	if strings.Contains(output, `<nav class="doc-chain"`) {
		t.Error("output should NOT contain doc-chain nav when no graph set")
	}
	if strings.Contains(output, `class="chain-stack"`) {
		t.Error("output should NOT contain chain-stack when no graph set")
	}
}

func TestCSVExportItems(t *testing.T) {
	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Export"),
		Nodes: []*model.Node{
			{Kind: model.KindContainer, Title: "Section One", Level: 1, Body: "Container prose."},
			{ID: "REQ-001", Attrs: map[string]any{"status": "approved"}},
			{Kind: model.KindInfo, Title: "Note", Level: 2, Body: "Some note."},
		},
	}

	var buf strings.Builder
	if err := (&CSV{}).Export(&buf, doc, propOrder); err != nil {
		t.Fatalf("CSV export failed: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (header + 3 nodes), got %d", len(lines))
	}
	wantRows := []string{
		"container,,Section One,,Container prose.,",
		"req,REQ-001,,approved,,",
		"info,,Note,,Some note.,",
	}
	for i, want := range wantRows {
		if lines[i+1] != want {
			t.Errorf("row %d = %q, want %q", i+1, lines[i+1], want)
		}
	}
}

func TestHTMLExportItems(t *testing.T) {
	propOrder := []string{"status"}
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

	var h HTML
	var buf strings.Builder
	if err := h.Export(&buf, doc, propOrder); err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}
	out := buf.String()

	// Container renders as a collapsible section with its title + body.
	if !strings.Contains(out, `<section class="container-item" id="info-1">`) {
		t.Error("container section missing")
	}
	if !strings.Contains(out, "Section One") {
		t.Error("container title missing")
	}
	if !strings.Contains(out, "Container prose.") {
		t.Error("container body missing")
	}
	// The nested requirement renders as a child card.
	if !strings.Contains(out, `<article class="req-card-child" id="REQ-001"`) {
		t.Error("child card missing")
	}
	// Info item renders as a plain block with its own anchor.
	if !strings.Contains(out, `<section class="info-item" id="info-2">`) {
		t.Error("info section missing")
	}
	if !strings.Contains(out, "Some note.") {
		t.Error("info body missing")
	}
	// Doc count mentions items.
	if !strings.Contains(out, "1 requirement · 2 items") {
		t.Error("doc count line missing")
	}
	// Index carries the item entries with type markers.
	if !strings.Contains(out, `"id":"info-1","type":"container"`) {
		t.Error("index missing container entry")
	}
}

func TestCSVExportVerdicts(t *testing.T) {
	propOrder := []string{"status"}
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Export"),
		Nodes: []*model.Node{
			{ID: "REQ-001", Attrs: map[string]any{"status": "approved"}, Body: "B"},
		},
	}

	var buf strings.Builder
	var c CSV
	c.SetVerdicts(map[string]VerdictInfo{
		"REQ-001": {
			Outcome: "fail",
			Source:  "ci/run.ctrf.json",
			Cases:   []string{"BOOT-TIME", "SUSPEND-OK"},
		},
	})
	if err := c.Export(&buf, doc, propOrder); err != nil {
		t.Fatalf("CSV export: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(lines))
	}
	header := lines[0]
	if !strings.Contains(header, "Verdict,Verdict Source,Verdict Cases") {
		t.Errorf("header missing verdict columns: %q", header)
	}
	row := lines[1]
	if !strings.Contains(row, "fail,ci/run.ctrf.json,BOOT-TIME; SUSPEND-OK") {
		t.Errorf("row missing verdict cells: %q", row)
	}
}

func TestHTMLExportEvidenceList(t *testing.T) {
	doc := model.Document{
		Path:   "/test/doc",
		Schema: schemaWithTitle("Test Export"),
		Nodes: []*model.Node{
			{
				ID:    "REQ-001",
				Attrs: map[string]any{"status": "approved"},
				Body:  "The system shall boot.",
			},
		},
	}

	var h HTML
	h.SetVerdicts(map[string]VerdictInfo{
		"REQ-001": {
			Outcome: "pass",
			Source:  "ci/run.ctrf.json",
			Evidence: []EvidenceItem{
				{
					Case:        "BOOT-TIME",
					Outcome:     "pass",
					File:        "ci/run.ctrf.json",
					Description: "Measures boot time.\n\n```go\nboot()\n```",
				},
			},
		},
	})
	var buf strings.Builder
	if err := h.Export(&buf, doc, nil); err != nil {
		t.Fatalf("HTML export: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "verdict-evidence") {
		t.Error("output missing verdict-evidence block")
	}
	if !strings.Contains(output, "Evidence (1)") {
		t.Error("output missing evidence summary count")
	}
	if !strings.Contains(output, "BOOT-TIME") {
		t.Error("output missing evidence case key")
	}
	if !strings.Contains(output, "evidence-desc") {
		t.Error("output missing rendered evidence description")
	}
}
