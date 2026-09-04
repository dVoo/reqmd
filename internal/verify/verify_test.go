package verify_test

import (
	"os"
	"path/filepath"
	"strings"

	"reqmd/internal/model"
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

// An uninstrumented test (no extra.x-reqmd block) is skipped silently:
// the old per-test "no x-reqmd.id" warning is gone so suites can adopt
// reqmd incrementally without spamming every uninstrumented test.
func TestLoadCTRF_UninstrumentedTestSkippedSilently(t *testing.T) {
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
		t.Errorf("results = %d, want 0 (uninstrumented)", len(results))
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %d, want 0 (silent skip)", len(warnings))
	}
}

// A test with an x-reqmd block but nothing bindable (no id, no verifies)
// yields an unbound-result warning.
func TestLoadCTRF_UnboundResultWarns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.ctrf.json")
	writeFixture(t, path, ctrfReport(`{
		"name": "TestOnlyCase",
		"status": "passed",
		"duration": 10,
		"extra": {"x-reqmd": {"case": "BOOT-TIME"}}
	}`))

	results, warnings, err := verify.LoadResults([]string{path})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("results = %d, want 0", len(results))
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %d, want 1", len(warnings))
	}
	if !strings.HasPrefix(warnings[0], "unbound-result:") {
		t.Errorf("warning should carry unbound-result prefix, got %q", warnings[0])
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
	if merged[verify.ResultKey{Target: "TST-001"}].Outcome != verify.OutcomePass {
		t.Errorf("latest outcome = %q, want pass", merged[verify.ResultKey{Target: "TST-001"}].Outcome)
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
	merged := map[verify.ResultKey]verify.Result{
		{Target: "TST-001"}: {MeasureID: "TST-001~3", Outcome: verify.OutcomePass, Source: "run.ctrf.json"},
	}
	doc := verify.Synthesize(merged)
	if !doc.Synthetic {
		t.Error("synthesized document must be marked Synthetic")
	}
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

// ---------------------------------------------------------------------------
// v2 binding (patterns C/D/E)
// ---------------------------------------------------------------------------

func TestLoadCTRF_IDWithCase_RecordsCaseKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.ctrf.json")
	writeFixture(t, path, ctrfReport(`{
		"name": "Boot time",
		"status": "failed",
		"duration": 10,
		"extra": {"x-reqmd": {"id": "TEST-001", "case": "BOOT-TIME"}}
	}`))

	results, _, err := verify.LoadResults([]string{path})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	r := results[0]
	if r.MeasureID != "TEST-001" {
		t.Errorf("MeasureID = %q, want TEST-001", r.MeasureID)
	}
	if r.CaseKey != "BOOT-TIME" {
		t.Errorf("CaseKey = %q, want BOOT-TIME", r.CaseKey)
	}
}

func TestLoadCTRF_CaseAndVerifies_SynthesizesCaseBinding(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.ctrf.json")
	writeFixture(t, path, ctrfReport(`{
		"name": "Boot time",
		"status": "passed",
		"duration": 10,
		"extra": {"x-reqmd": {
			"case": "BOOT-TIME",
			"verifies": ["SYS-001~2"],
			"description": "Measures boot time on reference HW."
		}}
	}`))

	results, _, err := verify.LoadResults([]string{path})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	r := results[0]
	if r.MeasureID != "TC:BOOT-TIME" {
		t.Errorf("MeasureID = %q, want TC:BOOT-TIME", r.MeasureID)
	}
	if r.CaseKey != "BOOT-TIME" {
		t.Errorf("CaseKey = %q, want BOOT-TIME", r.CaseKey)
	}
	if len(r.Verifies) != 1 || r.Verifies[0] != "SYS-001~2" {
		t.Errorf("Verifies = %v, want [SYS-001~2]", r.Verifies)
	}
	if r.Description == "" {
		t.Error("Description should be captured")
	}
}

func TestLoadCTRF_VerifiesOnly_ImplicitCaseFromSuiteName(t *testing.T) {
	dir := t.TempDir()
	// Two reports, two suites, same test name → distinct case keys.
	for _, suite := range []string{"boot", "media"} {
		writeFixture(t, filepath.Join(dir, suite+".ctrf.json"), `{
			"results": {
				"suite": "`+suite+`",
				"tests": [{
					"name": "Startup", "status": "passed", "duration": 1,
					"extra": {"x-reqmd": {"verifies": ["SYS-001"]}}
				}]
			}
		}`)
	}
	results, _, err := verify.LoadResults([]string{dir})
	if err != nil {
		t.Fatalf("LoadResults: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2 (one per suite)", len(results))
	}
	seen := map[string]bool{}
	for _, r := range results {
		if r.CaseKey != "boot/Startup" && r.CaseKey != "media/Startup" {
			t.Errorf("CaseKey = %q, want boot/Startup or media/Startup", r.CaseKey)
		}
		if r.MeasureID != "TC:"+r.CaseKey {
			t.Errorf("MeasureID = %q, want TC:%s", r.MeasureID, r.CaseKey)
		}
		seen[r.CaseKey] = true
	}
	if len(seen) != 2 {
		t.Errorf("distinct case keys = %d, want 2 (same test name in two suites stays distinct)", len(seen))
	}
}

func TestSynthesize_CaseKeyedCreatesTCAndCasedResults(t *testing.T) {
	merged := map[verify.ResultKey]verify.Result{
		{Target: "TEST-001", Case: "BOOT-TIME"}: {
			MeasureID: "TEST-001", CaseKey: "BOOT-TIME", Outcome: verify.OutcomeFail, Source: "run.ctrf.json",
		},
		{Target: "TC:BOOT-TIME", Case: "BOOT-TIME"}: {
			MeasureID:   "TC:BOOT-TIME",
			CaseKey:     "BOOT-TIME",
			Outcome:     verify.OutcomePass,
			Source:      "run.ctrf.json",
			Name:        "Boot time",
			Verifies:    []string{"SYS-001~2"},
			Description: "Measures boot time on reference HW.",
		},
	}
	doc := verify.Synthesize(merged)
	reqs := doc.Requirements()
	if len(reqs) != 3 {
		t.Fatalf("reqs = %d, want 3 (cased result, TC, TC result)", len(reqs))
	}
	byID := make(map[string]*model.Node, len(reqs))
	for _, r := range reqs {
		byID[r.ID] = r
	}
	if byID["RESULT:TEST-001#BOOT-TIME"] == nil {
		t.Error("missing RESULT:TEST-001#BOOT-TIME node")
	}
	tc := byID["TC:BOOT-TIME"]
	if tc == nil {
		t.Fatal("missing TC:BOOT-TIME node")
	}
	if tc.Body != "Measures boot time on reference HW." {
		t.Errorf("TC body = %q", tc.Body)
	}
	trace, _ := tc.Attrs["trace"].([]any)
	if len(trace) != 1 || trace[0] != "SYS-001~2" {
		t.Errorf("TC trace = %v, want [SYS-001~2]", trace)
	}
	resTC := byID["RESULT:TC:BOOT-TIME#BOOT-TIME"]
	if resTC == nil {
		t.Fatal("missing RESULT:TC:BOOT-TIME#BOOT-TIME node")
	}
	resTrace, _ := resTC.Attrs["trace"].([]any)
	if len(resTrace) != 1 || resTrace[0] != "TC:BOOT-TIME" {
		t.Errorf("TC result trace = %v, want [TC:BOOT-TIME]", resTrace)
	}
}

func TestIsTestCaseNode(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"TC:BOOT-TIME", true},
		{"TC-001", false},
		{"RESULT:TST-001", false},
	}
	for _, c := range cases {
		if got := verify.IsTestCaseNode(c.id); got != c.want {
			t.Errorf("IsTestCaseNode(%q) = %v, want %v", c.id, got, c.want)
		}
	}
}
