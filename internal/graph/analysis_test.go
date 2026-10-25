package graph_test

import (
	"testing"

	"github.com/dVoo/reqmd/internal/graph"
	"github.com/dVoo/reqmd/internal/model"
)

// analysisDoc builds a document with a document-id so qualified trace refs
// and coverage sources resolve.
func analysisDoc(dir, docID, level string, reqs ...*model.Node) model.Document {
	return docWithXReqmd(dir, &model.XReqmd{DocumentID: docID, Level: level}, reqs...)
}

func attrs(kv ...any) map[string]any {
	m := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

// TestCoverageStates_Semantic pins the coverage states to the same rules the
// requires-trace-from check applies: approved providers only, draft inbound
// requirements ignored, explicit opt-out, unknown sources unresolved.
func TestCoverageStates_Semantic(t *testing.T) {
	// The expectation lives on the requirement being covered: an approved
	// requirement from the named document must trace TO it.
	system := func(status string, extra ...*model.Node) model.Document {
		nodes := []*model.Node{
			req("SYS-001", "/docs/system/sys.md", attrs(
				model.AttrStatus, status,
				model.AttrRequiresTraceFrom, []any{"software"},
			)),
		}
		return analysisDoc("/docs/system", "system", "system", append(nodes, extra...)...)
	}
	software := func(status string, extra ...*model.Node) model.Document {
		nodes := []*model.Node{
			req("SW-001", "/docs/software/sw.md", attrs(
				model.AttrStatus, status,
				model.AttrTrace, []any{"system/SYS-001"},
			)),
		}
		return analysisDoc("/docs/software", "software", "software", append(nodes, extra...)...)
	}

	tests := []struct {
		name      string
		docs      []model.Document
		wantState string
		wantDraft int
	}{
		{
			name:      "approved downstream covers",
			docs:      []model.Document{system(model.StatusApproved), software(model.StatusApproved)},
			wantState: graph.CoverageCovered,
		},
		{
			name:      "draft downstream does not cover",
			docs:      []model.Document{system(model.StatusApproved), software(model.StatusDraft)},
			wantState: graph.CoverageMissing,
			wantDraft: 1,
		},
		{
			name: "explicit empty opt-out is not required",
			docs: []model.Document{
				software(model.StatusApproved),
				analysisDoc("/docs/system", "system", "system",
					req("SYS-001", "/docs/system/sys.md", attrs(
						model.AttrRequiresTraceFrom, []any{},
					)),
				),
			},
			wantState: graph.CoverageNotRequired,
		},
		{
			name: "unknown source is unresolved",
			docs: []model.Document{
				software(model.StatusApproved),
				analysisDoc("/docs/system", "system", "system",
					req("SYS-001", "/docs/system/sys.md", attrs(
						model.AttrRequiresTraceFrom, []any{"nope"},
					)),
				),
			},
			wantState: graph.CoverageUnresolved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := graph.New(tt.docs)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			cov := g.CoverageStates()["SYS-001"]
			if cov.State != tt.wantState {
				t.Errorf("state = %q, want %q (misses: %+v)", cov.State, tt.wantState, cov.Misses)
			}
			if cov.Drafts != tt.wantDraft {
				t.Errorf("drafts = %d, want %d", cov.Drafts, tt.wantDraft)
			}
		})
	}
}

// TestCoverageStates_GenericFallback checks the boundary-aware fallback used
// when a requirement declares no requires-trace-from expectation.
func TestCoverageStates_GenericFallback(t *testing.T) {
	// Root document: no upstream sources → coverage not required.
	root := docWithXReqmd("/docs/root", &model.XReqmd{DocumentID: "root", Level: "stakeholder"},
		req("STK-001", "/docs/root/stk.md", attrs(model.AttrStatus, model.StatusApproved)),
	)
	// Mid document, no explicit expectation: a requirement someone traces to
	// is covered; an orphan in the same document is missing coverage.
	mid := docWithXReqmd("/docs/mid", &model.XReqmd{
		DocumentID: "mid", Level: "system",
		Upstream: &model.TraceUpstream{Sources: []string{"/docs/root"}},
	},
		req("SYS-001", "/docs/mid/sys.md", attrs()),
		req("SYS-ORPHAN", "/docs/mid/orphan.md", attrs()),
	)
	// Leaf document so both nodes are not also boundary exemptions.
	leaf := docWithXReqmd("/docs/leaf", &model.XReqmd{
		DocumentID: "leaf", Level: "software",
		Upstream: &model.TraceUpstream{Sources: []string{"/docs/mid"}},
	},
		req("SW-001", "/docs/leaf/sw.md", attrs(model.AttrTrace, []any{"mid/SYS-001"})),
	)

	g, err := graph.New([]model.Document{root, mid, leaf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	states := g.CoverageStates()
	if got := states["STK-001"].State; got != graph.CoverageNotRequired {
		t.Errorf("root STK-001 = %q, want not-required", got)
	}
	if got := states["SYS-001"].State; got != graph.CoverageCovered {
		t.Errorf("traced SYS-001 = %q, want covered", got)
	}
	if got := states["SYS-ORPHAN"].State; got != graph.CoverageMissing {
		t.Errorf("SYS-ORPHAN = %q, want missing-coverage", got)
	}
}

// TestTraceRefs covers pin/version annotation and unresolved links.
func TestTraceRefs(t *testing.T) {
	up := analysisDoc("/docs/up", "up", "system",
		req("UP-001", "/docs/up/up.md", attrs(model.AttrVersion, 3)),
	)
	down := analysisDoc("/docs/down", "down", "software",
		req("SW-001", "/docs/down/sw.md", attrs(
			model.AttrTrace, []any{"up/UP-001~2", "up/MISSING", "UP-001"},
		)),
	)
	g, err := graph.New([]model.Document{up, down})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	links := g.TraceRefs("SW-001")
	if len(links) != 3 {
		t.Fatalf("got %d links, want 3: %+v", len(links), links)
	}
	first := links[0]
	if !first.Resolved || first.ID != "UP-001" || !first.Pinned || first.Pin != 2 || first.Version != 3 {
		t.Errorf("pinned link = %+v, want resolved UP-001 pin 2 version 3", first)
	}
	if links[1].Resolved || links[1].ID != "" {
		t.Errorf("dangling link = %+v, want unresolved", links[1])
	}
	if links[2].Pinned {
		t.Errorf("unpinned link = %+v, want Pinned=false", links[2])
	}
}

// TestVerificationStates checks that eligibility follows the missing-verdict
// gate and that unverified eligible measures are visible as empty outcomes.
func TestVerificationStates(t *testing.T) {
	measure := func(id, status string, extra ...any) *model.Node {
		a := attrs(append([]any{model.AttrVerify, "Test", model.AttrStatus, status}, extra...)...)
		return req(id, "/docs/test/"+id+".md", a)
	}
	docs := []model.Document{
		measureDoc(
			measure("TST-PASS", model.StatusApproved),
			measure("TST-NONE", model.StatusApproved),
			measure("TST-DRAFT", model.StatusDraft),
			measure("TST-DEFERRED", model.StatusApproved, model.AttrDisposition, "deferred"),
		),
		resultDoc(resultReq("TST-PASS", "pass"), resultReq("TST-DRAFT", "fail")),
	}
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	states := g.VerificationStates()

	if st := states["TST-PASS"]; !st.Eligible || st.Outcome != "pass" {
		t.Errorf("TST-PASS = %+v, want eligible pass", st)
	}
	if st := states["TST-NONE"]; !st.Eligible || st.Outcome != "" {
		t.Errorf("TST-NONE = %+v, want eligible with no verdict", st)
	}
	if st := states["TST-DRAFT"]; st.Eligible || st.Outcome != "fail" {
		t.Errorf("TST-DRAFT = %+v, want not eligible but verdict fail", st)
	}
	if st := states["TST-DEFERRED"]; st.Eligible {
		t.Errorf("TST-DEFERRED = %+v, want not eligible", st)
	}
	if _, ok := states["RESULT:TST-PASS"]; ok {
		t.Error("result pseudo-requirement must not appear in verification states")
	}
}
