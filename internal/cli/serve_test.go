package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// TestServeDiscover_WholeTree verifies the serve pipeline builds the graph
// and exports from the full tree: every requirement is served and coverage
// resolves across documents.
func TestServeDiscover_WholeTree(t *testing.T) {
	root := writeCoverageSpec(t)

	docs, _, g, err := discoverAndBuildGraph(root, nil)
	if err != nil {
		t.Fatalf("discoverAndBuildGraph: %v", err)
	}

	// Both documents and both requirements are served, unfiltered.
	got := map[string]bool{}
	for _, d := range docs {
		for _, r := range d.Requirements() {
			got[r.ID] = true
		}
	}
	if len(got) != 2 || !got["SYS-001"] || !got["SW-001"] {
		t.Fatalf("served requirements = %v, want SYS-001 and SW-001", got)
	}

	// Coverage is computed over the full tree: SW-001 traces to SYS-001.
	if findResult(g.CheckResults(), "SYS-001", "no upstream trace") {
		t.Fatal("SYS-001 should be covered by SW-001")
	}
}

// TestServeProjectHome verifies `/` serves the project home page instead of
// redirecting to the first document.
func TestServeProjectHome(t *testing.T) {
	root := writeCoverageSpec(t)
	docs, analysis, g, err := discoverAndBuildGraph(root, nil)
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
