package graph_test

import (
	"reqmd/internal/graph"
	"reqmd/internal/model"
	"testing"
)

// reqWithSuppressions builds a Requirement with a non-empty Suppressions
// slice — needed for tests that exercise reqmd-suppress handling, because
// the synthetic req() factory leaves Suppressions as nil.
func reqWithSuppressions(id string, source string, suppressions []string, attrs map[string]any) *model.Node {
	r := req(id, source, attrs)
	r.Suppressions = append(r.Suppressions, suppressions...)
	return r
}

// ---------------------------------------------------------------------------
// RepinDeltas: outdated pin
// ---------------------------------------------------------------------------

func TestRepinDeltas_Outdated(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("UP-001", "/docs/up/up.md", map[string]any{
			"title":   "Up",
			"version": 3,
		})),
		doc("/docs/down", req("DN-001", "/docs/down/dn.md", map[string]any{
			"title": "Down",
			"trace": []any{"UP-001~1"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	deltas := g.RepinDeltas(false)
	if len(deltas) != 1 {
		t.Fatalf("RepinDeltas = %d deltas, want 1: %+v", len(deltas), deltas)
	}
	d := deltas[0]
	if d.Kind != "outdated" {
		t.Errorf("Kind = %q, want outdated", d.Kind)
	}
	if d.ReqID != "DN-001" {
		t.Errorf("ReqID = %q, want DN-001", d.ReqID)
	}
	if d.TargetID != "UP-001" {
		t.Errorf("TargetID = %q, want UP-001", d.TargetID)
	}
	if d.SourceRef != "UP-001~1" {
		t.Errorf("SourceRef = %q, want UP-001~1", d.SourceRef)
	}
	if d.OldPin != 1 {
		t.Errorf("OldPin = %d, want 1", d.OldPin)
	}
	if d.NewPin != 3 {
		t.Errorf("NewPin = %d, want 3", d.NewPin)
	}
	if d.NewVersion != 3 {
		t.Errorf("NewVersion = %d, want 3", d.NewVersion)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: current pin → no delta
// ---------------------------------------------------------------------------

func TestRepinDeltas_Current(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("UP-001", "/docs/up/up.md", map[string]any{
			"version": 3,
		})),
		doc("/docs/down", req("DN-001", "/docs/down/dn.md", map[string]any{
			"trace": []any{"UP-001~3"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if deltas := g.RepinDeltas(false); len(deltas) != 0 {
		t.Errorf("RepinDeltas = %+v, want empty (pin == upstream.version)", deltas)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: predated pin → reported but not auto-fixable
// ---------------------------------------------------------------------------

func TestRepinDeltas_Predated(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("UP-001", "/docs/up/up.md", map[string]any{
			"version": 2,
		})),
		doc("/docs/down", req("DN-001", "/docs/down/dn.md", map[string]any{
			"trace": []any{"UP-001~5"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	deltas := g.RepinDeltas(false)
	if len(deltas) != 1 {
		t.Fatalf("RepinDeltas = %+v, want 1 predated delta", deltas)
	}
	if deltas[0].Kind != "predated" {
		t.Errorf("Kind = %q, want predated", deltas[0].Kind)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: unpinned ref (no promote) → empty
// ---------------------------------------------------------------------------

func TestRepinDeltas_UnpinnedNoPromote(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("UP-001", "/docs/up/up.md", map[string]any{
			"version": 2,
		})),
		doc("/docs/down", req("DN-001", "/docs/down/dn.md", map[string]any{
			"trace": []any{"UP-001"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if deltas := g.RepinDeltas(false); len(deltas) != 0 {
		t.Errorf("RepinDeltas(false) = %+v, want empty (no pin → skip without promote)", deltas)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: unpinned ref (promote) → unpinned delta
// ---------------------------------------------------------------------------

func TestRepinDeltas_UnpinnedPromote(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("UP-001", "/docs/up/up.md", map[string]any{
			"version": 4,
		})),
		doc("/docs/down", req("DN-001", "/docs/down/dn.md", map[string]any{
			"trace": []any{"UP-001"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	deltas := g.RepinDeltas(true)
	if len(deltas) != 1 {
		t.Fatalf("RepinDeltas(true) = %+v, want 1 unpinned delta", deltas)
	}
	d := deltas[0]
	if d.Kind != "unpinned" {
		t.Errorf("Kind = %q, want unpinned", d.Kind)
	}
	if d.SourceRef != "UP-001" {
		t.Errorf("SourceRef = %q, want UP-001 (no pin suffix)", d.SourceRef)
	}
	if d.OldPin != 0 {
		t.Errorf("OldPin = %d, want 0", d.OldPin)
	}
	if d.NewPin != 4 {
		t.Errorf("NewPin = %d, want 4", d.NewPin)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: external upstream → skipped
// ---------------------------------------------------------------------------

func TestRepinDeltas_ExternalSkipped(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/up", &model.XReqmd{Level: "test", External: true},
			req("UP-001", "/docs/up/up.md", map[string]any{"version": 3}),
		),
		doc("/docs/down", req("DN-001", "/docs/down/dn.md", map[string]any{
			"trace": []any{"UP-001~1"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if deltas := g.RepinDeltas(false); len(deltas) != 0 {
		t.Errorf("RepinDeltas = %+v, want empty (external target skipped)", deltas)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: unversioned upstream → skipped
// ---------------------------------------------------------------------------

func TestRepinDeltas_UnversionedSkipped(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("UP-001", "/docs/up/up.md", map[string]any{
			// no version
		})),
		doc("/docs/down", req("DN-001", "/docs/down/dn.md", map[string]any{
			"trace": []any{"UP-001~1"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if deltas := g.RepinDeltas(false); len(deltas) != 0 {
		t.Errorf("RepinDeltas = %+v, want empty (unversioned target skipped)", deltas)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: suppressed node → skipped
// ---------------------------------------------------------------------------

func TestRepinDeltas_Suppressed(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("UP-001", "/docs/up/up.md", map[string]any{
			"version": 3,
		})),
		doc("/docs/down", reqWithSuppressions("DN-001", "/docs/down/dn.md", []string{"version-pin"}, map[string]any{
			"trace": []any{"UP-001~1"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if deltas := g.RepinDeltas(false); len(deltas) != 0 {
		t.Errorf("RepinDeltas = %+v, want empty (suppressed node skipped)", deltas)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: qualified ref (doc-id/ID~N) → SourceRef preserves the form
// ---------------------------------------------------------------------------

func TestRepinDeltas_QualifiedSourceRef(t *testing.T) {
	g, err := graph.New([]model.Document{
		docWithXReqmd("/docs/up", &model.XReqmd{Level: "test", DocumentID: "upstream"},
			req("UP-001", "/docs/up/up.md", map[string]any{"version": 2}),
		),
		doc("/docs/down", req("DN-001", "/docs/down/dn.md", map[string]any{
			"trace": []any{"upstream/UP-001~1"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	deltas := g.RepinDeltas(false)
	if len(deltas) != 1 {
		t.Fatalf("RepinDeltas = %+v, want 1 delta", deltas)
	}
	if deltas[0].SourceRef != "upstream/UP-001~1" {
		t.Errorf("SourceRef = %q, want upstream/UP-001~1 (qualified form preserved)", deltas[0].SourceRef)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: result pseudo-requirement is not a source
// ---------------------------------------------------------------------------

func TestRepinDeltas_ResultSkipped(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("MEAS-001", "/docs/up/meas.md", map[string]any{
			"version": 2,
			"verify":  "Test",
		})),
		doc("/docs/results", req("RESULT:MEAS-001", "/docs/results/r.md", map[string]any{
			"trace":   []any{"MEAS-001~1"},
			"outcome": "pass",
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if deltas := g.RepinDeltas(false); len(deltas) != 0 {
		t.Errorf("RepinDeltas = %+v, want empty (RESULT: nodes are skipped)", deltas)
	}
}

// ---------------------------------------------------------------------------
// RepinDeltas: stable sort by file → reqID → sourceRef
// ---------------------------------------------------------------------------

func TestRepinDeltas_Sorted(t *testing.T) {
	g, err := graph.New([]model.Document{
		doc("/docs/up", req("UP-001", "/docs/up/up.md", map[string]any{
			"version": 5,
		})),
		// Two distinct IDs across two files, plus a same-ID duplicate
		// to verify graph.New first-wins behavior (the duplicate is dropped,
		// not turned into a delta).
		doc("/docs/z", req("Z-1", "/docs/z/z.md", map[string]any{
			"trace": []any{"UP-001~1"},
		})),
		doc("/docs/a", req("A-1", "/docs/a/a.md", map[string]any{
			"trace": []any{"UP-001~3"},
		})),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	deltas := g.RepinDeltas(false)
	if len(deltas) != 2 {
		t.Fatalf("RepinDeltas = %+v, want 2 deltas", deltas)
	}
	// Expected order: /docs/a/a.md, /docs/z/z.md
	if deltas[0].File != "/docs/a/a.md" {
		t.Errorf("deltas[0].File = %q, want /docs/a/a.md", deltas[0].File)
	}
	if deltas[1].File != "/docs/z/z.md" {
		t.Errorf("deltas[1].File = %q, want /docs/z/z.md", deltas[1].File)
	}
}
