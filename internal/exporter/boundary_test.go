package exporter

import (
	"testing"

	"reqmd/internal/model"
)

func TestComputeBoundaries_MixedTopology(t *testing.T) {
	// Doc A: root (no upstream sources) — referenced by B → not leaf.
	// Doc B: has upstream pointing at A → not root. Referenced by C → not leaf.
	// Doc C: has upstream pointing at B → not root. Not referenced → leaf.
	// Doc X: external → forced root. Not referenced → leaf.
	// Doc D: has upstream pointing at B (not root) AND referenced by E (not leaf).
	// Doc E: has upstream pointing at D (not root) AND not referenced → leaf.
	docs := []model.Document{
		{Path: "/spec/A"},
		{Path: "/spec/B", XReqmd: &model.XReqmd{Upstream: &model.TraceUpstream{Sources: []string{"../A/"}}}},
		{Path: "/spec/C", XReqmd: &model.XReqmd{Upstream: &model.TraceUpstream{Sources: []string{"../B/"}}}},
		{Path: "/spec/X", XReqmd: &model.XReqmd{External: true}},
		{Path: "/spec/D", XReqmd: &model.XReqmd{Upstream: &model.TraceUpstream{Sources: []string{"../B/"}}}},
		{Path: "/spec/E", XReqmd: &model.XReqmd{Upstream: &model.TraceUpstream{Sources: []string{"../D/"}}}},
	}

	got := ComputeDocBoundaries(docs)
	want := map[string]DocBoundary{
		"/spec/A": {IsRoot: true, IsLeaf: false},  // root, but referenced by B
		"/spec/B": {IsRoot: false, IsLeaf: false}, // middle: has up + referenced by C,D
		"/spec/C": {IsRoot: false, IsLeaf: true},  // leaf end
		"/spec/X": {IsRoot: true, IsLeaf: true},   // external + unreferenced
		"/spec/D": {IsRoot: false, IsLeaf: false}, // middle: has up + referenced by E
		"/spec/E": {IsRoot: false, IsLeaf: true},  // leaf end
	}
	if len(got) != len(want) {
		t.Fatalf("got %d docs, want %d", len(got), len(want))
	}
	for path, w := range want {
		g, ok := got[path]
		if !ok {
			t.Errorf("missing entry for %q", path)
			continue
		}
		if g != w {
			t.Errorf("%q: got %+v, want %+v", path, g, w)
		}
	}
	// Extra-keys guard: reject unexpected entries.
	for path := range got {
		if _, ok := want[path]; !ok {
			t.Errorf("unexpected extra entry for %q", path)
		}
	}
}

func TestComputeBoundaries_ExternalOverridesSources(t *testing.T) {
	// External doc with sources should still be root.
	docs := []model.Document{
		{Path: "/spec/ext", XReqmd: &model.XReqmd{
			External: true,
			Upstream: &model.TraceUpstream{Sources: []string{"../something/"}},
		}},
	}
	got := ComputeDocBoundaries(docs)
	if got["/spec/ext"].IsRoot != true {
		t.Errorf("external doc with sources should be root, got %+v", got["/spec/ext"])
	}
}

func TestComputeBoundaries_NilXReqmd(t *testing.T) {
	// Doc with no x-reqmd at all: should be root, and leaf if unreferenced.
	docs := []model.Document{
		{Path: "/spec/nil"},
	}
	got := ComputeDocBoundaries(docs)
	want := DocBoundary{IsRoot: true, IsLeaf: true}
	if got["/spec/nil"] != want {
		t.Errorf("nil-xreqmd standalone: got %+v, want %+v", got["/spec/nil"], want)
	}
}

func TestComputeBoundaries_EmptySources(t *testing.T) {
	// Doc with x-reqmd but no upstream sources: should still be root.
	docs := []model.Document{
		{Path: "/spec/empty", XReqmd: &model.XReqmd{Upstream: &model.TraceUpstream{Sources: []string{}}}},
	}
	got := ComputeDocBoundaries(docs)
	if got["/spec/empty"].IsRoot != true {
		t.Errorf("empty-sources doc should be root, got %+v", got["/spec/empty"])
	}
}
