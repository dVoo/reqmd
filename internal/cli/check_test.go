package cli

import (
	"encoding/json"
	"errors"
	"reqmd/internal/reporter"
	"strings"
	"testing"
)

// jsonCheckSummary captures the parts of `check --json` output that the
// filter/disjoint tests assert on.
type jsonCheckSummary struct {
	Summary struct {
		Filter   string `json:"filter"`
		Total    int    `json:"total"`
		Valid    int    `json:"valid"`
		Warnings int    `json:"warnings"`
	} `json:"summary"`
	ExitCode int `json:"exit_code"`
}

func TestCheckCmd_FilterScopesValidation(t *testing.T) {
	root := writeVariantSpec(t)

	// Unfiltered: both requirements are validated.
	out, err := runCmd(t, newCheckCmd(), "--json", root)
	if err != nil {
		t.Fatalf("unfiltered check failed: %v\n%s", err, out)
	}
	var rep jsonCheckSummary
	if err = json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	if rep.Summary.Total != 2 {
		t.Fatalf("unfiltered total = %d, want 2", rep.Summary.Total)
	}
	if rep.Summary.Valid != 2 {
		t.Fatalf("unfiltered valid = %d, want 2", rep.Summary.Valid)
	}
	if rep.Summary.Filter != "" {
		t.Fatalf("unfiltered filter = %q, want empty", rep.Summary.Filter)
	}

	// Filtered to Base: only SW-001 is validated, and the JSON summary
	// carries the filter expression (RFC §3.1).
	filter := `"Base" in variant`
	out, err = runCmd(t, newCheckCmd(), "--filter", filter, "--json", root)
	if err != nil {
		t.Fatalf("filtered check failed: %v\n%s", err, out)
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	if rep.Summary.Total != 1 {
		t.Fatalf("filtered total = %d, want 1", rep.Summary.Total)
	}
	if rep.Summary.Filter != filter {
		t.Fatalf("filter field = %q, want %q", rep.Summary.Filter, filter)
	}
}

func TestCheckCmd_FilterUnknownAttributeFails(t *testing.T) {
	root := writeVariantSpec(t)
	_, err := runCmd(t, newCheckCmd(), "--filter", `nonexistent == "x"`, root)
	if err == nil {
		t.Fatal("expected compile-time error for unknown attribute")
	}
	if !strings.Contains(err.Error(), "unknown attribute") {
		t.Fatalf("error = %q, want mention of unknown attribute", err)
	}
}

func TestCheckCmd_FilterSyntaxErrorFails(t *testing.T) {
	root := writeVariantSpec(t)
	_, err := runCmd(t, newCheckCmd(), "--filter", `status ==`, root)
	if err == nil {
		t.Fatal("expected compile-time syntax error")
	}
}

// TestCheckCmd_FilterBuiltinTraceVar verifies the RFC §2.3 built-in
// attribute `trace` is accepted by the typo check even though it is not
// declared in the fixture schema.
func TestCheckCmd_FilterBuiltinTraceVar(t *testing.T) {
	root := writeDisjointSpec(t)
	out, err := runCmd(t, newCheckCmd(), "--filter", `"sys/SYS-001" in trace`, "--json", root)
	if err != nil {
		t.Fatalf("filter on built-in trace should compile: %v\n%s", err, out)
	}
	var rep jsonCheckSummary
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	if rep.Summary.Total != 1 {
		t.Fatalf("trace filter total = %d, want 1 (only SW-001 traces to SYS-001)", rep.Summary.Total)
	}
}

func TestCheckCmd_DisjointCheckFlag(t *testing.T) {
	root := writeDisjointSpec(t)

	_, err := runCmd(t, newCheckCmd(), "--disjoint-check", "variant", root)
	if err == nil {
		t.Fatal("expected disjoint-attribute ERROR")
	}
	var exitErr *reporter.ExitCodeError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("error = %v, want ExitCodeError{Code: 1}", err)
	}
}

// TestCheckCmd_DisjointCheckClean verifies the disjoint check is off by
// default and passes when variants overlap.
func TestCheckCmd_DisjointCheckClean(t *testing.T) {
	root := writeDisjointSpec(t)
	if out, err := runCmd(t, newCheckCmd(), root); err != nil {
		t.Fatalf("check without --disjoint-check should pass: %v\n%s", err, out)
	}

	root = writeDisjointCleanSpec(t)
	if out, err := runCmd(t, newCheckCmd(), root); err != nil {
		t.Fatalf("disjoint check with overlapping variants should pass: %v\n%s", err, out)
	}
}

// TestCheckCmd_DisjointCheckSchemaDecl verifies the check is enabled via
// x-reqmd.disjoint-check in schema.yaml (RFC §3.5).
func TestCheckCmd_DisjointCheckSchemaDecl(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sys/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sys\n  disjoint-check: variant\n")
	writeFile(t, root, "sys/sys.md", "# System\n\n"+mdReq("SYS-001", "Platform", "status: approved\nvariant: [Premium]\n"))
	writeFile(t, root, "sw/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sw\n  disjoint-check: variant\n")
	writeFile(t, root, "sw/sw.md", "# Software\n\n"+mdReq("SW-001", "Base client", "status: approved\nvariant: [Base]\ntrace: [sys/SYS-001]\n"))

	_, err := runCmd(t, newCheckCmd(), root)
	if err == nil {
		t.Fatal("expected disjoint-attribute ERROR from schema declaration")
	}
	var exitErr *reporter.ExitCodeError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("error = %v, want ExitCodeError{Code: 1}", err)
	}
}

// TestCheckCmd_FilterWithResults verifies the --filter + --results
// interplay: a measure in the filter set reports its failing verdict even
// though the synthesized RESULT node is not itself in the filter set; a
// measure excluded by the filter is not reported.
func TestCheckCmd_FilterWithResults(t *testing.T) {
	root := writeMeasureSpec(t)
	results := writeManualResult(t, "SW-001", "fail")

	// Measure in the filter set → failing verdict reported (exit 1).
	_, err := runCmd(t, newCheckCmd(), "--results", results, "--filter", `"Base" in variant`, root)
	if err == nil {
		t.Fatal("expected failing-verdict ERROR for SW-001")
	}
	var exitErr *reporter.ExitCodeError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("error = %v, want ExitCodeError{Code: 1}", err)
	}

	// Measure excluded by the filter → no verdict checks, clean run.
	out, err := runCmd(t, newCheckCmd(), "--results", results, "--filter", `"Premium" in variant`, "--json", root)
	if err != nil {
		t.Fatalf("filtered-out measure should not fail: %v\n%s", err, out)
	}
	var rep jsonCheckSummary
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	if rep.Summary.Total != 0 {
		t.Fatalf("Premium-filtered total = %d, want 0", rep.Summary.Total)
	}
}

// TestCheckCmd_RequiresTraceFromDocumentDefault verifies the
// x-reqmd.requires-trace-from document default end-to-end through schema.yaml:
// a requirement without its own attribute inherits the default.
func TestCheckCmd_RequiresTraceFromDocumentDefault(t *testing.T) {
	writeSpec := func(traceToSystem bool) string {
		root := t.TempDir()
		writeFile(t, root, "sys/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sys\n  level: system\n  requires-trace-from: [software]\n")
		// SYS-001 declares no requires-trace-from; it inherits [software]
		// from the document default.
		writeFile(t, root, "sys/sys.md", "# System\n\n"+mdReq("SYS-001", "Platform", "status: approved\n"))
		writeFile(t, root, "sw/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sw\n  level: software\n  upstream:\n    level: system\n    sources:\n      - ../sys/\n")
		trace := ""
		if traceToSystem {
			trace = "trace: [sys/SYS-001]\n"
		}
		writeFile(t, root, "sw/sw.md", "# Software\n\n"+mdReq("SW-001", "Client", "status: approved\n"+trace))
		return root
	}

	// Covered: the inherited [software] default is satisfied.
	out, err := runCmd(t, newCheckCmd(), writeSpec(true))
	if err != nil {
		t.Fatalf("check with satisfied document default should pass: %v\n%s", err, out)
	}
	if strings.Contains(out, "no upstream trace") {
		t.Errorf("unexpected coverage warning with satisfied default:\n%s", out)
	}

	// Uncovered: no software requirement traces to SYS-001 → WARNING (exit 0).
	out, err = runCmd(t, newCheckCmd(), writeSpec(false))
	if err != nil {
		t.Fatalf("coverage warnings should not fail the run: %v\n%s", err, out)
	}
	if !strings.Contains(out, `no upstream trace: no approved requirement from "software" traces to this item`) {
		t.Errorf("expected inherited-default coverage warning, got:\n%s", out)
	}
}

// writeUnboundResultsDir creates a results dir containing one CTRF report
// whose test declares x-reqmd but binds to nothing (case only). Returns the
// results dir path.
func writeUnboundResultsDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "run.ctrf.json", `{
		"results": {
			"tool": {"name": "gotest"},
			"summary": {"tests": 1, "passed": 1},
			"tests": [{
				"name": "Boot", "status": "passed", "duration": 1,
				"extra": {"x-reqmd": {"case": "BOOT-TIME"}}
			}]
		}
	}`)
	return dir
}

// TestCheckCmd_IgnoreUnboundResults verifies unbound-result warnings surface
// by default and are suppressed by --ignore-unbound-results.
func TestCheckCmd_IgnoreUnboundResults(t *testing.T) {
	root := writeMeasureSpec(t)
	results := writeUnboundResultsDir(t)

	countWarnings := func(args ...string) int {
		t.Helper()
		out, err := runCmd(t, newCheckCmd(), append(args, "--json", root)...)
		if err != nil {
			t.Fatalf("check should pass: %v\n%s", err, out)
		}
		var rep jsonCheckSummary
		if err := json.Unmarshal([]byte(out), &rep); err != nil {
			t.Fatalf("unmarshal: %v\n%s", err, out)
		}
		return rep.Summary.Warnings
	}

	if w := countWarnings("--results", results); w != 1 {
		t.Errorf("warnings without flag = %d, want 1 (unbound-result)", w)
	}
	if w := countWarnings("--results", results, "--ignore-unbound-results"); w != 0 {
		t.Errorf("warnings with --ignore-unbound-results = %d, want 0", w)
	}
}
