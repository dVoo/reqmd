package verify_test

import (
	"os"
	"path/filepath"

	"reqmd/internal/verify"
	"testing"
	"time"
)

// writeFixture writes fixture content to a file, failing the test on error.
func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// ---------------------------------------------------------------------------
// CTRF parser
// ---------------------------------------------------------------------------

// ctrfReport builds a minimal CTRF JSON report with the given tests.
func ctrfReport(tests string) string {
	return `{
		"results": {
			"tool": {"name": "gotest"},
			"summary": {"tests": 1, "passed": 1, "failed": 0, "skipped": 0, "pending": 0, "other": 0, "start": 1700000000000, "stop": 1700000001000},
			"tests": [` + tests + `]
		}
	}`
}

func TestLoadCTRF_PassFail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.ctrf.json")
	writeFixture(t, path, ctrfReport(`{
		"name": "TestA",
		"status": "passed",
		"duration": 10,
		"stop": 1700000001000,
		"extra": {"x-reqmd": {"id": "TST-001"}}
	},{
		"name": "TestB",
		"status": "failed",
		"duration": 10,
		"stop": 1700000002000,
		"extra": {"x-reqmd": {"id": "TST-002"}}
	}`))

	results, warnings, err := verify.LoadResults([]string{path})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	want := map[string]verify.Outcome{
		"TST-001": verify.OutcomePass,
		"TST-002": verify.OutcomeFail,
	}
	for _, r := range results {
		if r.Outcome != want[r.MeasureID] {
			t.Errorf("%s outcome = %q, want %q", r.MeasureID, r.Outcome, want[r.MeasureID])
		}
	}
}

func TestLoadCTRF_UnmappedTestWarns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.ctrf.json")
	writeFixture(t, path, ctrfReport(`{
		"name": "TestNoReq",
		"status": "passed",
		"duration": 10,
		"extra": {}
	}`))

	results, warnings, err := verify.LoadResults([]string{path})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("results = %d, want 0 (unmapped)", len(results))
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %d, want 1", len(warnings))
	}
}

func TestLoadCTRF_VersionPinPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.ctrf.json")
	writeFixture(t, path, ctrfReport(`{
		"name": "TestPinned",
		"status": "passed",
		"duration": 10,
		"extra": {"x-reqmd": {"id": "TST-001~3"}}
	}`))

	results, _, err := verify.LoadResults([]string{path})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].MeasureID != "TST-001~3" {
		t.Errorf("MeasureID = %q, want TST-001~3 (pin preserved)", results[0].MeasureID)
	}
}

func TestLoadCTRF_FlakyIsInconclusive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.ctrf.json")
	writeFixture(t, path, ctrfReport(`{
		"name": "TestFlaky",
		"status": "passed",
		"duration": 10,
		"flaky": true,
		"extra": {"x-reqmd": {"id": "TST-001"}}
	}`))

	results, _, err := verify.LoadResults([]string{path})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if results[0].Outcome != verify.OutcomeInconclusive {
		t.Errorf("flaky outcome = %q, want inconclusive", results[0].Outcome)
	}
}

// ---------------------------------------------------------------------------
// Multi-file merge
// ---------------------------------------------------------------------------

func TestMergeLatest_LatestByVerifiedAt(t *testing.T) {
	results := []verify.Result{
		{MeasureID: "TST-001", Outcome: verify.OutcomeFail, VerifiedAt: time.UnixMilli(1000).UTC(), Source: "old.ctrf.json"},
		{MeasureID: "TST-001", Outcome: verify.OutcomePass, VerifiedAt: time.UnixMilli(2000).UTC(), Source: "new.ctrf.json"},
	}
	merged := verify.MergeLatest(results)
	if len(merged) != 1 {
		t.Fatalf("merged = %d, want 1", len(merged))
	}
	if merged["TST-001"].Outcome != verify.OutcomePass {
		t.Errorf("latest outcome = %q, want pass", merged["TST-001"].Outcome)
	}
}

func TestMergeLatest_PinStrippedFromKey(t *testing.T) {
	results := []verify.Result{
		{MeasureID: "TST-001~3", Outcome: verify.OutcomePass, VerifiedAt: time.UnixMilli(1000).UTC()},
		{MeasureID: "TST-001~4", Outcome: verify.OutcomeFail, VerifiedAt: time.UnixMilli(2000).UTC()},
	}
	merged := verify.MergeLatest(results)
	if len(merged) != 1 {
		t.Fatalf("merged = %d, want 1 (pin collapsed)", len(merged))
	}
}

// ---------------------------------------------------------------------------
// Auto-detection (non-CTRF json skipped)
// ---------------------------------------------------------------------------

func TestLoadResults_NonCTRFJsonSkipped(t *testing.T) {
	dir := t.TempDir()
	// coverage.json — not CTRF, must be skipped silently
	writeFixture(t, filepath.Join(dir, "coverage.json"), `{"coverage": 0.95}`)
	// a real CTRF report
	writeFixture(t, filepath.Join(dir, "run.ctrf.json"), ctrfReport(`{
		"name": "TestA", "status": "passed", "duration": 1,
		"extra": {"x-reqmd": {"id": "TST-001"}}
	}`))

	results, _, err := verify.LoadResults([]string{dir})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("results = %d, want 1 (coverage.json skipped)", len(results))
	}
}

func TestLoadResults_PlainJsonCTRFShapeAccepted(t *testing.T) {
	dir := t.TempDir()
	// plain .json with CTRF shape → accepted
	writeFixture(t, filepath.Join(dir, "report.json"), ctrfReport(`{
		"name": "TestA", "status": "passed", "duration": 1,
		"extra": {"x-reqmd": {"id": "TST-001"}}
	}`))

	results, _, err := verify.LoadResults([]string{dir})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("results = %d, want 1 (CTRF-shaped .json accepted)", len(results))
	}
}

// ---------------------------------------------------------------------------
// Synthesize
// ---------------------------------------------------------------------------

func TestSynthesize_ResultNodeIDAndTrace(t *testing.T) {
	merged := map[string]verify.Result{
		"TST-001": {MeasureID: "TST-001~3", Outcome: verify.OutcomePass, Source: "run.ctrf.json"},
	}
	doc := verify.Synthesize(merged)
	reqs := doc.Requirements()
	if len(reqs) != 1 {
		t.Fatalf("reqs = %d, want 1", len(reqs))
	}
	r := reqs[0]
	if r.ID != "RESULT:TST-001" {
		t.Errorf("ID = %q, want RESULT:TST-001", r.ID)
	}
	if got := verify.ResultOutcome(r.Attrs); got != verify.OutcomePass {
		t.Errorf("outcome = %q, want pass", got)
	}
	// trace carries the pinned measure ID so version-pin fires on stale
	trace, _ := r.Attrs["trace"].([]any)
	if len(trace) != 1 || trace[0] != "TST-001~3" {
		t.Errorf("trace = %v, want [TST-001~3]", trace)
	}
}

func TestIsResultNode(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"RESULT:TST-001", true},
		{"TST-001", false},
		{"", false},
	}
	for _, c := range cases {
		if got := verify.IsResultNode(c.id); got != c.want {
			t.Errorf("IsResultNode(%q) = %v, want %v", c.id, got, c.want)
		}
	}
}
