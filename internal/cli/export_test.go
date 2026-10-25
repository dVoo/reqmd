package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportCSV_Filter(t *testing.T) {
	root := writeVariantSpec(t)
	outDir := t.TempDir()

	if _, err := runCmd(t, NewRootCmd(""), "export", "csv", root, "-o", outDir, "--filter", `"Base" in variant`); err != nil {
		t.Fatalf("export csv --filter failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "sw-requirements.csv"))
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}
	csv := string(data)
	if !strings.Contains(csv, "SW-001") {
		t.Fatalf("csv should contain SW-001:\n%s", csv)
	}
	if strings.Contains(csv, "SW-002") {
		t.Fatalf("csv should not contain SW-002:\n%s", csv)
	}
}

func TestExportHTML_Filter(t *testing.T) {
	root := writeVariantSpec(t)
	outDir := t.TempDir()

	if _, err := runCmd(t, NewRootCmd(""), "export", "html", root, "-o", outDir, "--filter", `"Premium" in variant`); err != nil {
		t.Fatalf("export html --filter failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "sw-requirements.html"))
	if err != nil {
		t.Fatalf("reading html: %v", err)
	}
	html := string(data)
	if !strings.Contains(html, "SW-002") {
		t.Fatalf("html should contain SW-002:\n%s", html)
	}
	if strings.Contains(html, "SW-001") {
		t.Fatalf("html should not contain SW-001:\n%s", html)
	}
}

// TestExportHTML_ProjectIndex verifies that an export writes the project home
// alongside the document pages.
func TestExportHTML_ProjectIndex(t *testing.T) {
	root := writeCoverageSpec(t)
	outDir := t.TempDir()

	if _, err := runCmd(t, NewRootCmd(""), "export", "html", root, "-o", outDir); err != nil {
		t.Fatalf("export html failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("reading project index: %v", err)
	}
	index := string(data)
	for _, want := range []string{`data-page="project"`, "sys-requirements.html", "sw-requirements.html", "Needs attention"} {
		if !strings.Contains(index, want) {
			t.Errorf("project index missing %q", want)
		}
	}
}

// TestExportHTML_NestedTraceLinks verifies that, without -o, cross-document
// trace links are relative to each document's own directory.
func TestExportHTML_NestedTraceLinks(t *testing.T) {
	root := writeCoverageSpec(t)

	if _, err := runCmd(t, NewRootCmd(""), "export", "html", root); err != nil {
		t.Fatalf("export html failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "sys", "sys-requirements.html"))
	if err != nil {
		t.Fatalf("reading sys page: %v", err)
	}
	if !strings.Contains(string(data), `href="../sw/sw-requirements.html#SW-001"`) {
		t.Error("cross-document trace link is not relative to the page directory")
	}
	index, err := os.ReadFile(filepath.Join(root, "index.html"))
	if err != nil {
		t.Fatalf("reading project index: %v", err)
	}
	if !strings.Contains(string(index), `href="sys/sys-requirements.html"`) {
		t.Error("project catalog link is not relative to the project root")
	}
}
