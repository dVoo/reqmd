package graph

import (
	"github.com/dVoo/reqmd/internal/model"
)

// Coverage states reported by CoverageStates. They describe what the
// requires-trace-from check concluded for a single requirement, so exports
// never infer coverage from link presence.
const (
	CoverageCovered     = "covered"
	CoverageMissing     = "missing-coverage"
	CoverageNotRequired = "not-required"
	CoverageUnresolved  = "unresolved"
)

// CoverageMiss is one unmet `requires-trace-from` expectation. Unknown is
// true when the expected source names no known document-id or level.
type CoverageMiss struct {
	From    string
	Unknown bool
}

// NodeCoverage is the presentation view of one requirement's upstream
// coverage: the same conclusion checkRequiresTraceFromCoverage reaches,
// without the CheckResult wrapper. Misses is empty unless State is
// CoverageMissing or CoverageUnresolved; Drafts counts in-filter inbound
// requirements ignored by the status gate.
type NodeCoverage struct {
	State  string
	Misses []CoverageMiss
	Drafts int
}

// TraceLink is one declared `trace` reference with its resolution and
// version-pin state. Resolved is false for dangling and ambiguous refs; ID
// is empty in that case. Direction is "outdated", "predated", or "" for a
// current/unpinned link, using the same skip rules as the version-pin check.
type TraceLink struct {
	Ref       string
	ID        string
	Pin       int
	Version   int
	Direction string
	Pinned    bool
	Resolved  bool
}

// NodeVerification is the presentation view of one requirement's
// verification state. Eligible marks the measures the missing-verdict check
// expects a result for; Outcome is empty when such a measure has none.
type NodeVerification struct {
	Outcome  string
	Source   string
	Evidence []EvidenceItem
	Eligible bool
}

// CoverageStates returns the coverage state of every requirement node.
func (g *Graph) CoverageStates() map[string]NodeCoverage {
	out := make(map[string]NodeCoverage, len(g.nodes))
	for id, node := range g.nodes {
		out[id] = g.coverageState(node)
	}
	return out
}

// TraceRefs returns the declared trace references of one requirement in
// source order, each resolved to its target and annotated with the resolved
// version and the authored version pin. References that do not resolve are
// returned with Resolved=false so the caller can surface link health.
func (g *Graph) TraceRefs(id string) []TraceLink {
	node, ok := g.nodes[id]
	if !ok || len(node.Refs) == 0 {
		return nil
	}
	links := make([]TraceLink, 0, len(node.Refs))
	for _, ref := range node.Refs {
		_, pin, pinned := model.StripPin(ref)
		link := TraceLink{Ref: ref, Pin: pin, Pinned: pinned}
		if targetID, ok := g.ResolveTraceRef(ref); ok {
			link.ID = targetID
			link.Resolved = true
			if target := g.nodes[targetID]; target != nil {
				link.Version = target.Version
			}
			if direction, comparable := g.pinStatus(node, targetID, pin, pinned); comparable {
				link.Direction = direction
			}
		}
		links = append(links, link)
	}
	return links
}

// VersionOf returns the declared `version` of a requirement, or 0.
func (g *Graph) VersionOf(id string) int {
	if node, ok := g.nodes[id]; ok {
		return node.Version
	}
	return 0
}

// VerificationStates returns the verification state of every non-synthetic
// requirement. Eligible follows the missing-verdict gate (approved measure
// with a `verify` attribute that is not deferred or rejected), so a
// presentation denominator matches the check's expectation set.
func (g *Graph) VerificationStates() map[string]NodeVerification {
	out := make(map[string]NodeVerification, len(g.nodes))
	for id, node := range g.nodes {
		if node.IsResult || node.Synthetic {
			continue
		}
		st := NodeVerification{Eligible: g.measureGate(node)}
		if nv, ok := g.rollupVerdict(id); ok {
			st.Outcome = nv.Outcome
			st.Source = nv.Source
			st.Evidence = nv.Evidence
		}
		out[id] = st
	}
	return out
}
