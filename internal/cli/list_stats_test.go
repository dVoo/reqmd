package cli

import (
	"strings"
	"testing"
)

func TestListCmd_Filter(t *testing.T) {
	root := writeVariantSpec(t)

	out, err := runCmd(t, newListCmd(), root)
	if err != nil {
		t.Fatalf("ls failed: %v", err)
	}
	if !strings.Contains(out, "SW-001") || !strings.Contains(out, "SW-002") {
		t.Fatalf("unfiltered ls should list both requirements:\n%s", out)
	}

	out, err = runCmd(t, newListCmd(), "--filter", `"Base" in variant`, root)
	if err != nil {
		t.Fatalf("ls --filter failed: %v", err)
	}
	if !strings.Contains(out, "SW-001") {
		t.Fatalf("filtered ls should list SW-001:\n%s", out)
	}
	if strings.Contains(out, "SW-002") {
		t.Fatalf("filtered ls should not list SW-002:\n%s", out)
	}
}

func TestListCmd_FilterJSON(t *testing.T) {
	root := writeVariantSpec(t)

	out, err := runCmd(t, newListCmd(), "--filter", `"Premium" in variant`, "--json", root)
	if err != nil {
		t.Fatalf("ls --filter --json failed: %v", err)
	}
	if strings.Contains(out, "SW-001") {
		t.Fatalf("JSON ls should exclude SW-001:\n%s", out)
	}
	if !strings.Contains(out, "SW-002") {
		t.Fatalf("JSON ls should include SW-002:\n%s", out)
	}
}

func TestStatsCmd_Filter(t *testing.T) {
	root := writeVariantSpec(t)

	out, err := runCmd(t, newStatsCmd(), "--filter", `"Premium" in variant`, root)
	if err != nil {
		t.Fatalf("stats --filter failed: %v", err)
	}
	// Only SW-002 (Premium) is counted. The variant breakdown renders only
	// string-typed attributes, so assert on the requirement counts instead.
	if !strings.Contains(out, "Requirements: 1") {
		t.Fatalf("filtered stats should count 1 requirement:\n%s", out)
	}
	if strings.Contains(out, "Requirements: 2") {
		t.Fatalf("filtered stats should not count the Base requirement:\n%s", out)
	}
	if strings.Contains(out, "Base") {
		t.Fatalf("filtered stats should not include Base variant:\n%s", out)
	}
}

func TestListCmd_ShowsItems(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sw/schema.yaml", variantSchema)
	writeFile(t, root, "sw/sw.md", "# Software\n\nIntro prose for the doc.\n\n## Sub System\n\n### SW-001: Base parser\n```attr\nstatus: approved\nvariant: [Base]\n```\n")

	out, err := runCmd(t, newListCmd(), root)
	if err != nil {
		t.Fatalf("ls failed: %v", err)
	}
	if !strings.Contains(out, "container") {
		t.Fatalf("ls should show the ## Sub System container:\n%s", out)
	}
	if !strings.Contains(out, "Sub System") {
		t.Fatalf("ls should show the container title:\n%s", out)
	}
	if !strings.Contains(out, "SW-001") {
		t.Fatalf("ls should still list requirements:\n%s", out)
	}
	// Filtering scopes requirements but keeps the container visible.
	out, err = runCmd(t, newListCmd(), "--filter", `"Premium" in variant`, root)
	if err != nil {
		t.Fatalf("ls --filter failed: %v", err)
	}
	if !strings.Contains(out, "container") || !strings.Contains(out, "Sub System") {
		t.Fatalf("filtered ls should keep the container:\n%s", out)
	}
	if strings.Contains(out, "SW-001") {
		t.Fatalf("filtered ls should not list SW-001:\n%s", out)
	}
}

func TestListCmd_ItemsJSON(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sw/schema.yaml", variantSchema)
	writeFile(t, root, "sw/sw.md", "# Software\n\nIntro prose for the doc.\n\n## Sub System\n\n### SW-001: Base parser\n```attr\nstatus: approved\nvariant: [Base]\n```\n")

	out, err := runCmd(t, newListCmd(), "--json", root)
	if err != nil {
		t.Fatalf("ls --json failed: %v", err)
	}
	if !strings.Contains(out, `"type": "container"`) {
		t.Fatalf("ls --json should include the container row:\n%s", out)
	}
	if !strings.Contains(out, `"type": "req"`) {
		t.Fatalf("ls --json should include requirement rows:\n%s", out)
	}
}

func TestStatsCmd_ShowsItems(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sw/schema.yaml", variantSchema)
	writeFile(t, root, "sw/sw.md", "# Software\n\nIntro prose for the doc.\n\n## Sub System\n\n### SW-001: Base parser\n```attr\nstatus: approved\nvariant: [Base]\n```\n")

	out, err := runCmd(t, newStatsCmd(), root)
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if !strings.Contains(out, "1 reqs, 2 items") {
		t.Fatalf("stats should count the container and intro info items:\n%s", out)
	}
	if !strings.Contains(out, "type:") || !strings.Contains(out, "container") {
		t.Fatalf("stats should include a type breakdown:\n%s", out)
	}
}
