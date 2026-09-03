package exporter

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/fs"
	"reqmd/internal/model"
	"strings"

	katex "github.com/FurqanSoftware/goldmark-katex"
	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"go.abhg.dev/goldmark/mermaid"
)

//go:embed static
var staticFS embed.FS

// css and js are the embedded stylesheet and client script for the
// exported HTML. Each is read once at init and reused for every page.
// A leading "\n" is prepended so the inlined block sits on its own
// line after the <style> / <script> open tag, matching the previous
// byte-for-byte output.
var (
	css = mustReadStatic("static/style.css")
	js  = mustReadStatic("static/app.js")
)

func mustReadStatic(name string) string {
	b, err := fs.ReadFile(staticFS, name)
	if err != nil {
		panic(err)
	}
	return "\n" + string(b)
}

// VerdictInfo carries the verification result for a single measure.
type VerdictInfo struct {
	Outcome string // pass | fail | skipped | inconclusive
	Source  string // CTRF file path or manual-results markdown path
}

// HTML exports requirements as standalone HTML with card-based layout.
type HTML struct {
	verdicts      map[string]VerdictInfo
	docChainGraph DocChainGraph
	boundary      DocBoundary
}

// SetVerdicts stores the verification verdicts for rendering badges on
// measure cards. Called by the CLI when --results is supplied.
func (h *HTML) SetVerdicts(v map[string]VerdictInfo) {
	h.verdicts = v
}

// schemaTitle safely extracts the title field from a document's parsed
// schema. Returns "" when Schema is nil, not a map, or has no string title.
func schemaTitle(doc model.Document) string {
	m, ok := doc.Schema.(map[string]any)
	if !ok {
		return ""
	}
	s, _ := m["title"].(string)
	return s
}

// docTitle returns the title used for rendering a document page or card:
// the parsed file h1 (document title), falling back to the schema title.
func docTitle(doc model.Document) string {
	if doc.Title != "" {
		return doc.Title
	}
	return schemaTitle(doc)
}

// bodyRenderer renders requirement body markdown to HTML.
var bodyRenderer = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		&mermaid.Extender{},
		&fences.Extender{},
		highlighting.Highlighting,
		emoji.Emoji,
		&katex.Extender{},
	),
)

// renderMarkdown renders markdown content to safe HTML.
func renderMarkdown(content string) string {
	var buf bytes.Buffer
	if err := bodyRenderer.Convert([]byte(content), &buf); err != nil {
		return html.EscapeString(content)
	}
	return buf.String()
}

// SetDocChainGraph configures the tiered document graph used by the
// Confluence-Flow chain visualization. The graph holds the upstream tiers
// (above the current doc), the current doc card, and the downstream tiers
// (below the current doc). Pass a zero-value graph to disable the chain.
func (h *HTML) SetDocChainGraph(graph DocChainGraph) {
	h.docChainGraph = graph
}

// SetBoundary configures the document-level boundary flags (root/leaf
// position in the V-model chain) used by the header badges.
func (h *HTML) SetBoundary(b DocBoundary) {
	h.boundary = b
}

// Export writes requirements as a standalone HTML page.
func (h *HTML) Export(w io.Writer, doc model.Document, propOrder []string) error {
	return h.exportHTML(w, doc, propOrder, nil)
}

// ExportWithTraces writes HTML with upstream/downstream trace links.
// tc bundles the four callback hooks: Upstream / Downstream return
// requirement IDs reachable from the given ID; ResolveLink maps a
// requirement ID to a cross-file HTML anchor; TitleOf maps a
// requirement ID to its human-readable title. Nil fields are tolerated —
// the trace section is simply omitted for that side.
func (h *HTML) ExportWithTraces(w io.Writer, doc model.Document, propOrder []string, tc TraceResolver) error {
	return h.exportHTML(w, doc, propOrder, &tc)
}

func (h *HTML) exportHTML(w io.Writer, doc model.Document, propOrder []string, tc *TraceResolver) error {
	title := docTitle(doc)
	if title == "" {
		title = doc.Path
	}

	rd := buildRenderData(doc)
	traceCache := buildTraceCache(doc, tc)
	ignoreStatus := doc.XReqmd != nil && doc.XReqmd.IgnoreStatus

	b := bufio.NewWriterSize(w, 32*1024) // 32KB buffer

	renderDocHead(b, title)
	renderBodyOpen(b)
	renderDocChain(b, h.docChainGraph)
	renderDocHeader(b, doc, title, h.boundary.IsRoot, h.boundary.IsLeaf)
	renderToolbar(b, doc, ignoreStatus)

	_, _ = b.WriteString("<main>\n")
	for _, node := range rd.Nodes {
		renderContentNode(b, node, 0, rd, propOrder, traceCache, tc, ignoreStatus, h.verdicts)
	}
	_, _ = b.WriteString("</main>\n")

	renderScripts(b, rd)
	renderBodyClose(b)

	if err := b.Flush(); err != nil {
		return fmt.Errorf("flushing HTML output: %w", err)
	}
	return nil
}

// formatAttrLabel converts a snake_case or kebab-case attribute name to
// a human-readable title-case label (e.g. "disposition-reason" → "Disposition Reason").
func formatAttrLabel(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// formatAttrHTML renders an attribute value as rich HTML for the card grid.
// Arrays become inline pill tags. Nested maps become a compact <pre> JSON block.
// All output is already HTML-escaped where needed by the caller.
func formatAttrHTML(v any) string {
	switch val := v.(type) {
	case string:
		return html.EscapeString(val)
	case bool, int, int64, float64:
		return model.FormatScalar(v)
	case []any:
		if len(val) == 0 {
			return "—"
		}
		var b strings.Builder
		b.WriteString(`<span class="attr-pills">`)
		for i, item := range val {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(`<span class="attr-pill">`)
			switch it := item.(type) {
			case string:
				b.WriteString(html.EscapeString(it))
			case nil:
				b.WriteString(`<em class="attr-pill-null">null</em>`)
			default:
				b.WriteString(html.EscapeString(formatAttrHTML(it)))
			}
			b.WriteString(`</span>`)
		}
		b.WriteString(`</span>`)
		return b.String()
	case []string:
		if len(val) == 0 {
			return "—"
		}
		var b strings.Builder
		b.WriteString(`<span class="attr-pills">`)
		for i, s := range val {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(`<span class="attr-pill">`)
			b.WriteString(html.EscapeString(s))
			b.WriteString(`</span>`)
		}
		b.WriteString(`</span>`)
		return b.String()
	case map[string]any:
		b, _ := json.MarshalIndent(val, "", "  ")
		return fmt.Sprintf(`<pre class="attr-nested">%s</pre>`, html.EscapeString(string(b)))
	default:
		return ""
	}
}

// renderChainGraph writes the Document Stack chain as HTML to b.
// The output is pure HTML/CSS — no JavaScript.
//
// Layout: three vertically stacked zones (upstream / current / downstream)
// with short centered separator rules (chevron + lines) between the major
// zones. There are no per-card connectors, no merge/fork lines, and no
// sibling separators inside a zone.
func renderChainGraph(b *bufio.Writer, g DocChainGraph) {
	_, _ = b.WriteString("  <div class=\"chain-stack\">\n")

	// ── Upstream zone ──
	if len(g.Upstream) > 0 {
		b.WriteString("    <div class=\"chain-zone chain-zone--upstream\">\n")
		b.WriteString("      <div class=\"chain-zone-header\">Upstream Sources</div>\n")
		for _, tier := range g.Upstream {
			b.WriteString("      <div class=\"chain-tier\">\n")
			b.WriteString("        <div class=\"chain-tier-label\">")
			b.WriteString(html.EscapeString(tier.Label))
			b.WriteString("</div>\n")
			b.WriteString("        <div class=\"chain-tier-cards\">\n")
			for _, card := range tier.Cards {
				renderChainCard(b, card)
			}
			b.WriteString("        </div>\n")
			b.WriteString("      </div>\n")
		}
		b.WriteString("    </div>\n")

		b.WriteString("    <div class=\"chain-separator\" aria-hidden=\"true\">\n")
		b.WriteString("      <div class=\"chain-separator-line\"></div>\n")
		b.WriteString("      <div class=\"chain-separator-chevron\">&#9662;</div>\n")
		b.WriteString("      <div class=\"chain-separator-line\"></div>\n")
		b.WriteString("    </div>\n")
	}

	// ── Current zone ──
	b.WriteString("    <div class=\"chain-zone chain-zone--current\">\n")
	renderChainCard(b, g.Current)
	b.WriteString("    </div>\n")

	// ── Downstream zone ──
	if len(g.Downstream) > 0 {
		b.WriteString("    <div class=\"chain-separator\" aria-hidden=\"true\">\n")
		b.WriteString("      <div class=\"chain-separator-line\"></div>\n")
		b.WriteString("      <div class=\"chain-separator-chevron\">&#9662;</div>\n")
		b.WriteString("      <div class=\"chain-separator-line\"></div>\n")
		b.WriteString("    </div>\n")

		b.WriteString("    <div class=\"chain-zone chain-zone--downstream\">\n")
		b.WriteString("      <div class=\"chain-zone-header\">Downstream Consumers</div>\n")
		for _, tier := range g.Downstream {
			b.WriteString("      <div class=\"chain-tier\">\n")
			b.WriteString("        <div class=\"chain-tier-label\">")
			b.WriteString(html.EscapeString(tier.Label))
			b.WriteString("</div>\n")
			b.WriteString("        <div class=\"chain-tier-cards\">\n")
			for _, card := range tier.Cards {
				renderChainCard(b, card)
			}
			b.WriteString("        </div>\n")
			b.WriteString("      </div>\n")
		}
		b.WriteString("    </div>\n")
	}

	b.WriteString("  </div>\n")
}

// renderChainCard writes one <a> (or <span> for the current doc) card.
// The current document gets the chain-card--current class; external
// documents get the chain-card--external class.
func renderChainCard(b *bufio.Writer, card ChainCard) {
	classes := []string{"chain-card"}
	if card.IsCurrent {
		classes = append(classes, "chain-card--current")
	}
	if card.IsExternal {
		classes = append(classes, "chain-card--external")
	}
	display := card.Title
	if display == "" {
		display = card.DirName
	}
	escapedDisplay := html.EscapeString(display)
	escapedDir := html.EscapeString(card.DirName)
	escapedHref := html.EscapeString(card.Path)

	if card.IsCurrent || card.Path == "" {
		b.WriteString("      <span class=\"")
		b.WriteString(strings.Join(classes, " "))
		b.WriteString("\" aria-current=\"page\">\n")
		if card.IsExternal {
			b.WriteString("        <span class=\"chain-card-glyph\" aria-hidden=\"true\">&#8599;</span>\n")
		}
		b.WriteString("        <span class=\"chain-card-title\">")
		b.WriteString(escapedDisplay)
		b.WriteString("</span>\n")
		b.WriteString("        <span class=\"chain-card-meta\">")
		b.WriteString(escapedDir)
		b.WriteString("</span>\n")
		b.WriteString("      </span>\n")
		return
	}

	b.WriteString("      <a class=\"")
	b.WriteString(strings.Join(classes, " "))
	b.WriteString("\" href=\"")
	b.WriteString(escapedHref)
	b.WriteString("\">\n")
	if card.IsExternal {
		b.WriteString("        <span class=\"chain-card-glyph\" aria-hidden=\"true\">&#8599;</span>\n")
	}
	b.WriteString("        <span class=\"chain-card-title\">")
	b.WriteString(escapedDisplay)
	b.WriteString("</span>\n")
	b.WriteString("        <span class=\"chain-card-meta\">")
	b.WriteString(escapedDir)
	b.WriteString("</span>\n")
	b.WriteString("      </a>\n")
}
