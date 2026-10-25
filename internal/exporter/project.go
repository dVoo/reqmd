package exporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"sort"
	"strings"

	"github.com/dVoo/reqmd/internal/graph"
	"github.com/dVoo/reqmd/internal/model"
)

// ProjectDoc is one document row on the project home page.
type ProjectDoc struct {
	Path         string
	Href         string
	Title        string
	Level        string
	DocumentID   string
	Requirements int
	Coverage     CoverageSummary
	Verdicts     VerdictSummary
	Links        LinkSummary
	IDs          []string
}

// ProjectIndex is the project-level home page: document discovery,
// project-wide health, actionable queues, and cross-document search.
type ProjectIndex struct {
	analysis *Analysis
	titleMap map[string]string
	docs     []ProjectDoc
}

// NewProjectIndex builds the project home model. hrefFor maps a document
// directory path to the exported page href.
func NewProjectIndex(docs []model.Document, an *Analysis, titleMap map[string]string, hrefFor func(string) string) *ProjectIndex {
	p := &ProjectIndex{analysis: an, titleMap: titleMap}
	for _, doc := range docs {
		ids := idsOf(doc)
		row := ProjectDoc{
			Path:         doc.Path,
			Href:         hrefFor(doc.Path),
			Title:        docTitle(doc),
			Requirements: len(ids),
			IDs:          ids,
		}
		if row.Title == "" {
			row.Title = doc.Path
		}
		if doc.XReqmd != nil {
			row.Level = doc.XReqmd.Level
			row.DocumentID = doc.XReqmd.DocumentID
		}
		if an != nil {
			row.Coverage = an.CoverageOf(ids)
			row.Verdicts = an.VerdictsOf(ids)
			row.Links = an.LinksOf(ids)
		}
		p.docs = append(p.docs, row)
	}
	sort.SliceStable(p.docs, func(i, j int) bool { return p.docs[i].Path < p.docs[j].Path })
	return p
}

// Export writes the project home page.
func (p *ProjectIndex) Export(w io.Writer) error {
	b := bufio.NewWriterSize(w, pageBufferSize)

	renderDocHead(b, "Requirements project")
	_, _ = b.WriteString("<body data-page=\"project\">\n")
	_, _ = b.WriteString(`<header class="app-bar">
  <a class="app-bar__brand" href="index.html"><span class="app-bar__logo" aria-hidden="true"></span>reqmd</a>
  <span class="app-bar__sep" aria-hidden="true">/</span>
  <span class="app-bar__title">Project</span>
  <div class="app-bar__tools">
    <button type="button" class="icon-btn" data-action="theme" aria-label="Toggle theme" title="Toggle light/dark">◐</button>
  </div>
</header>
<div class="shell shell--project">
<div class="project">
`)
	_, _ = b.WriteString("<header class=\"doc-head\"><h1>Requirements</h1>\n")
	_, _ = b.WriteString(`<p class="doc-head__facts">Document catalog, project-wide health, and every requirement that needs attention.</p></header>` + "\n")

	p.renderTiles(b)
	p.renderSearch(b)
	p.renderQueues(b)
	p.renderCatalog(b)

	_, _ = b.WriteString("</div>\n</div>\n")
	renderProjectJSON(b, p)
	renderScript(b)
	renderBodyClose(b)

	if err := b.Flush(); err != nil {
		return fmt.Errorf("flushing HTML output: %w", err)
	}
	return nil
}

// renderTiles writes the project-wide summary tiles.
func (p *ProjectIndex) renderTiles(b *bufio.Writer) {
	var cov CoverageSummary
	var ver VerdictSummary
	var links LinkSummary
	var reqs int
	for _, d := range p.docs {
		reqs += d.Requirements
		cov.Covered += d.Coverage.Covered
		cov.Expected += d.Coverage.Expected
		cov.Missing += d.Coverage.Missing
		cov.Unresolved += d.Coverage.Unresolved
		cov.Drafts += d.Coverage.Drafts
		ver.Eligible += d.Verdicts.Eligible
		ver.Pass += d.Verdicts.Pass
		ver.Fail += d.Verdicts.Fail
		ver.Skipped += d.Verdicts.Skipped
		ver.Inconclusive += d.Verdicts.Inconclusive
		ver.NoVerdict += d.Verdicts.NoVerdict
		links.Links += d.Links.Links
		links.Unresolved += d.Links.Unresolved
		links.Outdated += d.Links.Outdated
		links.Predated += d.Links.Predated
	}

	_, _ = b.WriteString("<section class=\"tiles\" aria-label=\"Project health\">\n")
	tile := func(label, value, detail, state string) {
		cls := "tile"
		if state != "" {
			cls += " tile--" + state
		}
		_, _ = fmt.Fprintf(b, `<div class="%s"><span class="tile__label">%s</span><span class="tile__value">%s</span><span class="tile__detail">%s</span></div>`+"\n",
			cls, html.EscapeString(label), html.EscapeString(value), html.EscapeString(detail))
	}
	tile("Requirements", fmt.Sprintf("%d", reqs), fmt.Sprintf("%d documents", len(p.docs)), "")

	coverageValue := "not required"
	if cov.Expected > 0 {
		coverageValue = fmt.Sprintf("%d of %d", cov.Covered, cov.Expected)
	}
	state := ""
	if cov.Missing > 0 || cov.Unresolved > 0 {
		state = "attention"
	}
	tile("Coverage", coverageValue,
		fmt.Sprintf("%d missing · %d unresolved · %d draft downstreams ignored", cov.Missing, cov.Unresolved, cov.Drafts), state)

	verValue := "no results"
	if ver.Eligible > 0 {
		verValue = fmt.Sprintf("%d of %d pass", ver.Pass, ver.Eligible)
	}
	// An approval waiting on a recorded verdict is attention, not failure;
	// only an actual failing verdict is a hard problem.
	state = ""
	if ver.Fail > 0 {
		state = "attention"
	} else if ver.NoVerdict > 0 {
		state = "pending"
	}
	tile("Verification", verValue,
		fmt.Sprintf("%d fail · %d no verdict · %d inconclusive · %d skipped", ver.Fail, ver.NoVerdict, ver.Inconclusive, ver.Skipped), state)

	linkDetail := fmt.Sprintf("%d links · %d unresolved · %d stale", links.Links, links.Unresolved, links.Outdated+links.Predated)
	state = ""
	if links.Unresolved > 0 || links.Outdated+links.Predated > 0 {
		state = "attention"
	}
	tile("Trace links", fmt.Sprintf("%d", links.Links), linkDetail, state)
	_, _ = b.WriteString("</section>\n")
}

// renderSearch writes the cross-document search input and result list.
func (p *ProjectIndex) renderSearch(b *bufio.Writer) {
	_, _ = b.WriteString(`<section class="project-search" aria-label="Search requirements">
  <h2>Search requirements</h2>
  <input type="search" class="search" data-project-search placeholder="Search ID or title across every document…" autocomplete="off" spellcheck="false" aria-label="Search requirements">
  <ul class="results" data-project-results></ul>
</section>
`)
}

// queueItem is one actionable finding on the project page.
type queueItem struct {
	DocID string
	Href  string
	ID    string
	Title string
	Note  string
}

// queue builds one queue across every document, capped to keep the page
// bounded. The full count is returned so the section can say what was cut.
func (p *ProjectIndex) queue(spec queueSpec) (items []queueItem, total int) {
	const cap = 50
	for _, doc := range p.docs {
		for _, id := range doc.IDs {
			// Be explicit: a missing analysis entry must not be treated as a
			// zero NodeState, which some matchers would accept.
			st, known := p.analysis.Nodes[id]
			if !known {
				continue
			}
			note, ok := spec.match(st)
			if !ok {
				continue
			}
			total++
			if len(items) >= cap {
				continue
			}
			href := doc.Href + "#" + id
			if spec.query != "" {
				href = doc.Href + "?" + spec.query + "#" + id
			}
			items = append(items, queueItem{
				DocID: strings.TrimSuffix(filepathBase(doc.Href), ".html"),
				Href:  href,
				ID:    id,
				Title: p.titleMap[id],
				Note:  note,
			})
		}
	}
	return items, total
}

// filepathBase returns the last path element, tolerating URL separators.
func filepathBase(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}

// renderQueues writes the actionable exception lists.
func (p *ProjectIndex) renderQueues(b *bufio.Writer) {
	_, _ = b.WriteString("<section aria-label=\"Actionable findings\">\n")
	_, _ = b.WriteString("<h2>Needs attention</h2>\n")
	_, _ = b.WriteString(`<div class="queues">` + "\n")
	for _, spec := range projectQueueSpecs {
		p.renderQueue(b, spec)
	}
	_, _ = b.WriteString("</div>\n</section>\n")
}

// queueSpec is one actionable queue: a title, an explanation, and the matcher
// that decides which requirement states belong to it.
type queueSpec struct {
	title string
	body  string
	match func(st NodeState) (note string, ok bool)
	// query is the URL state appended to each item link so the destination
	// page opens with the same exception context the queue was built from.
	query string
}

// projectQueueSpecs is the fixed set of exception queues on the project home.
var projectQueueSpecs = []queueSpec{
	{
		title: "Missing coverage",
		body:  "Requirements whose expected downstream trace is missing or comes only from draft requirements.",
		match: missingCoverageNote,
		query: "view=document&f=state.coverage%3Amissing-coverage",
	},
	{
		title: "Unresolved coverage sources",
		body:  "Requirements expecting coverage from a document-id or level that does not exist.",
		match: unresolvedCoverageNote,
		query: "view=document&f=state.coverage%3Aunresolved",
	},
	{
		title: "Missing verdicts",
		body:  "Approved verification measures with no verification result.",
		match: missingVerdictNote,
		query: "view=coverage",
	},
	{
		title: "Failing verification",
		body:  "Approved measures whose rolled-up verdict is fail.",
		match: failingVerdictNote,
		query: "view=coverage",
	},
	{
		title: "Broken and stale traces",
		body:  "Unresolved trace references, outdated version pins, and predated pins.",
		match: brokenTraceNote,
		query: "view=coverage",
	},
}

// missingCoverageNote matches requirements whose coverage expectation is not
// met by an approved downstream requirement.
func missingCoverageNote(st NodeState) (string, bool) {
	if st.Coverage != graph.CoverageMissing {
		return "", false
	}
	note := strings.Join(st.Misses, ", ")
	if len(st.Unknown) > 0 {
		note = strings.Join(append(st.Misses, st.Unknown...), ", ")
	}
	return note, true
}

// unresolvedCoverageNote matches requirements expecting coverage from an
// unknown document-id or level.
func unresolvedCoverageNote(st NodeState) (string, bool) {
	if st.Coverage != graph.CoverageUnresolved {
		return "", false
	}
	return strings.Join(st.Unknown, ", "), true
}

// missingVerdictNote matches eligible measures with no result.
func missingVerdictNote(st NodeState) (string, bool) {
	if !st.VerdictEligible || st.Verdict != "" {
		return "", false
	}
	return "no verdict", true
}

// failingVerdictNote matches measures whose rolled-up verdict is fail.
func failingVerdictNote(st NodeState) (string, bool) {
	if st.Verdict != "fail" {
		return "", false
	}
	return st.VerdictSource, true
}

// brokenTraceNote matches requirements with unresolved, outdated, or predated
// outbound trace links.
func brokenTraceNote(st NodeState) (string, bool) {
	var notes []string
	for _, l := range st.Up {
		switch {
		case !l.Resolved:
			notes = append(notes, "unresolved: "+l.Ref)
		case l.Direction == graph.DirOutdated:
			notes = append(notes, fmt.Sprintf("outdated pin %s~%d (upstream v%d)", l.ID, l.Pin, l.Version))
		case l.Direction == graph.DirPredated:
			notes = append(notes, fmt.Sprintf("predated pin %s~%d (upstream v%d)", l.ID, l.Pin, l.Version))
		}
	}
	return strings.Join(notes, "; "), len(notes) > 0
}

// renderQueue writes one queue section, capped so the page stays bounded.
func (p *ProjectIndex) renderQueue(b *bufio.Writer, spec queueSpec) {
	items, total := p.queue(spec)
	if total == 0 {
		return
	}
	_, _ = fmt.Fprintf(b, `<div class="queue"><h3 class="queue__title">%s <span class="queue__count">%d</span></h3><p class="queue__body">%s</p>`+"\n",
		html.EscapeString(spec.title), total, html.EscapeString(spec.body))
	if len(items) == 0 {
		_, _ = b.WriteString("</div>\n")
		return
	}
	_, _ = b.WriteString("<ul class=\"queue__list\">\n")
	for _, it := range items {
		_, _ = fmt.Fprintf(b, `<li><a href="%s"><code>%s</code> <span class="queue__req-title">%s</span></a><span class="queue__note">%s</span></li>`+"\n",
			html.EscapeString(it.Href), html.EscapeString(it.ID), html.EscapeString(it.Title), html.EscapeString(it.Note))
	}
	_, _ = b.WriteString("</ul>\n")
	if total > len(items) {
		_, _ = fmt.Fprintf(b, `<p class="queue__more">Showing the first %d of %d.</p>`+"\n", len(items), total)
	}
	_, _ = b.WriteString("</div>\n")
}

// renderCatalog writes the document catalog table.
func (p *ProjectIndex) renderCatalog(b *bufio.Writer) {
	_, _ = b.WriteString(`<section class="catalog" aria-label="Documents">
  <h2>Documents</h2>
  <div class="table-scroll">
  <table class="grid" id="catalog">
    <thead><tr><th scope="col">Document</th><th scope="col">Level</th><th scope="col" class="num">Requirements</th><th scope="col" class="num" title="Covered of requirements that expect coverage">Coverage</th><th scope="col" class="num" title="Eligible measures with a verdict">Verified</th><th scope="col" class="num" title="Unresolved trace references">Broken</th></tr></thead>
    <tbody>
`)
	for _, d := range p.docs {
		covered := "—"
		if d.Coverage.Expected > 0 {
			covered = fmt.Sprintf("%d/%d", d.Coverage.Covered, d.Coverage.Expected)
		}
		verified := "—"
		if d.Verdicts.Eligible > 0 {
			verified = fmt.Sprintf("%d/%d", d.Verdicts.Pass, d.Verdicts.Eligible)
			if d.Verdicts.Fail > 0 {
				verified += fmt.Sprintf(" · %d fail", d.Verdicts.Fail)
			}
		}
		_, _ = fmt.Fprintf(b, `<tr data-href="%s"><td class="cell-title"><a href="%s">%s</a><span class="cell-sub">%s</span></td><td>%s</td><td class="num">%d</td><td class="num%s">%s</td><td class="num%s">%s</td><td class="num%s">%d</td></tr>`+"\n",
			html.EscapeString(d.Href),
			html.EscapeString(d.Href), html.EscapeString(d.Title), html.EscapeString(filepathBase(d.Path)),
			html.EscapeString(d.Level), d.Requirements,
			attentionClass(d.Coverage.Missing+d.Coverage.Unresolved > 0), covered,
			attentionClass(d.Verdicts.Fail > 0), verified,
			attentionClass(d.Links.Unresolved > 0), d.Links.Unresolved)
	}
	_, _ = b.WriteString("    </tbody>\n  </table>\n  </div>\n</section>\n")
}

// attentionClass returns a modifier class when a count should stand out.
func attentionClass(bad bool) string {
	if bad {
		return " cell--attention"
	}
	return ""
}

// projectIndexCapacity pre-sizes the cross-document search index.
const projectIndexCapacity = 512

// renderProjectJSON writes the cross-document requirement index used by the
// project search.
func renderProjectJSON(b *bufio.Writer, p *ProjectIndex) {
	type entry struct {
		ID    string `json:"id"`
		Title string `json:"title,omitempty"`
		Href  string `json:"href"`
		Doc   string `json:"doc"`
	}
	entries := make([]entry, 0, projectIndexCapacity)
	for _, d := range p.docs {
		for _, id := range d.IDs {
			entries = append(entries, entry{ID: id, Title: p.titleMap[id], Href: d.Href + "#" + id, Doc: d.Title})
		}
	}
	data, _ := json.Marshal(entries)
	_, _ = fmt.Fprintf(b, "<script type=\"application/json\" id=\"project-reqs\">%s</script>\n", data)
}
