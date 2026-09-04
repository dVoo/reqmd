// Package graph builds a pure Go in-memory adjacency graph from parsed
// requirement documents and runs Pass 2 trace validation checks: broken
// references, circular dependencies, requires-trace-from coverage,
// disposition, ID prefix, version-pin staleness, and outcome-gated
// checks (missing-verdict, failing-verdict) per the reqmd validation
// pipeline.
//
// The graph lives only for the duration of a single CLI invocation — no
// persistent database or temp files are used.
//
// Trace references resolve by document-id (preferred) or by bare
// requirement ID (when globally unique). Path-qualified references
// (e.g. "03-software/SW-001") are no longer supported — use document-id
// prefixed references (e.g. "software/SW-001") instead.
package graph

import (
	"fmt"
	"path/filepath"
	"reqmd/internal/model"
	"slices"
	"sort"
	"strings"
)

// Check result severity levels.
const (
	LevelError   = "ERROR"
	LevelWarning = "WARNING"
	LevelInfo    = "INFO"
)

// Machine-readable check codes and direction values for version-pin findings.
const (
	CodeVersionPin     = "version-pin"
	CodeMissingVerdict = "missing-verdict"
	CodeFailingVerdict = "failing-verdict"
	CodeVerdict        = "verdict"
	DirOutdated        = "outdated"
	DirPredated        = "predated"
)

// CheckResult describes a single Pass 2 trace validation finding.
type CheckResult struct {
	Level     string // one of LevelError, LevelWarning, LevelInfo
	Code      string // machine-readable check identifier (e.g. "version-pin"); empty for findings that don't need one
	Direction string // "outdated" | "predated" — only set for version-pin findings
	Outcome   string // verification verdict (pass|fail|skipped|inconclusive); only set for CodeVerdict
	ReqID     string
	File      string
	Message   string
}

// RepinDelta describes a single version-pin change that the repin
// command proposes or applies. One delta corresponds to one
// requirement's one trace ref. DeltaKind distinguishes "outdated"
// (pin < upstream.version) from "unpinned" (no pin against a
// versioned upstream) when promoteUnpinned is true. Predated
// findings (pin > upstream.version) are surfaced as DeltaPredated
// and never auto-fixed by repin.
type RepinDelta struct {
	Kind       string `json:"kind"`        // "outdated" | "unpinned" | "predated"
	ReqID      string `json:"req_id"`      // downstream requirement ID
	File       string `json:"file"`        // source file of the downstream requirement
	TargetID   string `json:"target_id"`   // upstream requirement ID being repinned
	SourceRef  string `json:"source_ref"`  // full ref as it appears in the source (e.g. "doc-id/ID~3" or "ID")
	OldPin     int    `json:"old_pin"`     // previous pin (0 for unpinned, also 0 for "~0")
	NewPin     int    `json:"new_pin"`     // new pin (== upstream.version)
	NewVersion int    `json:"new_version"` // upstream version this delta repins to
}

// CachedNode is an in-memory typed representation of a requirement node,
// populated once at graph creation. All check methods iterate this cache.
type CachedNode struct {
	OutboundPins         map[string]int
	suppressionsSet      map[string]struct{}
	OutboundRefs         map[string]string
	File                 string
	Dir                  string
	Disposition          string
	DispositionReason    string
	Verify               string
	Outcome              string
	Body                 string
	IDPrefix             string
	Status               string
	ReqID                string
	Inbound              []string
	Outbound             []string
	Suppressions         []string
	RequiresTraceFrom    []string
	Version              int
	MandatoryDisposition bool
	External             bool
	IsResult             bool
	IsChild              bool
	// Synthetic marks nodes that come from a synthetic document (produced
	// by internal/verify), not from authored Markdown on disk. Synthetic
	// nodes have no schema, no ID prefix, and no source file that tooling
	// may rewrite in place (repin skips them). RESULT: pseudo-requirements
	// are synthetic; future synthesized test-case nodes will be too.
	Synthetic bool
}

// Graph wraps an in-memory adjacency map with a typed node cache. It is
// built once from parsed documents and queried by Pass 2 check methods.
type Graph struct {
	docs             []model.Document          // original documents, kept for boundary inference in CheckResults
	nodes            map[string]*CachedNode    // reqID → CachedNode (first definition wins)
	nodesByIDDir     map[[2]string]*CachedNode // (docIDorDir, reqID) → CachedNode (used for both document-id and dir resolution)
	idDirs           map[string][]string       // reqID → list of directory paths containing it
	docByID          map[string]string         // document-id → directory path
	docByDir         map[string]string         // directory path → document-id
	docByLevel       map[string][]string       // level → list of directory paths
	dangling         []CheckResult
	duplicateIDs     []CheckResult   // duplicate requirement IDs across documents
	ignoreStatusDirs map[string]bool // dir path → true if document opts out of status lifecycle
	// Cached boundary maps computed once in CheckResults; RootDirs/LeafDirs
	// return these instead of recomputing inferBoundaries on every call.
	cachedRootDirs map[string]bool
	cachedLeafDirs map[string]bool
	// filter is the set of requirement IDs that match the active --filter.
	// nil means no filter is active (check everything — current behavior).
	// non-nil means only requirements in this set are checked/reported.
	filter map[string]struct{}
	// disjointAttrs names array-typed attributes whose values must overlap
	// between every trace-linked source and target. Empty = no disjoint check.
	disjointAttrs []string
	// hasResults is true when at least one synthesized result
	// pseudo-requirement (RESULT:*) is present. Set once in Pass 1 so the
	// outcome-gated checks short-circuit without re-scanning all nodes.
	hasResults bool
	// sevCache memoizes rolled-up evidence severity per requirement ID
	// within a check pass; invalidated whenever the active filter changes.
	sevCache map[string]severity
}

// SetFilter restricts per-requirement checks to the given set of requirement
// IDs. nil means no filter (check everything — the default); an empty
// non-nil set means "the filter matched nothing" — no requirement is
// checked. The distinction matters: a --filter expression that matches no
// requirements must scope the graph to the empty set, not reset to
// check-everything. The graph is always built from the full document tree
// so trace references to filtered-out requirements still resolve; only the
// per-req check/report loop is scoped.
func (g *Graph) SetFilter(ids map[string]struct{}) {
	g.filter = ids
	g.sevCache = nil // roll-up results depend on the active filter
}

// SetDisjointAttrs enables the disjoint-attribute check for the named
// array-typed attributes. Empty or nil disables it.
func (g *Graph) SetDisjointAttrs(attrs []string) {
	g.disjointAttrs = attrs
}

// inFilter reports whether a requirement ID is in the active filter set.
// Returns true when no filter is active (nil filter = check everything).
func (g *Graph) inFilter(reqID string) bool {
	if g.filter == nil {
		return true
	}
	_, ok := g.filter[reqID]
	return ok
}

// New builds a pure Go in-memory adjacency graph from all parsed documents.
// It creates one CachedNode per requirement and populates Inbound/Outbound
// adjacency from trace attributes. Broken trace references (target ID not
// found in any document) are collected as WARNING results.
//
// Trace reference resolution:
//   - "doc-id/ID" — prefix is resolved as a document-id first, falling back
//     to a directory path if no matching document-id is found.
//   - "ID" — resolved against the global ID map; emits an ERROR ambiguity
//     warning if the ID exists in multiple documents.
func New(docs []model.Document) (*Graph, error) {
	g := &Graph{
		docs:             docs,
		nodes:            make(map[string]*CachedNode),
		nodesByIDDir:     make(map[[2]string]*CachedNode),
		idDirs:           make(map[string][]string),
		docByID:          make(map[string]string),
		docByDir:         make(map[string]string),
		docByLevel:       make(map[string][]string),
		ignoreStatusDirs: make(map[string]bool),
	}

	// Build document-id → directory path map, level → directories map,
	// and the ignoreStatusDirs opt-out set.
	for _, doc := range docs {
		if doc.XReqmd == nil {
			continue
		}
		if doc.XReqmd.DocumentID != "" {
			g.docByID[doc.XReqmd.DocumentID] = doc.Path
			g.docByDir[doc.Path] = doc.XReqmd.DocumentID
		}
		if doc.XReqmd.Level != "" {
			g.docByLevel[doc.XReqmd.Level] = append(g.docByLevel[doc.XReqmd.Level], doc.Path)
		}
		if doc.XReqmd.IgnoreStatus {
			g.ignoreStatusDirs[doc.Path] = true
		}
	}

	// Pass 1: create one CachedNode per requirement
	for _, doc := range docs {
		var xr *model.XReqmd
		if doc.XReqmd != nil {
			xr = doc.XReqmd
		}
		for _, req := range doc.Requirements() {
			// Effective coverage expectation: the requirement's own
			// `requires-trace-from` attribute when declared; otherwise
			// inherit the document's x-reqmd.requires-trace-from default
			// when set. A nil result means neither declared anything, so
			// the generic boundary-inference fallback still applies.
			rtf := getStringSlice(req.Attrs, model.AttrRequiresTraceFrom)
			if rtf == nil && xr != nil && xr.RequiresTraceFrom != nil {
				rtf = slices.Clone(xr.RequiresTraceFrom)
			}
			node := &CachedNode{
				ReqID:             req.ID,
				File:              req.Source,
				Dir:               filepath.Dir(req.Source),
				IsChild:           req.ParentID != "",
				Disposition:       getString(req.Attrs, model.AttrDisposition),
				DispositionReason: getString(req.Attrs, model.AttrDispositionReason),
				Version:           getIntFromAttrs(req.Attrs, model.AttrVersion),
				RequiresTraceFrom: rtf,
				OutboundPins:      nil, // lazily allocated in Pass 2
				OutboundRefs:      nil, // lazily allocated in Pass 2
				Suppressions:      req.Suppressions,
				Status:            getString(req.Attrs, model.AttrStatus),
				Outcome:           getString(req.Attrs, model.AttrResultOutcome),
				Verify:            getString(req.Attrs, "verify"),
				Body:              req.Body,
				IsResult:          strings.HasPrefix(req.ID, "RESULT:"),
				Synthetic:         doc.Synthetic,
			}
			if len(req.Suppressions) > 0 {
				node.suppressionsSet = make(map[string]struct{}, len(req.Suppressions))
				for _, s := range req.Suppressions {
					node.suppressionsSet[s] = struct{}{}
				}
			}
			if xr != nil {
				node.External = xr.External
				node.MandatoryDisposition = xr.MandatoryDisposition
				node.IDPrefix = xr.IDPrefix
			}

			if existing, exists := g.nodes[req.ID]; exists {
				// Duplicate ID across documents — this is a data integrity error.
				// Two different requirements claiming the same ID causes undefined behavior.
				g.duplicateIDs = append(g.duplicateIDs, CheckResult{
					Level:   LevelError,
					ReqID:   req.ID,
					File:    req.Source,
					Message: fmt.Sprintf("duplicate requirement ID: already defined in %s", existing.File),
				})
				continue // skip overwrite — first definition wins
			}
			if node.IsResult {
				g.hasResults = true
			}
			g.nodes[req.ID] = node
			dirPath := node.Dir
			g.nodesByIDDir[[2]string{dirPath, req.ID}] = node
			g.idDirs[req.ID] = append(g.idDirs[req.ID], dirPath)
		}
	}

	// Pass 2: build Inbound/Outbound adjacency from trace attrs
	for _, doc := range docs {
		for _, req := range doc.Requirements() {
			traceVal, ok := req.Attrs[model.AttrTrace]
			if !ok {
				continue
			}
			traceList, ok := traceVal.([]any)
			if !ok {
				continue
			}
			node := g.nodes[req.ID]
			if node == nil {
				continue
			}
			for _, ref := range traceList {
				refStr, ok := ref.(string)
				if !ok {
					continue
				}
				// originalRef is the source-form ref exactly as written in
				// the attr block (e.g. "doc-id/ID~3" or "ID~3" or "ID").
				// We strip the pin below; the repin command needs the
				// unmutated form to drive text edits.
				originalRef := refStr

				// Strip optional version pin via the shared helper.
				var pin int
				var pinned bool
				refStr, pin, pinned = model.StripPin(refStr)

				var target *CachedNode
				var targetID string

				if strings.Contains(refStr, "/") {
					// Qualified reference: "<prefix>/<ID>"
					// Prefix is resolved as a document-id first, falling back to
					// directory path lookup.
					parts := strings.SplitN(refStr, "/", 2)
					prefix := parts[0]
					idPart := parts[1]

					if dir, ok := g.docByID[prefix]; ok {
						target = g.nodesByIDDir[[2]string{dir, idPart}]
					}
					if target == nil {
						target = g.nodesByIDDir[[2]string{prefix, idPart}]
					}
					if target == nil {
						g.dangling = append(g.dangling, CheckResult{
							Level:   LevelWarning,
							ReqID:   req.ID,
							File:    req.Source,
							Message: fmt.Sprintf("broken reference: target %q not found in %q", idPart, prefix),
						})
						continue
					}
					targetID = idPart
				} else {
					// Unqualified reference: "ID"
					target = g.nodes[refStr]
					if target == nil {
						g.dangling = append(g.dangling, CheckResult{
							Level:   LevelWarning,
							ReqID:   req.ID,
							File:    req.Source,
							Message: fmt.Sprintf("broken reference: target %q not found", refStr),
						})
						continue
					}
					// Ambiguous reference check: if ID exists in multiple directories
					dirs := g.idDirs[refStr]
					if len(dirs) > 1 {
						g.dangling = append(g.dangling, CheckResult{
							Level:   LevelError,
							ReqID:   req.ID,
							File:    req.Source,
							Message: fmt.Sprintf("ambiguous reference: %q exists in %d documents; use the qualified form %q/%s", refStr, len(dirs), g.firstDocID(dirs), refStr),
						})
						continue
					}
					targetID = refStr
				}
				node.Outbound = append(node.Outbound, targetID)
				target.Inbound = append(target.Inbound, req.ID)
				if pinned {
					if node.OutboundPins == nil {
						node.OutboundPins = make(map[string]int)
					}
					node.OutboundPins[targetID] = pin
				}
				// Record the original ref string for the repin command. Only
				// stored when the ref resolved to a real target (so dangling/
				// ambiguous refs don't pollute the map). First-wins semantics:
				// duplicate targetIDs in the same trace list keep the first form.
				if node.OutboundRefs == nil {
					node.OutboundRefs = make(map[string]string)
				}
				node.OutboundRefs[targetID] = originalRef
			}
		}
	}

	// Pass 3: build parent-child edges from heading hierarchy
	for _, doc := range docs {
		for _, req := range doc.Requirements() {
			if req.ParentID == "" {
				continue
			}
			node := g.nodes[req.ID]
			if node == nil {
				continue
			}
			target, exists := g.nodes[req.ParentID]
			if !exists {
				g.dangling = append(g.dangling, CheckResult{
					Level:   LevelError,
					ReqID:   req.ID,
					File:    req.Source,
					Message: fmt.Sprintf("parent requirement %q not found", req.ParentID),
				})
				continue
			}
			node.Outbound = append(node.Outbound, req.ParentID)
			target.Inbound = append(target.Inbound, req.ID)
		}
	}

	return g, nil
}

// firstDocID returns the first document-id associated with any of the given
// directory paths, or the first directory path itself if no document-id
// is declared.
func (g *Graph) firstDocID(dirs []string) string {
	for _, d := range dirs {
		if id, ok := g.docByDir[d]; ok {
			return id
		}
	}
	if len(dirs) == 0 {
		return ""
	}
	return dirs[0]
}

// inferBoundaries returns two maps describing the V-model topology:
//   - rootDirs: directories that are at the top (have no upstream sources,
//     or are declared external).
//   - leafDirs: directories that are at the bottom (are not referenced as
//     upstream sources by any other document).
//
// The returned maps are pure functions of the documents; callers should
// treat them as read-only.
func (g *Graph) inferBoundaries(docs []model.Document) (rootDirs, leafDirs map[string]bool) {
	rootDirs = make(map[string]bool)
	leafDirs = make(map[string]bool)

	referencedDirs := make(map[string]bool)
	for _, doc := range docs {
		if doc.XReqmd == nil || doc.XReqmd.Upstream == nil {
			continue
		}
		for _, src := range doc.XReqmd.Upstream.Sources {
			referencedDirs[src] = true
		}
	}

	for _, doc := range docs {
		// Root: no upstream sources declared (or no x-reqmd at all),
		// or external documents.
		isRoot := true
		if doc.XReqmd != nil {
			if doc.XReqmd.Upstream != nil && len(doc.XReqmd.Upstream.Sources) > 0 {
				isRoot = false
			}
			if doc.XReqmd.External {
				isRoot = true
			}
		}
		if isRoot {
			rootDirs[doc.Path] = true
		}

		// Leaf: not referenced as upstream by any other document.
		if !referencedDirs[doc.Path] {
			leafDirs[doc.Path] = true
		}
	}

	return rootDirs, leafDirs
}

// resolveCoverageSource resolves a coverage-expectation token (a document-id or a
// level name) to the set of directory paths that satisfy it. Returns
// nil/empty when nothing matches.
func (g *Graph) resolveCoverageSource(from string) []string {
	if dir, ok := g.docByID[from]; ok {
		return []string{dir}
	}
	if dirs, ok := g.docByLevel[from]; ok {
		// Return a copy to avoid mutating the map's slice.
		out := make([]string, len(dirs))
		copy(out, dirs)
		return out
	}
	return nil
}

// CheckResults runs all Pass 2 validation checks and returns the combined results.
// Checks performed:
//   - Broken reference (WARNING): trace target not found; also includes
//     duplicate ID and ambiguous reference errors collected during edge creation.
//   - Circular dependency (ERROR): cycles detected via DFS.
//   - Requires-trace-from coverage (WARNING): for each requirement that
//     declares `requires-trace-from`, verifies that at least one requirement
//     from each expected source (document-id or level) traces to it.
//   - Generic untraced / no-downstream (WARNING) — fallback for requirements
//     that do NOT declare `requires-trace-from`: the boundary inference
//     still applies so a single requirement stays clean when its document
//     is at the top or bottom of the V-model.
//   - Disposition without reason (WARNING).
//   - Mandatory disposition (ERROR).
//   - Version pin (ERROR): outdated (`V > N`) and predated (`V < N`) version
//     pins on trace references. Suppress per-node with
//     `reqmd-suppress: [version-pin]`. Code: "version-pin", Direction:
//     "outdated" or "predated".
//   - ID prefix (ERROR).
func (g *Graph) CheckResults() []CheckResult {
	var results []CheckResult
	results = append(results, g.duplicateIDs...)
	for _, d := range g.dangling {
		if node := g.nodes[d.ReqID]; node != nil && node.isSuppressed("broken-ref") {
			continue
		}
		results = append(results, d)
	}
	results = append(results, g.checkCircular()...)

	g.cachedRootDirs, g.cachedLeafDirs = g.inferBoundaries(g.docs)
	results = append(results, g.checkRequiresTraceFromCoverage(g.docs, g.cachedRootDirs, g.cachedLeafDirs)...)

	results = append(results, g.checkDispositionReason()...)
	results = append(results, g.checkMandatoryDisposition()...)
	results = append(results, g.checkVersionPins()...)
	results = append(results, g.checkIDPrefixes()...)
	results = append(results, g.checkMissingVerdict()...)
	results = append(results, g.checkFailingVerdict()...)
	results = append(results, g.checkVerdicts()...)
	results = append(results, g.checkDisjointAttribute()...)
	return results
}

// checkCircular detects cycles using DFS on the in-memory cached nodes.
// Returns ERROR results when a cycle is detected.
func (g *Graph) checkCircular() []CheckResult {
	var results []CheckResult
	visited := make(map[string]bool)
	onStack := make(map[string]bool)

	var dfs func(reqID string)
	dfs = func(reqID string) {
		if node := g.nodes[reqID]; node != nil && node.isSuppressed("circular") {
			visited[reqID] = true // mark visited so we don't re-process
			return
		}
		visited[reqID] = true
		onStack[reqID] = true
		node := g.nodes[reqID]
		if node != nil {
			for _, out := range node.Outbound {
				if !visited[out] {
					dfs(out)
				} else if onStack[out] {
					results = append(results, CheckResult{
						File:    node.File,
						ReqID:   reqID,
						Level:   LevelError,
						Message: fmt.Sprintf("circular dependency: chain involves %q", out),
					})
				}
			}
		}
		onStack[reqID] = false
	}

	for reqID := range g.nodes {
		if !visited[reqID] {
			dfs(reqID)
		}
	}
	return results
}

// checkRequiresTraceFromCoverage runs two related checks:
//  1. For each requirement that explicitly declares `requires-trace-from`
//     (even `requires-trace-from: []`), verify that at least one inbound
//     edge originates from each expected document-id or level. Unknown
//     expectations produce a WARNING.
//  2. For requirements that do NOT declare `requires-trace-from`, fall back
//     to the generic untraced / no-downstream checks using the V-model
//     boundary inference so a top-level or bottom-level document stays
//     clean.
func (g *Graph) checkRequiresTraceFromCoverage(docs []model.Document, rootDirs, leafDirs map[string]bool) []CheckResult {
	var results []CheckResult

	// Build a lookup from requirement source file to its document directory
	// for resolving inbound-edge origins against coverage-source directories.
	docByReqFile := make(map[string]string, len(docs))
	for _, doc := range docs {
		for _, req := range doc.Requirements() {
			docByReqFile[req.Source] = doc.Path
		}
	}

	for _, node := range g.nodes {
		// Filter guard: only check requirements in the active filter set.
		if !g.inFilter(node.ReqID) {
			continue
		}
		// Case 1: explicit `requires-trace-from` declared (nil = not set; non-nil = set)
		if node.RequiresTraceFrom != nil {
			if node.isSuppressed("requires-trace-from-coverage") {
				continue
			}
			// Empty `requires-trace-from: []` means "no coverage expected" — explicit opt-out.
			if len(node.RequiresTraceFrom) == 0 {
				continue
			}
			for _, from := range node.RequiresTraceFrom {
				dirs := g.resolveCoverageSource(from)
				if len(dirs) == 0 {
					results = append(results, CheckResult{
						Level:   LevelWarning,
						ReqID:   node.ReqID,
						File:    node.File,
						Message: fmt.Sprintf("unknown coverage source: %q", from),
					})
					continue
				}
				// Build a set for O(1) membership testing in the inbound loop.
				dirSet := make(map[string]bool, len(dirs))
				for _, d := range dirs {
					dirSet[d] = true
				}
				covered := false
				draftCount := 0
				for _, inboundID := range node.Inbound {
					inboundNode := g.nodes[inboundID]
					if inboundNode == nil {
						continue
					}
					// Filter guard: a filtered-out inbound cannot satisfy
					// coverage for a filtered-in requirement.
					if !g.inFilter(inboundID) {
						continue
					}
					// Status gate: a draft inbound does not satisfy
					// coverage. Count it for the message; do not consider
					// it a match.
					if !g.isCoverageProvider(inboundID) {
						draftCount++
						continue
					}
					inboundDir, ok := docByReqFile[inboundNode.File]
					if !ok {
						continue
					}
					if dirSet[inboundDir] {
						covered = true
						break
					}
				}
				if !covered {
					msg := fmt.Sprintf("no upstream trace: no approved requirement from %q traces to this item", from)
					if draftCount > 0 {
						msg += fmt.Sprintf(" (%d draft downstreams ignored)", draftCount)
					}
					results = append(results, CheckResult{
						Level:   LevelWarning,
						ReqID:   node.ReqID,
						File:    node.File,
						Message: msg,
					})
				}
			}
			continue
		}

		// Case 2: `requires-trace-from` not declared — fall back to generic
		// boundary-aware untraced and no-downstream checks.
		dir := docByReqFile[node.File]
		isRoot := rootDirs[dir]
		isLeaf := leafDirs[dir]

		// Untraced: no inbound edges from the active view, not at root
		// boundary, not a child req. Filter-aware: when a filter is active,
		// only inbounds that also match the filter count — a filtered-out
		// requirement cannot act as a coverage provider (RFC §3.2), so a
		// node whose only inbounds are filtered out is untraced in view.
		if node.isSuppressed("untraced") {
			// skip — per-requirement opt-out
		} else if !g.hasFilteredInbound(node) && !isRoot && !node.IsChild {
			results = append(results, CheckResult{
				Level:   LevelWarning,
				ReqID:   node.ReqID,
				File:    node.File,
				Message: "untraced: no downstream reference",
			})
		}

		// No downstream: no outbound edges to the active view, not at leaf
		// boundary. Filter-aware for the same reason as above.
		if !node.isSuppressed("no-downstream") {
			if !g.hasFilteredOutbound(node) && !isLeaf {
				results = append(results, CheckResult{
					Level:   LevelWarning,
					ReqID:   node.ReqID,
					File:    node.File,
					Message: "no upstream reference",
				})
			}
		}
	}
	return results
}

// hasFilteredInbound reports whether any inbound edge originates from a
// requirement in the active filter set. With no filter active (nil), it
// degenerates to "has any inbound" — preserving the pre-filter behavior
// exactly without iterating the inbound list.
func (g *Graph) hasFilteredInbound(node *CachedNode) bool {
	if g.filter == nil {
		return len(node.Inbound) > 0
	}
	return slices.ContainsFunc(node.Inbound, g.inFilter)
}

// hasFilteredOutbound reports whether any outbound edge targets a
// requirement in the active filter set. With no filter active (nil), it
// degenerates to "has any outbound" — preserving the pre-filter behavior.
func (g *Graph) hasFilteredOutbound(node *CachedNode) bool {
	if g.filter == nil {
		return len(node.Outbound) > 0
	}
	return slices.ContainsFunc(node.Outbound, g.inFilter)
}

// checkDispositionReason detects nodes where disposition is set to a value
// other than "implemented" but disposition-reason is missing or empty.
// Returns WARNING for each.
func (g *Graph) checkDispositionReason() []CheckResult {
	var results []CheckResult
	for _, node := range g.nodes {
		if !g.inFilter(node.ReqID) {
			continue
		}
		if node.isSuppressed(model.AttrDispositionReason) {
			continue
		}
		if node.Disposition == "" || node.Disposition == "implemented" {
			continue
		}
		if strings.TrimSpace(node.DispositionReason) == "" {
			results = append(results, CheckResult{
				Level:   LevelWarning,
				ReqID:   node.ReqID,
				File:    node.File,
				Message: fmt.Sprintf("disposition without reason for %q", node.Disposition),
			})
		}
	}
	return results
}

// checkMandatoryDisposition detects nodes in directories where
// mandatory-disposition is set and the requirement has no disposition value.
// Returns ERROR for each violation.
func (g *Graph) checkMandatoryDisposition() []CheckResult {
	var results []CheckResult
	for _, node := range g.nodes {
		if !g.inFilter(node.ReqID) {
			continue
		}
		if node.isSuppressed("mandatory-disposition") {
			continue
		}
		if !node.MandatoryDisposition {
			continue
		}
		if node.Disposition == "" {
			results = append(results, CheckResult{
				Level:   LevelError,
				ReqID:   node.ReqID,
				File:    node.File,
				Message: "disposition is required (mandatory-disposition is set)",
			})
		}
	}
	return results
}

// checkIDPrefixes validates that every requirement ID matches the declared
// prefix for its document directory and that no two directories share a prefix.
// Prefix collision is an ERROR: two document directories must not declare the
// same ID prefix, since that would make requirement IDs ambiguous across docs.
func (g *Graph) checkIDPrefixes() []CheckResult {
	var results []CheckResult
	prefixFiles := make(map[string]string) // prefix → first file seen for collision detection

	for _, node := range g.nodes {
		if !g.inFilter(node.ReqID) {
			continue
		}
		if node.isSuppressed("id-prefix") {
			continue
		}
		if node.IDPrefix == "" {
			continue
		}

		// Check: ID must start with the declared prefix
		if !strings.HasPrefix(node.ReqID, node.IDPrefix) {
			results = append(results, CheckResult{
				Level:   LevelError,
				ReqID:   node.ReqID,
				File:    node.File,
				Message: fmt.Sprintf("ID %q does not start with declared prefix %q", node.ReqID, node.IDPrefix),
			})
		}

		// Check: prefix collision — same prefix used by a different file
		prevFile, exists := prefixFiles[node.IDPrefix]
		if exists && prevFile != node.File {
			results = append(results, CheckResult{
				Level:   LevelError,
				ReqID:   node.ReqID,
				File:    node.File,
				Message: fmt.Sprintf("prefix %q already used by %s (collision)", node.IDPrefix, prevFile),
			})
			continue
		}
		prefixFiles[node.IDPrefix] = node.File
	}
	return results
}

// pinStatus classifies a single version pin against its target's
// current version. Shared by checkVersionPins (which emits findings)
// and RepinDeltas (which proposes edits). Returns the direction
// ("outdated", "predated", or "" for current) and ok=true when the
// pin is comparable to the target. Returns ok=false when the pin
// should be skipped (target missing, external, unversioned, or
// the node suppresses the version-pin check).
func (g *Graph) pinStatus(node *CachedNode, targetID string, pin int, pinned bool) (direction string, ok bool) {
	if node.isSuppressed(CodeVersionPin) {
		return "", false
	}
	target := g.nodes[targetID]
	if target == nil || target.External || target.Version == 0 {
		return "", false
	}
	if !pinned {
		return "", false
	}
	current := target.Version
	if pin == current {
		return "", false
	}
	if pin < current {
		return DirOutdated, true
	}
	return DirPredated, true
}

// checkVersionPins compares each downstream requirement's `~N` version pin
// on a trace reference against the upstream requirement's `version`.
// Findings:
//   - outdated — upstream version V > pinned N. Downstream must be re-verified.
//   - predated — upstream version V < pinned N. Downstream claims a version
//     that does not exist on the upstream; data integrity issue.
//
// Both emit ERROR by default. Outdated findings carry
// Code="version-pin" + Direction="outdated" and can be demoted to WARNING
// by the cmd layer when --relaxed-versions is set. Predated stays ERROR.
//
// Skip rules are shared with RepinDeltas via pinStatus.
func (g *Graph) checkVersionPins() []CheckResult {
	var results []CheckResult
	for _, node := range g.nodes {
		if !g.inFilter(node.ReqID) {
			continue
		}
		for targetID, pin := range node.OutboundPins {
			direction, ok := g.pinStatus(node, targetID, pin, true)
			if !ok {
				continue
			}
			current := g.nodes[targetID].Version
			switch direction {
			case DirOutdated:
				results = append(results, CheckResult{
					Code:      CodeVersionPin,
					Direction: DirOutdated,
					Level:     LevelError,
					ReqID:     node.ReqID,
					File:      node.File,
					Message:   fmt.Sprintf("outdated version pin: upstream %s is at v%d, this requirement still pins ~%d", targetID, current, pin),
				})
			case DirPredated:
				results = append(results, CheckResult{
					Code:      CodeVersionPin,
					Direction: DirPredated,
					Level:     LevelError,
					ReqID:     node.ReqID,
					File:      node.File,
					Message:   fmt.Sprintf("predated version pin: upstream %s is at v%d, this requirement pins ~%d", targetID, current, pin),
				})
			}
		}
	}
	return results
}

// checkMissingVerdict flags approved verification measures whose rolled-up
// evidence set is empty. A measure is any approved, non-result node carrying
// a `verify` attribute (see isMeasure). Evidence is the union of the
// measure's own attached results and the results of approved, in-filter
// nodes in its inbound trace closure (downstream test cases). Draft measures
// are skipped (a draft measure is not expected to have results yet).
//
// The check runs only when result pseudo-requirements are present in the
// graph (loaded via `check --results`). With no results loaded, no node
// has IsResult=true and the check is a no-op.
//
// Severity: WARNING. Suppress per-node with
// `reqmd-suppress: [missing-verdict]`. Code: "missing-verdict".
func (g *Graph) checkMissingVerdict() []CheckResult {
	var results []CheckResult
	// Short-circuit when no result nodes exist: the check is meaningful
	// only when results have been loaded.
	if !g.hasResultNodes() {
		return nil
	}
	for _, node := range g.nodes {
		if !g.inFilter(node.ReqID) {
			continue
		}
		if node.IsResult {
			continue
		}
		if node.isSuppressed(CodeMissingVerdict) {
			continue
		}
		if !g.isMeasure(node) {
			continue
		}
		// Draft measures are not expected to have results.
		if !g.isCoverageProvider(node.ReqID) {
			continue
		}
		// Deferred/rejected measures are not expected to be verified.
		if !expectsVerification(node) {
			continue
		}
		_, all, drafts := g.measureEvidence(node.ReqID)
		if len(all) > 0 {
			continue
		}
		msg := "missing verdict: no verification result or downstream test-case result covers this measure"
		if drafts > 0 {
			msg += fmt.Sprintf(" (%d draft downstreams ignored)", drafts)
		}
		results = append(results, CheckResult{
			Code:    CodeMissingVerdict,
			Level:   LevelWarning,
			ReqID:   node.ReqID,
			File:    node.File,
			Message: msg,
		})
	}
	return results
}

// checkFailingVerdict flags approved measures whose rolled-up verdict is
// fail. Like checkMissingVerdict it is a no-op when no result nodes are
// present. Draft measures are skipped (a draft measure is not expected to
// gate on results), mirroring the missing-verdict draft gate.
//
// The message names the failing case(s) and their source file(s).
//
// Severity: ERROR. Suppress per-node with
// `reqmd-suppress: [failing-verdict]`. Code: "failing-verdict".
func (g *Graph) checkFailingVerdict() []CheckResult {
	var results []CheckResult
	for _, node := range g.nodes {
		if !g.inFilter(node.ReqID) {
			continue
		}
		if node.IsResult {
			continue
		}
		if node.isSuppressed(CodeFailingVerdict) {
			continue
		}
		if !g.isMeasure(node) {
			continue
		}
		if !g.isCoverageProvider(node.ReqID) {
			continue
		}
		if !expectsVerification(node) {
			continue
		}
		if g.measureSeverity(node.ReqID) != sevFail {
			continue
		}
		failures := g.failingEvidence(node.ReqID)
		msg := "failing verdict: rolled-up verification result is fail"
		if len(failures) > 0 {
			parts := make([]string, 0, len(failures))
			for _, f := range failures {
				if f.Case != "" {
					parts = append(parts, fmt.Sprintf("%s (%s)", f.Case, f.File))
				} else {
					parts = append(parts, f.File)
				}
			}
			msg += ": " + strings.Join(parts, ", ")
		}
		results = append(results, CheckResult{
			Code:    CodeFailingVerdict,
			Level:   LevelError,
			ReqID:   node.ReqID,
			File:    node.File,
			Message: msg,
		})
	}
	return results
}

// hasResultNodes reports whether any synthesized result pseudo-requirement
// is present in the graph.
func (g *Graph) hasResultNodes() bool {
	return g.hasResults
}

// isMeasure reports whether a node is a verification measure — a
// requirement authored with a `verify` attribute (e.g. verify: Test,
// verify: Review). Whether a result currently traces to it is irrelevant to
// the definition (see checkMissingVerdict). Result nodes are not measures.
func (g *Graph) isMeasure(node *CachedNode) bool {
	if node.IsResult {
		return false
	}
	return node.Verify != ""
}

// expectsVerification reports whether a measure's disposition still expects
// verification results. A deferred or rejected measure is not expected to be
// verified, so it is exempt from the outcome-gated checks (mirroring how
// such dispositions suppress the generic trace warnings).
func expectsVerification(node *CachedNode) bool {
	switch node.Disposition {
	case "deferred", "rejected":
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Evidence-set roll-up (strict aggregation)
// ---------------------------------------------------------------------------

// severity ranks verification outcomes for strict roll-up aggregation:
// any fail → fail, else inconclusive → inconclusive, else skipped →
// skipped, else pass; an empty evidence set has severity none.
type severity int

const (
	sevNone severity = iota
	sevPass
	sevSkipped
	sevInconclusive
	sevFail
)

// outcomeSeverity maps an outcome string onto the severity lattice.
func outcomeSeverity(outcome string) severity {
	switch outcome {
	case "pass":
		return sevPass
	case "skipped":
		return sevSkipped
	case "inconclusive":
		return sevInconclusive
	case "fail":
		return sevFail
	default:
		return sevNone
	}
}

// severityOutcome maps a severity back to its outcome string ("" for none).
func severityOutcome(s severity) string {
	switch s {
	case sevPass:
		return "pass"
	case sevSkipped:
		return "skipped"
	case sevInconclusive:
		return "inconclusive"
	case sevFail:
		return "fail"
	default:
		return ""
	}
}

// EvidenceItem is one verification result contributing to a measure's
// evidence set: its case key ("" for direct, uncased results), outcome,
// source report file, and the synthesized test case's Markdown description
// (empty for direct results and authored test cases).
type EvidenceItem struct {
	Case        string
	Outcome     string
	File        string
	Description string
}

// evidenceItem derives an EvidenceItem from a synthesized result node.
func (g *Graph) evidenceItem(in *CachedNode) EvidenceItem {
	it := EvidenceItem{Case: g.evidenceCase(in), Outcome: in.Outcome, File: in.File}
	for _, outID := range in.Outbound {
		if out := g.nodes[outID]; out != nil && strings.HasPrefix(outID, "TC:") {
			it.Description = out.Body
			break
		}
	}
	return it
}

// evidenceCase resolves the case key of a result node: the `#case` suffix on
// a case-keyed result, or the case key of the synthesized test case the
// result traces to, or "" for a direct (uncased) result.
func (g *Graph) evidenceCase(in *CachedNode) string {
	if i := strings.IndexByte(in.ReqID, '#'); i >= 0 {
		return in.ReqID[i+1:]
	}
	for _, outID := range in.Outbound {
		if out := g.nodes[outID]; out != nil && out.Synthetic && strings.HasPrefix(outID, "TC:") {
			return strings.TrimPrefix(outID, "TC:")
		}
	}
	return ""
}

// measureEvidence gathers the evidence set of a requirement: its own
// attached results plus the results of approved, in-filter nodes in its
// inbound trace closure (downstream authored test cases and synthesized
// cases). Draft in-filter intermediaries are not contributors and are
// reported in the returned draft count. self holds the results attached
// directly to the requirement; items is the deduplicated union (by case
// key, keeping the higher severity).
func (g *Graph) measureEvidence(reqID string) (self, items []EvidenceItem, drafts int) {
	seen := make(map[string]bool)
	var rec func(id string) int
	rec = func(id string) int {
		node := g.nodes[id]
		if node == nil {
			return 0
		}
		d := 0
		for _, inID := range node.Inbound {
			in := g.nodes[inID]
			if in == nil {
				continue
			}
			if in.IsResult {
				it := g.evidenceItem(in)
				if id == reqID {
					self = append(self, it)
				}
				items = append(items, it)
				continue
			}
			if seen[inID] {
				continue
			}
			seen[inID] = true
			if !g.inFilter(inID) {
				continue
			}
			if !g.isCoverageProvider(inID) {
				d++
				continue
			}
			d += rec(inID)
		}
		return d
	}
	drafts = rec(reqID)

	// De-duplicate by case key, keeping the higher severity; sort for
	// deterministic message/export ordering.
	if len(items) > 1 {
		keep := make(map[string]EvidenceItem, len(items))
		for _, it := range items {
			cur, ok := keep[it.Case]
			if !ok || outcomeSeverity(it.Outcome) > outcomeSeverity(cur.Outcome) {
				keep[it.Case] = it
			}
		}
		items = items[:0]
		for _, it := range keep {
			items = append(items, it)
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Case < items[j].Case })
	}
	return self, items, drafts
}

// failingEvidence returns the evidence items whose outcome is fail.
func (g *Graph) failingEvidence(reqID string) []EvidenceItem {
	_, items, _ := g.measureEvidence(reqID)
	var fails []EvidenceItem
	for _, it := range items {
		if it.Outcome == "fail" {
			fails = append(fails, it)
		}
	}
	return fails
}

// severityOf computes the rolled-up severity of a node's evidence set,
// recursing through approved, in-filter intermediaries. Results are cached
// on the graph (invalidated by SetFilter) so a whole check pass stays cheap.
func (g *Graph) severityOf(reqID string, visiting map[string]bool) severity {
	if g.sevCache == nil {
		g.sevCache = make(map[string]severity)
	}
	if s, ok := g.sevCache[reqID]; ok {
		return s
	}
	if visiting[reqID] {
		return sevNone // cycle guard; circular refs are already errors
	}
	visiting[reqID] = true
	best := sevNone
	node := g.nodes[reqID]
	if node != nil {
		for _, inID := range node.Inbound {
			in := g.nodes[inID]
			if in == nil {
				continue
			}
			if in.IsResult {
				if s := outcomeSeverity(in.Outcome); s > best {
					best = s
				}
				continue
			}
			if !g.inFilter(inID) {
				continue
			}
			if !g.isCoverageProvider(inID) {
				continue
			}
			if s := g.severityOf(inID, visiting); s > best {
				best = s
			}
		}
	}
	visiting[reqID] = false
	g.sevCache[reqID] = best
	return best
}

// measureSeverity returns the rolled-up severity of a measure's evidence
// set (sevNone when the set is empty).
func (g *Graph) measureSeverity(reqID string) severity {
	return g.severityOf(reqID, make(map[string]bool))
}

// hasVerdictFor reports whether any result node traces to the given
// measure ID.
func (g *Graph) hasVerdictFor(measureID string) bool {
	for _, inID := range g.nodes[measureID].Inbound {
		if in, ok := g.nodes[inID]; ok && in.IsResult {
			return true
		}
	}
	return false
}

// latestOutcomeFor returns the outcome and source file of the result node
// tracing to the given measure. Since the result layer
// (internal/verify.MergeLatest) already collapses to one result per
// measure, there is at most one result node per measure. Returns
// ("", "", false) if no result traces to the measure.
func (g *Graph) latestOutcomeFor(measureID string) (string, string, bool) {
	node := g.nodes[measureID]
	if node == nil {
		return "", "", false
	}
	for _, inID := range node.Inbound {
		if in, ok := g.nodes[inID]; ok && in.IsResult {
			return in.Outcome, in.File, true
		}
	}
	return "", "", false
}

// checkVerdicts emits an INFO-level result for each measure that has a
// non-empty rolled-up evidence set, showing the rolled-up outcome. This
// makes passing verdicts visible in the output — without it, a clean run
// with --results looks identical to a run without --results. Measures
// without evidence are covered by checkMissingVerdict and emit no verdict
// line here.
//
// Severity: INFO (never affects exit code). Code: "verdict".
func (g *Graph) checkVerdicts() []CheckResult {
	var results []CheckResult
	if !g.hasResultNodes() {
		return nil
	}
	for _, node := range g.nodes {
		if !g.inFilter(node.ReqID) {
			continue
		}
		if node.IsResult {
			continue
		}
		if !g.isMeasure(node) {
			continue
		}
		if !expectsVerification(node) {
			continue
		}
		self, items, _ := g.measureEvidence(node.ReqID)
		if len(items) == 0 {
			continue
		}
		sev := sevNone
		file := ""
		for _, it := range items {
			if s := outcomeSeverity(it.Outcome); s > sev {
				sev = s
				file = it.File
			}
		}
		if len(self) > 0 {
			file = self[0].File
		}
		outcome := severityOutcome(sev)
		results = append(results, CheckResult{
			Code:    CodeVerdict,
			Level:   LevelInfo,
			Outcome: outcome,
			ReqID:   node.ReqID,
			File:    file,
			Message: fmt.Sprintf("verified: %s", outcome),
		})
	}
	return results
}

// MeasureOutcome returns the latest verification outcome for a measure
// (pass, fail, skipped, inconclusive) and whether a result exists. Returns
// ("", false) if the measure has no result or is not a measure.
func (g *Graph) MeasureOutcome(measureID string) (string, string, bool) {
	return g.latestOutcomeFor(measureID)
}

// NodeVerdict is the rolled-up verdict and a representative result source
// file for a single node, resolved by the graph. It feeds exporters and
// reporters; Outcome is the strict roll-up over the node's evidence set
// (own attached results plus approved downstream test cases), and Evidence
// lists the deduplicated items that make up that set (case, outcome, file,
// description).
type NodeVerdict struct {
	Outcome  string
	Source   string // representative result source file ("" when none)
	Evidence []EvidenceItem
}

// MeasureVerdicts returns, for every non-synthetic node with a non-empty
// rolled-up evidence set, the rolled-up outcome, a representative source
// file, and the deduplicated evidence items. The set is keyed by requirement
// ID and computed from graph adjacency, so a single source of truth feeds
// all outputs. Returns nil when no results are loaded.
func (g *Graph) MeasureVerdicts() map[string]NodeVerdict {
	if !g.hasResultNodes() {
		return nil
	}
	out := make(map[string]NodeVerdict)
	for id, node := range g.nodes {
		if node.IsResult || node.Synthetic {
			continue
		}
		self, items, _ := g.measureEvidence(id)
		if len(items) == 0 {
			continue
		}
		sev := sevNone
		file := ""
		for _, it := range items {
			if s := outcomeSeverity(it.Outcome); s > sev {
				sev = s
				file = it.File
			}
		}
		if len(self) > 0 {
			file = self[0].File
		}
		out[id] = NodeVerdict{
			Outcome:  severityOutcome(sev),
			Source:   file,
			Evidence: items,
		}
	}
	return out
}

// HasResults reports whether the graph contains any synthesized result
// pseudo-requirements (i.e. `--results` was supplied).
func (g *Graph) HasResults() bool {
	return g.hasResultNodes()
}

// RootDirs returns the document directory paths at the top of the V-model
// (no upstream sources declared, or external), sorted. Uses the cached
// boundaries from CheckResults; returns empty if CheckResults has not been
// called.
func (g *Graph) RootDirs() []string {
	return mapKeysSorted(g.cachedRootDirs)
}

// LeafDirs returns a sorted slice of document directory paths that are at
// the bottom of the V-model (not referenced as upstream sources by any
// other document).
// Uses the cached boundaries from CheckResults; returns empty if CheckResults
// has not been called.
func (g *Graph) LeafDirs() []string {
	return mapKeysSorted(g.cachedLeafDirs)
}

func mapKeysSorted(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ---------------------------------------------------------------------------
// Query / inspection methods on *Graph
// ---------------------------------------------------------------------------

// NodeCount returns the number of requirement nodes in the graph.
func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

// UpstreamNeighbors returns the requirement IDs that the given requirement
// traces TO — its parents/upstream requirements. Returns nil if not found.
func (g *Graph) UpstreamNeighbors(id string) []string {
	if node, ok := g.nodes[id]; ok {
		return node.Outbound
	}
	return nil
}

// DownstreamNeighbors returns the requirement IDs that trace TO the given
// requirement — its children/downstream requirements. Returns nil if not found.
func (g *Graph) DownstreamNeighbors(id string) []string {
	if node, ok := g.nodes[id]; ok {
		return node.Inbound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Property access helpers
// ---------------------------------------------------------------------------

func getString(props map[string]any, key string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getStringSlice reads a []any / []string attr as []string. Returns nil if
// the key is absent (so callers can distinguish "not set" from
// "explicitly empty" by comparing against nil). Returns []string{} when
// the key is present but the array is empty.
func getStringSlice(props map[string]any, key string) []string {
	v, ok := props[key]
	if !ok {
		return nil
	}
	switch arr := v.(type) {
	case []any:
		if arr == nil {
			return []string{}
		}
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		if arr == nil {
			return []string{}
		}
		out := make([]string, len(arr))
		copy(out, arr)
		return out
	}
	return nil
}

// getIntFromAttrs reads an integer attr. Handles the types that appear in
// practice: int (yaml.v3 decode), int64, float64 (JSON round-trip),
// and json.Number (defensive). Returns 0 if absent or undecodable.
func getIntFromAttrs(props map[string]any, key string) int {
	v, ok := props[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

// isSuppressed returns true if the given check name is in this node's
// suppression set. O(1) via the precomputed map; falls back to the
// linear scan only if the set was not built (defensive).
func (n *CachedNode) isSuppressed(check string) bool {
	if n.suppressionsSet != nil {
		_, ok := n.suppressionsSet[check]
		return ok
	}
	return slices.Contains(n.Suppressions, check)
}

// effectiveStatus returns the built-in status value, falling back to
// model.StatusDefault ("approved") when no status attr is declared.
func (n *CachedNode) effectiveStatus() string {
	if s := n.Status; s != "" {
		return s
	}
	return model.StatusDefault
}

// isCoverageProvider reports whether a node counts as a valid upstream
// coverage source for downstream requirements. A node is a coverage
// provider when:

// RepinDeltas computes the list of version-pin changes the repin
// command would propose (or apply) against this graph.
//
// Without promoteUnpinned, only outdated findings are returned —
// refs whose pin is strictly below the upstream's current version.
// Predated findings (pin strictly above upstream version) are
// surfaced separately so the user can see them; repin never
// auto-fixes a predated pin because it is a data integrity error,
// not a process error.
//
// With promoteUnpinned=true, refs that have no pin at all
// (`~N` not present in the source) against a versioned upstream
// also produce a delta. This converts "no claim" into "claimed at
// current version" — the caller is expected to opt in deliberately.
//
// Rules shared with checkVersionPins:
//   - Skip synthetic nodes (result pseudo-requirements and future
//     synthesized test cases — they exist in memory only, so there is no
//     source file to rewrite). IsResult is kept for nodes built directly
//     in tests that predate the Synthetic document marker.
//   - Skip nodes that suppress the version-pin check.
//   - Skip refs whose target is external (versions out of our control).
//   - Skip refs whose target has no version (no ground truth).
//
// Returns deltas sorted for stable output: by source file, then
// requirement ID, then source ref.
func (g *Graph) RepinDeltas(promoteUnpinned bool) []RepinDelta {
	var deltas []RepinDelta
	for _, node := range g.nodes {
		if node.Synthetic || node.IsResult {
			continue
		}
		for targetID, originalRef := range node.OutboundRefs {
			pin, pinned := node.OutboundPins[targetID]

			// Unpinned promote path: opt-in only.
			if !pinned {
				if !promoteUnpinned {
					continue
				}
				target := g.nodes[targetID]
				if target == nil || target.External || target.Version == 0 {
					continue
				}
				if node.isSuppressed(CodeVersionPin) {
					continue
				}
				deltas = append(deltas, RepinDelta{
					Kind:       "unpinned",
					ReqID:      node.ReqID,
					File:       node.File,
					TargetID:   targetID,
					SourceRef:  originalRef,
					NewPin:     target.Version,
					NewVersion: target.Version,
				})
				continue
			}

			// Pinned path: use the shared pinStatus predicate so the
			// skip rules and direction logic stay in sync with
			// checkVersionPins.
			direction, ok := g.pinStatus(node, targetID, pin, true)
			if !ok {
				continue
			}
			current := g.nodes[targetID].Version
			deltas = append(deltas, RepinDelta{
				Kind:       direction,
				ReqID:      node.ReqID,
				File:       node.File,
				TargetID:   targetID,
				SourceRef:  originalRef,
				OldPin:     pin,
				NewPin:     current,
				NewVersion: current,
			})
		}
	}
	sort.Slice(deltas, func(i, j int) bool {
		if deltas[i].File != deltas[j].File {
			return deltas[i].File < deltas[j].File
		}
		if deltas[i].ReqID != deltas[j].ReqID {
			return deltas[i].ReqID < deltas[j].ReqID
		}
		return deltas[i].SourceRef < deltas[j].SourceRef
	})
	return deltas
}

func (g *Graph) isCoverageProvider(reqID string) bool {
	node, ok := g.nodes[reqID]
	if !ok {
		return false
	}
	if g.ignoreStatusDirs[node.Dir] {
		return true
	}
	return strings.EqualFold(node.effectiveStatus(), model.StatusApproved)
}

// CodeDisjointAttribute is the check code for the disjoint-attribute check.
const CodeDisjointAttribute = "disjoint-attribute"

// checkDisjointAttribute verifies that for every trace link, the source and
// target requirements have at least one overlapping value for each declared
// disjoint-check attribute. Requirements with an empty/absent value for the
// attribute are exempt ("applies to all"). Both non-empty with zero
// intersection → ERROR. Filter-aware: only checks source reqs in the active
// filter set.
func (g *Graph) checkDisjointAttribute() []CheckResult {
	if len(g.disjointAttrs) == 0 {
		return nil
	}
	// Build reqID → attr → []string lookup once from g.docs. The map is
	// allocated lazily per requirement: most requirements do not declare
	// the disjoint attribute, and an absent value means "applies to all"
	// (exempt) — an empty map entry is indistinguishable from a miss.
	type attrVals = map[string][]string
	lookup := make(map[string]attrVals, len(g.nodes))
	for _, doc := range g.docs {
		for _, req := range doc.Requirements() {
			var vals attrVals
			for _, attr := range g.disjointAttrs {
				if s := getStringSlice(req.Attrs, attr); len(s) > 0 {
					if vals == nil {
						vals = make(attrVals, len(g.disjointAttrs))
					}
					vals[attr] = s
				}
			}
			if vals != nil {
				lookup[req.ID] = vals
			}
		}
	}

	var results []CheckResult
	for _, node := range g.nodes {
		if node.IsResult {
			continue
		}
		if !g.inFilter(node.ReqID) {
			continue
		}
		if node.isSuppressed(CodeDisjointAttribute) {
			continue
		}
		srcVals := lookup[node.ReqID]
		for _, targetID := range node.Outbound {
			target, ok := g.nodes[targetID]
			if !ok || target.IsResult {
				continue
			}
			tgtVals := lookup[targetID]
			for _, attr := range g.disjointAttrs {
				src := srcVals[attr]
				tgt := tgtVals[attr]
				// Empty/absent = "applies to all" → exempt.
				if len(src) == 0 || len(tgt) == 0 {
					continue
				}
				if !intersects(src, tgt) {
					results = append(results, CheckResult{
						Code:    CodeDisjointAttribute,
						Level:   LevelError,
						ReqID:   node.ReqID,
						File:    node.File,
						Message: fmt.Sprintf("disjoint-attribute: traces to %s (%s: %v) but %s's %s %v has no overlap — no valid configuration includes both", targetID, attr, tgt, node.ReqID, attr, src),
					})
				}
			}
		}
	}
	return results
}

// intersects reports whether two string slices have at least one element
// in common. Both slices are assumed non-empty.
func intersects(a, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		if _, ok := set[s]; ok {
			return true
		}
	}
	return false
}
