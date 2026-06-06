package exporter

import (
	"path/filepath"

	"reqmd/internal/model"
)

// RenderContext holds precomputed document chain and link resolution data
// for cross-document HTML traceability. It is shared by the export and serve commands.
type RenderContext struct {
	docs      []model.Document
	root      string
	reqToHTML map[string]string // reqID → output HTML path
	docLinks  []docLinkInfo
	absPaths  []string
	absMap    map[string]absDocInfo
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
		htmlPath := filepath.Join(doc.Path, htmlName)

		for _, req := range doc.Requirements {
			rctx.reqToHTML[req.ID] = htmlPath
		}
		title := ""
		if t, ok := doc.Schema.(map[string]any)["title"]; ok {
			if s, ok := t.(string); ok {
				title = s
			}
		}
		info := docLinkInfo{
			id:    dirName,
			title: title,
			path:  htmlPath,
		}
		rctx.docLinks = append(rctx.docLinks, info)

		absPath, err := filepath.Abs(doc.Path)
		if err != nil {
			return nil, err
		}
		rctx.absMap[absPath] = absDocInfo{info: info, doc: doc}
		rctx.absPaths = append(rctx.absPaths, absPath)
	}

	return rctx, nil
}

// ResolveLink returns a function that maps a requirement ID to an HTML anchor link
// relative to the given output path.
func (r *RenderContext) ResolveLink(currentOutPath string) func(string) string {
	return func(reqID string) string {
		targetPath, ok := r.reqToHTML[reqID]
		if !ok || targetPath == currentOutPath {
			return "#" + reqID
		}
		rel, err := filepath.Rel(filepath.Dir(currentOutPath), targetPath)
		if err != nil {
			return "#" + reqID
		}
		return rel + "#" + reqID
	}
}

// ChainCard represents a single document in the chain graph visualization.
type ChainCard struct {
	Title     string // schema title (or dirName if empty)
	Path      string // output HTML path (empty for current doc, or unknown)
	DirName   string // directory basename
	IsCurrent bool   // true for the doc this graph was built for
	IsExternal bool  // true for docs marked x-reqmd.external: true
}

// ChainTier is a horizontal row of docs at the same trace distance.
type ChainTier struct {
	Level  int         // distance from current (negative upstream, positive downstream, 0 = current)
	Label  string      // human label, e.g. "Upstream tier 2", "Downstream tier 1", "Current"
	Cards  []ChainCard // docs in this tier (1 for current tier, n for parallel branches)
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
			// Walk every other doc in the registry; if it traces to abs
			// (directly) and is not yet visited, it is a downstream doc.
			for _, otherAbs := range r.absPaths {
				if visited[otherAbs] {
					continue
				}
				otherInfo := r.absMap[otherAbs]
				if otherInfo.doc.XReqmd == nil || otherInfo.doc.XReqmd.Upstream == nil {
					continue
				}
				sources := otherInfo.doc.XReqmd.Upstream.Sources
				hit := false
				for _, src := range sources {
					target := filepath.Clean(filepath.Join(otherAbs, src))
					if target == abs {
						hit = true
						break
					}
				}
				if !hit {
					continue
				}
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
		return "Upstream · " + itoa(n) + " levels above"
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
		return "Downstream · " + itoa(level) + " levels below"
	}
}

// itoa is a tiny int-to-string helper to avoid pulling strconv into the label path.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
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
