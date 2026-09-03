package exporter

import (
	"fmt"
	"path/filepath"
	"reqmd/internal/model"
	"strconv"
)

// BuildTitleMap builds a reqID → title map across all documents, for use as
// trace-link labels. Requirements with an empty title are omitted.
func BuildTitleMap(docs []model.Document) map[string]string {
	titleMap := make(map[string]string)
	for _, d := range docs {
		for _, req := range d.Requirements() {
			if req.Title != "" {
				titleMap[req.ID] = req.Title
			}
		}
	}
	return titleMap
}

// RenderContext holds precomputed document chain and link resolution data
// for cross-document HTML traceability. It is shared by the export and serve commands.
type RenderContext struct {
	reqToHTML    map[string]string
	absMap       map[string]absDocInfo
	downstreamOf map[string][]string
	root         string
	docs         []model.Document
	docLinks     []docLinkInfo
	absPaths     []string
}

type docLinkInfo struct {
	id    string // dirName (basename of doc.Path)
	title string
	path  string // full output path
}

type absDocInfo struct {
	info docLinkInfo
	doc  model.Document
}

// NewRenderContext precomputes the cross-document link map and trace chain data.
func NewRenderContext(docs []model.Document, root string) (*RenderContext, error) {
	rctx := &RenderContext{
		docs:      docs,
		root:      root,
		reqToHTML: make(map[string]string),
		absMap:    make(map[string]absDocInfo),
	}

	for _, doc := range docs {
		dirName := filepath.Base(doc.Path)
		htmlName := dirName + "-requirements.html"
		htmlPath := htmlName

		for _, req := range doc.Requirements() {
			rctx.reqToHTML[req.ID] = htmlPath
		}
		info := docLinkInfo{
			id:    dirName,
			title: docTitle(doc),
			path:  htmlPath,
		}
		rctx.docLinks = append(rctx.docLinks, info)

		absPath, err := filepath.Abs(doc.Path)
		if err != nil {
			return nil, fmt.Errorf("resolving %s: %w", doc.Path, err)
		}
		rctx.absMap[absPath] = absDocInfo{info: info, doc: doc}
		rctx.absPaths = append(rctx.absPaths, absPath)
	}

	rctx.downstreamOf = make(map[string][]string, len(rctx.absPaths))
	for _, abs := range rctx.absPaths {
		info := rctx.absMap[abs]
		if info.doc.XReqmd == nil || info.doc.XReqmd.Upstream == nil {
			continue
		}
		for _, src := range info.doc.XReqmd.Upstream.Sources {
			target := filepath.Clean(filepath.Join(abs, src))
			if _, ok := rctx.absMap[target]; !ok {
				continue
			}
			rctx.downstreamOf[target] = append(rctx.downstreamOf[target], abs)
		}
	}

	return rctx, nil
}

// ResolveLink returns a function that maps a requirement ID to an HTML
// anchor link relative to the given output path. Since all HTML files
// are written as flat siblings in the output directory, the relative
// path between any two files is just the target filename. The current
// file is detected by comparing basenames so absolute vs relative path
// mismatches don't cause false negatives.
func (r *RenderContext) ResolveLink(currentOutPath string) func(string) string {
	currentBase := filepath.Base(currentOutPath)
	return func(reqID string) string {
		targetPath, ok := r.reqToHTML[reqID]
		if !ok || targetPath == currentBase {
			return "#" + reqID
		}
		return targetPath + "#" + reqID
	}
}

// ChainCard represents a single document in the chain graph visualization.
type ChainCard struct {
	Title      string // document title (parsed h1, or schema title, or dirName if empty)
	Path       string // output HTML path (empty for current doc, or unknown)
	DirName    string // directory basename
	IsCurrent  bool   // true for the doc this graph was built for
	IsExternal bool   // true for docs marked x-reqmd.external: true
}

// ChainTier is a horizontal row of docs at the same trace distance.
type ChainTier struct {
	Label string
	Cards []ChainCard
	Level int
}

// DocChainGraph is the full tiered document graph for a single current doc.
// It is the source of truth for the Confluence-Flow visualization.
type DocChainGraph struct {
	Upstream   []ChainTier // tiers with negative levels, ordered from farthest (lowest) to nearest (highest)
	Current    ChainCard   // the doc the graph was built for
	Downstream []ChainTier // tiers with positive levels, ordered from nearest (level 1) to farthest
}

// HasContent reports whether the graph has any tier that should be rendered.
func (g DocChainGraph) HasContent() bool {
	return len(g.Upstream) > 0 || len(g.Downstream) > 0
}

// BuildChainGraph constructs the tiered document trace graph for the given doc.
//
// Upstream tiers (negative levels) are computed by BFS outward from the current
// doc following upstream.sources references. Docs at the same shortest
// distance are grouped into a single horizontal tier so parallel branches at
// the same level render side-by-side. External docs (x-reqmd.external: true)
// are flagged so the renderer can mark them.
//
// Downstream tiers are computed symmetrically by BFS following inverted
// references (docs whose upstream.sources point at the current doc, or any
// other already-reached downstream doc).
//
// Self-references and visited nodes are tracked globally to terminate cycles.
func (r *RenderContext) BuildChainGraph(doc model.Document) DocChainGraph {
	docAbsPath, err := filepath.Abs(doc.Path)
	if err != nil {
		return DocChainGraph{}
	}

	selfInfo := r.absMap[docAbsPath]
	current := ChainCard{
		Title:     selfInfo.info.title,
		Path:      "",
		DirName:   selfInfo.info.id,
		IsCurrent: true,
	}

	g := DocChainGraph{Current: current}

	// ── Upstream BFS ──
	// visited tracks every absolute path we have already added to any tier.
	// current doc is in visited so we never include it in upstream tiers.
	visited := map[string]bool{docAbsPath: true}
	// frontier holds the abs paths at the current BFS depth.
	frontier := []string{docAbsPath}
	distance := 0

	for len(frontier) > 0 {
		distance--
		nextFrontier := []string{}
		tier := ChainTier{
			Level: distance,
			Label: upstreamTierLabel(distance),
		}
		for _, abs := range frontier {
			info, ok := r.absMap[abs]
			if !ok {
				continue
			}
			if info.doc.XReqmd == nil || info.doc.XReqmd.Upstream == nil {
				continue
			}
			for _, src := range info.doc.XReqmd.Upstream.Sources {
				resolved := filepath.Clean(filepath.Join(abs, src))
				if visited[resolved] {
					continue
				}
				targetInfo, ok := r.absMap[resolved]
				if !ok {
					continue
				}
				visited[resolved] = true
				nextFrontier = append(nextFrontier, resolved)
				card := ChainCard{
					Title:      targetInfo.info.title,
					Path:       targetInfo.info.path,
					DirName:    targetInfo.info.id,
					IsCurrent:  false,
					IsExternal: targetInfo.doc.XReqmd != nil && targetInfo.doc.XReqmd.External,
				}
				tier.Cards = append(tier.Cards, card)
			}
		}
		if len(tier.Cards) > 0 {
			g.Upstream = append(g.Upstream, tier)
		}
		frontier = nextFrontier
	}

	// Upstream tiers currently read [closest tier first, farthest tier last].
	// We want the visual order to be [farthest at top, closest just above current].
	for i, j := 0, len(g.Upstream)-1; i < j; i, j = i+1, j-1 {
		g.Upstream[i], g.Upstream[j] = g.Upstream[j], g.Upstream[i]
	}

	// ── Downstream BFS ──
	visited = map[string]bool{docAbsPath: true}
	frontier = []string{docAbsPath}
	distance = 0

	for len(frontier) > 0 {
		distance++
		nextFrontier := []string{}
		tier := ChainTier{
			Level: distance,
			Label: downstreamTierLabel(distance),
		}
		for _, abs := range frontier {
			// Direct adjacency lookup: downstreamOf[abs] is exactly the set of
			// docs whose upstream.sources resolve to abs.
			for _, otherAbs := range r.downstreamOf[abs] {
				if visited[otherAbs] {
					continue
				}
				otherInfo := r.absMap[otherAbs]
				visited[otherAbs] = true
				nextFrontier = append(nextFrontier, otherAbs)
				card := ChainCard{
					Title:      otherInfo.info.title,
					Path:       otherInfo.info.path,
					DirName:    otherInfo.info.id,
					IsCurrent:  false,
					IsExternal: otherInfo.doc.XReqmd != nil && otherInfo.doc.XReqmd.External,
				}
				tier.Cards = append(tier.Cards, card)
			}
		}
		if len(tier.Cards) > 0 {
			g.Downstream = append(g.Downstream, tier)
		}
		frontier = nextFrontier
	}

	return g
}

func upstreamTierLabel(level int) string {
	if level == -1 {
		return "Upstream"
	}
	return upstreamTierPlural(-level)
}

func upstreamTierPlural(n int) string {
	switch n {
	case 2:
		return "Upstream · 2 levels above"
	case 3:
		return "Upstream · 3 levels above"
	default:
		// fallback for deeper graphs
		return "Upstream · " + strconv.Itoa(n) + " levels above"
	}
}

func downstreamTierLabel(level int) string {
	if level == 1 {
		return "Downstream"
	}
	switch level {
	case 2:
		return "Downstream · 2 levels below"
	case 3:
		return "Downstream · 3 levels below"
	default:
		return "Downstream · " + strconv.Itoa(level) + " levels below"
	}
}

// ResolveChainOutputs replaces every non-current card's Path — which the
// render context seeds with the flat output file name (e.g.
// "01-stakeholder-requirements.html") — with the real output path for that
// document, as reported by outPath(dirName). Callers know where each page
// is actually written (flat into an output dir, or into each document's own
// dir), so they supply the mapping. Relative card paths are then produced
// by RelativizeChainGraph.
func ResolveChainOutputs(g *DocChainGraph, outPath func(dirName string) string) {
	if g == nil || outPath == nil {
		return
	}
	resolveTier := func(t *ChainTier) {
		for i := range t.Cards {
			c := &t.Cards[i]
			if c.IsCurrent || c.DirName == "" {
				continue
			}
			c.Path = outPath(c.DirName)
		}
	}
	for i := range g.Upstream {
		resolveTier(&g.Upstream[i])
	}
	for i := range g.Downstream {
		resolveTier(&g.Downstream[i])
	}
}

// RelativizeChainGraph rewrites every non-current card's Path to be
// relative to baseDir. This is a free function (not a method) so callers
// in other packages can use it without exposing the unexported fields
// of DocChainGraph.
func RelativizeChainGraph(g *DocChainGraph, baseDir string) {
	if g == nil {
		return
	}
	relativizeTier := func(t *ChainTier) {
		for i := range t.Cards {
			c := &t.Cards[i]
			if c.IsCurrent || c.Path == "" {
				continue
			}
			rel, err := filepath.Rel(baseDir, c.Path)
			if err == nil {
				c.Path = rel
			}
		}
	}
	for i := range g.Upstream {
		relativizeTier(&g.Upstream[i])
	}
	for i := range g.Downstream {
		relativizeTier(&g.Downstream[i])
	}
}

// BasenameChainGraph collapses every non-current card's Path to its
// basename. Useful when all HTML files are served at a single URL root
// (the dev server).
func BasenameChainGraph(g *DocChainGraph) {
	if g == nil {
		return
	}
	basenameTier := func(t *ChainTier) {
		for i := range t.Cards {
			c := &t.Cards[i]
			if c.IsCurrent || c.Path == "" {
				continue
			}
			c.Path = filepath.Base(c.Path)
		}
	}
	for i := range g.Upstream {
		basenameTier(&g.Upstream[i])
	}
	for i := range g.Downstream {
		basenameTier(&g.Downstream[i])
	}
}
