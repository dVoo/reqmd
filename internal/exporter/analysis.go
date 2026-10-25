package exporter

import (
	"sort"
	"strings"

	"github.com/dVoo/reqmd/internal/graph"
	"github.com/dVoo/reqmd/internal/model"
)

// LinkState is one trace link as the views present it. Title is the target's
// human-readable title; Pin/Direction describe the authored version pin and
// its health against the target's current Version.
type LinkState struct {
	ID    string
	Ref   string
	Title string
	// Href is the cross-document target of the linked requirement, filled in
	// by the page that links to it (empty for same-page links).
	Href      string
	Direction string
	Pin       int
	Version   int
	Pinned    bool
	Resolved  bool
}

// NodeState is the presentation state of one requirement, adapted from graph
// analysis. Views render these states instead of inferring coverage or
// verification from link presence.
type NodeState struct {
	Coverage        string // graph.Coverage* value
	Misses          []string
	Unknown         []string
	Drafts          int
	Verdict         string // pass|fail|skipped|inconclusive, empty for none
	VerdictEligible bool
	VerdictSource   string
	Evidence        []EvidenceItem
	Up              []LinkState // declared trace refs (what this traces to)
	Down            []LinkState // reverse edges (what traces to this)
}

// Analysis is the semantic layer shared by every HTML view: coverage,
// verification, and trace link health adapted from the graph so the export
// cannot disagree with reqmd check.
type Analysis struct {
	Nodes      map[string]NodeState
	HasResults bool
	// hrefFor resolves cross-document links for requirements that live on
	// another page. A nil resolver leaves those links without a target, which
	// the client then renders as plain text.
	hrefFor func(string) string
}

// SetHrefResolver lets the analysis attach cross-document targets to links
// whose requirement lives on another page (the client cannot build them for
// nodes it never received).
func (a *Analysis) SetHrefResolver(fn func(string) string) {
	a.hrefFor = fn
}

// NewAnalysis adapts graph results into presentation state. titleMap supplies
// link titles; missing entries fall back to the bare ID.
func NewAnalysis(g *graph.Graph, titleMap map[string]string) *Analysis {
	a := &Analysis{
		Nodes:      make(map[string]NodeState, g.NodeCount()),
		HasResults: g.HasResults(),
	}
	real := realRequirementIDs(g)
	coverage := g.CoverageStates()
	verification := g.VerificationStates()

	for _, doc := range g.Documents() {
		if doc.Synthetic {
			continue
		}
		for _, req := range doc.Requirements() {
			if !real[req.ID] {
				continue
			}
			a.Nodes[req.ID] = nodeState(coverage[req.ID], verification[req.ID], real, titleMap, g.TraceRefs(req.ID))
		}
	}
	a.mirrorDownstream(real, titleMap)
	return a
}

// realRequirementIDs returns the IDs of authored requirements. Synthetic nodes
// (RESULT: pseudo-requirements and synthesized test cases) are evidence, not
// requirements: they have no page and must not appear as trace links.
func realRequirementIDs(g *graph.Graph) map[string]bool {
	real := make(map[string]bool, g.NodeCount())
	for _, doc := range g.Documents() {
		if doc.Synthetic {
			continue
		}
		for _, req := range doc.Requirements() {
			if isSyntheticID(req.ID) {
				continue
			}
			real[req.ID] = true
		}
	}
	return real
}

// nodeState assembles one requirement's presentation state.
func nodeState(cov graph.NodeCoverage, ver graph.NodeVerification, real map[string]bool, titleMap map[string]string, links []graph.TraceLink) NodeState {
	st := NodeState{
		Coverage:        cov.State,
		Drafts:          cov.Drafts,
		Verdict:         ver.Outcome,
		VerdictEligible: ver.Eligible,
		VerdictSource:   ver.Source,
		Evidence:        evidenceItems(ver.Evidence),
		Up:              upLinks(links, real, titleMap),
	}
	for _, miss := range cov.Misses {
		if miss.Unknown {
			st.Unknown = append(st.Unknown, miss.From)
		} else {
			st.Misses = append(st.Misses, miss.From)
		}
	}
	return st
}

// evidenceItems converts graph evidence into the exporter's render type.
func evidenceItems(items []graph.EvidenceItem) []EvidenceItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]EvidenceItem, 0, len(items))
	for _, ev := range items {
		out = append(out, EvidenceItem{
			Case:        ev.Case,
			Outcome:     ev.Outcome,
			File:        ev.File,
			Description: ev.Description,
		})
	}
	return out
}

// upLinks keeps the authored trace refs that point at a real requirement (or
// are unresolved, so broken links stay visible).
func upLinks(links []graph.TraceLink, real map[string]bool, titleMap map[string]string) []LinkState {
	out := make([]LinkState, 0, len(links))
	for _, link := range links {
		if link.Resolved && !real[link.ID] {
			continue
		}
		out = append(out, LinkState{
			ID:        link.ID,
			Ref:       link.Ref,
			Title:     titleMap[link.ID],
			Direction: link.Direction,
			Pin:       link.Pin,
			Version:   link.Version,
			Pinned:    link.Pinned,
			Resolved:  link.Resolved,
		})
	}
	return out
}

// mirrorDownstream copies every resolved outbound link into its target's
// downstream list, keeping the authored pin and its health, then sorts each
// downstream list for deterministic output.
func (a *Analysis) mirrorDownstream(real map[string]bool, titleMap map[string]string) {
	for id, st := range a.Nodes {
		for _, link := range st.Up {
			if !link.Resolved || !real[link.ID] {
				continue
			}
			target, ok := a.Nodes[link.ID]
			if !ok {
				continue
			}
			down := LinkState{
				ID:        id,
				Ref:       link.Ref,
				Title:     titleMap[id],
				Direction: link.Direction,
				Pin:       link.Pin,
				Version:   link.Version,
				Pinned:    link.Pinned,
				Resolved:  true,
			}
			if a.hrefFor != nil {
				down.Href = a.hrefFor(id)
			}
			target.Down = append(target.Down, down)
			a.Nodes[link.ID] = target
		}
	}
	for id, st := range a.Nodes {
		sort.SliceStable(st.Down, func(i, j int) bool { return st.Down[i].ID < st.Down[j].ID })
		a.Nodes[id] = st
	}
}

// CoverageSummary counts the coverage states of the given requirements.
type CoverageSummary struct {
	Requirements int
	Expected     int // requirements whose coverage is expected (not not-required)
	Covered      int
	Missing      int
	Unresolved   int
	NotRequired  int
	Drafts       int
}

// CoverageOf summarizes the given requirement states.
func (a *Analysis) CoverageOf(ids []string) CoverageSummary {
	var s CoverageSummary
	for _, id := range ids {
		st, ok := a.Nodes[id]
		if !ok {
			continue
		}
		s.Requirements++
		s.Drafts += st.Drafts
		switch st.Coverage {
		case graph.CoverageCovered:
			s.Covered++
			s.Expected++
		case graph.CoverageMissing:
			s.Missing++
			s.Expected++
		case graph.CoverageUnresolved:
			s.Unresolved++
			s.Expected++
		default:
			s.NotRequired++
		}
	}
	return s
}

// VerdictSummary counts verification states over the eligible measures of the
// given requirements. Unverified eligible measures are counted as NoVerdict,
// so percentages always divide by the eligible total.
type VerdictSummary struct {
	Eligible     int
	Pass         int
	Fail         int
	Skipped      int
	Inconclusive int
	NoVerdict    int
}

// VerdictsOf summarizes the verification states of the given requirements.
func (a *Analysis) VerdictsOf(ids []string) VerdictSummary {
	var s VerdictSummary
	for _, id := range ids {
		st, ok := a.Nodes[id]
		if !ok || !st.VerdictEligible {
			continue
		}
		s.Eligible++
		switch st.Verdict {
		case "pass":
			s.Pass++
		case "fail":
			s.Fail++
		case "skipped":
			s.Skipped++
		case "inconclusive":
			s.Inconclusive++
		default:
			s.NoVerdict++
		}
	}
	return s
}

// LinkSummary counts trace link health across the given requirements.
type LinkSummary struct {
	Links      int
	Unresolved int
	Outdated   int
	Predated   int
}

// LinksOf summarizes the declared outbound trace links of the given
// requirements.
func (a *Analysis) LinksOf(ids []string) LinkSummary {
	var s LinkSummary
	for _, id := range ids {
		for _, l := range a.Nodes[id].Up {
			s.Links++
			if !l.Resolved {
				s.Unresolved++
			}
			switch l.Direction {
			case graph.DirOutdated:
				s.Outdated++
			case graph.DirPredated:
				s.Predated++
			}
		}
	}
	return s
}

// isSyntheticID reports whether an ID belongs to a synthesized node (result
// pseudo-requirement or test case) rather than an authored requirement.
func isSyntheticID(id string) bool {
	return strings.HasPrefix(id, "RESULT:") || strings.HasPrefix(id, "TC:")
}

// idsOf returns the given document's requirement IDs in document order.
func idsOf(doc model.Document) []string {
	reqs := doc.Requirements()
	ids := make([]string, len(reqs))
	for i, req := range reqs {
		ids[i] = req.ID
	}
	return ids
}
