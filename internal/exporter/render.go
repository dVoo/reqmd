package exporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"reqmd/internal/model"
	"sort"
	"strconv"
	"strings"
)

// TraceResolver bundles the four callback hooks used to render upstream /
// downstream trace links in the exported HTML. Any nil field is treated
// as "no data" — the trace section is simply omitted for that side.
type TraceResolver struct {
	// Upstream returns the requirement IDs that the given requirement
	// traces TO (its parents).
	Upstream func(string) []string
	// Downstream returns the requirement IDs that trace TO the given
	// requirement (its children).
	Downstream func(string) []string
	// ResolveLink maps a requirement ID to a cross-file HTML anchor
	// (e.g. "02-system-requirements.html#SYS-FMT-001"). When nil or it
	// returns "", same-file fragment links ("#ID") are used.
	ResolveLink func(string) string
	// TitleOf maps a requirement ID to its human-readable title. When
	// nil or it returns "", only the ID is shown.
	TitleOf func(string) string
}

// linkHref returns the resolved HTML anchor for the given req ID, or
// the same-file fragment fallback when the trace context is missing or
// the resolver returns an empty string.
func (tc *TraceResolver) linkHref(id string) string {
	if tc != nil && tc.ResolveLink != nil {
		if ref := tc.ResolveLink(id); ref != "" {
			return ref
		}
	}
	return "#" + id
}

// linkTitle returns the resolved title for the given req ID, or "" when
// the trace context is missing or the titleOf function is not set.
func (tc *TraceResolver) linkTitle(id string) string {
	if tc != nil && tc.TitleOf != nil {
		return tc.TitleOf(id)
	}
	return ""
}

// indexEntry is one node in the embedded JSON index used by the Alpine
// search/filter logic and the sidebar TOC. The JSON tags match the wire
// format the browser-side JS reads. ID is the DOM anchor: a requirement
// ID for requirements, or a generated "info-N" anchor for containers and
// info items.
type indexEntry struct {
	ID        string         `json:"id"`
	Type      string         `json:"type,omitempty"` // "container" | "info"; omitted for requirements
	Title     string         `json:"title,omitempty"`
	Body      string         `json:"body"`
	Rationale string         `json:"rationale"`
	Attrs     map[string]any `json:"attrs"`
	ParentID  string         `json:"parentId,omitempty"`
	Children  []string       `json:"children,omitempty"`
}

// filterSet is the per-attribute sorted unique values collected across
// all requirements, used to populate the search toolbar filter menus.
type filterSet map[string][]string

// renderData is the single O(n) intermediate produced by
// buildRenderData; it powers the embedded JSON index, the search/filter
// toolbar, and the tree-walk rendering loop.
type renderData struct {
	Filters   filterSet
	Anchors   map[*model.Node]string
	StatusCnt map[string]int
	Index     []indexEntry
	Nodes     []*model.Node
}

// tracePair holds the precomputed upstream/downstream trace links for a
// single requirement. nil entries are treated as "no trace data".
type tracePair struct {
	up, down []traceLink
}

// traceLink is a single upstream or downstream trace reference,
// pairing a requirement ID with its optional human-readable title.
type traceLink struct {
	ID    string
	Title string
}

// buildTraceCache precomputes upstream/downstream trace links for every
// requirement in doc. When tc is nil the result is also nil.
func buildTraceCache(doc model.Document, tc *TraceResolver) map[string]tracePair {
	if tc == nil {
		return nil
	}
	reqs := doc.Requirements()
	cache := make(map[string]tracePair, len(reqs))
	for _, req := range reqs {
		cache[req.ID] = tracePair{
			up:   toTraceLinks(tc.Upstream(req.ID), tc),
			down: toTraceLinks(tc.Downstream(req.ID), tc),
		}
	}
	return cache
}

func toTraceLinks(ids []string, tc *TraceResolver) []traceLink {
	links := make([]traceLink, len(ids))
	for i, id := range ids {
		links[i] = traceLink{ID: id, Title: tc.linkTitle(id)}
	}
	return links
}

// normalizeAttrValue collapses []any / []string / scalar attribute
// values into a uniform []string, dropping nil entries and nested maps
// (the latter because they cannot be represented as filter values).
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

// buildRenderData runs a single O(n) pass over the document content tree
// to produce every per-document data structure the rendering helpers
// need: the JSON index (all nodes, for search/filter and the TOC), the
// filter set, the per-status counts, and the per-node DOM anchors.
//
// Every node gets a stable anchor: requirements anchor on their ID,
// containers and info items on a generated "info-N" anchor (N = document
// order sequence). The same anchors are used by the index and by the
// rendered DOM so TOC links and scroll-spy stay consistent.
//
// The `status` filter is special-cased so built-in values (draft,
// approved) appear first in canonical order, followed by declared
// additions A→Z, followed by any unrecognized values A→Z. All other
// attrs are emitted sorted A→Z. Attributes with fewer than two distinct
// values are omitted (single-value filters are not useful).
func buildRenderData(doc model.Document) *renderData {
	reqs := doc.Requirements()
	rd := &renderData{
		Index:     make([]indexEntry, 0, len(reqs)),
		Filters:   buildFilters(reqs, doc),
		Nodes:     doc.Nodes,
		Anchors:   make(map[*model.Node]string, len(reqs)),
		StatusCnt: countStatuses(reqs),
	}

	// Pass 1: assign a stable DOM anchor to every node.
	itemSeq := 0
	var assign func(nodes []*model.Node)
	assign = func(nodes []*model.Node) {
		for _, n := range nodes {
			if n.Kind == model.KindRequirement {
				rd.Anchors[n] = n.ID
			} else {
				itemSeq++
				rd.Anchors[n] = fmt.Sprintf("info-%d", itemSeq)
			}
			assign(n.Children)
		}
	}
	assign(doc.Nodes)

	// Pass 2: build index entries (parent + children via the anchor map).
	var build func(nodes []*model.Node, parentAnchor string)
	build = func(nodes []*model.Node, parentAnchor string) {
		for _, n := range nodes {
			rd.Index = append(rd.Index, buildIndexEntry(n, rd.Anchors[n], parentAnchor, rd.Anchors))
			build(n.Children, rd.Anchors[n])
		}
	}
	build(doc.Nodes, "")

	return rd
}

// buildFilters collects per-attribute unique values across all requirements
// into a filterSet. The `status` filter is special-cased so built-in values
// (draft, approved) appear first in canonical order, followed by declared
// additions A→Z, then any unrecognized values A→Z. All other attrs are emitted
// sorted A→Z. Attributes with fewer than two distinct values are omitted
// (single-value filters are not useful). Status values are skipped when the
// document opts out of the status lifecycle.
func buildFilters(reqs []*model.Node, doc model.Document) filterSet {
	ignoreStatus := doc.XReqmd != nil && doc.XReqmd.IgnoreStatus
	raw := make(map[string]map[string]struct{})
	for _, req := range reqs {
		for k, v := range req.Attrs {
			if k == model.AttrTrace {
				continue
			}
			if k == model.AttrStatus && ignoreStatus {
				continue
			}
			if raw[k] == nil {
				raw[k] = make(map[string]struct{})
			}
			for _, s := range normalizeAttrValue(v) {
				raw[k][s] = struct{}{}
			}
		}
	}

	filters := make(filterSet, len(raw))
	for k, set := range raw {
		if len(set) < 2 {
			continue
		}
		vals := make([]string, 0, len(set))
		for v := range set {
			vals = append(vals, v)
		}
		// Special-case `status`: built-in [draft, approved] first in
		// canonical order, then declared additions A→Z, then any
		// unrecognized values A→Z.
		if k == model.AttrStatus {
			filters[k] = orderStatusFilterValues(vals, doc)
			continue
		}
		sort.Strings(vals)
		filters[k] = vals
	}
	return filters
}

// countStatuses tallies status values across all requirements. Requirements
// without a recognized status string count as model.StatusDefault. The tally
// is independent of filter-set visibility: the header breakdown is always
// rendered unless the document opts out of the lifecycle.
func countStatuses(reqs []*model.Node) map[string]int {
	statusCounts := make(map[string]int)
	for _, req := range reqs {
		sv, ok := req.Attrs[model.AttrStatus].(string)
		if !ok || sv == "" {
			sv = model.StatusDefault
		}
		statusCounts[sv]++
	}
	return statusCounts
}

// sortedStatusValues returns the given status values in canonical order:
// built-ins [draft, approved] first, then declared additions A→Z, then any
// unrecognized values A→Z. Both countStatusBreakdown and the filter-set
// ordering use this; the test in render_test.go covers it.
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
	out = append(out, rest...)
	return out
}

// orderStatusFilterValues is a thin wrapper around sortedStatusValues
// that pulls the declared additions from doc.XReqmd (nil-safe).
func orderStatusFilterValues(vals []string, doc model.Document) []string {
	var additions []string
	if doc.XReqmd != nil {
		additions = doc.XReqmd.AdditionalStatusValues
	}
	return sortedStatusValues(vals, additions)
}

// statusBreakdownItem describes one status value + its count for the
// header breakdown row.
type statusBreakdownItem struct {
	Value string
	Count int
}

// countStatusBreakdown tallies status values across all requirements in a
// document. Returns the items in display order: built-in
// (draft, approved) first, then additions A→Z. Values not declared are
// treated as StatusDefault ("approved"). When the document opts out of
// the status lifecycle (x-reqmd.ignore-status: true), callers should
// skip rendering the breakdown entirely.
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

// statusColor returns the CSS color expression for a status value used
// in the header breakdown dots and card borders. Built-ins return a
// CSS variable so the active theme (light/dark) selects the right
// value; extensions return a fixed hash-derived muted color (rare,
// small dots — fixed color is acceptable).
func statusColor(value string) string {
	switch strings.ToLower(value) {
	case model.StatusDraft:
		return "var(--status-draft)"
	case model.StatusApproved:
		return "var(--status-approved)"
	}
	// DJB2a hash → hue in one of 4 collision-free regions:
	// 0-50, 80-120, 160-260, 280-340.
	var h uint32 = 5381
	for i := range len(value) {
		h = (h*33 + uint32(value[i])) & 0xFFFFFFFF
	}
	region := h % 4
	offset := (h >> 8) % 30 // 0-29
	var hue int
	switch region {
	case 0:
		hue = 0 + int(offset) // 0-29
	case 1:
		hue = 80 + int(offset) // 80-109
	case 2:
		hue = 160 + int(offset) // 160-189
	case 3:
		hue = 280 + int(offset) // 280-309
	}
	return fmt.Sprintf("hsl(%d, 28%%, 55%%)", hue)
}

// buildIndexEntry produces a single index entry for one node. The body
// is truncated to 80 runes + "..." for the sidebar preview; the full
// body still appears in the card itself. Requirements carry their attrs;
// containers and info items carry no attrs and a type marker.
func buildIndexEntry(n *model.Node, anchor, parentAnchor string, anchors map[*model.Node]string) indexEntry {
	body := n.Body
	runes := []rune(body)
	if len(runes) > 80 {
		body = string(runes[:80]) + "..."
	}
	entry := indexEntry{
		ID:        anchor,
		Title:     n.Title,
		Body:      body,
		Rationale: n.Rationale,
		ParentID:  parentAnchor,
	}
	if n.Kind == model.KindRequirement {
		rawAttrs := make(map[string]any, len(n.Attrs))
		for k, v := range n.Attrs {
			if k == model.AttrTrace {
				continue
			}
			rawAttrs[k] = v
		}
		entry.Attrs = rawAttrs
	} else {
		entry.Type = n.Kind.String()
	}
	if len(n.Children) > 0 {
		entry.Children = make([]string, 0, len(n.Children))
		for _, c := range n.Children {
			entry.Children = append(entry.Children, anchors[c])
		}
	}
	return entry
}

// renderDocHeader writes the <header class="doc-header">...</header>
// block: title, count line, status breakdown, boundary badges, and the
// optional description from YAML frontmatter.
func renderDocHeader(b *bufio.Writer, doc model.Document, title string, isRoot, isLeaf bool) {
	_, _ = b.WriteString("<header class=\"doc-header\">\n")
	_, _ = fmt.Fprintf(b, "<h1>%s</h1>\n", html.EscapeString(title))

	count := len(doc.Requirements())
	var parentCount, childCount int
	for _, req := range doc.Requirements() {
		if req.ParentID == "" {
			parentCount++
		} else {
			childCount++
		}
	}
	label := fmt.Sprintf("%d requirements (%d parents, %d children)", count, parentCount, childCount)
	if count == 1 {
		label = "1 requirement"
	}
	if items := countItems(doc.Nodes); items > 0 {
		label += fmt.Sprintf(" · %d items", items)
	}
	_, _ = fmt.Fprintf(b, "<p class=\"doc-count\">%s</p>\n", label)

	if doc.XReqmd == nil || !doc.XReqmd.IgnoreStatus {
		var additions []string
		if doc.XReqmd != nil {
			additions = doc.XReqmd.AdditionalStatusValues
		}
		breakdown := countStatusBreakdown(doc.Requirements(), additions)
		if len(breakdown) > 0 {
			b.WriteString(`<div class="status-breakdown">`)
			for _, item := range breakdown {
				_, _ = fmt.Fprintf(b,
					`<span class="status-breakdown__item"><span class="status-breakdown__dot" style="background:%s"></span><span class="status-breakdown__label">%s</span><span class="status-breakdown__count">%d</span></span>`,
					html.EscapeString(statusColor(item.Value)),
					html.EscapeString(item.Value),
					item.Count,
				)
			}
			b.WriteString(`</div>`)
		}
	}

	if isRoot || isLeaf {
		b.WriteString(`<div class="boundary-badges">`)
		if isRoot {
			b.WriteString(`<span class="boundary-badge boundary-badge--root" title="Root boundary — this document has no upstream sources">Root</span>`)
		}
		if isLeaf {
			b.WriteString(`<span class="boundary-badge boundary-badge--leaf" title="Leaf boundary — this document is not referenced as upstream by any other document">Leaf</span>`)
		}
		b.WriteString(`</div>`)
	}

	if desc, ok := doc.Meta["description"]; ok {
		if s, ok := desc.(string); ok && s != "" {
			fmt.Fprintf(b, "<div class=\"doc-description\">\n%s\n</div>\n", renderMarkdown(s))
		}
	}

	b.WriteString("</header>\n")
}

// renderToolbar writes the search/filter toolbar (search input, active
// filter chips, and the per-attribute filter menus). The Alpine
// component reads data-default-status and data-additional-status-values
// from the root attributes.
func renderToolbar(b *bufio.Writer, doc model.Document, ignoreStatus bool) {
	var additionsJSON []byte
	if doc.XReqmd != nil {
		additionsJSON, _ = json.Marshal(doc.XReqmd.AdditionalStatusValues)
	}
	defaultStatus := model.StatusApproved
	if ignoreStatus {
		defaultStatus = "" // no status filter applied for ignore-status docs
	}
	toolbarAttrs := fmt.Sprintf(
		` class="search-toolbar" data-default-status=%q data-additional-status-values=%q`,
		defaultStatus,
		string(additionsJSON),
	)
	b.WriteString(`<div` + toolbarAttrs + `>
  <div class="search-row">
    <input type="search" class="search-input" placeholder="Search ID, body, rationale, attributes..." autocomplete="off"
      x-model.debounce.150ms="searchTerm">
    <span class="search-count" x-text="visibleCount + '/' + totalCount"></span>
  </div>
  <div class="filter-row">
    <template x-for="(values, attr) in activeFiltersByAttr" :key="attr">
      <template x-for="v in values" :key="attr + ':' + v">
        <span class="filter-chip">
          <span x-text="prettyName(attr) + ': ' + v"></span>
          <button @click="removeFilter(attr, v)" aria-label="Remove filter">×</button>
        </span>
      </template>
    </template>
    <template x-if="Object.keys(activeFiltersByAttr).length > 0">
      <button class="clear-btn" @click="clearAll()">× Clear all</button>
    </template>
  </div>
  <div class="filter-row" id="filter-row">
    <template x-for="(values, attr) in filterOptions" :key="attr">
      <div class="filter-group"
           x-data="{ open: false, closeMenu() { if (this.open) { this.open = false; this.$nextTick(() => { if (this.$refs.toggleBtn) this.$refs.toggleBtn.focus(); }); } } }"
           @click.outside="closeMenu()"
           @keydown.escape.window="closeMenu()">
        <button class="filter-toggle" type="button" @click.prevent="open = !open" :aria-expanded="open" :aria-controls="'filter-menu-' + attr" x-ref="toggleBtn">
          <span x-text="prettyName(attr)"></span>
          <span class="filter-badge" x-show="activeFiltersByAttr[attr] && activeFiltersByAttr[attr].length" x-text="(activeFiltersByAttr[attr] || []).length"></span>
          <span class="filter-arrow" x-text="open ? '▲' : '▾'"></span>
        </button>
        <div class="filter-menu" :id="'filter-menu-' + attr" x-show="open" x-cloak x-transition x-ref="menu" x-effect="if (open) { $nextTick(() => { const cb = $refs.menu.querySelector('input'); if (cb) cb.focus(); }); }">
          <template x-for="v in values" :key="v">
            <label class="filter-option">
              <input type="checkbox" :checked="(activeFiltersByAttr[attr] || []).indexOf(v) !== -1"
                     @change="toggleFilter(attr, v)">
              <span x-text="v"></span>
            </label>
          </template>
        </div>
      </div>
    </template>
  </div>
</div>
`)
}

// countItems tallies non-requirement nodes (containers + info items)
// in the document tree.
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

// renderContentNode renders one tree node and its children depth-first:
// requirements become cards, containers become collapsible folder
// sections, info items become plain blocks. depth selects the child
// modifier classes (root vs nested).
func renderContentNode(b *bufio.Writer, node *model.Node, depth int, rd *renderData, propOrder []string, traceCache map[string]tracePair, tc *TraceResolver, ignoreStatus bool, verdicts map[string]VerdictInfo) {
	isChild := depth > 0
	switch node.Kind {
	case model.KindRequirement:
		renderCard(b, node, propOrder, traceCache, tc, ignoreStatus, isChild, verdicts)
	case model.KindContainer:
		renderContainerBlock(b, node, rd.Anchors[node], isChild)
	case model.KindInfo:
		renderInfoBlock(b, node, rd.Anchors[node], isChild)
	}
	for _, c := range node.Children {
		renderContentNode(b, c, depth+1, rd, propOrder, traceCache, tc, ignoreStatus, verdicts)
	}
}

// itemHeadingTag returns the HTML heading tag for an item's native
// markdown level. Level 1 is reserved for the document title (rendered by
// the doc header's h1), so content headings start at level 2 and render at
// their native level.
func itemHeadingTag(level int) string {
	tag := min(max(level, 2), 6)
	return "h" + strconv.Itoa(tag)
}

// renderContainerBlock writes a collapsible folder-like section for a
// container node: a <details> whose summary carries the heading, an
// anchor link, and the container body. Children are rendered by the
// caller.
func renderContainerBlock(b *bufio.Writer, node *model.Node, anchor string, isChild bool) {
	class := "container-item"
	if isChild {
		class += " container-item--child"
	}
	escTitle := html.EscapeString(node.Title)
	escAnchor := html.EscapeString(anchor)
	headingTag := itemHeadingTag(node.Level)
	fmt.Fprintf(b, "<section class=\"%s\" id=\"%s\">\n", class, escAnchor)
	b.WriteString("<details open>\n")
	fmt.Fprintf(b, "<summary class=\"container-item-summary\">\n<%s class=\"container-item-title\">%s</%s>\n", headingTag, escTitle, headingTag)
	fmt.Fprintf(b, "<a class=\"container-item-anchor\" href=\"#%s\" title=\"Link to this section\">#</a>\n", escAnchor)
	b.WriteString("</summary>\n")
	if node.Body != "" {
		fmt.Fprintf(b, "<div class=\"container-item-body\">\n%s\n</div>\n", renderMarkdown(node.Body))
	}
	b.WriteString("</details>\n")
	b.WriteString("</section>\n")
}

// renderInfoBlock writes a plain block for a leaf info item: an optional
// heading (native level) and the rendered body. Level-0 items (leading
// prose without a heading) render body only.
func renderInfoBlock(b *bufio.Writer, node *model.Node, anchor string, isChild bool) {
	class := "info-item"
	if isChild {
		class += " info-item--child"
	}
	escAnchor := html.EscapeString(anchor)
	fmt.Fprintf(b, "<section class=\"%s\" id=\"%s\">\n", class, escAnchor)
	if node.Title != "" {
		headingTag := itemHeadingTag(node.Level)
		escTitle := html.EscapeString(node.Title)
		fmt.Fprintf(b, "<%s class=\"info-item-title\">%s</%s>\n", headingTag, escTitle, headingTag)
	}
	if node.Body != "" {
		fmt.Fprintf(b, "<div class=\"info-item-body\">\n%s\n</div>\n", renderMarkdown(node.Body))
	}
	b.WriteString("</section>\n")
}

// renderCard writes a single <article> for one requirement. The
// `isChild` flag controls CSS class names ("req-card" vs
// "req-card-child") and the heading level (h2 vs h3). The trace
// section is rendered only when the requirement has at least one
// upstream or downstream link and tc is non-nil.
func renderCard(b *bufio.Writer, req *model.Node, propOrder []string, traceCache map[string]tracePair, tc *TraceResolver, ignoreStatus, isChild bool, verdicts map[string]VerdictInfo) {
	escID := html.EscapeString(req.ID)
	renderedBody := renderMarkdown(req.Body)

	cardClass := "req-card"
	if isChild {
		cardClass = "req-card-child"
	}
	// Status-driven card decoration: any non-approved value (including
	// the implicit default → approved) shows a 3px left border. Skip
	// when the document opts out of the status lifecycle.
	if !ignoreStatus {
		statusVal, hasStatus := req.Attrs[model.AttrStatus].(string)
		if !hasStatus || statusVal == "" {
			statusVal = model.StatusDefault
		}
		if !strings.EqualFold(statusVal, model.StatusApproved) {
			cardClass += " req-card--status-draft"
		}
	}
	// `x-show` runs the per-card visibility check; `x-cloak` keeps the
	// card hidden until Alpine has booted and the check has run for the
	// first time (avoids a flash of unfiltered content).
	fmt.Fprintf(b, "<article class=\"%s\" id=\"%s\" x-show=\"isCardVisible('%s')\" x-cloak>\n", cardClass, escID, escID)

	renderCardHeader(b, req, isChild)
	renderVerdictBadge(b, req, verdicts, isChild)
	renderEvidenceList(b, req, verdicts, isChild)
	renderAttrGrid(b, req, propOrder, isChild)

	// Body
	bodyClass := "req-body"
	if isChild {
		bodyClass = "req-body req-body--child"
	}
	if req.Body != "" {
		fmt.Fprintf(b, "<div class=\"%s\">\n%s\n</div>\n", bodyClass, renderedBody)
	}

	renderRationale(b, req, isChild)
	renderTraceLinks(b, req, traceCache, tc, isChild)

	b.WriteString("</article>\n")
}

// renderCardHeader writes the header row: the ID (+ title if present)
// heading and the anchor link. The heading level (h2 vs h3) and the
// req-id--child modifier are selected by isChild.
func renderCardHeader(b *bufio.Writer, req *model.Node, isChild bool) {
	escID := html.EscapeString(req.ID)
	escTitle := html.EscapeString(req.Title)

	b.WriteString("<div class=\"req-header\">\n")
	headingTag := "h2"
	headingExtra := ""
	if isChild {
		headingTag = "h3"
		headingExtra = " req-id--child"
	}
	if escTitle != "" {
		fmt.Fprintf(b, "<%s class=\"req-id%s\"><code>%s</code>: %s</%s>\n", headingTag, headingExtra, escID, escTitle, headingTag)
	} else {
		fmt.Fprintf(b, "<%s class=\"req-id%s\"><code>%s</code></%s>\n", headingTag, headingExtra, escID, headingTag)
	}
	ariaLabel := ""
	if !isChild {
		ariaLabel = fmt.Sprintf(" aria-label=\"Link to %s\"", escID)
	}
	fmt.Fprintf(b, "<a class=\"req-anchor\" href=\"#%s\" title=\"Link to this requirement\"%s>#</a>\n", escID, ariaLabel)
	b.WriteString("</div>\n")
}

// renderVerdictBadge writes a verification verdict badge next to the
// requirement header when verification results were loaded (--results).
// Measures without results show nothing (missing-verdict is a check
// output, not a badge). The badge is color-coded: green for pass, red
// for fail, orange for inconclusive, gray for skipped.
func renderVerdictBadge(b *bufio.Writer, req *model.Node, verdicts map[string]VerdictInfo, isChild bool) {
	if verdicts == nil {
		return
	}
	v, ok := verdicts[req.ID]
	if !ok {
		return
	}
	class := "verdict-badge"
	if isChild {
		class += " verdict-badge--child"
	}
	switch v.Outcome {
	case "pass":
		class += " verdict-badge--pass"
	case "fail":
		class += " verdict-badge--fail"
	case "skipped":
		class += " verdict-badge--skipped"
	case "inconclusive":
		class += " verdict-badge--inconclusive"
	}
	title := fmt.Sprintf("verified: %s", v.Outcome)
	if v.Source != "" {
		title += fmt.Sprintf(" — %s", v.Source)
	}
	fmt.Fprintf(b, "<span class=\"%s\" title=\"%s\">%s</span>\n",
		class, html.EscapeString(title), html.EscapeString(v.Outcome))
}

// renderEvidenceList writes an expandable evidence list under a measure
// card showing the rolled-up results that determined its verdict: case,
// outcome, source file, and the synthesized test case's rendered Markdown
// description. Synthesized case descriptions are rendered through the same
// goldmark pipeline as authored prose.
func renderEvidenceList(b *bufio.Writer, req *model.Node, verdicts map[string]VerdictInfo, isChild bool) {
	if verdicts == nil {
		return
	}
	v, ok := verdicts[req.ID]
	if !ok || len(v.Evidence) == 0 {
		return
	}
	detailsClass := "verdict-evidence"
	if isChild {
		detailsClass += " verdict-evidence--child"
	}
	fmt.Fprintf(b, "<details class=\"%s\">\n<summary>Evidence (%d)</summary>\n", detailsClass, len(v.Evidence))
	for _, it := range v.Evidence {
		b.WriteString("<div class=\"evidence-item\">\n")
		fmt.Fprintf(b, "<span class=\"evidence-outcome evidence-outcome--%s\">%s</span>\n",
			html.EscapeString(it.Outcome), html.EscapeString(it.Outcome))
		if it.Case != "" {
			fmt.Fprintf(b, "<code class=\"evidence-case\">%s</code>\n", html.EscapeString(it.Case))
		}
		if it.File != "" {
			fmt.Fprintf(b, "<span class=\"evidence-file\">%s</span>\n", html.EscapeString(it.File))
		}
		if it.Description != "" {
			fmt.Fprintf(b, "<div class=\"evidence-desc\">\n%s\n</div>\n", renderMarkdown(it.Description))
		}
		b.WriteString("</div>\n")
	}
	b.WriteString("</details>\n")
}

// renderAttrGrid writes the <dl class="req-attrs"> attribute grid,
// iterating propOrder and emitting one <div class="attr"> per present
// attribute. The trace attribute is skipped (shown in the traces
// section). The req-attrs--child modifier is selected by isChild.
func renderAttrGrid(b *bufio.Writer, req *model.Node, propOrder []string, isChild bool) {
	dlClass := "req-attrs"
	if isChild {
		dlClass = "req-attrs req-attrs--child"
	}
	hasAttrs := false
	for _, prop := range propOrder {
		if prop == model.AttrTrace {
			continue // trace shown in traces section
		}
		if val, ok := req.Attrs[prop]; ok {
			if !hasAttrs {
				fmt.Fprintf(b, "<dl class=\"%s\">\n", dlClass)
				hasAttrs = true
			}
			label := formatAttrLabel(prop)
			valHTML := formatAttrHTML(val)
			fmt.Fprintf(b, "<div class=\"attr\">\n<dt>%s</dt>\n<dd>%s</dd>\n</div>\n", html.EscapeString(label), valHTML)
		}
	}
	if hasAttrs {
		b.WriteString("</dl>\n")
	}
}

// renderRationale writes the rationale block when the requirement has
// rationale text. The req-rationale--child modifier is selected by
// isChild.
func renderRationale(b *bufio.Writer, req *model.Node, isChild bool) {
	if req.Rationale == "" {
		return
	}
	rationaleClass := "req-rationale"
	if isChild {
		rationaleClass = "req-rationale req-rationale--child"
	}
	renderedRationale := renderMarkdown(req.Rationale)
	fmt.Fprintf(b, "<div class=\"%s\">\n<strong>Rationale:</strong> %s\n</div>\n", rationaleClass, renderedRationale)
}

// renderTraceLinks writes the upstream/downstream trace section when
// the requirement has at least one link and tc is non-nil. The
// trace-link--child modifier is selected by isChild.
func renderTraceLinks(b *bufio.Writer, req *model.Node, traceCache map[string]tracePair, tc *TraceResolver, isChild bool) {
	if traceCache == nil {
		return
	}
	pair := traceCache[req.ID]
	if len(pair.up)+len(pair.down) == 0 {
		return
	}
	linkClass := "trace-link"
	if isChild {
		linkClass = "trace-link trace-link--child"
	}
	renderTraces(b, pair.up, pair.down, tc, linkClass)
}

// renderTraces writes the upstream/downstream trace columns. The link
// class is supplied by the caller ("trace-link" for parents, plus
// "trace-link--child" for child cards).
func renderTraces(b *bufio.Writer, up, down []traceLink, tc *TraceResolver, linkClass string) {
	b.WriteString("<div class=\"req-traces\">\n")
	b.WriteString("<div class=\"trace-columns\">\n")

	b.WriteString("<div class=\"trace-column\">\n")
	fmt.Fprintf(b, "<div class=\"trace-column-header\">Upstream (%d)</div>\n", len(up))
	for _, link := range up {
		href := html.EscapeString(tc.linkHref(link.ID))
		escID := html.EscapeString(link.ID)
		if link.Title != "" {
			escTitle := html.EscapeString(link.Title)
			fmt.Fprintf(b, "<a class=\"%s\" href=\"%s\"><span class=\"trace-link-id\">%s</span><span class=\"trace-link-title\">%s</span></a>\n", linkClass, href, escID, escTitle)
		} else {
			fmt.Fprintf(b, "<a class=\"%s\" href=\"%s\"><span class=\"trace-link-id\">%s</span></a>\n", linkClass, href, escID)
		}
	}
	b.WriteString("</div>\n")

	b.WriteString("<div class=\"trace-column\">\n")
	fmt.Fprintf(b, "<div class=\"trace-column-header\">Downstream (%d)</div>\n", len(down))
	for _, link := range down {
		href := html.EscapeString(tc.linkHref(link.ID))
		escID := html.EscapeString(link.ID)
		if link.Title != "" {
			escTitle := html.EscapeString(link.Title)
			fmt.Fprintf(b, "<a class=\"%s\" href=\"%s\"><span class=\"trace-link-id\">%s</span><span class=\"trace-link-title\">%s</span></a>\n", linkClass, href, escID, escTitle)
		} else {
			fmt.Fprintf(b, "<a class=\"%s\" href=\"%s\"><span class=\"trace-link-id\">%s</span></a>\n", linkClass, href, escID)
		}
	}
	b.WriteString("</div>\n")

	b.WriteString("</div>\n")
	b.WriteString("</div>\n")
}

// renderDocChain writes the optional Confluence-Flow nav block that
// shows the upstream/downstream doc chain for the current document.
// Renders nothing when the graph has no upstream and no downstream
// tiers.
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
	totalTiers := len(g.Upstream) + len(g.Downstream)
	isDeep := totalTiers > 6

	b.WriteString("<nav class=\"doc-chain")
	if isDeep {
		b.WriteString(" doc-chain--deep")
	}
	b.WriteString("\" aria-label=\"Requirement traceability chain\">\n")

	b.WriteString("  <details class=\"chain-details\"")
	if !isDeep {
		b.WriteString(" open")
	}
	b.WriteString(">\n")

	b.WriteString("    <summary class=\"chain-summary\">\n")
	b.WriteString("      <span class=\"chain-summary__label\">Traceability Chain</span>\n")
	b.WriteString("      <span class=\"chain-summary__count\">")
	fmt.Fprintf(b, "%d upstream", totalUp)
	if totalDown > 0 {
		fmt.Fprintf(b, " · %d downstream", totalDown)
	}
	b.WriteString("</span>\n")
	b.WriteString("    </summary>\n")

	renderChainGraph(b, g)

	b.WriteString("  </details>\n")
	b.WriteString("</nav>\n")
}

// renderDocHead writes the <!DOCTYPE> + <html> + <head>...</head> block,
// inlining the embedded CSS and loading Alpine.js from CDN.
func renderDocHead(b *bufio.Writer, title string) {
	b.WriteString("<!DOCTYPE html>\n")
	b.WriteString("<html lang=\"en\">\n")
	b.WriteString("<head>\n")
	b.WriteString("<meta charset=\"UTF-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	fmt.Fprintf(b, "<title>%s</title>\n", html.EscapeString(title))
	b.WriteString("<style>\n")
	b.WriteString(css)
	b.WriteString("</style>\n")
	// Alpine.js — powers the search/filter toolbar, theme toggle, and mobile TOC.
	// Loaded with `defer` so it runs after the DOM is parsed and parsed
	// before `DOMContentLoaded`.
	b.WriteString("<script defer src=\"https://cdn.jsdelivr.net/npm/alpinejs@3.14.3/dist/cdn.min.js\"></script>\n")
	b.WriteString("</head>\n")
}

// renderBodyOpen writes the <body ...> element, the sidebar TOC
// placeholder, the mobile toggle button, the content-wrapper div, and
// the theme toggle. The Alpine `reqmdApp()` component owns runtime
// state for all of these.
func renderBodyOpen(b *bufio.Writer) {
	b.WriteString("<body x-data=\"reqmdApp()\" x-init=\"initApp()\" @keydown.escape.window=\"tocOpen = false\">\n")

	b.WriteString(`<aside class="req-toc" id="req-toc" :class="{'req-toc--open': tocOpen}" @click.away="tocOpen = false">
  <div class="req-toc-header">
    <div class="req-toc-title">Requirements</div>
    <div class="req-toc-count" id="req-toc-count"></div>
  </div>
  <div class="req-toc-list" id="req-toc-list">
  </div>
</aside>
`)

	b.WriteString(`<button id="toc-toggle-btn" class="toc-toggle-btn" aria-label="Toggle requirements list" @click="tocOpen = !tocOpen">&#9776;</button>
`)

	b.WriteString("<div class=\"content-wrapper\">\n")

	b.WriteString(`<button id="theme-toggle" class="theme-toggle" aria-label="Toggle theme" title="Toggle light/dark mode"
  @click="toggleTheme()" x-text="theme === 'dark' ? '☀️' : '🌙'">🌙</button>
`)
}

// renderBodyClose writes the closing tags for content-wrapper,
// body, and html.
func renderBodyClose(b *bufio.Writer) {
	b.WriteString("</div>\n") // end content-wrapper
	b.WriteString("</body>\n")
	b.WriteString("</html>\n")
}

// renderScripts writes the embedded JS, the requirement index JSON
// (consumed by Alpine and the sidebar TOC builder), and the filter
// options JSON (consumed by Alpine for the toolbar).
func renderScripts(b *bufio.Writer, rd *renderData) {
	b.WriteString("<script>\n")
	b.WriteString(js)
	b.WriteString("</script>\n")

	indexJSON, _ := json.Marshal(rd.Index)
	fmt.Fprintf(b, "<script type=\"application/json\" id=\"req-index\">%s</script>\n", indexJSON)

	filterJSON, _ := json.Marshal(rd.Filters)
	fmt.Fprintf(b, "<script type=\"application/json\" id=\"req-filter-index\">%s</script>\n", filterJSON)
}
