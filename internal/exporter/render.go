package exporter

import (
	"bufio"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"

	"github.com/dVoo/reqmd/internal/graph"
	"github.com/dVoo/reqmd/internal/model"
)

// The HTML export has one server-rendered reading view and three analytical
// views rendered on demand from the embedded page model:
//
//	document — quiet card-per-requirement reading view (works without JS)
//	table    — dense sortable grid with a column chooser and row expansion
//	trace    — selected-requirement trace explorer with a bounded matrix
//	coverage — actionable coverage, verification, and link-health lists
//
// Every view, count, aggregate, and outline reads the same filtered
// requirement set, so the toolbar means one thing.

// normalizeAttrValue collapses []any / []string / scalar attribute
// values into a uniform []string, dropping nil entries and nested maps.
func normalizeAttrValue(v any) []string {
	switch val := v.(type) {
	case []any:
		if val == nil {
			return nil
		}
		out := make([]string, 0, len(val))
		for _, item := range val {
			if item == nil {
				continue
			}
			if _, isMap := item.(map[string]any); isMap {
				continue
			}
			out = append(out, fmt.Sprintf("%v", item))
		}
		return out
	case []string:
		if val == nil {
			return nil
		}
		out := make([]string, len(val))
		copy(out, val)
		return out
	default:
		if val == nil {
			return nil
		}
		if _, isMap := val.(map[string]any); isMap {
			return nil
		}
		return []string{fmt.Sprintf("%v", val)}
	}
}

// buildFilters collects per-attribute unique values across all requirements
// into a filterSet. Attributes with fewer than two distinct values are omitted
// (single-value filters are not useful).
func buildFilters(reqs []*model.Node, doc model.Document) filterSet {
	raw := collectFilterValues(reqs, doc)
	filters := make(filterSet, len(raw))
	for attr, set := range raw {
		if len(set) < 2 {
			continue
		}
		filters[attr] = orderFilterValues(attr, set, doc)
	}
	return filters
}

// collectFilterValues gathers the distinct values of every filterable
// attribute. The trace attribute never becomes a filter; status values are
// skipped when the document opts out of the status lifecycle.
func collectFilterValues(reqs []*model.Node, doc model.Document) map[string]map[string]struct{} {
	ignoreStatus := doc.XReqmd != nil && doc.XReqmd.IgnoreStatus
	raw := make(map[string]map[string]struct{})
	for _, req := range reqs {
		for attr, v := range req.Attrs {
			if attr == model.AttrTrace {
				continue
			}
			if attr == model.AttrStatus && ignoreStatus {
				continue
			}
			if raw[attr] == nil {
				raw[attr] = make(map[string]struct{})
			}
			for _, value := range normalizeAttrValue(v) {
				raw[attr][value] = struct{}{}
			}
		}
	}
	return raw
}

// orderFilterValues returns one attribute's values in display order: status is
// canonical (draft, approved, declared additions, then A→Z); everything else
// is sorted A→Z.
func orderFilterValues(attr string, set map[string]struct{}, doc model.Document) []string {
	vals := make([]string, 0, len(set))
	for v := range set {
		vals = append(vals, v)
	}
	if attr == model.AttrStatus {
		return orderStatusFilterValues(vals, doc)
	}
	sort.Strings(vals)
	return vals
}

// sortedStatusValues returns the given status values in canonical order:
// built-ins [draft, approved] first, then declared additions A→Z, then any
// unrecognized values A→Z.
func sortedStatusValues(values []string, additions []string) []string {
	priority := []string{model.StatusDraft, model.StatusApproved}
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, v := range priority {
		if _, ok := set[v]; ok {
			out = append(out, v)
			seen[v] = true
		}
	}
	for _, v := range additions {
		if _, ok := set[v]; ok && !seen[v] {
			out = append(out, v)
			seen[v] = true
		}
	}
	var rest []string
	for _, v := range values {
		if !seen[v] {
			rest = append(rest, v)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}

// orderStatusFilterValues pulls the declared additions from doc.XReqmd
// (nil-safe) and returns the values in canonical order.
func orderStatusFilterValues(vals []string, doc model.Document) []string {
	var additions []string
	if doc.XReqmd != nil {
		additions = doc.XReqmd.AdditionalStatusValues
	}
	return sortedStatusValues(vals, additions)
}

// statusBreakdownItem describes one status value and its count.
type statusBreakdownItem struct {
	Value string
	Count int
}

// countStatusBreakdown tallies status values across all requirements in a
// document, in display order: built-in (draft, approved) first, then
// additions A→Z. Values not declared are treated as StatusDefault.
func countStatusBreakdown(reqs []*model.Node, additions []string) []statusBreakdownItem {
	counts := make(map[string]int)
	for _, req := range reqs {
		v, ok := req.Attrs[model.AttrStatus].(string)
		if !ok || v == "" {
			v = model.StatusDefault
		}
		counts[v]++
	}
	if len(counts) == 0 {
		return nil
	}
	present := make([]string, 0, len(counts))
	for v := range counts {
		present = append(present, v)
	}
	items := make([]statusBreakdownItem, 0, len(present))
	for _, v := range sortedStatusValues(present, additions) {
		items = append(items, statusBreakdownItem{Value: v, Count: counts[v]})
	}
	return items
}

// statusColor returns the CSS color expression for a status value. Built-ins
// return a CSS variable so the active theme selects the right value.
func statusColor(value string) string {
	switch strings.ToLower(value) {
	case model.StatusDraft:
		return "var(--status-draft)"
	case model.StatusApproved:
		return "var(--status-approved)"
	}
	var h uint32 = 5381
	for i := range len(value) {
		h = (h*33 + uint32(value[i])) & 0xFFFFFFFF
	}
	region := h % 4
	offset := (h >> 8) % 30
	var hue int
	switch region {
	case 0:
		hue = 0 + int(offset)
	case 1:
		hue = 80 + int(offset)
	case 2:
		hue = 160 + int(offset)
	case 3:
		hue = 280 + int(offset)
	}
	return fmt.Sprintf("hsl(%d, 28%%, 55%%)", hue)
}

// effectiveStatus returns the requirement's status, defaulting to approved.
func effectiveStatus(req *model.Node) string {
	if s, ok := req.Attrs[model.AttrStatus].(string); ok && s != "" {
		return s
	}
	return model.StatusDefault
}

// countItems tallies non-requirement nodes (containers + info items).
func countItems(nodes []*model.Node) int {
	n := 0
	for _, node := range nodes {
		if node.Kind != model.KindRequirement {
			n++
		}
		n += countItems(node.Children)
	}
	return n
}

// Content headings start at h2 (h1 is the document title) and stop at h6.
const (
	minHeadingLevel = 2
	maxHeadingLevel = 6
)

// itemHeadingTag returns the HTML heading tag for an item's native markdown
// level. Top-level items start at h2 (h1 is the document title); nested items
// keep their native depth, floored one level below their parent.
func itemHeadingTag(level int, isChild bool) string {
	floor := minHeadingLevel
	if isChild {
		floor = level
	}
	return "h" + strconv.Itoa(min(max(level, floor), maxHeadingLevel))
}

// linkHref resolves a requirement ID to an href for this page: the
// cross-document resolver when available, else a same-page fragment.
func linkHref(resolveLink func(string) string, id string) string {
	if resolveLink != nil {
		if h := resolveLink(id); h != "" {
			return h
		}
	}
	if id == "" {
		return ""
	}
	return "#" + id
}

// renderDocHead writes the <!DOCTYPE> + <html> + <head> block. The
// stylesheet and client script are inlined: there is no runtime CDN request,
// so an exported page is fully offline. The theme is applied by a small head
// script before first paint to avoid a light/dark flash.
func renderDocHead(b *bufio.Writer, title string) {
	_, _ = b.WriteString("<!DOCTYPE html>\n")
	_, _ = b.WriteString("<html lang=\"en\">\n")
	_, _ = b.WriteString("<head>\n")
	_, _ = b.WriteString("<meta charset=\"UTF-8\">\n")
	_, _ = b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	_, _ = fmt.Fprintf(b, "<title>%s</title>\n", html.EscapeString(title))
	_, _ = b.WriteString(`<script>(function(){try{var t=localStorage.getItem('reqmd-theme');var d=t?t==='dark':window.matchMedia&&window.matchMedia('(prefers-color-scheme: dark)').matches;if(d)document.documentElement.classList.add('dark');}catch(e){}})();</script>` + "\n")
	_, _ = b.WriteString("<style>\n")
	_, _ = b.WriteString(css)
	_, _ = b.WriteString("</style>\n")
	_, _ = b.WriteString("</head>\n")
}

// renderBodyOpen writes the <body>, the app bar, the shell open, and the
// server-rendered outline.
func renderBodyOpen(b *bufio.Writer, pageTitle string) {
	_, _ = b.WriteString("<body data-page=\"document\">\n")
	_, _ = fmt.Fprintf(b, `<header class="app-bar">
  <button type="button" class="icon-btn toc-toggle" data-action="toggle-toc" aria-label="Toggle outline" aria-controls="toc" aria-expanded="false">&#9776;</button>
  <a class="app-bar__brand" href="index.html"><span class="app-bar__logo" aria-hidden="true"></span>reqmd</a>
  <span class="app-bar__sep" aria-hidden="true">/</span>
  <span class="app-bar__title">%s</span>
  <div class="app-bar__tools">
    <button type="button" class="icon-btn" data-action="palette" aria-label="Command palette" title="Command palette (Ctrl+K)" data-mod-hint="palette">Ctrl K</button>
    <button type="button" class="icon-btn" data-action="theme" aria-label="Toggle theme" title="Toggle light/dark">◐</button>
  </div>
</header>
<div class="shell">
`, html.EscapeString(pageTitle))
}

// renderTOC writes the server-rendered outline. It works without JavaScript;
// the client hides items that fall outside the filtered requirement set.
func renderTOC(b *bufio.Writer, doc model.Document, rd *renderData) {
	_, _ = b.WriteString(`<nav class="toc" id="toc" aria-label="Outline">
  <div class="toc__head"><span>Outline</span><span class="toc__count" data-toc-count></span></div>
  <ul class="toc__list">
`)
	var walk func(nodes []*model.Node, child bool)
	walk = func(nodes []*model.Node, child bool) {
		for _, n := range nodes {
			anchor := html.EscapeString(rd.Anchors[n])
			class := "toc__item"
			if child {
				class += " toc__item--child"
			}
			_, _ = fmt.Fprintf(b, `<li class="%s" data-toc-for="%s">`, class, anchor)
			_, _ = fmt.Fprintf(b, `<a class="toc__link" href="#%s">`, anchor)
			if n.Kind == model.KindRequirement {
				_, _ = fmt.Fprintf(b, `<code>%s</code>`, html.EscapeString(n.ID))
				if n.Title != "" {
					_, _ = fmt.Fprintf(b, `<span class="toc__title">%s</span>`, html.EscapeString(n.Title))
				}
			} else {
				_, _ = fmt.Fprintf(b, `<span class="toc__title">%s</span>`, html.EscapeString(n.Title))
			}
			_, _ = b.WriteString("</a>")
			if len(n.Children) > 0 {
				_, _ = b.WriteString("<ul class=\"toc__list\">")
				walk(n.Children, true)
				_, _ = b.WriteString("</ul>")
			}
			_, _ = b.WriteString("</li>\n")
		}
	}
	walk(doc.Nodes, false)
	_, _ = b.WriteString("  </ul>\n</nav>\n")
}

// renderDocHeader writes the document title, count line, status breakdown,
// boundary badges, and the frontmatter description. The trace chain lives at
// the end of the Document view so it can never push a reader past the toolbar.
func renderDocHeader(b *bufio.Writer, doc model.Document, title string, isRoot, isLeaf bool, rd *renderData) {
	_, _ = b.WriteString("<header class=\"doc-head\">\n")
	_, _ = fmt.Fprintf(b, "<h1>%s</h1>\n", html.EscapeString(title))

	reqs := doc.Requirements()
	count := len(reqs)
	var childCount int
	for _, req := range reqs {
		if req.ParentID != "" {
			childCount++
		}
	}
	label := fmt.Sprintf("%d requirements", count)
	if count == 1 {
		label = "1 requirement"
	}
	if childCount > 0 {
		label += fmt.Sprintf(" · %d nested", childCount)
	}
	if items := countItems(doc.Nodes); items > 0 {
		label += fmt.Sprintf(" · %d items", items)
	}
	if len(rd.StatusBreakdown) > 0 {
		_, _ = b.WriteString(`<span class="doc-head__status">`)
		for _, item := range rd.StatusBreakdown {
			_, _ = fmt.Fprintf(b, `<span class="doc-head__status-item"><span class="dot" style="background:%s" aria-hidden="true"></span>%s %d</span>`,
				html.EscapeString(statusColor(item.Value)), html.EscapeString(item.Value), item.Count)
		}
		_, _ = b.WriteString(`</span>`)
	}
	if isRoot || isLeaf {
		_, _ = b.WriteString(`<span class="doc-head__badges">`)
		if isRoot {
			_, _ = b.WriteString(`<span class="badge" title="Root boundary — this document has no upstream sources">Root</span>`)
		}
		if isLeaf {
			_, _ = b.WriteString(`<span class="badge" title="Leaf boundary — this document is not referenced as upstream by any other document">Leaf</span>`)
		}
		_, _ = b.WriteString(`</span>`)
	}
	_, _ = fmt.Fprintf(b, "<p class=\"doc-head__facts\">%s</p>\n", label)

	if desc, ok := doc.Meta["description"]; ok {
		if s, ok := desc.(string); ok && s != "" {
			_, _ = fmt.Fprintf(b, "<div class=\"doc-head__desc prose\">\n%s\n</div>\n", renderMarkdown(s))
		}
	}
	_, _ = b.WriteString("</header>\n")
}

// renderDocChain writes the optional document chain navigation block.
func renderDocChain(b *bufio.Writer, g DocChainGraph) {
	if !g.HasContent() {
		return
	}
	totalUp, totalDown := 0, 0
	for _, t := range g.Upstream {
		totalUp += len(t.Cards)
	}
	for _, t := range g.Downstream {
		totalDown += len(t.Cards)
	}
	// Orientation, not the task: always collapsed so the reading view keeps
	// the reader's attention.
	_, _ = b.WriteString("<nav class=\"doc-chain doc-chain--deep\" aria-label=\"Document traceability chain\">\n")
	_, _ = b.WriteString("  <details class=\"chain-details\">\n")

	_, _ = b.WriteString("    <summary class=\"chain-summary\">\n")
	_, _ = b.WriteString("      <span class=\"chain-summary__label\">Traceability chain</span>\n")
	_, _ = b.WriteString("      <span class=\"chain-summary__count\">")
	_, _ = fmt.Fprintf(b, "%d upstream", totalUp)
	if totalDown > 0 {
		_, _ = fmt.Fprintf(b, " · %d downstream", totalDown)
	}
	_, _ = b.WriteString("</span>\n")
	_, _ = b.WriteString("    </summary>\n")

	renderChainGraph(b, g)

	_, _ = b.WriteString("  </details>\n")
	_, _ = b.WriteString("</nav>\n")
}

// renderToolbar writes the view switcher, search field, active-filter chips
// placeholder, and the per-attribute filter menus.
func renderToolbar(b *bufio.Writer, doc model.Document, rd *renderData) {
	statusEnabled := doc.XReqmd == nil || !doc.XReqmd.IgnoreStatus
	defaultStatus := ""
	if statusEnabled {
		defaultStatus = model.StatusApproved
	}

	_, _ = b.WriteString("<div class=\"toolbar\">\n")
	_, _ = b.WriteString("  <div class=\"toolbar__row\">\n")
	_, _ = b.WriteString(`    <div class="viewswitch" role="group" aria-label="View">
      <button type="button" class="viewswitch__btn is-active" data-view="document" aria-pressed="true" title="Document view (g d)">Document</button>
      <button type="button" class="viewswitch__btn" data-view="table" aria-pressed="false" title="Table view (g t)">Table</button>
      <button type="button" class="viewswitch__btn" data-view="trace" aria-pressed="false" title="Trace explorer (g r)">Trace</button>
      <button type="button" class="viewswitch__btn" data-view="coverage" aria-pressed="false" title="Coverage view (g c)">Coverage</button>
    </div>
`)
	_, _ = b.WriteString(`    <div class="toolbar__tools">
      <button type="button" class="icon-btn" data-action="density" aria-label="Toggle density" title="Toggle density (f)">⇕</button>
    </div>
  </div>` + "\n")
	_, _ = fmt.Fprintf(b, `  <div class="toolbar__row">
    <input type="search" class="search" data-search placeholder="Search ID, title, body, attributes…" autocomplete="off" spellcheck="false" aria-label="Search requirements">
    <span class="count" data-count role="status" aria-live="polite"></span>
  </div>
  <div class="toolbar__chips" data-active-filters hidden></div>
  <div class="filters" data-filters data-default-status=%q>
`, defaultStatus)

	// Status is the lifecycle filter; keep it first for stable ordering.
	attrs := sortedKeys(rd.Filters)
	sort.SliceStable(attrs, func(i, j int) bool {
		if attrs[i] == model.AttrStatus {
			return true
		}
		if attrs[j] == model.AttrStatus {
			return false
		}
		return attrs[i] < attrs[j]
	})
	for _, attr := range attrs {
		_, _ = fmt.Fprintf(b, `    <details class="filter" data-filter-group=%q>
      <summary class="filter__toggle">%s<span class="filter__count" data-filter-count=%q></span></summary>
      <div class="filter__menu">
`, attr, html.EscapeString(formatAttrLabel(attr)), attr)
		for _, value := range rd.Filters[attr] {
			checked := ""
			if attr == model.AttrStatus && statusEnabled && value == model.StatusApproved {
				checked = " checked"
			}
			_, _ = fmt.Fprintf(b, `        <label class="filter__option"><input type="checkbox" value="%s" data-filter-attr="%s"%s><span>%s</span></label>`+"\n",
				html.EscapeString(value), attr, checked, html.EscapeString(value))
		}
		_, _ = b.WriteString("      </div>\n    </details>\n")
	}
	_, _ = b.WriteString("  </div>\n")
	_, _ = b.WriteString(`<noscript><p class="filter-note">The Table, Trace, and Coverage views and the search filters need JavaScript. The Document view is complete without it.</p></noscript>` + "\n")
	_, _ = b.WriteString("</div>\n")
}

// renderDocumentView writes the server-rendered reading view: one card per
// requirement, containers as collapsible sections, and info items as plain
// blocks. Contents are visible by default (no x-cloak), so the reading view
// survives a JavaScript failure or an offline file:// open.
func renderDocumentView(b *bufio.Writer, doc model.Document, rd *renderData, propOrder []string, an *Analysis, resolveLink func(string) string, chain DocChainGraph) {
	_, _ = b.WriteString("<section class=\"view view--document\" data-view-panel=\"document\">\n")
	_, _ = b.WriteString("<main class=\"doc\">\n")
	if len(doc.Nodes) == 0 {
		renderEmptyState(b, "No requirements", "This document contains no requirement nodes.")
	}
	var walk func(nodes []*model.Node, depth int)
	walk = func(nodes []*model.Node, depth int) {
		for _, n := range nodes {
			anchor := rd.Anchors[n]
			switch n.Kind {
			case model.KindRequirement:
				renderCard(b, n, propOrder, an, resolveLink, depth > 0)
			case model.KindContainer:
				renderContainerBlock(b, n, anchor, depth > 0)
			case model.KindInfo:
				renderInfoBlock(b, n, anchor, depth > 0)
			}
			walk(n.Children, depth+1)
		}
	}
	walk(doc.Nodes, 0)
	renderDocChain(b, chain)
	_, _ = b.WriteString("</main>\n</section>\n")
}

// renderEmptyState writes a consistent empty-state block.
func renderEmptyState(b *bufio.Writer, title, body string) {
	_, _ = fmt.Fprintf(b, "<div class=\"empty\"><p class=\"empty__title\">%s</p><p>%s</p></div>\n",
		html.EscapeString(title), html.EscapeString(body))
}

// renderContainerBlock writes a collapsible section for a container node.
func renderContainerBlock(b *bufio.Writer, node *model.Node, anchor string, isChild bool) {
	class := "item item--container"
	if isChild {
		class += " item--child"
	}
	escAnchor := html.EscapeString(anchor)
	_, _ = fmt.Fprintf(b, "<section class=\"%s\" id=\"%s\" data-anchor=\"%s\">\n", class, escAnchor, escAnchor)
	_, _ = b.WriteString("<details open>\n")
	tag := itemHeadingTag(node.Level, isChild)
	_, _ = fmt.Fprintf(b, "<summary class=\"item__head\"><%s class=\"item__title\">%s</%s><a class=\"anchor\" href=\"#%s\" aria-label=\"Link to this section\">#</a></summary>\n",
		tag, html.EscapeString(node.Title), tag, escAnchor)
	if node.Body != "" {
		_, _ = fmt.Fprintf(b, "<div class=\"item__body prose\">\n%s\n</div>\n", renderMarkdown(node.Body))
	}
	_, _ = b.WriteString("</details>\n</section>\n")
}

// renderInfoBlock writes a plain block for a leaf info item.
func renderInfoBlock(b *bufio.Writer, node *model.Node, anchor string, isChild bool) {
	class := "item"
	if isChild {
		class += " item--child"
	}
	escAnchor := html.EscapeString(anchor)
	_, _ = fmt.Fprintf(b, "<section class=\"%s\" id=\"%s\" data-anchor=\"%s\">\n", class, escAnchor, escAnchor)
	if node.Title != "" {
		tag := itemHeadingTag(node.Level, isChild)
		_, _ = fmt.Fprintf(b, "<%s class=\"item__title\">%s</%s>\n", tag, html.EscapeString(node.Title), tag)
	}
	if node.Body != "" {
		_, _ = fmt.Fprintf(b, "<div class=\"item__body prose\">\n%s\n</div>\n", renderMarkdown(node.Body))
	}
	_, _ = b.WriteString("</section>\n")
}

// renderCard writes one requirement card. Semantic state is carried on data
// attributes (used by CSS and the client) plus at most a few state chips, so
// the reading view stays quiet.
func renderCard(b *bufio.Writer, req *model.Node, propOrder []string, an *Analysis, resolveLink func(string) string, isChild bool) {
	escID := html.EscapeString(req.ID)
	var renderedBody string
	if req.Body != "" {
		renderedBody = renderMarkdown(req.Body)
	}

	var st NodeState
	hasState := false
	if an != nil {
		st, hasState = an.Nodes[req.ID]
	}

	class := "req"
	if isChild {
		class += " req--child"
	}
	_, _ = fmt.Fprintf(b, "<article class=\"%s\" id=\"%s\" data-id=\"%s\" data-status=\"%s\"",
		class, escID, escID, html.EscapeString(effectiveStatus(req)))
	if hasState {
		if st.Coverage != "" {
			_, _ = fmt.Fprintf(b, " data-coverage=\"%s\"", html.EscapeString(st.Coverage))
		}
		if st.Verdict != "" {
			_, _ = fmt.Fprintf(b, " data-verdict=\"%s\"", html.EscapeString(st.Verdict))
		}
	}
	_, _ = b.WriteString(">\n")

	renderCardHeader(b, req, isChild)
	renderCardFlags(b, req, st, hasState, an != nil && an.HasResults)

	if req.Body != "" {
		_, _ = fmt.Fprintf(b, "<div class=\"req__body prose\">\n%s\n</div>\n", renderedBody)
	}

	rows := attrRows(req, propOrder)
	hasTraces := len(st.Up)+len(st.Down) > 0
	if len(rows) > 0 || req.Rationale != "" || len(st.Evidence) > 0 || hasTraces {
		_, _ = fmt.Fprintf(b, "<details class=\"req__detail\"><summary class=\"req__detail-toggle\">Details</summary>\n<div class=\"req__detail-body\">\n")
		renderAttrGrid(b, rows)
		renderRationale(b, req)
		renderEvidenceList(b, st)
		renderTraceLinks(b, st, resolveLink)
		_, _ = b.WriteString("</div>\n</details>\n")
	}

	_, _ = b.WriteString("</article>\n")
}

// renderCardHeader writes the ID/title heading and the anchor link.
func renderCardHeader(b *bufio.Writer, req *model.Node, isChild bool) {
	escID := html.EscapeString(req.ID)
	escTitle := html.EscapeString(req.Title)
	tag := "h2"
	if isChild {
		tag = "h3"
	}
	_, _ = b.WriteString("<header class=\"req__head\">\n")
	if escTitle != "" {
		_, _ = fmt.Fprintf(b, "<%s class=\"req__title\"><code class=\"req__id\">%s</code> <span class=\"req__name\">%s</span></%s>\n", tag, escID, escTitle, tag)
	} else {
		_, _ = fmt.Fprintf(b, "<%s class=\"req__title\"><code class=\"req__id\">%s</code></%s>\n", tag, escID, tag)
	}
	_, _ = fmt.Fprintf(b, "<a class=\"anchor\" href=\"#%s\" aria-label=\"Link to %s\">#</a>\n", escID, escID)
	_, _ = b.WriteString("</header>\n")
}

// renderCardFlags writes the state chips. Only exceptional or decision-
// relevant states are shown: draft/custom status, coverage problems, and any
// verdict. A covered requirement with an approved status adds no chrome.
func renderCardFlags(b *bufio.Writer, req *model.Node, st NodeState, hasState, hasResults bool) {
	var chips []string
	if chip := statusChip(req); chip != "" {
		chips = append(chips, chip)
	}
	if hasState {
		if chip := coverageChip(st); chip != "" {
			chips = append(chips, chip)
		}
		if chip := verdictChip(st, hasResults); chip != "" {
			chips = append(chips, chip)
		}
	}
	if len(chips) == 0 {
		return
	}
	_, _ = b.WriteString("<p class=\"req__flags\">")
	_, _ = b.WriteString(strings.Join(chips, " "))
	_, _ = b.WriteString("</p>\n")
}

// statusChip renders a chip only for non-approved statuses: an approved
// requirement adds no chrome.
func statusChip(req *model.Node) string {
	status := effectiveStatus(req)
	if strings.EqualFold(status, model.StatusApproved) {
		return ""
	}
	return fmt.Sprintf(`<span class="chip chip--status" style="--chip:%s">%s</span>`,
		html.EscapeString(statusColor(status)), html.EscapeString(status))
}

// coverageChip renders coverage problems only (missing or unresolved); a
// covered or not-required requirement adds no chrome.
func coverageChip(st NodeState) string {
	switch st.Coverage {
	case graph.CoverageMissing:
		label := "missing coverage"
		if len(st.Misses) > 0 {
			label += " from " + strings.Join(st.Misses, ", ")
		}
		if len(st.Unknown) > 0 {
			label += " (unknown source: " + strings.Join(st.Unknown, ", ") + ")"
		}
		if st.Drafts > 0 {
			label += fmt.Sprintf(" (%d draft downstreams ignored)", st.Drafts)
		}
		return fmt.Sprintf(`<span class="chip chip--coverage chip--coverage-missing">%s</span>`, html.EscapeString(label))
	case graph.CoverageUnresolved:
		if len(st.Unknown) == 0 {
			return ""
		}
		return fmt.Sprintf(`<span class="chip chip--coverage chip--coverage-unresolved">unknown coverage source: %s</span>`,
			html.EscapeString(strings.Join(st.Unknown, ", ")))
	default:
		return ""
	}
}

// verdictChip renders the rolled-up verdict, or an explicit no-verdict chip
// for an eligible measure when results were loaded.
func verdictChip(st NodeState, hasResults bool) string {
	if st.Verdict == "" {
		if hasResults && st.VerdictEligible {
			return `<span class="chip chip--no-verdict">no verdict</span>`
		}
		return ""
	}
	title := "verified: " + st.Verdict
	if st.VerdictSource != "" {
		title += " — " + st.VerdictSource
	}
	return fmt.Sprintf(`<span class="chip chip--%s" title="%s">%s</span>`,
		html.EscapeString(st.Verdict), html.EscapeString(title), html.EscapeString(st.Verdict))
}

// attrRow is one rendered attribute cell in a requirement card.
type attrRow struct {
	Name       string
	Value      any
	Undeclared bool
}

// attrRows returns the attribute cells to render, in display order: the
// schema property order first, then any remaining attributes actually present
// on the requirement sorted A→Z. The trace attribute is skipped.
func attrRows(req *model.Node, propOrder []string) []attrRow {
	rows := make([]attrRow, 0, len(req.Attrs))
	seen := make(map[string]struct{}, len(propOrder))
	for _, prop := range propOrder {
		seen[prop] = struct{}{}
		if prop == model.AttrTrace {
			continue
		}
		if val, ok := req.Attrs[prop]; ok {
			rows = append(rows, attrRow{Name: prop, Value: val})
		}
	}
	var extra []string
	for name := range req.Attrs {
		if name == model.AttrTrace {
			continue
		}
		if _, declared := seen[name]; !declared {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	for _, name := range extra {
		rows = append(rows, attrRow{Name: name, Value: req.Attrs[name], Undeclared: true})
	}
	return rows
}

// renderAttrGrid writes the attribute definition list.
func renderAttrGrid(b *bufio.Writer, rows []attrRow) {
	if len(rows) == 0 {
		return
	}
	_, _ = b.WriteString("<dl class=\"attrs\">\n")
	for _, row := range rows {
		class := "attrs__row"
		if row.Undeclared {
			class += " attrs__row--undeclared"
		}
		_, _ = fmt.Fprintf(b, "<div class=\"%s\" data-attr=\"%s\"><dt>%s</dt><dd>%s</dd></div>\n",
			class, html.EscapeString(row.Name), html.EscapeString(formatAttrLabel(row.Name)), formatAttrHTML(row.Value))
	}
	_, _ = b.WriteString("</dl>\n")
}

// renderRationale writes the rationale block when present.
func renderRationale(b *bufio.Writer, req *model.Node) {
	if req.Rationale == "" {
		return
	}
	_, _ = fmt.Fprintf(b, "<div class=\"rationale\"><strong>Rationale:</strong> %s</div>\n", renderMarkdown(req.Rationale))
}

// renderEvidenceList writes the rolled-up verification evidence.
func renderEvidenceList(b *bufio.Writer, st NodeState) {
	if len(st.Evidence) == 0 {
		return
	}
	_, _ = fmt.Fprintf(b, "<details class=\"evidence\"><summary class=\"evidence__toggle\">Evidence (%d)</summary>\n", len(st.Evidence))
	for _, it := range st.Evidence {
		_, _ = b.WriteString("<div class=\"evidence__item\">\n")
		if it.Outcome != "" {
			_, _ = fmt.Fprintf(b, "<span class=\"chip chip--%s\">%s</span>\n", html.EscapeString(it.Outcome), html.EscapeString(it.Outcome))
		}
		if it.Case != "" {
			_, _ = fmt.Fprintf(b, "<code class=\"evidence__case\">%s</code>\n", html.EscapeString(it.Case))
		}
		if it.File != "" {
			_, _ = fmt.Fprintf(b, "<span class=\"evidence__file\">%s</span>\n", html.EscapeString(it.File))
		}
		if it.Description != "" {
			_, _ = fmt.Fprintf(b, "<div class=\"evidence__desc prose\">\n%s\n</div>\n", renderMarkdown(it.Description))
		}
		_, _ = b.WriteString("</div>\n")
	}
	_, _ = b.WriteString("</details>\n")
}

// renderTraceLinks writes the upstream/downstream trace columns with the
// authored reference, title, pin, and resolution state.
func renderTraceLinks(b *bufio.Writer, st NodeState, resolveLink func(string) string) {
	if len(st.Up) == 0 && len(st.Down) == 0 {
		return
	}
	_, _ = b.WriteString("<div class=\"traces\">\n")
	renderTraceColumn(b, "Traces to", st.Up, resolveLink)
	renderTraceColumn(b, "Traced from", st.Down, resolveLink)
	_, _ = b.WriteString("</div>\n")
}

// renderTraceColumn writes one direction column.
func renderTraceColumn(b *bufio.Writer, header string, links []LinkState, resolveLink func(string) string) {
	if len(links) == 0 {
		return
	}
	_, _ = fmt.Fprintf(b, "<div class=\"traces__col\"><div class=\"traces__head\">%s (%d)</div>\n", header, len(links))
	for _, link := range links {
		renderTraceLink(b, link, resolveLink)
	}
	_, _ = b.WriteString("</div>\n")
}

// renderTraceLink writes one link with its pin/version state.
func renderTraceLink(b *bufio.Writer, link LinkState, resolveLink func(string) string) {
	if !link.Resolved {
		_, _ = fmt.Fprintf(b, "<span class=\"link link--broken\" title=\"unresolved reference\"><code>%s</code><span class=\"link__title\">unresolved</span></span>\n",
			html.EscapeString(link.Ref))
		return
	}
	href := html.EscapeString(linkHref(resolveLink, link.ID))
	_, _ = fmt.Fprintf(b, "<a class=\"link\" href=\"%s\"><code>%s</code>", href, html.EscapeString(link.ID))
	if link.Title != "" {
		_, _ = fmt.Fprintf(b, "<span class=\"link__title\">%s</span>", html.EscapeString(link.Title))
	}
	if link.Pinned {
		pin := fmt.Sprintf("~%d", link.Pin)
		if link.Version > 0 {
			pin += fmt.Sprintf(" → v%d", link.Version)
		}
		class := "link__pin"
		switch link.Direction {
		case graph.DirOutdated:
			class += " link__pin--outdated"
		case graph.DirPredated:
			class += " link__pin--predated"
		}
		_, _ = fmt.Fprintf(b, "<span class=\"%s\">%s</span>", class, html.EscapeString(pin))
	}
	_, _ = b.WriteString("</a>\n")
}

// renderPalette writes the command-palette overlay. Overlays are the only
// content hidden until the client boots.
func renderPalette(b *bufio.Writer) {
	_, _ = b.WriteString(`<div class="overlay" id="palette" role="dialog" aria-modal="true" aria-label="Command palette" hidden>
  <div class="overlay__backdrop" data-action="close-overlays"></div>
  <div class="overlay__panel">
    <div class="palette">
      <input class="palette__input" data-palette-input type="text" placeholder="Jump to a requirement or run a command…" autocomplete="off" spellcheck="false" aria-label="Palette search" aria-controls="palette-list">
      <ul class="palette__list" id="palette-list" data-palette-list role="listbox" aria-label="Results"></ul>
      <p class="palette__empty" data-palette-empty hidden>No matching requirements or commands.</p>
      <div class="palette__foot"><span><kbd>↑</kbd><kbd>↓</kbd> navigate</span><span><kbd>↵</kbd> open</span><span><kbd>esc</kbd> close</span><span><kbd>?</kbd> shortcuts</span></div>
    </div>
  </div>
</div>
`)
}

// renderShortcuts writes the keyboard-shortcut dialog.
func renderShortcuts(b *bufio.Writer) {
	_, _ = b.WriteString(`<div class="overlay" id="shortcuts" role="dialog" aria-modal="true" aria-label="Keyboard shortcuts" hidden>
  <div class="overlay__backdrop" data-action="close-overlays"></div>
  <div class="overlay__panel overlay__panel--narrow" tabindex="-1" data-shortcuts-panel>
    <div class="shortcuts">
      <div class="shortcuts__head">
        <h2 class="shortcuts__title">Keyboard shortcuts</h2>
        <button type="button" class="icon-btn" data-action="close-overlays" aria-label="Close shortcuts">×</button>
      </div>
      <dl class="shortcuts__grid">
        <div><dt data-mod-keys="palette"><kbd>Ctrl</kbd><kbd>K</kbd></dt><dd>Command palette</dd></div>
        <div><dt><kbd>/</kbd></dt><dd>Focus search</dd></div>
        <div><dt><kbd>g</kbd><kbd>d</kbd></dt><dd>Document view</dd></div>
        <div><dt><kbd>g</kbd><kbd>t</kbd></dt><dd>Table view</dd></div>
        <div><dt><kbd>g</kbd><kbd>r</kbd></dt><dd>Trace explorer</dd></div>
        <div><dt><kbd>g</kbd><kbd>c</kbd></dt><dd>Coverage view</dd></div>
        <div><dt><kbd>f</kbd></dt><dd>Toggle density</dd></div>
        <div><dt><kbd>t</kbd></dt><dd>Toggle theme</dd></div>
        <div><dt><kbd>?</kbd></dt><dd>This help</dd></div>
        <div><dt><kbd>esc</kbd></dt><dd>Close overlays</dd></div>
      </dl>
      <p class="shortcuts__note">Shortcuts are ignored while typing in a field.</p>
    </div>
  </div>
</div>
`)
}

// renderBodyClose writes the end of the shell and document.
func renderBodyClose(b *bufio.Writer) {
	_, _ = b.WriteString("</body>\n")
	_, _ = b.WriteString("</html>\n")
}
