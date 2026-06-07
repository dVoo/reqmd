// Package graph builds a pure Go in-memory adjacency graph from parsed
// requirement documents and runs Pass 2 trace validation checks (broken
// references, circular dependencies, requires-trace-from coverage, etc.)
// per the reqmd validation pipeline.
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
	"sort"
	"strconv"
	"strings"

	"reqmd/internal/model"
)

// Check result severity levels.
const (
	LevelError   = "ERROR"
	LevelWarning = "WARNING"
)

// CheckResult describes a single Pass 2 trace validation finding.
type CheckResult struct {
	Level     string // one of LevelError, LevelWarning
	Code      string // machine-readable check identifier (e.g. "version-pin"); empty for findings that don't need one
	Direction string // "outdated" | "predated" — only set for version-pin findings
	ReqID     string
	File      string
	Message   string
}

// CachedNode is an in-memory typed representation of a requirement node,
// populated once at graph creation. All check methods iterate this cache.
type CachedNode struct {
	ReqID                string
	File                 string
	IsChild              bool
	Disposition          string
	DispositionReason    string
	External             bool
	MandatoryDisposition bool
	IDPrefix             string
	Version              int      // 0 = not declared
	RequiresTraceFrom    []string // nil = not set, []string{} = explicitly empty, [..] = expected coverage sources
	Inbound              []string
	Outbound             []string
	OutboundPins         map[string]int // target ID → ~N pin; key absence = no pin
	Suppressions         []string
	// Status is the raw value of the built-in `status` attribute.
	// Empty means the attribute was not declared — effectiveStatus()
	// falls back to model.StatusDefault in that case.
	Status string
	// Dir is the directory path of the requirement's source file.
	// Precomputed to avoid repeated filepath.Dir(node.File) calls.
	Dir string
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
		for _, req := range doc.Requirements {
			node := &CachedNode{
				ReqID:             req.ID,
				File:              req.Source,
				Dir:               filepath.Dir(req.Source),
				IsChild:           req.ParentID != "",
				Disposition:       getString(req.Attrs, model.AttrDisposition),
				DispositionReason: getString(req.Attrs, model.AttrDispositionReason),
				Version:           getIntFromAttrs(req.Attrs, model.AttrVersion),
				RequiresTraceFrom: getStringSlice(req.Attrs, model.AttrRequiresTraceFrom),
				OutboundPins:      make(map[string]int),
				Suppressions:      req.Suppressions,
				Status:            getString(req.Attrs, model.AttrStatus),
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
			g.nodes[req.ID] = node
			dirPath := node.Dir
			g.nodesByIDDir[[2]string{dirPath, req.ID}] = node
			g.idDirs[req.ID] = append(g.idDirs[req.ID], dirPath)
		}
	}

	// Pass 2: build Inbound/Outbound adjacency from trace attrs
	for _, doc := range docs {
		for _, req := range doc.Requirements {
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

				// Strip optional version pin: "doc-id/ID~N" or "ID~N" → "doc-id/ID" / "ID", pin=N
				// A successful Atoi means the user wrote a pin (even if N == 0).
				pin := 0
				pinned := false
				if idx := strings.LastIndex(refStr, "~"); idx >= 0 {
					if n, err := strconv.Atoi(refStr[idx+1:]); err == nil {
						pin = n
						pinned = true
						refStr = refStr[:idx]
					}
				}

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
					node.OutboundPins[targetID] = pin
				}
			}
		}
	}

	// Pass 3: build parent-child edges from heading hierarchy
	for _, doc := range docs {
		for _, req := range doc.Requirements {
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

	rootDirs, leafDirs := g.inferBoundaries(g.docs)
	results = append(results, g.checkRequiresTraceFromCoverage(g.docs, rootDirs, leafDirs)...)

	results = append(results, g.checkDispositionReason()...)
	results = append(results, g.checkMandatoryDisposition()...)
	results = append(results, g.checkVersionPins()...)
	results = append(results, g.checkIDPrefixes()...)
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
		for _, req := range doc.Requirements {
			docByReqFile[req.Source] = doc.Path
		}
	}

	for _, node := range g.nodes {
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
				covered := false
				draftCount := 0
				for _, inboundID := range node.Inbound {
					inboundNode := g.nodes[inboundID]
					if inboundNode == nil {
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
					for _, dir := range dirs {
						if inboundDir == dir {
							covered = true
							break
						}
					}
					if covered {
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

		// Untraced: no inbound edges, not at root boundary, not a child req.
		if node.isSuppressed("untraced") {
			// skip — per-requirement opt-out
		} else if len(node.Inbound) == 0 && !isRoot && !node.IsChild {
			results = append(results, CheckResult{
				Level:   LevelWarning,
				ReqID:   node.ReqID,
				File:    node.File,
				Message: "untraced: no downstream reference",
			})
		}

		// No downstream: no outbound edges, not at leaf boundary.
		if !node.isSuppressed("no-downstream") {
			if len(node.Outbound) == 0 && !isLeaf {
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

// checkDispositionReason detects nodes where disposition is set to a value
// other than "implemented" but disposition-reason is missing or empty.
// Returns WARNING for each.
func (g *Graph) checkDispositionReason() []CheckResult {
	var results []CheckResult
	for _, node := range g.nodes {
		if node.isSuppressed("disposition-reason") {
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
// Rules:
//   - Skip if the node has `reqmd-suppress: [version-pin]`.
//   - Skip if the target is external (versions are out of our control).
//   - Skip if the upstream has no `version` (Version == 0): we have no
//     ground truth to compare against, so a pin to an unversioned upstream
//     is harmless. The pin is still recorded for inspection.
//   - A pin of 0 against a versioned upstream is reported as outdated:
//     the user wrote "~0" and the upstream has since been bumped.
func (g *Graph) checkVersionPins() []CheckResult {
	var results []CheckResult
	for _, node := range g.nodes {
		if node.isSuppressed("version-pin") {
			continue
		}
		for targetID, pin := range node.OutboundPins {
			target := g.nodes[targetID]
			if target == nil || target.External {
				continue
			}
			if target.Version == 0 {
				continue
			}
			current := target.Version
			if pin == current {
				continue
			}
			if pin < current {
				results = append(results, CheckResult{
					Code:      "version-pin",
					Direction: "outdated",
					Level:     LevelError,
					ReqID:     node.ReqID,
					File:      node.File,
					Message:   fmt.Sprintf("outdated version pin: upstream %s is at v%d, this requirement still pins ~%d", targetID, current, pin),
				})
			} else {
				results = append(results, CheckResult{
					Code:      "version-pin",
					Direction: "predated",
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

// ---------------------------------------------------------------------------
// Boundary inference accessors — used by HTML export to drive visual
// root/leaf styling without re-computing topology.
// ---------------------------------------------------------------------------

// RootDirs returns a sorted slice of document directory paths that are at
// the top of the V-model (no upstream sources declared, or external).
func (g *Graph) RootDirs(docs []model.Document) []string {
	rootDirs, _ := g.inferBoundaries(docs)
	return mapKeysSorted(rootDirs)
}

// LeafDirs returns a sorted slice of document directory paths that are at
// the bottom of the V-model (not referenced as upstream sources by any
// other document).
func (g *Graph) LeafDirs(docs []model.Document) []string {
	_, leafDirs := g.inferBoundaries(docs)
	return mapKeysSorted(leafDirs)
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

// isSuppressed returns true if the given check name is in this node's suppression list.
func (n *CachedNode) isSuppressed(check string) bool {
	for _, s := range n.Suppressions {
		if s == check {
			return true
		}
	}
	return false
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
//   - it lives in a directory that opts out of the status lifecycle
//     (x-reqmd.ignore-status: true), OR
//   - its status (case-insensitive) is "approved".
//
// All other values (draft, additions, or anything unrecognised) are NOT
// coverage providers — they remain visible but do not satisfy traceability.
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
