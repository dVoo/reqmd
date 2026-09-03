package exporter

import (
	"reqmd/internal/model"
	"testing"
)

// makeDoc constructs a minimal model.Document with the given path, title and
// optional x-reqmd block. Requirements are empty (the chain graph doesn't need
// any) and schema is the bare map containing the title.
func makeDoc(path, title string, xr *model.XReqmd) model.Document {
	return model.Document{
		Path:   path,
		Schema: map[string]any{"title": title, "type": "object"},
		XReqmd: xr,
	}
}

// makeXReqmd is a tiny constructor for XReqmd pointer values.
func makeXReqmd(sources []string) *model.XReqmd {
	return &model.XReqmd{
		Upstream: &model.TraceUpstream{
			Sources: sources,
		},
	}
}

func TestBuildChainGraphSimpleLinear(t *testing.T) {
	// root → mid → leaf
	docs := []model.Document{
		makeDoc("/docs/root", "Root", nil),
		makeDoc("/docs/mid", "Mid", makeXReqmd([]string{"../root"})),
		makeDoc("/docs/leaf", "Leaf", makeXReqmd([]string{"../mid"})),
	}
	rctx, err := NewRenderContext(docs, "/docs")
	if err != nil {
		t.Fatalf("NewRenderContext: %v", err)
	}

	midDoc := docs[1]
	g := rctx.BuildChainGraph(midDoc)

	if g.Current.Title != "Mid" {
		t.Errorf("Current.Title = %q, want %q", g.Current.Title, "Mid")
	}
	if !g.Current.IsCurrent {
		t.Error("Current.IsCurrent should be true")
	}
	if len(g.Upstream) != 1 {
		t.Fatalf("Upstream tiers = %d, want 1", len(g.Upstream))
	}
	if len(g.Upstream[0].Cards) != 1 || g.Upstream[0].Cards[0].Title != "Root" {
		t.Errorf("upstream tier 0 = %+v, want [Root]", g.Upstream[0].Cards)
	}
	if len(g.Downstream) != 1 {
		t.Fatalf("Downstream tiers = %d, want 1", len(g.Downstream))
	}
	if len(g.Downstream[0].Cards) != 1 || g.Downstream[0].Cards[0].Title != "Leaf" {
		t.Errorf("downstream tier 0 = %+v, want [Leaf]", g.Downstream[0].Cards)
	}
}

func TestBuildChainGraphParallelBranches(t *testing.T) {
	// root_a, root_b → mid
	docs := []model.Document{
		makeDoc("/docs/root_a", "Root A", nil),
		makeDoc("/docs/root_b", "Root B", nil),
		makeDoc("/docs/mid", "Mid", makeXReqmd([]string{"../root_a", "../root_b"})),
	}
	rctx, err := NewRenderContext(docs, "/docs")
	if err != nil {
		t.Fatalf("NewRenderContext: %v", err)
	}

	g := rctx.BuildChainGraph(docs[2])

	if len(g.Upstream) != 1 {
		t.Fatalf("Upstream tiers = %d, want 1 (parallel branches in same tier)", len(g.Upstream))
	}
	if len(g.Upstream[0].Cards) != 2 {
		t.Fatalf("upstream tier 0 cards = %d, want 2", len(g.Upstream[0].Cards))
	}
	titles := map[string]bool{
		g.Upstream[0].Cards[0].Title: true,
		g.Upstream[0].Cards[1].Title: true,
	}
	if !titles["Root A"] || !titles["Root B"] {
		t.Errorf("upstream tier 0 titles = %v, want {Root A, Root B}", titles)
	}
}

func TestBuildChainGraphExternalFlag(t *testing.T) {
	docs := []model.Document{
		makeDoc("/docs/ext", "External Spec", &model.XReqmd{External: true}),
		makeDoc("/docs/mid", "Mid", makeXReqmd([]string{"../ext"})),
	}
	rctx, err := NewRenderContext(docs, "/docs")
	if err != nil {
		t.Fatalf("NewRenderContext: %v", err)
	}

	g := rctx.BuildChainGraph(docs[1])
	if len(g.Upstream) != 1 || len(g.Upstream[0].Cards) != 1 {
		t.Fatalf("unexpected upstream: %+v", g.Upstream)
	}
	if !g.Upstream[0].Cards[0].IsExternal {
		t.Error("external upstream card should have IsExternal=true")
	}
}

func TestBuildChainGraphCycle(t *testing.T) {
	// a → b → a (cycle)
	docs := []model.Document{
		makeDoc("/docs/a", "A", makeXReqmd([]string{"../b"})),
		makeDoc("/docs/b", "B", makeXReqmd([]string{"../a"})),
	}
	rctx, err := NewRenderContext(docs, "/docs")
	if err != nil {
		t.Fatalf("NewRenderContext: %v", err)
	}

	g := rctx.BuildChainGraph(docs[0])
	// a's upstream: b; b's upstream: a (but a is the current → already visited) →
	// no further tiers. So a sees exactly 1 upstream card.
	if len(g.Upstream) != 1 || len(g.Upstream[0].Cards) != 1 {
		t.Fatalf("unexpected upstream for cycle: %+v", g.Upstream)
	}
	if g.Upstream[0].Cards[0].Title != "B" {
		t.Errorf("upstream card = %q, want B", g.Upstream[0].Cards[0].Title)
	}
}

func TestBuildChainGraphDeeperLevels(t *testing.T) {
	// root → a → b → leaf (3 levels away from leaf)
	docs := []model.Document{
		makeDoc("/docs/root", "Root", nil),
		makeDoc("/docs/a", "A", makeXReqmd([]string{"../root"})),
		makeDoc("/docs/b", "B", makeXReqmd([]string{"../a"})),
		makeDoc("/docs/leaf", "Leaf", makeXReqmd([]string{"../b"})),
	}
	rctx, err := NewRenderContext(docs, "/docs")
	if err != nil {
		t.Fatalf("NewRenderContext: %v", err)
	}

	// Looking from B: upstream = [Root, A] (Root farthest, A closest);
	// downstream = [Leaf].
	g := rctx.BuildChainGraph(docs[2])
	if len(g.Upstream) != 2 {
		t.Fatalf("Upstream tiers = %d, want 2", len(g.Upstream))
	}
	// Farthest tier first.
	if g.Upstream[0].Cards[0].Title != "Root" {
		t.Errorf("upstream[0] title = %q, want Root (farthest)", g.Upstream[0].Cards[0].Title)
	}
	if g.Upstream[1].Cards[0].Title != "A" {
		t.Errorf("upstream[1] title = %q, want A (closest)", g.Upstream[1].Cards[0].Title)
	}
	if len(g.Downstream) != 1 || g.Downstream[0].Cards[0].Title != "Leaf" {
		t.Errorf("downstream = %+v, want [Leaf]", g.Downstream)
	}
}

func TestBuildChainGraphNoUpstream(t *testing.T) {
	// A root doc with no upstream sources: empty upstream, no downstream either.
	docs := []model.Document{
		makeDoc("/docs/root", "Root", nil),
	}
	rctx, err := NewRenderContext(docs, "/docs")
	if err != nil {
		t.Fatalf("NewRenderContext: %v", err)
	}

	g := rctx.BuildChainGraph(docs[0])
	if g.HasContent() {
		t.Errorf("root-only graph should have no content, got %+v", g)
	}
}

func TestResolveAndRelativizeChainOutputs(t *testing.T) {
	g := &DocChainGraph{
		Upstream: []ChainTier{{Cards: []ChainCard{
			{DirName: "01-stakeholder", Path: "01-stakeholder-requirements.html"},
		}}},
		Downstream: []ChainTier{{Cards: []ChainCard{
			{IsCurrent: true, DirName: "02-system", Path: ""},
			{DirName: "03-software", Path: "03-software-requirements.html"},
		}}},
	}
	ResolveChainOutputs(g, func(dirName string) string {
		return "/export/" + dirName + "-requirements.html"
	})
	if got := g.Upstream[0].Cards[0].Path; got != "/export/01-stakeholder-requirements.html" {
		t.Errorf("upstream Path = %q, want resolved output path", got)
	}
	if got := g.Downstream[0].Cards[1].Path; got != "/export/03-software-requirements.html" {
		t.Errorf("downstream Path = %q, want resolved output path", got)
	}
	// Current card is untouched.
	if got := g.Downstream[0].Cards[0].Path; got != "" {
		t.Errorf("current Path = %q, want unchanged", got)
	}

	RelativizeChainGraph(g, "/export")
	if got := g.Upstream[0].Cards[0].Path; got != "01-stakeholder-requirements.html" {
		t.Errorf("relative upstream Path = %q, want flat sibling name", got)
	}
	if got := g.Downstream[0].Cards[1].Path; got != "03-software-requirements.html" {
		t.Errorf("relative downstream Path = %q, want flat sibling name", got)
	}
}
