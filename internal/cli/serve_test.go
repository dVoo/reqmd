package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/internal/exporter"
	"github.com/dVoo/reqmd/internal/graph"
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
	_, _, g, err := discoverAndBuildGraph(root, nil, "", false)
	if err != nil {
		t.Fatalf("discoverAndBuildGraph: %v", err)
	}
	if findResult(g.CheckResults(), "SYS-001", "no upstream trace") {
		t.Fatal("unfiltered: SYS-001 should be covered")
	}

	// Filter "Base": SW-001 (Premium) is excluded and cannot act as a
	// coverage provider → SYS-001 reports missing coverage.
	docs, _, g, err := discoverAndBuildGraph(root, nil, `"Base" in variant`, false)
	if err != nil {
		t.Fatalf("discoverAndBuildGraph with filter: %v", err)
	}
	if !findResult(g.CheckResults(), "SYS-001", "no upstream trace") {
		t.Fatal("filtered: SYS-001 should report missing coverage")
	}

	// The returned docs are scoped too: the sw doc has no matching reqs.
	for _, d := range docs {
		if filepath.Base(d.Path) == "sw" {
			if len(d.Requirements()) != 0 {
				t.Fatalf("filtered sw doc should have 0 reqs, got %d", len(d.Requirements()))
			}
		}
	}
}

// TestServeDiscover_FilterUnknownAttrFails verifies compile-time filter
// errors propagate out of the serve pipeline (strict mode).
func TestServeDiscover_FilterUnknownAttrFails(t *testing.T) {
	root := writeVariantSpec(t)
	_, _, _, err := discoverAndBuildGraph(root, nil, `nonexistent == "x"`, false)
	if err == nil {
		t.Fatal("expected error for unknown attribute")
	}
}

// TestServeDiscover_RelaxedFilterAllowsUndeclared verifies --relaxed-filter
// lets the serve pipeline filter on an attribute that is present on
// requirements but not declared in any schema.yaml.
func TestServeDiscover_RelaxedFilterAllowsUndeclared(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "spec/schema.yaml", "$schema: \"https://json-schema.org/draft/2020-12/schema\"\n"+
		"title: Relaxed\ntype: object\nrequired: [status]\n"+
		"properties:\n  status:\n    type: string\n")
	// `owner` is used on a requirement but never declared in the schema.
	writeFile(t, root, "spec/a.md", "# A\n\n## REQ-1\n\n```attr\nstatus: approved\nowner: alice\n```\n\nOne.\n"+
		"\n## REQ-2\n\n```attr\nstatus: approved\nowner: bob\n```\n\nTwo.\n")

	// Strict mode rejects the undeclared attribute.
	if _, _, _, err := discoverAndBuildGraph(root, nil, `owner == "alice"`, false); err == nil {
		t.Fatal("strict mode should reject an undeclared attribute")
	}

	// Relaxed mode accepts it and scopes the docs to the matching req.
	docs, _, _, err := discoverAndBuildGraph(root, nil, `owner == "alice"`, true)
	if err != nil {
		t.Fatalf("relaxed mode should allow an undeclared attribute: %v", err)
	}
	var ids []string
	for _, d := range docs {
		for _, r := range d.Requirements() {
			ids = append(ids, r.ID)
		}
	}
	if len(ids) != 1 || ids[0] != "REQ-1" {
		t.Fatalf("relaxed filter should keep only REQ-1, got %v", ids)
	}
}

// TestServeProjectHome verifies `/` serves the project home page instead of
// redirecting to the first document.
func TestServeProjectHome(t *testing.T) {
	root := writeCoverageSpec(t)
	docs, analysis, g, err := discoverAndBuildGraph(root, nil, "", false)
	if err != nil {
		t.Fatalf("discoverAndBuildGraph: %v", err)
	}
	rctx, err := exporter.NewRenderContext(docs, root)
	if err != nil {
		t.Fatalf("NewRenderContext: %v", err)
	}
	pages, _, _, err := exportServePages(docs, analysis, g, rctx, exporter.ComputeDocBoundaries(docs), exporter.BuildTitleMap(docs))
	if err != nil {
		t.Fatalf("exportServePages: %v", err)
	}
	if _, ok := pages["index.html"]; !ok {
		t.Fatal("serve pages missing index.html")
	}

	srv := &serveData{pages: pages}
	rec := httptest.NewRecorder()
	srv.fileHandler(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/ status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-page="project"`) {
		t.Error("/ did not serve the project home")
	}
	if !strings.Contains(body, "sys-requirements.html") || !strings.Contains(body, "sw-requirements.html") {
		t.Error("project home missing document catalog links")
	}
}
