package graph

import (
	"reflect"
	"testing"

	"github.com/dVoo/reqmd/internal/model"
)

// evidenceChain builds the three-level chain SYS-001 ← TEST-001 ← RESULT,
// plus a direct result on SYS-001, with the tests document status supplied
// by the caller (approved is a coverage provider, draft is not).
func evidenceChain(t *testing.T, testStatus string) *Graph {
	t.Helper()
	sys := model.Document{
		Path:   "/docs/sys",
		Nodes:  []*model.Node{evidenceReq("SYS-001", "/docs/sys/sys.md", map[string]any{"verify": "Test"})},
		XReqmd: &model.XReqmd{Level: "system-requirements"},
	}
	tests := model.Document{
		Path: "/docs/tests",
		Nodes: []*model.Node{evidenceReq("TEST-001", "/docs/tests/t.md", map[string]any{
			"verify": "Test",
			"status": testStatus,
			"trace":  []any{"SYS-001"},
		})},
		XReqmd: &model.XReqmd{Level: "test-specs"},
	}
	results := model.Document{
		Path: "<verify-results>",
		Nodes: []*model.Node{
			evidenceReq("RESULT:SYS-001", "sys.ctrf.json", map[string]any{
				model.AttrResultOutcome: "pass",
				model.AttrTrace:         []any{"SYS-001"},
			}),
			evidenceReq("RESULT:TEST-001#BOOT-TIME", "tests.ctrf.json", map[string]any{
				model.AttrResultOutcome: "fail",
				model.AttrTrace:         []any{"TEST-001"},
			}),
		},
		XReqmd: &model.XReqmd{Level: "verify-results"},
	}
	g, err := New([]model.Document{sys, tests, results})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return g
}

// evidenceReq builds a requirement node for the graph.
func evidenceReq(id, source string, attrs map[string]any) *model.Node {
	return &model.Node{Kind: model.KindRequirement, ID: id, Source: source, Attrs: attrs}
}

// TestMeasureEvidence_CachedMatchesUncached compares the cached roll-up with
// the freshly computed one (first call uncached, second served from the
// cache) across a multi-level chain.
func TestMeasureEvidence_CachedMatchesUncached(t *testing.T) {
	for _, status := range []string{model.StatusApproved, model.StatusDraft} {
		t.Run(status, func(t *testing.T) {
			g := evidenceChain(t, status)
			self1, items1, drafts1 := g.measureEvidence("SYS-001")
			self2, items2, drafts2 := g.measureEvidence("SYS-001")
			if !reflect.DeepEqual(self1, self2) || !reflect.DeepEqual(items1, items2) || drafts1 != drafts2 {
				t.Errorf("cached result differs: (%v,%v,%d) vs (%v,%v,%d)",
					self1, items1, drafts1, self2, items2, drafts2)
			}
			if len(self1) != 1 || self1[0].File != "sys.ctrf.json" {
				t.Errorf("self = %+v, want the direct SYS-001 result", self1)
			}
			wantDrafts := 0
			if status == model.StatusDraft {
				wantDrafts = 1
			}
			if drafts1 != wantDrafts {
				t.Errorf("drafts = %d, want %d", drafts1, wantDrafts)
			}
			if status == model.StatusApproved && (len(items1) != 2 || items1[1].Case != "BOOT-TIME") {
				t.Errorf("items = %+v, want direct result plus the downstream case", items1)
			}
			if status == model.StatusDraft && len(items1) != 1 {
				t.Errorf("items = %+v, want the direct result only", items1)
			}
		})
	}
}

// TestMeasureEvidence_Memoized proves the roll-up is served from the cache:
// after the first call, dropping the chain's inbound edges does not change
// the answer until the cache is invalidated.
func TestMeasureEvidence_Memoized(t *testing.T) {
	g := evidenceChain(t, model.StatusApproved)
	_, before, _ := g.measureEvidence("SYS-001")

	g.nodes["TEST-001"].Inbound = nil
	if _, items, _ := g.measureEvidence("SYS-001"); !reflect.DeepEqual(items, before) {
		t.Errorf("second call recomputed instead of using the cache: %+v", items)
	}

	g.SetFilter(nil) // invalidates the evidence cache
	if _, items, _ := g.measureEvidence("SYS-001"); len(items) != 1 {
		t.Errorf("items after invalidation = %+v, want the direct result only", items)
	}
}

// TestMeasureEvidence_SetFilterInvalidates ensures filter-scoped evidence is
// recomputed after SetFilter instead of reusing the pre-filter answer.
func TestMeasureEvidence_SetFilterInvalidates(t *testing.T) {
	g := evidenceChain(t, model.StatusApproved)
	if _, items, _ := g.measureEvidence("SYS-001"); len(items) != 2 {
		t.Fatalf("unfiltered items = %+v, want 2", items)
	}

	g.SetFilter(map[string]struct{}{"SYS-001": {}})
	_, items, _ := g.measureEvidence("SYS-001")
	if len(items) != 1 || items[0].Outcome != "pass" {
		t.Errorf("filtered items = %+v, want only the direct passing result", items)
	}
	if nv, ok := g.MeasureVerdicts()["SYS-001"]; !ok || nv.Outcome != "pass" {
		t.Errorf("post-filter verdict = %+v (ok=%v), want pass", nv, ok)
	}
}

// TestVerificationStates_NoResultsSkipsRollup pins the early exit: without
// result nodes no evidence is gathered and every outcome stays empty.
func TestVerificationStates_NoResultsSkipsRollup(t *testing.T) {
	g, err := New([]model.Document{{
		Path:   "/docs/sys",
		Nodes:  []*model.Node{evidenceReq("SYS-001", "/docs/sys/sys.md", map[string]any{"verify": "Test"})},
		XReqmd: &model.XReqmd{Level: "system-requirements"},
	}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	states := g.VerificationStates()
	if st := states["SYS-001"]; !st.Eligible || st.Outcome != "" || st.Evidence != nil {
		t.Errorf("SYS-001 = %+v, want eligible with no verdict", st)
	}
	if g.evidenceCache != nil {
		t.Errorf("evidence cache populated without result nodes: %+v", g.evidenceCache)
	}
}
