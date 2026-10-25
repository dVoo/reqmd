package exporter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dVoo/reqmd/internal/model"
)

// filterSet is the per-attribute sorted unique values collected across
// all requirements, used to populate the search toolbar filter menus.
type filterSet map[string][]string

// pageLink is one trace link in the client model. Href is resolved for the
// page that embeds the model; Unresolved links carry the authored Ref only.
type pageLink struct {
	ID        string `json:"id,omitempty"`
	Ref       string `json:"ref"`
	Title     string `json:"title,omitempty"`
	Href      string `json:"href,omitempty"`
	Direction string `json:"direction,omitempty"`
	Pin       int    `json:"pin,omitempty"`
	Version   int    `json:"version,omitempty"`
	Pinned    bool   `json:"pinned,omitempty"`
	Resolved  bool   `json:"resolved"`
}

// pageNode is one content node in the shared view model. Body and rationale
// are deliberately absent: the server-rendered Document view is their single
// representation and analytical views read them from the DOM.
type pageNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type,omitempty"` // omitted for requirements
	Title    string         `json:"title,omitempty"`
	Attrs    map[string]any `json:"attrs,omitempty"`
	Coverage string         `json:"coverage,omitempty"`
	Verdict  string         `json:"verdict,omitempty"`
	Eligible bool           `json:"eligible,omitempty"`
	Drafts   int            `json:"drafts,omitempty"`
	Up       []pageLink     `json:"up,omitempty"`
	Down     []pageLink     `json:"down,omitempty"`
}

// pageDocMeta carries the document facts the client toolbar and views show.
type pageDocMeta struct {
	Title      string `json:"title"`
	Dir        string `json:"dir"`
	DocumentID string `json:"documentId,omitempty"`
	Level      string `json:"level,omitempty"`
}

// pageModel is the single view model for one exported document page. Every
// analytical view, count, aggregate, and outline is derived from it, and the
// filters are the only filtering input.
type pageModel struct {
	Doc           pageDocMeta `json:"doc"`
	Filters       filterSet   `json:"filters"`
	DefaultStatus string      `json:"defaultStatus,omitempty"`
	StatusEnabled bool        `json:"statusEnabled"`
	HasResults    bool        `json:"hasResults"`
	TableCols     []string    `json:"attributeColumns,omitempty"`
	Nodes         []pageNode  `json:"nodes"`
}

// renderData is the single O(n) intermediate produced by buildRenderData: it
// powers the server-rendered TOC, the filter menus, and the view model.
type renderData struct {
	Filters         filterSet
	StatusBreakdown []statusBreakdownItem
	Anchors         map[*model.Node]string
	Nodes           []*model.Node
}

// buildRenderData runs a single O(n) pass over the document content tree to
// produce the per-document data the rendering helpers need: the filter set,
// the per-status breakdown, and the stable DOM anchors.
//
// Every node gets a stable anchor: requirements anchor on their ID,
// containers and info items on a generated "info-N" anchor (N = document
// order sequence). The same anchors are used by the TOC and by the rendered
// DOM so scroll-spy stays consistent.
func buildRenderData(doc model.Document) *renderData {
	reqs := doc.Requirements()
	rd := &renderData{
		Filters: buildFilters(reqs, doc),
		Anchors: make(map[*model.Node]string, len(reqs)),
		Nodes:   doc.Nodes,
	}

	// Stable DOM anchor for every node.
	itemSeq := 0
	var assign func(nodes []*model.Node)
	assign = func(nodes []*model.Node) {
		for _, n := range nodes {
			if n.Kind == model.KindRequirement {
				rd.Anchors[n] = n.ID
			} else {
				itemSeq++
				rd.Anchors[n] = fmt.Sprintf("item-%d", itemSeq)
			}
			assign(n.Children)
		}
	}
	assign(doc.Nodes)

	if doc.XReqmd == nil || !doc.XReqmd.IgnoreStatus {
		var additions []string
		if doc.XReqmd != nil {
			additions = doc.XReqmd.AdditionalStatusValues
		}
		rd.StatusBreakdown = countStatusBreakdown(reqs, additions)
	}
	return rd
}

// buildPageModel turns the document tree and graph analysis into the shared
// view model. resolveLink maps a requirement ID to a cross-document href; a
// nil resolver yields same-page fragments. propOrder is the schema property
// order used to order the table's attribute columns.
func buildPageModel(doc model.Document, rd *renderData, an *Analysis, resolveLink func(string) string, propOrder []string) pageModel {
	p := pageMeta(doc, rd, an, propOrder)
	href := modelHref(resolveLink)
	var walk func(nodes []*model.Node)
	walk = func(nodes []*model.Node) {
		for _, n := range nodes {
			p.Nodes = append(p.Nodes, pageNodeFor(n, rd.Anchors[n], an, href))
			walk(n.Children)
		}
	}
	walk(rd.Nodes)
	return p
}

// pageMeta fills the document-level model fields.
func pageMeta(doc model.Document, rd *renderData, an *Analysis, propOrder []string) pageModel {
	p := pageModel{
		Doc: pageDocMeta{
			Title: docTitle(doc),
			Dir:   doc.Path,
		},
		Filters:       rd.Filters,
		StatusEnabled: doc.XReqmd == nil || !doc.XReqmd.IgnoreStatus,
		HasResults:    an != nil && an.HasResults,
		TableCols:     attributeColumns(doc, propOrder),
		Nodes:         make([]pageNode, 0, len(rd.Nodes)),
	}
	if p.StatusEnabled {
		p.DefaultStatus = model.StatusApproved
	}
	if doc.XReqmd != nil {
		p.Doc.DocumentID = doc.XReqmd.DocumentID
		p.Doc.Level = doc.XReqmd.Level
	}
	return p
}

// modelHref resolves cross-document links for the model, falling back to an
// empty href (the client uses the same-page anchor).
func modelHref(resolveLink func(string) string) func(string) string {
	return func(id string) string {
		if resolveLink == nil {
			return ""
		}
		return resolveLink(id)
	}
}

// pageNodeFor converts one content node, attaching graph-derived state for
// requirement nodes.
func pageNodeFor(n *model.Node, anchor string, an *Analysis, href func(string) string) pageNode {
	pn := pageNode{ID: anchor, Title: n.Title}
	if n.Kind != model.KindRequirement {
		pn.Type = n.Kind.String()
		return pn
	}
	pn.ID = n.ID
	pn.Attrs = attrsForModel(n.Attrs)
	if an == nil {
		return pn
	}
	if st, ok := an.Nodes[n.ID]; ok {
		pn.Coverage = st.Coverage
		pn.Verdict = st.Verdict
		pn.Eligible = st.VerdictEligible
		pn.Drafts = st.Drafts
		pn.Up = pageLinks(st.Up, href)
		pn.Down = pageLinks(st.Down, href)
	}
	return pn
}

// attributeColumns returns every attribute present on at least one
// requirement, in schema property order followed by undeclared extras A→Z.
// `trace` and `status` are excluded: both have dedicated presentation
// (trace link columns and the status chip/column).
func attributeColumns(doc model.Document, propOrder []string) []string {
	reqs := doc.Requirements()
	seen := make(map[string]struct{}, len(propOrder))
	cols := make([]string, 0, len(propOrder))
	for _, prop := range propOrder {
		if prop == model.AttrTrace || prop == model.AttrStatus {
			continue
		}
		seen[prop] = struct{}{}
		for _, req := range reqs {
			if _, ok := req.Attrs[prop]; ok {
				cols = append(cols, prop)
				break
			}
		}
	}
	var extra []string
	for _, req := range reqs {
		for k := range req.Attrs {
			if k == model.AttrTrace || k == model.AttrStatus {
				continue
			}
			if _, declared := seen[k]; !declared {
				seen[k] = struct{}{}
				extra = append(extra, k)
			}
		}
	}
	sort.Strings(extra)
	return append(cols, extra...)
}

// attrsForModel copies a requirement's attributes for the client model,
// dropping `trace` (represented by the Up links).
func attrsForModel(attrs map[string]any) map[string]any {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		if k == model.AttrTrace {
			continue
		}
		out[k] = v
	}
	return out
}

// pageLinks converts analysis links for JSON. Same-page links resolve to
// "#id", which needs no href in the model; links to another document keep
// their page-relative target so the client can navigate to them.
func pageLinks(links []LinkState, href func(string) string) []pageLink {
	if len(links) == 0 {
		return nil
	}
	out := make([]pageLink, len(links))
	for i, l := range links {
		target := l.Href
		if target == "" && href != nil {
			target = href(l.ID)
		}
		if strings.HasPrefix(target, "#") {
			target = ""
		}
		out[i] = pageLink{
			ID:        l.ID,
			Ref:       l.Ref,
			Title:     l.Title,
			Href:      target,
			Direction: l.Direction,
			Pin:       l.Pin,
			Version:   l.Version,
			Pinned:    l.Pinned,
			Resolved:  l.Resolved,
		}
	}
	return out
}

// sortedKeys returns the keys of a string-keyed map in sorted order.
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
