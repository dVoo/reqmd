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
	"strings"

	"github.com/dVoo/reqmd/internal/model"

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
// line after the <style> / <script> open tag.
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

// EvidenceItem is one verification result contributing to a measure's
// rolled-up evidence set, ready for rendering in exports.
type EvidenceItem struct {
	Case        string
	Outcome     string
	File        string
	Description string
}

// VerdictInfo carries the rolled-up verification verdict for a single
// measure: the outcome, a representative source file, the case keys that
// contributed (for CSV), and the full evidence list (for HTML).
type VerdictInfo struct {
	Outcome  string // pass | fail | skipped | inconclusive
	Source   string // CTRF file path or manual-results markdown path
	Cases    []string
	Evidence []EvidenceItem
}

// HTML exports requirements as standalone HTML: a server-rendered Document
// view plus on-demand analytical views driven by one shared view model.
type HTML struct {
	analysis      *Analysis
	resolveLink   func(string) string
	docChainGraph DocChainGraph
	boundary      DocBoundary
}

// SetAnalysis supplies the graph-derived coverage, verification, and link
// state every view renders. A nil analysis renders requirements without
// semantic status (still valid output, just less information).
func (h *HTML) SetAnalysis(a *Analysis) {
	h.analysis = a
}

// SetResolveLink supplies the cross-document href resolver. A nil resolver
// produces same-page fragments.
func (h *HTML) SetResolveLink(fn func(string) string) {
	h.resolveLink = fn
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
// document-chain navigation. Pass a zero-value graph to disable the chain.
func (h *HTML) SetDocChainGraph(graph DocChainGraph) {
	h.docChainGraph = graph
}

// SetBoundary configures the document-level boundary flags (root/leaf
// position in the V-model chain) used by the header badges.
func (h *HTML) SetBoundary(b DocBoundary) {
	h.boundary = b
}

// pageBufferSize is the buffered-writer size for one exported page.
const pageBufferSize = 32 * 1024

// Export writes requirements as a standalone HTML page. The Document view is
// server-rendered and fully visible without JavaScript; Table, Trace, and
// Coverage render on demand from the embedded view model.
func (h *HTML) Export(w io.Writer, doc model.Document, propOrder []string) error {
	title := docTitle(doc)
	if title == "" {
		title = doc.Path
	}

	rd := buildRenderData(doc)
	pageData := buildPageModel(doc, rd, h.analysis, h.resolveLink, propOrder)

	b := bufio.NewWriterSize(w, pageBufferSize)

	renderDocHead(b, title)
	renderBodyOpen(b, title)
	renderTOC(b, doc, rd)
	_, _ = b.WriteString(`<div class="shell__main">` + "\n")
	renderDocHeader(b, doc, title, h.boundary.IsRoot, h.boundary.IsLeaf, rd)
	renderToolbar(b, doc, rd)
	_, _ = b.WriteString("<main class=\"views\">\n")
	renderDocumentView(b, doc, rd, propOrder, h.analysis, h.resolveLink, h.docChainGraph)
	// Analytical views are empty shells filled on demand by the client.
	_, _ = b.WriteString(`<section class="view" data-view-panel="table" hidden></section>` + "\n")
	_, _ = b.WriteString(`<section class="view" data-view-panel="trace" hidden></section>` + "\n")
	_, _ = b.WriteString(`<section class="view" data-view-panel="coverage" hidden></section>` + "\n")
	_, _ = b.WriteString("</main>\n")
	_, _ = b.WriteString("</div>\n</div>\n")

	renderPalette(b)
	renderShortcuts(b)

	renderModelJSON(b, pageData)
	renderScript(b)
	renderBodyClose(b)

	if err := b.Flush(); err != nil {
		return fmt.Errorf("flushing HTML output: %w", err)
	}
	return nil
}

// renderModelJSON writes the shared view model the client renders from.
func renderModelJSON(b *bufio.Writer, m pageModel) {
	data, _ := json.Marshal(m)
	_, _ = fmt.Fprintf(b, "<script type=\"application/json\" id=\"req-model\">%s</script>\n", data)
}

func renderScript(b *bufio.Writer) {
	_, _ = b.WriteString("<script>\n")
	_, _ = b.WriteString(js)
	_, _ = b.WriteString("</script>\n")
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
		var sb strings.Builder
		sb.WriteString(`<span class="attr__pills">`)
		for i, item := range val {
			if i > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(`<span class="attr__pill">`)
			switch it := item.(type) {
			case string:
				sb.WriteString(html.EscapeString(it))
			case nil:
				sb.WriteString(`<em>null</em>`)
			default:
				sb.WriteString(html.EscapeString(formatAttrHTML(it)))
			}
			sb.WriteString(`</span>`)
		}
		sb.WriteString(`</span>`)
		return sb.String()
	case []string:
		if len(val) == 0 {
			return "—"
		}
		var sb strings.Builder
		sb.WriteString(`<span class="attr__pills">`)
		for i, s := range val {
			if i > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(`<span class="attr__pill">`)
			sb.WriteString(html.EscapeString(s))
			sb.WriteString(`</span>`)
		}
		sb.WriteString(`</span>`)
		return sb.String()
	case map[string]any:
		b, _ := json.MarshalIndent(val, "", "  ")
		return fmt.Sprintf(`<pre class="attr__nested">%s</pre>`, html.EscapeString(string(b)))
	default:
		return ""
	}
}

// renderChainGraph writes the Document Stack chain as HTML to b.
// The output is pure HTML/CSS — no JavaScript.
func renderChainGraph(b *bufio.Writer, g DocChainGraph) {
	_, _ = b.WriteString("  <div class=\"chain-stack\">\n")

	if len(g.Upstream) > 0 {
		renderChainZone(b, "chain-zone--upstream", "Upstream Sources", g.Upstream)
		renderChainSeparator(b)
	}

	_, _ = b.WriteString("    <div class=\"chain-zone chain-zone--current\">\n")
	renderChainCard(b, g.Current)
	_, _ = b.WriteString("    </div>\n")

	if len(g.Downstream) > 0 {
		renderChainSeparator(b)
		renderChainZone(b, "chain-zone--downstream", "Downstream Consumers", g.Downstream)
	}

	_, _ = b.WriteString("  </div>\n")
}

// renderChainZone writes one horizontal zone (upstream or downstream)
// with its header and stacked tiers.
func renderChainZone(b *bufio.Writer, zoneClass, header string, tiers []ChainTier) {
	_, _ = fmt.Fprintf(b, "    <div class=\"chain-zone %s\">\n"+
		"      <div class=\"chain-zone-header\">%s</div>\n",
		zoneClass, header)
	for _, tier := range tiers {
		_, _ = fmt.Fprintf(b, "      <div class=\"chain-tier\">\n"+
			"        <div class=\"chain-tier-label\">%s</div>\n"+
			"        <div class=\"chain-tier-cards\">\n",
			html.EscapeString(tier.Label))
		for _, card := range tier.Cards {
			renderChainCard(b, card)
		}
		_, _ = b.WriteString("        </div>\n      </div>\n")
	}
	_, _ = b.WriteString("    </div>\n")
}

// renderChainSeparator writes the chevron rule between chain zones.
func renderChainSeparator(b *bufio.Writer) {
	_, _ = b.WriteString("    <div class=\"chain-separator\" aria-hidden=\"true\">\n" +
		"      <div class=\"chain-separator-line\"></div>\n" +
		"      <div class=\"chain-separator-chevron\">&#9662;</div>\n" +
		"      <div class=\"chain-separator-line\"></div>\n" +
		"    </div>\n")
}

// renderChainCard writes one <a> (or <span> for the current doc) card.
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
		_, _ = b.WriteString("      <span class=\"")
		_, _ = b.WriteString(strings.Join(classes, " "))
		_, _ = b.WriteString("\" aria-current=\"page\">\n")
		if card.IsExternal {
			_, _ = b.WriteString("        <span class=\"chain-card-glyph\" aria-hidden=\"true\">&#8599;</span>\n")
		}
		_, _ = b.WriteString("        <span class=\"chain-card-title\">")
		_, _ = b.WriteString(escapedDisplay)
		_, _ = b.WriteString("</span>\n")
		_, _ = b.WriteString("        <span class=\"chain-card-meta\">")
		_, _ = b.WriteString(escapedDir)
		_, _ = b.WriteString("</span>\n")
		_, _ = b.WriteString("      </span>\n")
		return
	}

	_, _ = b.WriteString("      <a class=\"")
	_, _ = b.WriteString(strings.Join(classes, " "))
	_, _ = b.WriteString("\" href=\"")
	_, _ = b.WriteString(escapedHref)
	_, _ = b.WriteString("\">\n")
	if card.IsExternal {
		_, _ = b.WriteString("        <span class=\"chain-card-glyph\" aria-hidden=\"true\">&#8599;</span>\n")
	}
	_, _ = b.WriteString("        <span class=\"chain-card-title\">")
	_, _ = b.WriteString(escapedDisplay)
	_, _ = b.WriteString("</span>\n")
	_, _ = b.WriteString("        <span class=\"chain-card-meta\">")
	_, _ = b.WriteString(escapedDir)
	_, _ = b.WriteString("</span>\n")
	_, _ = b.WriteString("      </a>\n")
}
