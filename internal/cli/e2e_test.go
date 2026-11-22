package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/internal/model"
	"github.com/dVoo/reqmd/internal/parser"
	"github.com/dVoo/reqmd/reqmd-import/exchange"
)

// bundleFixture is a real exchange bundle: sections grouping requirements, an
// internal link, an external ref, a declared schema and x-reqmd metadata.
const bundleFixture = "../../reqmd-import/exchange/testdata/valid/doors-sections.json"

// renderBundle runs a bundle through the renderer and writes the files it
// produces into dir, exactly as `reqmd-import` does.
func renderBundle(t *testing.T, dir string) {
	t.Helper()
	raw, err := os.ReadFile(bundleFixture)
	if err != nil {
		t.Fatalf("read bundle fixture: %v", err)
	}
	b, err := exchange.Decode(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("decode bundle: %v", err)
	}
	files, err := b.Build()
	if err != nil {
		t.Fatalf("build bundle: %v", err)
	}
	for _, f := range files {
		writeFile(t, dir, f.Path, string(f.Data))
	}
}

// TestE2E_ExchangeBundleToCheckAndExports is the whole chain in one test: an
// exchange bundle becomes reqmd documents, those documents validate, and the
// exports carry the structure the bundle declared. It asserts behaviour across
// module boundaries (exchange → parser → graph → exporter) instead of inside a
// single package, which is where the pipeline actually breaks.
func TestE2E_ExchangeBundleToCheckAndExports(t *testing.T) {
	root := t.TempDir()
	renderBundle(t, root)

	// 1. The rendered tree validates.
	out, err := runCmd(t, NewRootCmd(""), "check", root)
	if err != nil {
		t.Fatalf("check failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "0 invalid") {
		t.Fatalf("check reported invalid requirements:\n%s", out)
	}

	// 2. A section is structure, an attr-block heading is content: the same
	//    split must hold in the parsed tree.
	docs, err := parser.Discover(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("documents = %d, want 1", len(docs))
	}
	nodes := docs[0].Nodes
	if len(nodes) != 1 || nodes[0].Kind != model.KindSection || nodes[0].Title != "Hydraulics" {
		t.Fatalf("root nodes = %+v, want one Hydraulics section", nodes)
	}
	if got := len(nodes[0].Children); got != 2 {
		t.Fatalf("section children = %d, want 2 requirements", got)
	}
	for _, child := range nodes[0].Children {
		if child.ID == "" || child.Kind.String() != "req" {
			t.Errorf("child = %+v, want an identified requirement", child)
		}
	}

	// 3. NDJSON carries requirements only, with the section path as context.
	outDir := filepath.Join(t.TempDir(), "ndjson")
	if _, err := runCmd(t, NewRootCmd(""), "export", "json", root, "-o", outDir); err != nil {
		t.Fatalf("export json: %v", err)
	}
	rows := ndjsonRows(t, filepath.Join(outDir, "requirements.ndjson"))
	if len(rows) != 2 {
		t.Fatalf("requirement rows = %d, want 2 (sections are not facts):\n%v", len(rows), rows)
	}
	byID := map[string]map[string]any{}
	for _, r := range rows {
		byID[r["id"].(string)] = r
	}
	for _, id := range []string{"DOORS-17", "DOORS-42"} {
		row := byID[id]
		if row == nil {
			t.Fatalf("row %s missing; got %v", id, keysOf(byID))
		}
		if row["section_path"] != "Hydraulics" {
			t.Errorf("%s section_path = %v, want Hydraulics", id, row["section_path"])
		}
		if _, has := row["type"]; has {
			t.Errorf("%s still carries a type column", id)
		}
	}
	if byID["DOORS-17"]["ordinal"] != float64(1) || byID["DOORS-42"]["ordinal"] != float64(2) {
		t.Errorf("ordinals = %v/%v, want 1/2 (document order)",
			byID["DOORS-17"]["ordinal"], byID["DOORS-42"]["ordinal"])
	}

	// The bundle's internal link and external ref both became trace refs.
	trace, ok := byID["DOORS-17"]["trace"].([]any)
	if !ok || len(trace) != 2 {
		t.Fatalf("DOORS-17 trace = %#v, want two refs", byID["DOORS-17"]["trace"])
	}
	if trace[0] != "SYS-BRAKE-001" || trace[1] != "DOORS-42" {
		t.Errorf("DOORS-17 trace = %v, want [SYS-BRAKE-001 DOORS-42]", trace)
	}

	// 4. CSV keeps the section row: it is a review artifact, not a fact table.
	csvDir := filepath.Join(t.TempDir(), "csv")
	if _, err := runCmd(t, NewRootCmd(""), "export", "csv", root, "-o", csvDir); err != nil {
		t.Fatalf("export csv: %v", err)
	}
	csvData, err := os.ReadFile(filepath.Join(csvDir, "brakes-requirements.csv"))
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(csvData), "section,,Hydraulics") {
		t.Errorf("CSV should keep the section row:\n%s", csvData)
	}
	if !strings.Contains(string(csvData), "req,DOORS-17") {
		t.Errorf("CSV should list DOORS-17 as a requirement:\n%s", csvData)
	}
}

// TestE2E_SectionsSurviveParseAndFilter covers the section model end to end in
// reqmd core: a real document is parsed into a tree, and a filtered view drops
// a folder whose requirements were all pruned while keeping leaf sections.
func TestE2E_SectionsSurviveParseAndFilter(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sys/schema.yaml", `$schema: "https://json-schema.org/draft/2020-12/schema"
type: object
properties:
  status:
    type: string
  variant:
    type: array
    items:
      type: string
`)
	writeFile(t, root, "sys/sys.md", `# System

Preamble prose.

## Brakes

Chapter intro.

### SYS-001: Braking

`+"```attr\nstatus: approved\nvariant: [Base]\n```\n"+`
### SYS-002: Pressure

`+"```attr\nstatus: approved\nvariant: [Premium]\n```\n"+`
## Legacy

Superseded chapter.

### SYS-003: Old behaviour

`+"```attr\nstatus: approved\nvariant: [Base]\n```\n"+`
## Notes

A leaf information item with no requirements.
`)

	docs, err := parser.Discover(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	nodes := docs[0].Nodes
	// Preamble, Brakes (two children), Legacy (one), Notes (leaf).
	if len(nodes) != 4 {
		t.Fatalf("root nodes = %d, want 4 (preamble, Brakes, Legacy, Notes): %+v", len(nodes), nodes)
	}
	if nodes[0].Level != 0 || nodes[0].Title != "" || nodes[0].Body == "" {
		t.Errorf("nodes[0] = %+v, want the headingless preamble", nodes[0])
	}
	if nodes[1].Title != "Brakes" || len(nodes[1].Children) != 2 {
		t.Errorf("nodes[1] = %+v, want Brakes with 2 requirements", nodes[1])
	}
	if nodes[2].Title != "Legacy" || len(nodes[2].Children) != 1 {
		t.Errorf("nodes[2] = %+v, want Legacy with one requirement", nodes[2])
	}
	if nodes[3].Title != "Notes" || len(nodes[3].Children) != 0 {
		t.Errorf("nodes[3] = %+v, want a leaf section", nodes[3])
	}

	// Unfiltered CSV lists everything in document order.
	csvDir := filepath.Join(t.TempDir(), "all")
	if _, err := runCmd(t, NewRootCmd(""), "export", "csv", root, "-o", csvDir); err != nil {
		t.Fatalf("export csv: %v", err)
	}
	all, err := os.ReadFile(filepath.Join(csvDir, "sys-requirements.csv"))
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	for _, want := range []string{"section,,Brakes", "req,SYS-001", "req,SYS-002", "section,,Legacy", "req,SYS-003", "section,,Notes"} {
		if !strings.Contains(string(all), want) {
			t.Errorf("unfiltered CSV missing %q:\n%s", want, all)
		}
	}

	// Filtering to Premium prunes SYS-001 and SYS-003. Brakes keeps SYS-002 and
	// therefore its place; Legacy loses its only requirement and is dropped;
	// the leaf section with prose survives regardless.
	filteredDir := filepath.Join(t.TempDir(), "filtered")
	if _, err := runCmd(t, NewRootCmd(""), "export", "csv", root, "-o", filteredDir,
		"--filter", `"Premium" in variant`); err != nil {
		t.Fatalf("export csv --filter: %v", err)
	}
	filtered, err := os.ReadFile(filepath.Join(filteredDir, "sys-requirements.csv"))
	if err != nil {
		t.Fatalf("read filtered csv: %v", err)
	}
	if !strings.Contains(string(filtered), "section,,Brakes") || !strings.Contains(string(filtered), "SYS-002") {
		t.Errorf("a folder that still leads to a match must survive with it:\n%s", filtered)
	}
	if strings.Contains(string(filtered), "Legacy") || strings.Contains(string(filtered), "SYS-003") {
		t.Errorf("a folder whose requirements were all pruned must be dropped:\n%s", filtered)
	}
	if !strings.Contains(string(filtered), "section,,Notes") {
		t.Errorf("a leaf section with prose must survive filtering:\n%s", filtered)
	}
}

// TestE2E_CSVImporterToCheckChain runs the documented flagship flow when a
// Python interpreter is available: DOORS CSV → reqmd_io bundle → renderer →
// check. It is skipped rather than failed when python3 is missing, the same way
// the client-side JS tests skip without node.
func TestE2E_CSVImporterToCheckChain(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not installed")
	}
	csvPath := "../../quickstart/15-custom-importers/brakes.csv"
	if _, err := os.Stat(csvPath); err != nil {
		t.Skipf("example CSV missing: %v", err)
	}

	bundlePath := filepath.Join(t.TempDir(), "brakes.json")
	cmd := exec.Command(python, "../../python/examples/csv_import.py", csvPath,
		"-o", bundlePath,
		"--key", "brakes", "--dir", "brakes", "--id-prefix", "DOORS-",
		"--title", "Brakes Module", "--level", "stakeholder-requirements", "--external",
		"--id-column", "Absolute Number", "--title-column", "Object Heading",
		"--body-column", "Object Text", "--parent-column", "Parent",
		"--trace-column", "Links", "--attr", "priority=Priority",
		"--kind-column", "Object Type")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("csv importer failed: %v\n%s", err, out)
	}

	raw, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	b, err := exchange.Decode(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("decode bundle: %v", err)
	}
	files, err := b.Build()
	if err != nil {
		t.Fatalf("build bundle: %v", err)
	}
	root := t.TempDir()
	for _, f := range files {
		writeFile(t, root, f.Path, string(f.Data))
	}

	out, err := runCmd(t, NewRootCmd(""), "check", root)
	if err != nil {
		t.Fatalf("check failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "0 invalid") {
		t.Fatalf("imported tree is invalid:\n%s", out)
	}
}

// TestE2E_ExportJSONTables pins what a plain `export json` guarantees at the
// CLI boundary: the four tables always exist, results stay empty without
// --results, authored text survives the JSON encoding, and a second export is
// byte-identical.
func TestE2E_ExportJSONTables(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sys/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sys\n")
	writeFile(t, root, "sys/sys.md", "# System\n\n"+
		mdReq("SYS-001", "Flow & pressure", "status: approved\nvariant: [Base]\n"))

	outDir := filepath.Join(t.TempDir(), "ndjson")
	if _, err := runCmd(t, NewRootCmd(""), "export", "json", root, "-o", outDir); err != nil {
		t.Fatalf("export json: %v", err)
	}

	tables := []string{"requirements.ndjson", "traces.ndjson", "results.ndjson", "findings.ndjson"}
	for _, name := range tables {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("%s missing after a plain export: %v", name, err)
		}
	}
	if rows := ndjsonRows(t, filepath.Join(outDir, "results.ndjson")); len(rows) != 0 {
		t.Errorf("results rows without --results = %d, want 0: %v", len(rows), rows)
	}
	rows := ndjsonRows(t, filepath.Join(outDir, "requirements.ndjson"))
	if len(rows) != 1 {
		t.Fatalf("requirement rows = %d, want 1: %v", len(rows), rows)
	}
	// The encoder escapes &, <, > in the bytes while decoding restores the
	// authored text: the table is JSON, not hand-built text.
	raw, err := os.ReadFile(filepath.Join(outDir, "requirements.ndjson"))
	if err != nil {
		t.Fatalf("read requirements.ndjson: %v", err)
	}
	if !strings.Contains(string(raw), `\u0026`) {
		t.Errorf("title should be HTML-escaped in the raw bytes:\n%s", raw)
	}
	if rows[0]["title"] != "Flow & pressure" {
		t.Errorf("title = %v, want the authored text", rows[0]["title"])
	}

	// A pure function of the tree: re-exporting changes no byte.
	snapshot := func() map[string]string {
		t.Helper()
		files := make(map[string]string, len(tables))
		for _, name := range tables {
			data, err := os.ReadFile(filepath.Join(outDir, name))
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			files[name] = string(data)
		}
		return files
	}
	first := snapshot()
	if _, err := runCmd(t, NewRootCmd(""), "export", "json", root, "-o", outDir); err != nil {
		t.Fatalf("second export json: %v", err)
	}
	second := snapshot()
	for _, name := range tables {
		if second[name] != first[name] {
			t.Errorf("%s changed between two exports:\nfirst:\n%s\nsecond:\n%s", name, first[name], second[name])
		}
	}
}

// TestE2E_ExportJSONFilter covers --filter at the CLI boundary: it scopes the
// requirements table and, because the command hands the filtered docs to the
// trace writer, traces.ndjson follows it too. Findings still cover the whole
// graph, so that table is written regardless.
func TestE2E_ExportJSONFilter(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sys/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sys\n")
	writeFile(t, root, "sys/sys.md", "# System\n\n"+
		mdReq("SYS-001", "Platform", "status: approved\nvariant: [Premium]\n")+
		mdReq("SYS-002", "Telemetry", "status: approved\nvariant: [Premium]\n"))
	writeFile(t, root, "sw/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sw\n")
	writeFile(t, root, "sw/sw.md", "# Software\n\n"+
		mdReq("SW-001", "Base client", "status: approved\nvariant: [Base]\ntrace: [sys/SYS-001]\n")+
		mdReq("SW-002", "Premium client", "status: approved\nvariant: [Premium]\ntrace: [sys/SYS-002]\n"))

	// Baseline: unfiltered, both trace refs become rows.
	allDir := filepath.Join(t.TempDir(), "all")
	if _, err := runCmd(t, NewRootCmd(""), "export", "json", root, "-o", allDir); err != nil {
		t.Fatalf("export json: %v", err)
	}
	if rows := ndjsonRows(t, filepath.Join(allDir, "traces.ndjson")); len(rows) != 2 {
		t.Fatalf("unfiltered trace rows = %d, want 2: %v", len(rows), rows)
	}

	filteredDir := filepath.Join(t.TempDir(), "filtered")
	if _, err := runCmd(t, NewRootCmd(""), "export", "json", root, "-o", filteredDir,
		"--filter", `"Premium" in variant`); err != nil {
		t.Fatalf("export json --filter: %v", err)
	}
	ids := map[string]bool{}
	for _, row := range ndjsonRows(t, filepath.Join(filteredDir, "requirements.ndjson")) {
		ids[row["id"].(string)] = true
	}
	for _, want := range []string{"SYS-001", "SYS-002", "SW-002"} {
		if !ids[want] {
			t.Errorf("matching requirement %s missing from requirements.ndjson: %v", want, keysOf(ids))
		}
	}
	if ids["SW-001"] {
		t.Errorf("filtered-out requirement SW-001 is still exported: %v", keysOf(ids))
	}

	traces := ndjsonRows(t, filepath.Join(filteredDir, "traces.ndjson"))
	if len(traces) != 1 || traces[0]["from"] != "SW-002" {
		t.Fatalf("filtered trace rows = %v, want only SW-002's", traces)
	}
	if _, err := os.Stat(filepath.Join(filteredDir, "findings.ndjson")); err != nil {
		t.Errorf("findings.ndjson missing on a filtered export: %v", err)
	}
}

// TestE2E_ExportJSONResults covers --results: a manual result bound to an
// authored measure becomes a verdict row while the requirement and trace
// tables keep the authored tree only.
func TestE2E_ExportJSONResults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sys/schema.yaml", variantSchema+"\nx-reqmd:\n  document-id: sys\n")
	writeFile(t, root, "sys/sys.md", "# System\n\n"+
		mdReq("SYS-001", "Platform", "status: approved\nvariant: [Base]\n"))
	writeFile(t, root, "sw/schema.yaml", verifySchema+"\nx-reqmd:\n  document-id: sw\n")
	writeFile(t, root, "sw/sw.md", "# Software\n\n"+
		mdReq("SW-001", "Base measure", "status: approved\nvariant: [Base]\nverify: Test\ntrace: [sys/SYS-001]\n"))
	results := writeManualResult(t, "SW-001", "pass")

	outDir := filepath.Join(t.TempDir(), "ndjson")
	if _, err := runCmd(t, NewRootCmd(""), "export", "json", root, "-o", outDir, "--results", results); err != nil {
		t.Fatalf("export json --results: %v", err)
	}

	// Rows are keyed by requirement: a measure gets its own evidence, and an
	// upstream requirement rolls the same evidence up.
	byMeasure := map[string]map[string]any{}
	for _, row := range ndjsonRows(t, filepath.Join(outDir, "results.ndjson")) {
		byMeasure[row["measure_id"].(string)] = row
	}
	measure := byMeasure["SW-001"]
	if measure == nil {
		t.Fatalf("no verdict row for SW-001: %v", keysOf(byMeasure))
	}
	if measure["outcome"] != "pass" || measure["verify"] != "Test" {
		t.Errorf("SW-001 verdict = %v, want pass with verify Test", measure)
	}
	if up := byMeasure["SYS-001"]; up == nil || up["outcome"] != "pass" {
		t.Errorf("SYS-001 verdict = %v, want the downstream pass rolled up", up)
	}
	traces := ndjsonRows(t, filepath.Join(outDir, "traces.ndjson"))
	if len(traces) != 1 || traces[0]["from"] != "SW-001" {
		t.Errorf("authored trace rows = %v, want only SW-001's", traces)
	}
}

// TestE2E_ExportJSONDisjointFinding covers --disjoint-check: a trace pair with
// no overlapping attribute value surfaces as an ERROR in findings.ndjson.
func TestE2E_ExportJSONDisjointFinding(t *testing.T) {
	root := writeDisjointSpec(t)

	outDir := filepath.Join(t.TempDir(), "ndjson")
	if _, err := runCmd(t, NewRootCmd(""), "export", "json", root, "-o", outDir, "--disjoint-check", "variant"); err != nil {
		t.Fatalf("export json --disjoint-check: %v", err)
	}

	var found bool
	for _, row := range ndjsonRows(t, filepath.Join(outDir, "findings.ndjson")) {
		if row["code"] != "disjoint-attribute" {
			continue
		}
		found = true
		if row["level"] != "ERROR" {
			t.Errorf("disjoint finding level = %v, want ERROR", row["level"])
		}
		if row["req_id"] != "SW-001" {
			t.Errorf("disjoint finding req_id = %v, want SW-001", row["req_id"])
		}
	}
	if !found {
		t.Fatalf("no disjoint-attribute finding in findings.ndjson:\n%v",
			ndjsonRows(t, filepath.Join(outDir, "findings.ndjson")))
	}
}

// ndjsonRows reads a JSON-lines table.
func ndjsonRows(t *testing.T, path string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var rows []map[string]any
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("bad NDJSON line: %v", err)
		}
		rows = append(rows, row)
	}
	return rows
}

// keysOf lists map keys for failure messages.
func keysOf[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
