package graph_test

import (
	"reqmd/internal/graph"
	"reqmd/internal/model"
	"strings"
	"testing"
)

// resultReq builds a synthesized verification-result pseudo-requirement
// tracing to measureID with the given outcome.
func resultReq(measureID, outcome string) *model.Node {
	return &model.Node{
		ID:     "RESULT:" + measureID,
		Source: measureID + ".ctrf.json",
		Attrs: map[string]any{
			"outcome":        outcome,
			model.AttrTrace:  []any{measureID},
			model.AttrStatus: model.StatusApproved,
		},
	}
}

// measureReq builds a verification-measure requirement (an authored req
// that a result traces to). By default it has status approved.
func measureReq(status string) *model.Node {
	attrs := map[string]any{"verify": "Test"}
	if status != "" {
		attrs[model.AttrStatus] = status
	}
	return &model.Node{
		ID:     "TST-001",
		Source: "/docs/TST-001.md",
		Attrs:  attrs,
	}
}

func resultDoc(reqs ...*model.Node) model.Document {
	return model.Document{
		Path:   "<verify-results>",
		Nodes:  reqs,
		XReqmd: &model.XReqmd{Level: "verify-results"},
	}
}

func measureDoc(reqs ...*model.Node) model.Document {
	return docWithXReqmd("/docs/test-specs", &model.XReqmd{Level: "test-specs"}, reqs...)
}

// ---------------------------------------------------------------------------
// No result nodes → outcome checks are no-ops
// ---------------------------------------------------------------------------

func TestCheckResults_NoResults_NoOutcomeChecks(t *testing.T) {
	g, err := graph.New([]model.Document{
		measureDoc(measureReq(model.StatusApproved)),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	for _, r := range results {
		if r.Code == graph.CodeMissingVerdict || r.Code == graph.CodeFailingVerdict {
			t.Errorf("outcome check fired without results: %+v", r)
		}
	}
}

// ---------------------------------------------------------------------------
// missing-verdict: approved measure with no result → WARNING
// ---------------------------------------------------------------------------

func TestMissingVerdict_MeasureWithoutResult(t *testing.T) {
	g, err := graph.New([]model.Document{
		measureDoc(measureReq(model.StatusApproved)),
		resultDoc(resultReq("TST-002", "pass")), // result for a *different* measure
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	var found bool
	for _, r := range results {
		if r.Code == graph.CodeMissingVerdict && r.ReqID == "TST-001" {
			found = true
			if r.Level != graph.LevelWarning {
				t.Errorf("level = %q, want WARNING", r.Level)
			}
		}
	}
	if !found {
		t.Errorf("missing-verdict not fired for TST-001; got %v", results)
	}
}

func TestMissingVerdict_MeasureWithResult_NoFinding(t *testing.T) {
	g, err := graph.New([]model.Document{
		measureDoc(measureReq(model.StatusApproved)),
		resultDoc(resultReq("TST-001", "pass")),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	for _, r := range results {
		if r.Code == graph.CodeMissingVerdict && r.ReqID == "TST-001" {
			t.Errorf("missing-verdict fired for covered measure: %+v", r)
		}
	}
}

func TestMissingVerdict_DraftMeasureSkipped(t *testing.T) {
	g, err := graph.New([]model.Document{
		measureDoc(measureReq(model.StatusDraft)),
		resultDoc(resultReq("TST-002", "pass")), // result for a different measure
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	for _, r := range results {
		if r.Code == graph.CodeMissingVerdict && r.ReqID == "TST-001" {
			t.Errorf("missing-verdict fired for draft measure: %+v", r)
		}
	}
}

// ---------------------------------------------------------------------------
// failing-verdict: measure with fail result → ERROR
// ---------------------------------------------------------------------------

func TestFailingVerdict_MeasureFails(t *testing.T) {
	g, err := graph.New([]model.Document{
		measureDoc(measureReq(model.StatusApproved)),
		resultDoc(resultReq("TST-001", "fail")),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	var found bool
	for _, r := range results {
		if r.Code == graph.CodeFailingVerdict && r.ReqID == "TST-001" {
			found = true
			if r.Level != graph.LevelError {
				t.Errorf("level = %q, want ERROR", r.Level)
			}
		}
	}
	if !found {
		t.Errorf("failing-verdict not fired; got %v", results)
	}
}

func TestFailingVerdict_MeasurePasses_NoFinding(t *testing.T) {
	g, err := graph.New([]model.Document{
		measureDoc(measureReq(model.StatusApproved)),
		resultDoc(resultReq("TST-001", "pass")),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	for _, r := range results {
		if r.Code == graph.CodeFailingVerdict {
			t.Errorf("failing-verdict fired for passing measure: %+v", r)
		}
	}
}

// ---------------------------------------------------------------------------
// version-pin reuse: stale result pin triggers existing check
// ---------------------------------------------------------------------------

func TestVersionPin_StaleResultPin(t *testing.T) {
	// measure at version 4, result pins ~3 → outdated
	measure := measureReq(model.StatusApproved)
	measure.Attrs[model.AttrVersion] = 4
	g, err := graph.New([]model.Document{
		measureDoc(measure),
		resultDoc(staleResultReq("TST-001~3", "pass")),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := g.CheckResults()
	var found bool
	for _, r := range results {
		if r.Code == graph.CodeVersionPin && r.ReqID == "RESULT:TST-001" {
			found = true
			if r.Direction != graph.DirOutdated {
				t.Errorf("direction = %q, want outdated", r.Direction)
			}
		}
	}
	if !found {
		t.Errorf("version-pin not fired for stale result pin; got %v", results)
	}
}

// staleResultReq builds a result whose trace carries a ~N pin.
func staleResultReq(pinnedMeasureID, outcome string) *model.Node {
	strip := strings.SplitN(pinnedMeasureID, "~", 2)[0]
	return &model.Node{
		ID:     "RESULT:" + strip,
		Source: "run.ctrf.json",
		Attrs: map[string]any{
			"outcome":        outcome,
			model.AttrTrace:  []any{pinnedMeasureID},
			model.AttrStatus: model.StatusApproved,
		},
	}
}
