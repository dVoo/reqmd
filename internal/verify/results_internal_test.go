package verify

import (
	"os"
	"path/filepath"
	"reqmd/internal/model"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Manual-results loading (loadManualDir)
// ---------------------------------------------------------------------------

// TestLoadManualDir_ResultFields creates a temp results dir with a schema.yaml
// and a .md file holding a manual-results requirement (outcome, trace,
// verifier, evidence, verified-at), then asserts loadManualDir returns one
// Result with every field populated correctly.
func TestLoadManualDir_ResultFields(t *testing.T) {
	dir := t.TempDir()

	schema := `x-reqmd:
  level: verify-results
$id: "results/schema.yaml"
title: "Manual Results"
type: object
required:
  - x-reqmd.outcome
  - x-reqmd.verifier
  - x-reqmd.verified-at
properties:
  x-reqmd.outcome:
    type: string
    enum: [pass, fail, skipped, inconclusive]
  x-reqmd.verifier:
    type: string
  x-reqmd.evidence:
    type: string
  x-reqmd.verified-at:
    type: string
additionalProperties: false
`
	if err := os.WriteFile(filepath.Join(dir, "schema.yaml"), []byte(schema), 0o644); err != nil {
		t.Fatalf("write schema.yaml: %v", err)
	}

	md := `# Manual Verification Results

## RES-001: Review result
` + "```attr" + `
x-reqmd.outcome: pass
x-reqmd.verifier: Jane
x-reqmd.evidence: report.md
x-reqmd.verified-at: "2024-01-15"
trace:
  - MEAS-001~3
` + "```" + `
The review confirmed the requirement is satisfied.
`
	mdPath := filepath.Join(dir, "results.md")
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		t.Fatalf("write results.md: %v", err)
	}

	results, err := loadManualDir(dir)
	if err != nil {
		t.Fatalf("loadManualDir: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1: %+v", len(results), results)
	}
	r := results[0]
	if r.MeasureID != "MEAS-001~3" {
		t.Errorf("MeasureID = %q, want MEAS-001~3", r.MeasureID)
	}
	if r.Outcome != OutcomePass {
		t.Errorf("Outcome = %q, want pass", r.Outcome)
	}
	if r.Verifier != "Jane" {
		t.Errorf("Verifier = %q, want Jane", r.Verifier)
	}
	if r.Evidence != "report.md" {
		t.Errorf("Evidence = %q, want report.md", r.Evidence)
	}
	wantAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !r.VerifiedAt.Equal(wantAt) {
		t.Errorf("VerifiedAt = %v, want %v", r.VerifiedAt, wantAt)
	}
	// Source is the markdown file path the parser recorded on the requirement.
	if r.Source != mdPath {
		t.Errorf("Source = %q, want %q", r.Source, mdPath)
	}
}

// ---------------------------------------------------------------------------
// resultFromReq — direct construction tests
// ---------------------------------------------------------------------------

func TestResultFromReq_HappyPath(t *testing.T) {
	req := &model.Node{
		Kind:   model.KindRequirement,
		ID:     "RES-001",
		Source: "/results/r.md",
		Attrs: map[string]any{
			model.AttrResultOutcome:    "pass",
			model.AttrResultVerifier:   "Jane",
			model.AttrResultEvidence:   "report.md",
			model.AttrResultVerifiedAt: "2024-01-15",
			model.AttrTrace: []any{
				"MEAS-001~3",
			},
		},
	}
	r, ok := resultFromReq(req)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if r.MeasureID != "MEAS-001~3" {
		t.Errorf("MeasureID = %q, want MEAS-001~3", r.MeasureID)
	}
	if r.Outcome != OutcomePass {
		t.Errorf("Outcome = %q, want pass", r.Outcome)
	}
	if r.Verifier != "Jane" {
		t.Errorf("Verifier = %q, want Jane", r.Verifier)
	}
	if r.Evidence != "report.md" {
		t.Errorf("Evidence = %q, want report.md", r.Evidence)
	}
	wantAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !r.VerifiedAt.Equal(wantAt) {
		t.Errorf("VerifiedAt = %v, want %v", r.VerifiedAt, wantAt)
	}
	if r.Source != "/results/r.md" {
		t.Errorf("Source = %q, want /results/r.md", r.Source)
	}
}

func TestResultFromReq_NoOutcomeNotAResult(t *testing.T) {
	req := &model.Node{
		Kind:   model.KindRequirement,
		ID:     "REQ-001",
		Source: "/docs/r.md",
		Attrs: map[string]any{
			model.AttrResultVerifier: "Jane",
			model.AttrTrace:          []any{"MEAS-001~3"},
		},
	}
	if _, ok := resultFromReq(req); ok {
		t.Error("ok = true, want false (no outcome attr → not a result)")
	}
}

func TestResultFromReq_EmptyTraceFallsBackToReqID(t *testing.T) {
	req := &model.Node{
		Kind:   model.KindRequirement,
		ID:     "RES-001",
		Source: "/results/r.md",
		Attrs: map[string]any{
			model.AttrResultOutcome:    "pass",
			model.AttrResultVerifier:   "Jane",
			model.AttrResultVerifiedAt: "2024-01-15",
			// no trace attr → MeasureID must fall back to the requirement ID
		},
	}
	r, ok := resultFromReq(req)
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if r.MeasureID != "RES-001" {
		t.Errorf("MeasureID = %q, want RES-001 (fallback to req.ID)", r.MeasureID)
	}
}
