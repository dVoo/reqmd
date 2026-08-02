package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"reqmd/internal/graph"
)

// findResult reports whether any check result for reqID mentions substr.
func findResult(checks []graph.CheckResult, reqID, substr string) bool {
	for _, c := range checks {
		if c.ReqID == reqID && strings.Contains(c.Message, substr) {
			return true
		}
	}
	return false
}

// TestServeDiscover_Filter verifies the serve pipeline's filter wiring
// (RFC §3.2): the graph is built from the full tree so refs resolve, but
// coverage is computed over the filtered subset, and the returned docs
// are scoped to matching requirements.
func TestServeDiscover_Filter(t *testing.T) {
	root := writeCoverageSpec(t)

	// Unfiltered: SYS-001 is covered by SW-001 (approved, from software).
	_, _, g, err := discoverAndBuildGraph(root, nil, "")
	if err != nil {
		t.Fatalf("discoverAndBuildGraph: %v", err)
	}
	if findResult(g.CheckResults(), "SYS-001", "no upstream trace") {
		t.Fatal("unfiltered: SYS-001 should be covered")
	}

	// Filter "Base": SW-001 (Premium) is excluded and cannot act as a
	// coverage provider → SYS-001 reports missing coverage.
	docs, _, g, err := discoverAndBuildGraph(root, nil, `"Base" in variant`)
	if err != nil {
		t.Fatalf("discoverAndBuildGraph with filter: %v", err)
	}
	if !findResult(g.CheckResults(), "SYS-001", "no upstream trace") {
		t.Fatal("filtered: SYS-001 should report missing coverage")
	}

	// The returned docs are scoped too: the sw doc has no matching reqs.
	for _, d := range docs {
		if filepath.Base(d.Path) == "sw" {
			if len(d.Requirements) != 0 {
				t.Fatalf("filtered sw doc should have 0 reqs, got %d", len(d.Requirements))
			}
		}
	}
}

// TestServeDiscover_FilterUnknownAttrFails verifies compile-time filter
// errors propagate out of the serve pipeline.
func TestServeDiscover_FilterUnknownAttrFails(t *testing.T) {
	root := writeVariantSpec(t)
	_, _, _, err := discoverAndBuildGraph(root, nil, `nonexistent == "x"`)
	if err == nil {
		t.Fatal("expected error for unknown attribute")
	}
}
