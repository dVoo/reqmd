package exporter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/internal/model"
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
	if !strings.Contains(output, `<nav class="doc-chain`) {
		t.Error("output should contain doc-chain nav element")
	}
	// The chain is orientation only: always collapsed, so it can never push
	// the reader past the toolbar.
	if strings.Contains(output, `<details class="chain-details" open>`) {
		t.Error("chain details must always start collapsed")
	}
	if !strings.Contains(output, `<details class="chain-details">`) {
		t.Error("output should contain the chain-details folding wrapper")
	}
	// Summary header.
	if !strings.Contains(output, `chain-summary`) {
		t.Error("output should contain chain-summary header")
	}
	if !strings.Contains(output, `Traceability chain`) {
		t.Error("output should contain 'Traceability chain' summary label")
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
