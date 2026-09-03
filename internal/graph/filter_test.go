package graph_test

import (
	"reqmd/internal/graph"
	"reqmd/internal/model"
	"testing"
)

// ---------------------------------------------------------------------------
// Filter-aware coverage checking (RFC §3.2)
// ---------------------------------------------------------------------------

// TestFilterAwareCoverage_FilteredOutInboundNoFalseFailure verifies that
// a requirement excluded by the filter does not cause a false coverage
// failure for a filtered-in requirement. This is the core behavioral
// guarantee of RFC §3.2: "A requirement excluded by the filter cannot be
// required as, or count as, a coverage provider for another excluded
// requirement."
func TestFilterAwareCoverage_FilteredOutInboundNoFalseFailure(t *testing.T) {
	// Setup: SW-001 (variant: Sport) traces to SYS-001 (variant: Base).
	// Without filter: SW-001 would get a "no downstream reference" type
	// warning if it were untraced. But here we test coverage: SYS-001
	// declares requires-trace-from expecting coverage from the software
	// level. SW-001 traces to SYS-001, but SW-001 is Sport-only.
	//
	// With filter `"Base" in variant`: SYS-001 (no variant = common) is in,
	// SW-001 (Sport) is out. Without filter-awareness, SYS-001 would appear
	// to lack coverage (SW-001 was filtered out). With filter-awareness,
	// SYS-001's inbound from SW-001 is ignored, but SYS-001 is also
	// filtered-in, so the coverage check should NOT report a false failure
	// because... well, actually SYS-001 does lack coverage from the filtered
	// set. Let's test the actual guarantee: the filtered-out inbound doesn't
	// count, and the filtered-in req gets a proper warning (not a broken ref).
	//
	// Better test: SYS-001 has two inbound — SW-001 (Sport, filtered out)
	// and SW-002 (Base, filtered in). SYS-001 should be covered (SW-002
	// counts). Without filter-awareness, this works too. The key test is:
	// SYS-001 has ONLY SW-001 (Sport) as inbound. With filter "Base",
	// SYS-001's coverage check should report the missing coverage, NOT a
	// broken ref (SW-001 exists in the graph, just filtered out).

	docs := []model.Document{
		docWithXReqmd("/spec/sys", &model.XReqmd{
			Level:      "system",
			DocumentID: "sys",
		},
			req("SYS-001", "/spec/sys/sys.md", map[string]any{
				"status":              "approved",
				"variant":             []any{"Base"},
				"requires-trace-from": []any{"software"},
			}),
		),
		docWithXReqmd("/spec/sw", &model.XReqmd{
			Level:      "software",
			DocumentID: "software",
			Upstream:   &model.TraceUpstream{Level: "system", Sources: []string{"/spec/sys"}},
		},
			req("SW-001", "/spec/sw/sw.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Sport"},
				"trace":   []any{"sys/SYS-001"},
			}),
		),
	}

	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Without filter: SYS-001 is covered by SW-001 (status approved, from software level).
	results := g.CheckResults()
	coverageWarnings := filterByMessage(results, "no upstream trace")
	if len(coverageWarnings) != 0 {
		t.Fatalf("without filter: expected no coverage warnings, got %d: %v", len(coverageWarnings), coverageWarnings)
	}

	// With filter "Base" in variant: SYS-001 is in (Base), SW-001 is out (Sport).
	// SYS-001 should now report missing coverage (SW-001 filtered out).
	g.SetFilter(map[string]struct{}{"SYS-001": {}})
	results = g.CheckResults()
	coverageWarnings = filterByMessage(results, "no upstream trace")
	if len(coverageWarnings) == 0 {
		t.Fatal("with filter: expected coverage warning for SYS-001 (SW-001 filtered out)")
	}

	// Crucially, no broken-ref warning — SW-001 exists in the graph, just filtered out.
	brokenRefs := filterByMessage(results, "broken reference")
	if len(brokenRefs) != 0 {
		t.Fatalf("with filter: expected no broken-ref warnings, got %d", len(brokenRefs))
	}
}

// TestFilterAwareCoverage_FilteredInProviderSatisfies verifies that a
// filtered-in inbound requirement properly satisfies coverage.
func TestFilterAwareCoverage_FilteredInProviderSatisfies(t *testing.T) {
	docs := []model.Document{
		docWithXReqmd("/spec/sys", &model.XReqmd{
			Level:      "system",
			DocumentID: "sys",
		},
			req("SYS-001", "/spec/sys/sys.md", map[string]any{
				"status":              "approved",
				"variant":             []any{"Base"},
				"requires-trace-from": []any{"software"},
			}),
		),
		docWithXReqmd("/spec/sw", &model.XReqmd{
			Level:      "software",
			DocumentID: "software",
			Upstream:   &model.TraceUpstream{Level: "system", Sources: []string{"/spec/sys"}},
		},
			req("SW-001", "/spec/sw/sw.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Base"},
				"trace":   []any{"sys/SYS-001"},
			}),
		),
	}

	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Filter: both SYS-001 and SW-001 have "Base" in variant → both in set.
	g.SetFilter(map[string]struct{}{"SYS-001": {}, "SW-001": {}})
	results := g.CheckResults()

	coverageWarnings := filterByMessage(results, "no upstream trace")
	if len(coverageWarnings) != 0 {
		t.Fatalf("expected no coverage warnings (SW-001 is a valid provider), got %d", len(coverageWarnings))
	}
}

// TestFilterAwareCoverage_DocumentDefault verifies filter-aware coverage
// interacts with the x-reqmd.requires-trace-from document default: an
// inherited expectation behaves like a per-requirement one under --filter.
func TestFilterAwareCoverage_DocumentDefault(t *testing.T) {
	docs := []model.Document{
		docWithXReqmd("/spec/sys", &model.XReqmd{
			Level:             "system",
			DocumentID:        "sys",
			RequiresTraceFrom: []string{"software"},
		},
			req("SYS-001", "/spec/sys/sys.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Base"},
			}),
		),
		docWithXReqmd("/spec/sw", &model.XReqmd{
			Level:      "software",
			DocumentID: "software",
			Upstream:   &model.TraceUpstream{Level: "system", Sources: []string{"/spec/sys"}},
		},
			req("SW-001", "/spec/sw/sw.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Sport"},
				"trace":   []any{"sys/SYS-001"},
			}),
		),
	}

	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Without filter: the sole inbound SW-001 satisfies the inherited default.
	if w := filterByMessage(g.CheckResults(), "no upstream trace"); len(w) != 0 {
		t.Fatalf("without filter: expected no coverage warnings, got %v", w)
	}

	// Filter "Base": SYS-001 in, SW-001 (Sport) out → inherited default now
	// reports missing coverage, with no broken-ref warning.
	g.SetFilter(map[string]struct{}{"SYS-001": {}})
	results := g.CheckResults()
	if w := filterByMessage(results, "no upstream trace"); len(w) != 1 {
		t.Fatalf("with filter: expected 1 coverage warning, got %d: %v", len(w), w)
	}
	if w := filterByMessage(results, "broken reference"); len(w) != 0 {
		t.Fatalf("with filter: expected no broken-ref warnings, got %d", len(w))
	}
}

// TestFilter_NoFilter_ParityWithUnfiltered verifies that an empty filter
// set produces identical results to no filter at all.
func TestFilter_NoFilter_ParityWithUnfiltered(t *testing.T) {
	docs := []model.Document{
		doc("/spec/a",
			req("REQ-001", "/spec/a/a.md", map[string]any{
				"status": "approved",
				"trace":  []any{"REQ-002"},
			}),
			req("REQ-002", "/spec/a/b.md", map[string]any{
				"status": "approved",
			}),
		),
	}

	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	unfiltered := g.CheckResults()

	// SetFilter(nil) should behave identically to no filter.
	g.SetFilter(nil)
	filtered := g.CheckResults()

	if len(unfiltered) != len(filtered) {
		t.Fatalf("parity broken: unfiltered has %d results, filtered(nil) has %d", len(unfiltered), len(filtered))
	}
	for i := range unfiltered {
		if unfiltered[i].Message != filtered[i].Message {
			t.Fatalf("parity broken at result %d: %q vs %q", i, unfiltered[i].Message, filtered[i].Message)
		}
	}

	// An empty non-nil set means "the filter matched nothing" — no
	// requirement is checked. This is distinct from nil.
	g.SetFilter(map[string]struct{}{})
	none := g.CheckResults()
	if len(none) != 0 {
		t.Fatalf("empty filter set should suppress all checks, got %d results", len(none))
	}
}

// ---------------------------------------------------------------------------
// Disjoint-attribute check (RFC §3.5)
// ---------------------------------------------------------------------------

// TestDisjointAttribute_ZeroIntersection_Errors verifies that a trace link
// between two requirements with non-overlapping variant arrays produces an
// ERROR.
func TestDisjointAttribute_ZeroIntersection_Errors(t *testing.T) {
	docs := []model.Document{
		doc("/spec/sys",
			req("SYS-014", "/spec/sys/sys.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Premium", "Sport"},
			}),
		),
		doc("/spec/sw",
			req("SW-042", "/spec/sw/sw.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Base"},
				"trace":   []any{"SYS-014"},
			}),
		),
	}

	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	g.SetDisjointAttrs([]string{"variant"})

	results := g.CheckResults()
	disjointErrors := filterByCode(results)
	if len(disjointErrors) == 0 {
		t.Fatal("expected disjoint-attribute ERROR for SW-042→SYS-014 with zero overlap")
	}
	if disjointErrors[0].ReqID != "SW-042" {
		t.Fatalf("expected SW-042, got %s", disjointErrors[0].ReqID)
	}
	if disjointErrors[0].Level != graph.LevelError {
		t.Fatalf("expected ERROR level, got %s", disjointErrors[0].Level)
	}
}

// TestDisjointAttribute_PartialOverlap_NoError verifies that overlapping
// values do not trigger an error.
func TestDisjointAttribute_PartialOverlap_NoError(t *testing.T) {
	docs := []model.Document{
		doc("/spec/sys",
			req("SYS-014", "/spec/sys/sys.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Base", "Premium"},
			}),
		),
		doc("/spec/sw",
			req("SW-042", "/spec/sw/sw.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Base", "Sport"},
				"trace":   []any{"SYS-014"},
			}),
		),
	}

	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	g.SetDisjointAttrs([]string{"variant"})

	results := g.CheckResults()
	disjointErrors := filterByCode(results)
	if len(disjointErrors) != 0 {
		t.Fatalf("expected no disjoint error (partial overlap on Base), got %d", len(disjointErrors))
	}
}

// TestDisjointAttribute_EmptyAttribute_Exempt verifies that a requirement
// with an empty/absent variant attribute is exempt (treated as "applies to all").
func TestDisjointAttribute_EmptyAttribute_Exempt(t *testing.T) {
	docs := []model.Document{
		doc("/spec/sys",
			// SYS-014 has no variant attribute — exempt.
			req("SYS-014", "/spec/sys/sys.md", map[string]any{
				"status": "approved",
			}),
		),
		doc("/spec/sw",
			req("SW-042", "/spec/sw/sw.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Base"},
				"trace":   []any{"SYS-014"},
			}),
		),
	}

	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	g.SetDisjointAttrs([]string{"variant"})

	results := g.CheckResults()
	disjointErrors := filterByCode(results)
	if len(disjointErrors) != 0 {
		t.Fatalf("expected no disjoint error (SYS-014 has no variant = exempt), got %d", len(disjointErrors))
	}
}

// TestDisjointAttribute_NoAttrsSet_NoCheck verifies that without
// SetDisjointAttrs, the check is a no-op.
func TestDisjointAttribute_NoAttrsSet_NoCheck(t *testing.T) {
	docs := []model.Document{
		doc("/spec/sys",
			req("SYS-014", "/spec/sys/sys.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Premium"},
			}),
		),
		doc("/spec/sw",
			req("SW-042", "/spec/sw/sw.md", map[string]any{
				"status":  "approved",
				"variant": []any{"Base"},
				"trace":   []any{"SYS-014"},
			}),
		),
	}

	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// No SetDisjointAttrs — check should be off.
	results := g.CheckResults()
	disjointErrors := filterByCode(results)
	if len(disjointErrors) != 0 {
		t.Fatalf("expected no disjoint check without SetDisjointAttrs, got %d", len(disjointErrors))
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func filterByMessage(results []graph.CheckResult, substr string) []graph.CheckResult {
	var out []graph.CheckResult
	for _, r := range results {
		if contains(r.Message, substr) {
			out = append(out, r)
		}
	}
	return out
}

func filterByCode(results []graph.CheckResult) []graph.CheckResult {
	var out []graph.CheckResult
	for _, r := range results {
		if r.Code == graph.CodeDisjointAttribute {
			out = append(out, r)
		}
	}
	return out
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
