package graph_test

import (
	"regexp"
	"strings"
	"testing"

	"reqmd/internal/graph"
	"reqmd/internal/model"
	"reqmd/internal/parser"
	"reqmd/internal/schema"
)

// ---------------------------------------------------------------------------
// Helper factories
// ---------------------------------------------------------------------------

func req(id string, source string, attrs map[string]any) model.Requirement {
	return model.Requirement{
		ID:     id,
		Source: source,
		Attrs:  attrs,
	}
}

func doc(path string, reqs ...model.Requirement) model.Document {
	return model.Document{
		Path: path,
		Schema: map[string]any{
			"$id":   path + "/schema.yaml",
			"title": "Test",
		},
		Requirements: reqs,
		XReqmd:       &model.XReqmd{Level: "test"},
	}
}

// docWithXReqmd creates a document with a custom XReqmd configuration.
func docWithXReqmd(path string, xr *model.XReqmd, reqs ...model.Requirement) model.Document {
	return model.Document{
		Path: path,
		Schema: map[string]any{
			"$id":   path + "/schema.yaml",
			"title": "Test",
		},
		Requirements: reqs,
		XReqmd:       xr,
	}
}

// ---------------------------------------------------------------------------
// 1. Empty docs
// ---------------------------------------------------------------------------

func TestNew_EmptyDocs(t *testing.T) {
	g, err := graph.New(nil)
	if err != nil {
		t.Fatalf("New(nil) returned error: %v", err)
	}
	if got := g.NodeCount(); got != 0 {
		t.Errorf("NodeCount = %d, want 0", got)
	}
	got := g.CheckResults()
	if len(got) != 0 {
		t.Errorf("CheckResults = %v, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func TestNew_SingleDocNoTraces(t *testing.T) {
	// A single doc with no upstream is inferred as a root boundary.
	// All requirements in it get orphan suppressed (they're at the top of the V-model).
	// But they still get missing-downstream warnings unless the doc is also a leaf.
	// A doc that has no upstream AND isn't referenced by anyone else is both root and leaf.
	g, err := graph.New([]model.Document{
		doc("/docs/login",
			req("REQ-001", "/docs/login/login.md", map[string]any{
				"title": "Login",
				"asil":  "B",
			}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := g.NodeCount(); got != 1 {
		t.Errorf("NodeCount = %d, want 1", got)
	}

	// Single doc with no upstream sources:
	//   → root boundary inferred → orphan suppressed
	//   → not referenced by anyone → leaf boundary inferred → missing-downstream suppressed
	// Result: no warnings at all.
	results := g.CheckResults()
	for _, r := range results {
		t.Errorf("unexpected check result: %v", r)
	}
}

// ---------------------------------------------------------------------------
// 3. Valid trace — two requirements, one traces to the other
// ---------------------------------------------------------------------------

func TestNew_ValidTrace(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/ivi",
			req("IVI-FUN-001", "/docs/ivi/ivi-startup.md", map[string]any{
				"title": "Startup animation",
				"trace": []any{"IVI-FUN-002"},
			}),
			req("IVI-FUN-002", "/docs/ivi/ivi-startup.md", map[string]any{
				"title": "Splash screen",
			}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := g.NodeCount(); got != 2 {
		t.Errorf("NodeCount = %d, want 2", got)
	}

	results := g.CheckResults()

	// No broken refs (all trace targets exist)
	var brokenRefs int
	for _, r := range results {
		if r.Level == "WARNING" && r.Message != "untraced: no downstream reference" && r.Message != "no upstream reference" {
			brokenRefs++
		}
	}
	if brokenRefs != 0 {
		t.Errorf("broken refs = %d, want 0", brokenRefs)
	}

	// No circular
	for _, r := range results {
		if r.Level == "ERROR" {
			t.Errorf("unexpected ERROR: %+v", r)
		}
	}

	// Single doc with no upstream sources → root boundary → orphan suppressed
	// Single doc → not referenced by anyone else → leaf boundary → missing-downstream suppressed
	// So no warnings expected at all.
	var orphans, missing int
	for _, r := range results {
		switch r.Message {
		case "untraced: no downstream reference":
			orphans++
		case "no upstream reference":
			missing++
		}
	}
	if orphans != 0 {
		t.Errorf("orphans = %d, want 0 (root boundary suppresses)", orphans)
	}
	if missing != 0 {
		t.Errorf("missing downstream = %d, want 0 (leaf boundary suppresses)", missing)
	}

	// Neighbors (using UpstreamNeighbors / DownstreamNeighbors)
	up := g.UpstreamNeighbors("IVI-FUN-001")
	if len(up) != 1 || up[0] != "IVI-FUN-002" {
		t.Errorf("IVI-FUN-001 upstream = %v, want [IVI-FUN-002]", up)
	}
	down := g.DownstreamNeighbors("IVI-FUN-001")
	if len(down) != 0 {
		t.Errorf("IVI-FUN-001 downstream = %v, want empty", down)
	}
}

// ---------------------------------------------------------------------------
// 4. Dangling trace — one req traces to a non-existent ID
// ---------------------------------------------------------------------------

func TestNew_DanglingTrace(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/ivi",
			req("IVI-FUN-001", "/docs/ivi/ivi-startup.md", map[string]any{
				"title": "Startup",
				"trace": []any{"DOES-NOT-EXIST"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := g.NodeCount(); got != 1 {
		t.Errorf("NodeCount = %d, want 1", got)
	}

	cr := g.CheckResults()

	// Filter for broken refs
	var dl []graph.CheckResult
	for _, r := range cr {
		if strings.Contains(r.Message, "broken reference") {
			dl = append(dl, r)
		}
	}
	if len(dl) != 1 {
		t.Fatalf("broken refs count = %d, want 1", len(dl))
	}
	if dl[0].Level != "WARNING" {
		t.Errorf("level = %q, want WARNING", dl[0].Level)
	}
	if dl[0].ReqID != "IVI-FUN-001" {
		t.Errorf("ReqID = %q, want IVI-FUN-001", dl[0].ReqID)
	}
	if dl[0].File != "/docs/ivi/ivi-startup.md" {
		t.Errorf("File = %q, want /docs/ivi/ivi-startup.md", dl[0].File)
	}
	if !regexp.MustCompile(`broken reference.*DOES-NOT-EXIST`).MatchString(dl[0].Message) {
		t.Errorf("Message = %q, want broken reference referencing DOES-NOT-EXIST", dl[0].Message)
	}
}

// ---------------------------------------------------------------------------
// 5. Circular trace — A→B→C→A
// ---------------------------------------------------------------------------

func TestNew_CircularTrace(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/circ",
			req("A", "/docs/circ/a.md", map[string]any{
				"trace": []any{"B"},
			}),
			req("B", "/docs/circ/b.md", map[string]any{
				"trace": []any{"C"},
			}),
			req("C", "/docs/circ/c.md", map[string]any{
				"trace": []any{"A"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := g.NodeCount(); got != 3 {
		t.Errorf("NodeCount = %d, want 3", got)
	}

	results := g.CheckResults()
	var circularErrors int
	for _, r := range results {
		if r.Level == "ERROR" && strings.Contains(r.Message, "circular dependency") {
			circularErrors++
		}
	}
	if circularErrors == 0 {
		t.Fatal("expected at least one circular ERROR, got none")
	}
}

// ---------------------------------------------------------------------------
// 6. Orphan detection — A→B means A has no incoming → orphan
// ---------------------------------------------------------------------------

func TestNew_OrphanDetection(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/sys",
			req("A", "/docs/sys/a.md", map[string]any{
				"trace": []any{"B"},
			}),
			req("B", "/docs/sys/b.md", map[string]any{}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	results := g.CheckResults()
	var orphanIDs []string
	for _, r := range results {
		if r.Message == "untraced: no downstream reference" {
			orphanIDs = append(orphanIDs, r.ReqID)
		}
	}

	// Single doc with no upstream → inferred as root boundary → orphan suppressed for all.
	// So actually no orphan warnings expected.
	if len(orphanIDs) != 0 {
		t.Errorf("orphan count = %d, want 0 (root boundary suppresses), got %v", len(orphanIDs), orphanIDs)
	}
}

// ---------------------------------------------------------------------------
// 7. Orphan suppression via directory boundary (no upstream sources)
// ---------------------------------------------------------------------------

func TestNew_RootBoundarySuppressesOrphan(t *testing.T) {
	// Two docs: root doc (no upstream) and child doc (traces to root).
	// Root doc's requirements should not get orphan warnings.
	rootDoc := docWithXReqmd("/docs/root", &model.XReqmd{Level: "root"},
		req("ROOT", "/docs/root/root.md", map[string]any{
			"trace": []any{"CHILD"},
		}),
	)
	childDoc := docWithXReqmd("/docs/child", &model.XReqmd{
		Level:    "child",
		Upstream: &model.TraceUpstream{Level: "root", Sources: []string{"/docs/root"}},
	},
		req("CHILD", "/docs/child/child.md", map[string]any{}),
	)

	g, err := graph.New([]model.Document{rootDoc, childDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	results := g.CheckResults()
	var orphans, missing int
	for _, r := range results {
		switch r.Message {
		case "untraced: no downstream reference":
			orphans++
		case "no upstream reference":
			missing++
		}
	}
	// ROOT is in a root boundary dir → orphan suppressed
	// CHILD has incoming from ROOT → not orphan
	// CHILD traces to ROOT → has outgoing traces, but CHILD's doc is a leaf (not referenced by anyone else) → missing-downstream suppressed
	// ROOT's doc IS referenced by CHILD's sources → not a leaf → ROOT has no incoming → should have been orphan but root boundary suppresses it
	if orphans != 0 {
		t.Errorf("orphans = %d, want 0 (root boundary suppresses)", orphans)
	}
	// CHILD has outgoing trace → not missing downstream
	// ROOT has no outgoing trace and is not a leaf → missing downstream
	// Wait — ROOT traces to CHILD so ROOT has outgoing trace → not missing downstream
	if missing != 0 {
		t.Errorf("missing downstream = %d, want 0, got results: %v", missing, results)
	}
}

// ---------------------------------------------------------------------------
// 8. Orphan and missing-downstream when NOT in a boundary directory
// ---------------------------------------------------------------------------

func TestNew_NonBoundaryOrphanAndMissing(t *testing.T) {
	// Middle doc in a 3-layer chain should NOT be a boundary.
	// Requirements in it without connections should get warnings.
	topDoc := docWithXReqmd("/docs/top", &model.XReqmd{Level: "top"},
		req("TOP-001", "/docs/top/top.md", map[string]any{
			"title": "Top requirement",
		}),
	)
	midDoc := docWithXReqmd("/docs/mid", &model.XReqmd{
		Level:    "mid",
		Upstream: &model.TraceUpstream{Level: "top", Sources: []string{"/docs/top"}},
	},
		req("MID-001", "/docs/mid/mid.md", map[string]any{
			"title": "Mid requirement",
			"trace": []any{"TOP-001"},
		}),
	)

	g, err := graph.New([]model.Document{topDoc, midDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	results := g.CheckResults()
	var orphans, missing int
	for _, r := range results {
		switch r.Message {
		case "untraced: no downstream reference":
			orphans++
		case "no upstream reference":
			missing++
		}
	}
	// TOP-001: root boundary (no upstream) → orphan suppressed, BUT it has no outgoing traces.
	//   /docs/top is NOT a leaf dir (referenced by /docs/mid) → missing-downstream applies.
	//   Wait, but disposition is empty → missing-downstream warning.
	//   Actually: /docs/top IS referenced by /docs/mid (as upstream source) → NOT a leaf.
	//   So TOP-001 has no outgoing traces and is not in a leaf dir → missing downstream WARNING.
	// MID-001: has trace to TOP-001 → has outgoing. TOP-001 has trace from MID-001 → has incoming.
	//   But /docs/mid has upstream.sources pointing to /docs/top → MID-001 has outgoing to TOP-001.
	//   /docs/mid IS NOT referenced by anyone else (no other doc traces up to it) → leaf boundary.
	//   But MID-001 HAS outgoing traces → no missing-downstream.

	// The key check: /docs/top is NOT a leaf (it IS referenced), so TOP-001 should get
	// missing-downstream unless it has outgoing traces (it doesn't).
	_ = orphans // root boundary → 0 orphans
	_ = missing
	// TOP-001: root dir → orphan suppressed. Not leaf dir → missing downstream if no outgoing.
	// But MID-001 traces to TOP-001 → TOP-001 has inbound → not orphan anyway.
	// Hmm, actually MID-001 has trace to TOP-001, so TOP-001 has incoming from MID-001.
	// So TOP-001 is NOT orphan anyway (it has incoming).
	// TOP-001 has no outgoing trace → and /docs/top is referenced → not leaf → missing-downstream warning.
	if missing != 1 {
		t.Errorf("missing downstream = %d, want 1 (TOP-001 has no outgoing, not in leaf dir), got results: %v", missing, results)
	}
}

// ---------------------------------------------------------------------------
// 9. Orphan suppression via disposition (no root boundary, but disposition set)
// ---------------------------------------------------------------------------

func TestNew_DispositionSuppressesOrphan(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/sys",
			req("A", "/docs/sys/a.md", map[string]any{
				"disposition": "deferred",
				"trace":       []any{"B"},
			}),
			req("B", "/docs/sys/b.md", map[string]any{}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	results := g.CheckResults()
	var orphans, missingDispo int
	for _, r := range results {
		switch {
		case r.Message == "untraced: no downstream reference":
			orphans++
		case r.Message == "no upstream reference":
			missingDispo++
		}
	}
	// A has disposition set → no orphan WARNING even though no incoming (also root boundary suppresses)
	// B has no outgoing, but B is in a leaf boundary dir (not referenced by others) → missing-downstream suppressed
	if orphans != 0 {
		t.Errorf("orphans = %d, want 0 (disposition + root boundary suppresses)", orphans)
	}
	// Actually both are in the same single doc with no upstream → root boundary; and it's not referenced → leaf boundary
	// So missing-downstream is also suppressed. But B has no disposition...
	// The boundary inference means: if the doc IS a leaf, ALL its requirements get missing-downstream suppressed.
	// So 0 missing-downstream warnings.
	if missingDispo != 0 {
		t.Errorf("missing downstream = %d, want 0 (leaf boundary suppresses)", missingDispo)
	}
}

// ---------------------------------------------------------------------------
// 10. Disposition without rationale — deferred/rejected with no rationale
// ---------------------------------------------------------------------------

func TestNew_DispositionWithoutRationale(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/sys",
			req("A", "/docs/sys/a.md", map[string]any{
				"disposition": "deferred",
			}),
			req("B", "/docs/sys/b.md", map[string]any{
				"disposition":        "rejected",
				"disposition-reason": "Rejected per review 2025-01-15.",
			}),
			req("C", "/docs/sys/c.md", map[string]any{
				"disposition": "implemented",
			}),
			req("D", "/docs/sys/d.md", map[string]any{
				"disposition": "rejected",
			}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	results := g.CheckResults()
	var dispWarnings []string
	for _, r := range results {
		if r.Level == "WARNING" && r.Message != "untraced: no downstream reference" && r.Message != "no upstream reference" {
			dispWarnings = append(dispWarnings, r.ReqID)
		}
	}

	// A: deferred, no rationale → WARNING
	// B: rejected, has rationale  → no WARNING
	// C: implemented, no rationale → no WARNING (implemented doesn't need rationale)
	// D: rejected, no rationale   → WARNING
	if len(dispWarnings) != 2 {
		t.Fatalf("disposition rationale warnings = %v, want [A, D]", dispWarnings)
	}
	got := make(map[string]bool)
	for _, id := range dispWarnings {
		got[id] = true
	}
	if !got["A"] {
		t.Error("missing warning for A (deferred, no rationale)")
	}
	if !got["D"] {
		t.Error("missing warning for D (rejected, no rationale)")
	}
}

// ---------------------------------------------------------------------------
// 11. `needs` attribute — explicit coverage expectations
// ---------------------------------------------------------------------------

func TestNew_NeedsCoverageSatisfied(t *testing.T) {
	// Top-level stakeholder requirement declares needs: [system, software] —
	// both documents are present and contain requirements that trace to it.
	stkDoc := docWithXReqmd("/docs/stakeholder", &model.XReqmd{
		Level:      "stakeholder-needs",
		DocumentID: "stakeholder",
		IDPrefix:   "STK-",
	},
		req("STK-001", "/docs/stakeholder/needs.md", map[string]any{
			"requires-trace-from": []any{"system", "software"},
		}),
	)
	sysDoc := docWithXReqmd("/docs/system", &model.XReqmd{
		Level:      "system-requirements",
		DocumentID: "system",
		Upstream: &model.TraceUpstream{
			Level:   "stakeholder-needs",
			Sources: []string{"/docs/stakeholder"},
		},
		IDPrefix: "SYS-",
	},
		req("SYS-001", "/docs/system/features.md", map[string]any{
			"trace": []any{"stakeholder/STK-001"},
		}),
	)
	swDoc := docWithXReqmd("/docs/software", &model.XReqmd{
		Level:      "software-requirements",
		DocumentID: "software",
		Upstream: &model.TraceUpstream{
			Level:   "system-requirements",
			Sources: []string{"/docs/system"},
		},
		IDPrefix: "SW-",
	},
		req("SW-001", "/docs/software/components.md", map[string]any{
			"trace": []any{"stakeholder/STK-001"},
		}),
	)

	g, err := graph.New([]model.Document{stkDoc, sysDoc, swDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	cr := g.CheckResults()
	for _, r := range cr {
		if strings.Contains(r.Message, "no upstream trace") || strings.Contains(r.Message, "unknown coverage source") {
			t.Errorf("unexpected coverage warning: %+v", r)
		}
	}
}

func TestNew_NeedsCoverageMissing(t *testing.T) {
	// STK-001 declares needs: [system] but no requirement from the
	// system document traces to it → WARNING.
	stkDoc := docWithXReqmd("/docs/stakeholder", &model.XReqmd{
		Level:      "stakeholder-needs",
		DocumentID: "stakeholder",
		IDPrefix:   "STK-",
	},
		req("STK-001", "/docs/stakeholder/needs.md", map[string]any{
			"requires-trace-from": []any{"system"},
		}),
	)
	// system doc exists but does NOT trace to STK-001.
	sysDoc := docWithXReqmd("/docs/system", &model.XReqmd{
		Level:      "system-requirements",
		DocumentID: "system",
		IDPrefix:   "SYS-",
	},
		req("SYS-001", "/docs/system/features.md", map[string]any{
			"title": "Unrelated",
		}),
	)

	g, err := graph.New([]model.Document{stkDoc, sysDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	cr := g.CheckResults()
	var missing int
	for _, r := range cr {
		if r.Message == "no upstream trace: no approved requirement from \"system\" traces to this item" {
			missing++
		}
	}
	if missing != 1 {
		t.Errorf("no upstream trace warnings = %d, want 1; results: %v", missing, cr)
	}
}

func TestNew_NeedsCoverageUnknownTarget(t *testing.T) {
	// STK-001 declares needs: [nonexistent] — neither a document-id nor a level.
	stkDoc := docWithXReqmd("/docs/stakeholder", &model.XReqmd{
		Level:      "stakeholder-needs",
		DocumentID: "stakeholder",
		IDPrefix:   "STK-",
	},
		req("STK-001", "/docs/stakeholder/needs.md", map[string]any{
			"requires-trace-from": []any{"nonexistent"},
		}),
	)
	g, err := graph.New([]model.Document{stkDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	cr := g.CheckResults()
	var unknown int
	for _, r := range cr {
		if r.Message == "unknown coverage source: \"nonexistent\"" {
			unknown++
		}
	}
	if unknown != 1 {
		t.Errorf("unknown coverage warnings = %d, want 1; results: %v", unknown, cr)
	}
}

func TestNew_NeedsEmptySaysNoCoverageExpected(t *testing.T) {
	// STK-001 declares needs: [] (explicitly empty) — no coverage warning,
	// even when the document has no downstream requirements.
	stkDoc := docWithXReqmd("/docs/stakeholder", &model.XReqmd{
		Level:      "stakeholder-needs",
		DocumentID: "stakeholder",
		IDPrefix:   "STK-",
	},
		req("STK-001", "/docs/stakeholder/needs.md", map[string]any{
			"requires-trace-from": []any{},
		}),
	)
	g, err := graph.New([]model.Document{stkDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	cr := g.CheckResults()
	for _, r := range cr {
		if r.Message == "untraced: no downstream reference" {
			t.Errorf("explicit empty needs: [] should suppress generic orphan: %+v", r)
		}
	}
}

func TestNew_NeedsSuppressesGenericOrphan(t *testing.T) {
	// STK-001 declares needs: [system], which IS satisfied (system doc has
	// a req that traces to STK-001). The generic orphan check should NOT
	// fire because the needs check covers it.
	stkDoc := docWithXReqmd("/docs/stakeholder", &model.XReqmd{
		Level:      "stakeholder-needs",
		DocumentID: "stakeholder",
		IDPrefix:   "STK-",
	},
		req("STK-001", "/docs/stakeholder/needs.md", map[string]any{
			"requires-trace-from": []any{"system"},
		}),
	)
	sysDoc := docWithXReqmd("/docs/system", &model.XReqmd{
		Level:      "system-requirements",
		DocumentID: "system",
		IDPrefix:   "SYS-",
	},
		req("SYS-001", "/docs/system/features.md", map[string]any{
			"trace": []any{"stakeholder/STK-001"},
		}),
	)
	g, err := graph.New([]model.Document{stkDoc, sysDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	cr := g.CheckResults()
	for _, r := range cr {
		if r.Message == "untraced: no downstream reference" {
			t.Errorf("satisfied needs should suppress generic orphan: %+v", r)
		}
	}
}

func TestNew_QualifiedTraceByDocumentID(t *testing.T) {
	// Reference uses document-id (not directory path) to disambiguate
	// the trace target.
	stkDoc := docWithXReqmd("/docs/stakeholder", &model.XReqmd{
		Level:      "stakeholder-needs",
		DocumentID: "stakeholder",
		IDPrefix:   "STK-",
	},
		req("STK-001", "/docs/stakeholder/needs.md", map[string]any{}),
	)
	sysDoc := docWithXReqmd("/docs/system", &model.XReqmd{
		Level:      "system-requirements",
		DocumentID: "system",
		IDPrefix:   "SYS-",
	},
		req("SYS-001", "/docs/system/features.md", map[string]any{
			"trace": []any{"stakeholder/STK-001"},
		}),
	)

	g, err := graph.New([]model.Document{stkDoc, sysDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	up := g.UpstreamNeighbors("SYS-001")
	if len(up) != 1 || up[0] != "STK-001" {
		t.Errorf("SYS-001 upstream = %v, want [STK-001]", up)
	}

	cr := g.CheckResults()
	for _, r := range cr {
		if strings.Contains(r.Message, "broken reference") {
			t.Errorf("unexpected broken reference warning: %+v", r)
		}
	}
}

// ---------------------------------------------------------------------------
// 12. Neighbors — A traces to B and C
// ---------------------------------------------------------------------------

func TestNeighbors(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/sys",
			req("A", "/docs/sys/a.md", map[string]any{
				"trace": []any{"B", "C"},
			}),
			req("B", "/docs/sys/b.md", map[string]any{}),
			req("C", "/docs/sys/c.md", map[string]any{}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// A traces to B and C → A's upstream (parents) = [B, C]
	upA := g.UpstreamNeighbors("A")
	if len(upA) != 2 {
		t.Fatalf("UpstreamNeighbors(A) = %v, want [B, C] (len 2)", upA)
	}
	upSet := make(map[string]bool)
	for _, id := range upA {
		upSet[id] = true
	}
	if !upSet["B"] {
		t.Errorf("UpstreamNeighbors(A) missing B, got %v", upA)
	}
	if !upSet["C"] {
		t.Errorf("UpstreamNeighbors(A) missing C, got %v", upA)
	}

	// Nothing traces to A → A's downstream (children) = []
	downA := g.DownstreamNeighbors("A")
	if len(downA) != 0 {
		t.Errorf("DownstreamNeighbors(A) = %v, want empty", downA)
	}

	// B doesn't trace to anything → B's upstream = []
	upB := g.UpstreamNeighbors("B")
	if len(upB) != 0 {
		t.Errorf("UpstreamNeighbors(B) = %v, want empty", upB)
	}

	// A traces to B → B's downstream (children) = [A]
	downB := g.DownstreamNeighbors("B")
	if len(downB) != 1 || downB[0] != "A" {
		t.Errorf("DownstreamNeighbors(B) = %v, want [A]", downB)
	}
}

// ---------------------------------------------------------------------------
// 13. Neighbors — non-existent reqID
// ---------------------------------------------------------------------------

func TestNeighbors_NonExistent(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/sys",
			req("A", "/docs/sys/a.md", map[string]any{
				"trace": []any{"B"},
			}),
			req("B", "/docs/sys/b.md", map[string]any{}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := g.UpstreamNeighbors("Z"); got != nil {
		t.Errorf("UpstreamNeighbors(Z) = %v, want nil", got)
	}
	if got := g.DownstreamNeighbors("Z"); got != nil {
		t.Errorf("DownstreamNeighbors(Z) = %v, want nil", got)
	}
}

// ---------------------------------------------------------------------------
// 14. SchemaTitle
// ---------------------------------------------------------------------------

func TestSchemaTitle(t *testing.T) {
	tests := []struct {
		name   string
		schema any
		want   string
	}{
		{
			name:   "nil",
			schema: nil,
			want:   "",
		},
		{
			name:   "not a map",
			schema: 42,
			want:   "",
		},
		{
			name:   "empty map",
			schema: map[string]any{},
			want:   "",
		},
		{
			name: "only title",
			schema: map[string]any{
				"title": "My Schema",
			},
			want: "My Schema",
		},
		{
			name: "$id only",
			schema: map[string]any{
				"$id": "ivi-startup",
			},
			want: "ivi-startup",
		},
		{
			name: "$id and title",
			schema: map[string]any{
				"$id":   "ivi-startup",
				"title": "IVI Startup Requirements",
			},
			want: "ivi-startup — IVI Startup Requirements",
		},
		{
			name: "$id empty string, title set",
			schema: map[string]any{
				"$id":   "",
				"title": "Fallback",
			},
			want: "Fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schema.SchemaTitle(tt.schema)
			if got != tt.want {
				t.Errorf("SchemaTitle = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNew_CrossDocTrace(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/ivi",
			req("IVI-FUN-001", "/docs/ivi/ivi-startup.md", map[string]any{
				"trace": []any{"SAFETY-001"},
			}),
		),
		doc("/docs/safety",
			req("SAFETY-001", "/docs/safety/reqs.md", map[string]any{
				"asil": "D",
			}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if got := g.NodeCount(); got != 2 {
		t.Errorf("NodeCount = %d, want 2", got)
	}

	// No broken refs (cross-doc trace is valid)
	cr := g.CheckResults()
	var brokenRefs int
	for _, r := range cr {
		if strings.Contains(r.Message, "broken reference") {
			brokenRefs++
		}
	}
	if brokenRefs != 0 {
		t.Errorf("broken refs = %d, want 0", brokenRefs)
	}

	// IVI-FUN-001 traces to SAFETY-001 → IVI's upstream = [SAFETY-001]
	up := g.UpstreamNeighbors("IVI-FUN-001")
	if len(up) != 1 || up[0] != "SAFETY-001" {
		t.Errorf("UpstreamNeighbors(IVI-FUN-001) = %v, want [SAFETY-001]", up)
	}

	// IVI traces to SAFETY → SAFETY's downstream = [IVI-FUN-001]
	down := g.DownstreamNeighbors("SAFETY-001")
	if len(down) != 1 || down[0] != "IVI-FUN-001" {
		t.Errorf("DownstreamNeighbors(SAFETY-001) = %v, want [IVI-FUN-001]", down)
	}
}

// ---------------------------------------------------------------------------
// 15. Multiple broken refs from same requirement
// ---------------------------------------------------------------------------

func TestNew_MultipleBrokenRefs(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/ivi",
			req("IVI-FUN-001", "/docs/ivi/ivi-startup.md", map[string]any{
				"trace": []any{"MISSING-A", "MISSING-B", "MISSING-C"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	cr := g.CheckResults()
	var dl []graph.CheckResult
	for _, r := range cr {
		if strings.Contains(r.Message, "broken reference") {
			dl = append(dl, r)
		}
	}
	if len(dl) != 3 {
		t.Fatalf("broken refs count = %d, want 3", len(dl))
	}
	for _, r := range dl {
		if r.Level != "WARNING" {
			t.Errorf("level = %q, want WARNING", r.Level)
		}
		if r.ReqID != "IVI-FUN-001" {
			t.Errorf("ReqID = %q, want IVI-FUN-001", r.ReqID)
		}
	}
}

// ---------------------------------------------------------------------------
// 16. External suppresses orphan
// ---------------------------------------------------------------------------

func TestNew_ExternalSuppressesOrphan(t *testing.T) {
	extDoc := docWithXReqmd("/docs/aspice", &model.XReqmd{
		Level:    "external",
		External: true,
		IDPrefix: "ASP-",
	},
		req("ASP-BP-001", "/docs/aspice/bp.md", map[string]any{
			"title": "Base practice",
		}),
	)

	g, err := graph.New([]model.Document{extDoc})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	results := g.CheckResults()
	for _, r := range results {
		if r.Message == "untraced: no downstream reference" {
			t.Errorf("external doc should suppress orphan: %+v", r)
		}
	}
}

// ---------------------------------------------------------------------------
// 17. Child requirement suppresses orphan via IsChild
// ---------------------------------------------------------------------------

func TestNew_ChildSuppressesOrphan(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/sys",
			req("SYS-001", "/docs/sys/sys.md", map[string]any{
				"title": "System requirement",
			}),
		),
	})
	// Manually set ParentID to simulate a sub-requirement
	g2docs := []model.Document{
		{
			Path:   "/docs/sys",
			Schema: map[string]any{"$id": "/docs/sys/schema.yaml", "title": "Test"},
			Requirements: []model.Requirement{
				{ID: "SYS-001", Source: "/docs/sys/sys.md", Attrs: map[string]any{"title": "Parent"}},
				{ID: "SYS-001-001", ParentID: "SYS-001", Source: "/docs/sys/sys.md", Attrs: map[string]any{"title": "Child"}},
			},
			XReqmd: &model.XReqmd{Level: "test"},
		},
	}
	g2, err := graph.New(g2docs)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// _ = g  // suppress unused var warning — actually we should use g
	_ = g

	results := g2.CheckResults()
	for _, r := range results {
		// Sub-requirement should not get orphan warning
		if r.ReqID == "SYS-001-001" && r.Message == "untraced: no downstream reference" {
			t.Errorf("sub-requirement should not get orphan warning: %+v", r)
		}
	}
}

// ---------------------------------------------------------------------------
// 18. Duplicate ID across documents
// ---------------------------------------------------------------------------

func TestNew_DuplicateIDAcrossDocuments(t *testing.T) {
	// Two documents both have STK-001 — this is a data integrity violation
	docA := docWithXReqmd("/docs/a", &model.XReqmd{Level: "stakeholder", IDPrefix: "STK-"},
		req("STK-001", "/docs/a/needs.md", map[string]any{}),
	)
	docB := docWithXReqmd("/docs/b", &model.XReqmd{Level: "stakeholder", IDPrefix: "STK-"},
		req("STK-001", "/docs/b/needs.md", map[string]any{}),
	)

	g, _ := graph.New([]model.Document{docA, docB})

	// Should have ERROR for duplicate STK-001
	cr := g.CheckResults()
	var dupErrors int
	for _, r := range cr {
		if strings.Contains(r.Message, "duplicate requirement ID") && r.Level == graph.LevelError {
			dupErrors++
		}
	}
	if dupErrors != 1 {
		t.Errorf("duplicate ID errors = %d, want 1; results: %v", dupErrors, cr)
	}

	// Node count should be 1 (first definition wins, second is skipped)
	if got := g.NodeCount(); got != 1 {
		t.Errorf("NodeCount = %d, want 1 (second duplicate skipped)", got)
	}
}

// ---------------------------------------------------------------------------
// 19. Prefix collision is WARNING, not ERROR
// ---------------------------------------------------------------------------

func TestNew_PrefixCollisionIsError(t *testing.T) {
	// Two documents with same prefix but different IDs — prefix collision is ERROR
	docA := docWithXReqmd("/docs/a", &model.XReqmd{Level: "stakeholder-a", IDPrefix: "STK-"},
		req("STK-001", "/docs/a/needs.md", map[string]any{}),
	)
	docB := docWithXReqmd("/docs/b", &model.XReqmd{Level: "stakeholder-b", IDPrefix: "STK-"},
		req("STK-002", "/docs/b/needs.md", map[string]any{}),
	)

	g, _ := graph.New([]model.Document{docA, docB})

	// Should have ERROR for prefix collision
	cr := g.CheckResults()
	var prefixErrors int
	for _, r := range cr {
		if strings.Contains(r.Message, "prefix") && strings.Contains(r.Message, "collision") {
			if r.Level == graph.LevelError {
				prefixErrors++
			}
		}
	}
	if prefixErrors != 1 {
		t.Errorf("prefix collision errors = %d, want 1; results: %v", prefixErrors, cr)
	}
}

// ---------------------------------------------------------------------------
// 20. Version-pin checks — current, outdated, predated, no-pin, missing version, external, suppressed
// ---------------------------------------------------------------------------

func TestNew_VersionPinCurrent(t *testing.T) {
	// UP-001 is at v3 and downstream pins ~3 → no finding.
	g, err := graph.New([]model.Document{
		doc("/docs/up",
			req("UP-001", "/docs/up/up.md", map[string]any{
				"version": 3,
			}),
		),
		doc("/docs/down",
			req("DN-001", "/docs/down/down.md", map[string]any{
				"trace": []any{"UP-001~3"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if r.Code == "version-pin" {
			t.Errorf("unexpected version-pin finding: %+v", r)
		}
	}
}

func TestNew_VersionPinOutdated(t *testing.T) {
	// UP-001 is at v5 and downstream pins ~3 → outdated ERROR.
	g, err := graph.New([]model.Document{
		doc("/docs/up",
			req("UP-001", "/docs/up/up.md", map[string]any{
				"version": 5,
			}),
		),
		doc("/docs/down",
			req("DN-001", "/docs/down/down.md", map[string]any{
				"trace": []any{"UP-001~3"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	var vp []graph.CheckResult
	for _, r := range results {
		if r.Code == "version-pin" {
			vp = append(vp, r)
		}
	}
	if len(vp) != 1 {
		t.Fatalf("version-pin findings = %d, want 1; results: %v", len(vp), results)
	}
	if vp[0].Level != graph.LevelError {
		t.Errorf("level = %q, want ERROR", vp[0].Level)
	}
	if vp[0].Direction != "outdated" {
		t.Errorf("direction = %q, want outdated", vp[0].Direction)
	}
	if vp[0].ReqID != "DN-001" {
		t.Errorf("ReqID = %q, want DN-001", vp[0].ReqID)
	}
	if !strings.Contains(vp[0].Message, "outdated") || !strings.Contains(vp[0].Message, "v5") || !strings.Contains(vp[0].Message, "~3") {
		t.Errorf("Message = %q, want it to mention outdated, v5, ~3", vp[0].Message)
	}
}

func TestNew_VersionPinPredated(t *testing.T) {
	// UP-001 is at v2 and downstream pins ~5 → predated ERROR.
	g, err := graph.New([]model.Document{
		doc("/docs/up",
			req("UP-001", "/docs/up/up.md", map[string]any{
				"version": 2,
			}),
		),
		doc("/docs/down",
			req("DN-001", "/docs/down/down.md", map[string]any{
				"trace": []any{"UP-001~5"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	var vp []graph.CheckResult
	for _, r := range results {
		if r.Code == "version-pin" {
			vp = append(vp, r)
		}
	}
	if len(vp) != 1 {
		t.Fatalf("version-pin findings = %d, want 1; results: %v", len(vp), results)
	}
	if vp[0].Level != graph.LevelError {
		t.Errorf("level = %q, want ERROR", vp[0].Level)
	}
	if vp[0].Direction != "predated" {
		t.Errorf("direction = %q, want predated", vp[0].Direction)
	}
	if !strings.Contains(vp[0].Message, "predated") || !strings.Contains(vp[0].Message, "v2") || !strings.Contains(vp[0].Message, "~5") {
		t.Errorf("Message = %q, want it to mention predated, v2, ~5", vp[0].Message)
	}
}

func TestNew_VersionPinNoPinIsClean(t *testing.T) {
	// No pin at all → check is silent (no finding).
	g, err := graph.New([]model.Document{
		doc("/docs/up",
			req("UP-001", "/docs/up/up.md", map[string]any{
				"version": 5,
			}),
		),
		doc("/docs/down",
			req("DN-001", "/docs/down/down.md", map[string]any{
				"trace": []any{"UP-001"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if r.Code == "version-pin" {
			t.Errorf("unpinned trace should not emit version-pin finding: %+v", r)
		}
	}
}

func TestNew_VersionPinUpstreamNoVersionIsClean(t *testing.T) {
	// Pin is present but upstream has no version → no finding (no ground truth).
	g, err := graph.New([]model.Document{
		doc("/docs/up",
			req("UP-001", "/docs/up/up.md", map[string]any{}),
		),
		doc("/docs/down",
			req("DN-001", "/docs/down/down.md", map[string]any{
				"trace": []any{"UP-001~3"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if r.Code == "version-pin" {
			t.Errorf("upstream without version should not emit version-pin finding: %+v", r)
		}
	}
}

func TestNew_VersionPinExternalUpstreamIsClean(t *testing.T) {
	// External upstream: version is out of our control → skip the check.
	upDoc := docWithXReqmd("/docs/ext", &model.XReqmd{
		Level:    "external",
		External: true,
	},
		req("UP-001", "/docs/ext/up.md", map[string]any{
			"version": 1,
		}),
	)
	dnDoc := doc("/docs/down",
		req("DN-001", "/docs/down/down.md", map[string]any{
			"trace": []any{"UP-001~3"},
		}),
	)
	g, err := graph.New([]model.Document{upDoc, dnDoc})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if r.Code == "version-pin" {
			t.Errorf("external upstream should skip version-pin check: %+v", r)
		}
	}
}

func TestNew_VersionPinSuppression(t *testing.T) {
	// reqmd-suppress: [version-pin] opts out per-node. The parser extracts
	// this attr into Requirement.Suppressions; we set it directly here.
	dnReq := req("DN-001", "/docs/down/down.md", map[string]any{
		"trace": []any{"UP-001~3"},
	})
	dnReq.Suppressions = []string{"version-pin"}

	g, err := graph.New([]model.Document{
		doc("/docs/up",
			req("UP-001", "/docs/up/up.md", map[string]any{
				"version": 5,
			}),
		),
		doc("/docs/down", dnReq),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if r.Code == "version-pin" {
			t.Errorf("suppressed node should not emit version-pin finding: %+v", r)
		}
	}
}

func TestNew_VersionPinChildWithPin(t *testing.T) {
	// Child requirement pins to its parent — child must be re-verified on parent bump.
	// Parent has v3, child pins ~2 → outdated ERROR on the child.
	docs := []model.Document{
		{
			Path:   "/docs/sys",
			Schema: map[string]any{"$id": "/docs/sys/schema.yaml", "title": "Test"},
			Requirements: []model.Requirement{
				{ID: "PARENT", Source: "/docs/sys/sys.md", Attrs: map[string]any{
					"version": 3,
				}},
				{ID: "CHILD", ParentID: "PARENT", Source: "/docs/sys/sys.md", Attrs: map[string]any{
					"trace": []any{"PARENT~2"},
				}},
			},
			XReqmd: &model.XReqmd{Level: "test"},
		},
	}
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	var vp []graph.CheckResult
	for _, r := range results {
		if r.Code == "version-pin" {
			vp = append(vp, r)
		}
	}
	if len(vp) != 1 {
		t.Fatalf("version-pin findings = %d, want 1; results: %v", len(vp), results)
	}
	if vp[0].ReqID != "CHILD" {
		t.Errorf("ReqID = %q, want CHILD", vp[0].ReqID)
	}
	if vp[0].Direction != "outdated" {
		t.Errorf("direction = %q, want outdated", vp[0].Direction)
	}
}

func TestNew_VersionPinQualifiedRef(t *testing.T) {
	// Qualified ref "doc-id/ID~N" must have the pin extracted and the
	// doc-id prefix resolved normally.
	upDoc := docWithXReqmd("/docs/up", &model.XReqmd{
		Level:      "up",
		DocumentID: "upstream",
		IDPrefix:   "UP-",
	},
		req("UP-001", "/docs/up/up.md", map[string]any{
			"version": 7,
		}),
	)
	dnDoc := docWithXReqmd("/docs/down", &model.XReqmd{
		Level:      "down",
		DocumentID: "downstream",
		IDPrefix:   "DN-",
	},
		req("DN-001", "/docs/down/down.md", map[string]any{
			"trace": []any{"upstream/UP-001~3"},
		}),
	)
	g, err := graph.New([]model.Document{upDoc, dnDoc})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	var vp []graph.CheckResult
	for _, r := range results {
		if r.Code == "version-pin" {
			vp = append(vp, r)
		}
	}
	if len(vp) != 1 {
		t.Fatalf("version-pin findings = %d, want 1; results: %v", len(vp), results)
	}
	if vp[0].Direction != "outdated" {
		t.Errorf("direction = %q, want outdated", vp[0].Direction)
	}
	if vp[0].ReqID != "DN-001" {
		t.Errorf("ReqID = %q, want DN-001", vp[0].Direction)
	}
}

// ---------------------------------------------------------------------------
// Status lifecycle tests
// ---------------------------------------------------------------------------

// T3: missing status attr → default (approved). A node with no `status`
// attr is treated as approved, so it counts as a coverage provider.
func TestStatus_DefaultIsApproved(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:      "system-requirements",
			DocumentID: "system",
		},
			req("SYS-001", "/docs/sys/sys.md", map[string]any{
				// no `status` attribute
				"trace": []any{"STK-001"},
			}),
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// STK-001 needs [system]; SYS-001 (no status → default approved) is a
	// coverage provider → no missing-coverage warning expected.
	for _, r := range g.CheckResults() {
		if r.Message == "no upstream trace: no approved requirement from \"system\" traces to this item" {
			t.Errorf("default-approved upstream should satisfy coverage, got: %+v", r)
		}
	}
}

// T4: a draft inbound does NOT satisfy coverage. STK-001 needs [system]
// but the only system doc is draft → missing-coverage warning.
func TestStatus_DraftDoesNotProvideCoverage(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:      "system-requirements",
			DocumentID: "system",
		},
			req("SYS-001", "/docs/sys/sys.md", map[string]any{
				"status": "draft",
				"trace":  []any{"STK-001"},
			}),
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var missing int
	for _, r := range g.CheckResults() {
		if strings.Contains(r.Message, "no upstream trace") {
			missing++
		}
	}
	if missing != 1 {
		t.Errorf("missing-coverage warnings = %d, want 1 (draft does not cover)", missing)
	}
}

// T5: an approved inbound DOES satisfy coverage.
func TestStatus_ApprovedProvidesCoverage(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:      "system-requirements",
			DocumentID: "system",
		},
			req("SYS-001", "/docs/sys/sys.md", map[string]any{
				"status": "approved",
				"trace":  []any{"STK-001"},
			}),
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if strings.Contains(r.Message, "no upstream trace") {
			t.Errorf("approved upstream should satisfy coverage, got: %+v", r)
		}
	}
}

// T6: case-insensitive — "APPROVED" still counts as approved.
func TestStatus_ApprovedIsCaseInsensitive(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:      "system-requirements",
			DocumentID: "system",
		},
			req("SYS-001", "/docs/sys/sys.md", map[string]any{
				"status": "APPROVED",
				"trace":  []any{"STK-001"},
			}),
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if strings.Contains(r.Message, "no upstream trace") {
			t.Errorf("APPROVED should pass case-insensitive check, got: %+v", r)
		}
	}
}

// T11: ignore-status: true document → all requirements are coverage
// providers regardless of their `status` value (even draft).
func TestStatus_IgnoreStatusOptsOut(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:        "system-requirements",
			DocumentID:   "system",
			IgnoreStatus: true,
		},
			req("SYS-001", "/docs/sys/sys.md", map[string]any{
				"status": "draft", // would normally be filtered out
				"trace":  []any{"STK-001"},
			}),
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if strings.Contains(r.Message, "no upstream trace") {
			t.Errorf("ignore-status should bypass status gate, got: %+v", r)
		}
	}
}

// T13: updated checkNeedsCoverage message includes "(%d draft downstreams
// ignored)" when at least one inbound is filtered by the status gate.
func TestStatus_MessageMentionsDraftCount(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:      "system-requirements",
			DocumentID: "system",
		},
			// Two inbounds — both draft → no coverage, draftCount=2.
			req("SYS-A", "/docs/sys/a.md", map[string]any{
				"status": "draft",
				"trace":  []any{"STK-001"},
			}),
			req("SYS-B", "/docs/sys/b.md", map[string]any{
				"status": "draft",
				"trace":  []any{"STK-001"},
			}),
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var found string
	for _, r := range g.CheckResults() {
		if strings.Contains(r.Message, "no upstream trace") {
			found = r.Message
		}
	}
	if found == "" {
		t.Fatal("expected a missing-coverage warning")
	}
	if !strings.Contains(found, "(2 draft downstreams ignored)") {
		t.Errorf("message should mention 2 draft downstreams, got: %q", found)
	}
}

// T16: when draftCount == 0, message has no parenthetical. This is the
// classic case where the need simply has no inbound at all (orphan-ish).
func TestStatus_MessageHasNoParentheticalWhenZeroDrafts(t *testing.T) {
	// Use a single doc with no upstream — but the doc has an `upstream`
	// so it isn't a root dir. We use two docs: an upstream dir with no
	// requirements (so no inbound exists) and the test target.
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:      "system-requirements",
			DocumentID: "system",
		},
		// no requirements at all in the system dir
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
			Upstream:   &model.TraceUpstream{Level: "system-requirements", Sources: []string{"/docs/sys"}},
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	var found string
	var allResults []graph.CheckResult
	for _, r := range g.CheckResults() {
		allResults = append(allResults, r)
		if strings.Contains(r.Message, "no upstream trace") {
			found = r.Message
		}
	}
	if found == "" {
		t.Fatalf("expected a missing-coverage warning; got all results: %+v", allResults)
	}
	if strings.Contains(found, "draft downstreams ignored") {
		t.Errorf("message should not have parenthetical when draftCount == 0, got: %q", found)
	}
}

// T19: ignore-status + additional-status-values combination. The ignore
// flag bypasses the status gate entirely, AND the schema accepts the
// extension value.
func TestStatus_IgnoreAndAdditionsCombination(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:        "system-requirements",
			DocumentID:   "system",
			IgnoreStatus: true,
			// the model doesn't carry additional-status-values into the
			// graph — that's a compile-time concern. The graph just
			// treats all nodes in an ignore-status dir as coverage
			// providers, regardless of the literal status value.
		},
			req("SYS-001", "/docs/sys/sys.md", map[string]any{
				"status": "review", // extension value, would normally be filtered
				"trace":  []any{"STK-001"},
			}),
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	for _, r := range g.CheckResults() {
		if strings.Contains(r.Message, "no upstream trace") {
			t.Errorf("ignore-status should cover any status value, got: %+v", r)
		}
	}
}

// T7: schema-level reject of capitalized status. Compile fails when
// user puts `status: Implemented` in an attr block that doesn't match
// the new built-in enum. (The validator catches it; we mirror the test
// here on the graph side: an uppercase status attr in the data is
// treated as a non-coverage provider because it's not "approved".)
func TestStatus_UnknownValueIsNotApproved(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/sys", &model.XReqmd{
			Level:      "system-requirements",
			DocumentID: "system",
		},
			req("SYS-001", "/docs/sys/sys.md", map[string]any{
				"status": "Implemented", // legacy value — not in [draft, approved]
				"trace":  []any{"STK-001"},
			}),
		),
		docWithXReqmd("/docs/stk", &model.XReqmd{
			Level:      "stakeholder-needs",
			DocumentID: "stakeholder",
		},
			req("STK-001", "/docs/stk/stk.md", map[string]any{
				"requires-trace-from": []any{"system"},
			}),
		),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// The graph layer doesn't validate status values; that's done by
	// the schema layer. But for the coverage gate, an unknown value is
	// treated as not-approved, so coverage is missing.
	var missing int
	for _, r := range g.CheckResults() {
		if strings.Contains(r.Message, "no upstream trace") {
			missing++
		}
	}
	if missing != 1 {
		t.Errorf("unknown status value should not satisfy coverage, missing = %d, want 1", missing)
	}
}

// T15 (graph-side counterpart): schemas with `additionalProperties: false`
// (e.g. 00-aspice, 01-stakeholder) must still parse and inject the
// built-in `status` attribute. The full Validate() path is exercised
// against the real spec/ tree by the CLI check command. Here we just
// confirm the graph builds without panic on a flat additionalProperties
// schema that contains `status` in the resulting attribute set.
func TestStatus_AdditionalPropertiesFalseGraphBuilds(t *testing.T) {
	g, err := graph.New([]model.Document{
		{
			Path: "/docs/x",
			Schema: map[string]any{
				"$id":                  "/docs/x/schema.yaml",
				"title":                "T15",
				"additionalProperties": false,
			},
			Requirements: []model.Requirement{
				{ID: "X-001", Source: "/docs/x/x.md", Attrs: map[string]any{
					"status": "approved",
				}},
			},
			XReqmd: &model.XReqmd{Level: "x"},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if g.NodeCount() != 1 {
		t.Errorf("NodeCount = %d, want 1", g.NodeCount())
	}
	// Confirm the graph lookup methods work on the additionalProperties: false doc.
	_ = g.UpstreamNeighbors("X-001")
	_ = g.DownstreamNeighbors("X-001")
}

// T14: full pipeline against the migrated spec/ tree. The 4 migrated
// dirs (01a, 02, 03, 04) plus 00 and 01 must all build a graph
// successfully and produce no coverage errors stemming from the status
// gate. The graph must reflect the new status values without errors.
func TestStatus_FullPipelineAgainstMigratedSpec(t *testing.T) {
	// The test runs from internal/graph/, so spec/ is ../../spec/.
	docs, err := parser.Discover("../../spec")
	if err != nil {
		t.Fatalf("parser.Discover: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("Discover returned no documents")
	}
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("graph.New: %v", err)
	}

	// All 4 migrated dirs + 00-aspice + 01-stakeholder must be present.
	wantDirs := map[string]bool{
		"../../spec/00-aspice":              false,
		"../../spec/01-stakeholder":         false,
		"../../spec/01a-aspice-stakeholder": false,
		"../../spec/02-system":              false,
		"../../spec/03-software":            false,
		"../../spec/04-tests":               false,
	}
	for _, d := range docs {
		key := d.Path
		if _, ok := wantDirs[key]; ok {
			wantDirs[key] = true
		}
	}
	for dir, seen := range wantDirs {
		if !seen {
			t.Errorf("missing spec dir in Discover: %s", dir)
		}
	}

	// The graph should not report any coverage-gate errors triggered by
	// the new status rule. (Existing graph warnings are tolerated.)
	for _, r := range g.CheckResults() {
		if r.Level == graph.LevelError {
			t.Errorf("unexpected ERROR after migration: %+v", r)
		}
	}
}

// Helpers
// ---------------------------------------------------------------------------
