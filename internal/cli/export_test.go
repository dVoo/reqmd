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
