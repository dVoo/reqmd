package exporter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"reqmd/internal/graph"
	"reqmd/internal/model"
)

// jsonDocs builds a small two-document tree: one upstream with a version,
// one downstream that traces to it with a pin and a dangling ref. Enough to
// exercise every table.
func jsonDocs() []model.Document {
	up := model.Document{
		Path:  "/spec/up",
		Title: "Upstream",
		XReqmd: &model.XReqmd{
			Level:      "upstream-level",
			DocumentID: "upstream",
		},
		Nodes: []*model.Node{
			{
				ID:     "UP-001",
				Title:  "Up",
				Kind:   model.KindRequirement,
				Source: "/spec/up/up.md",
				Attrs:  map[string]any{"status": "approved", "version": 3},
			},
		},
	}
	down := model.Document{
		Path:  "/spec/down",
		Title: "Downstream",
		XReqmd: &model.XReqmd{
			Level:      "downstream-level",
			DocumentID: "downstream",
		},
		Nodes: []*model.Node{
			{
				ID:     "DN-001",
				Title:  "Down",
				Kind:   model.KindRequirement,
				Source: "/spec/down/dn.md",
				Attrs: map[string]any{
					"status":   "approved",
					"priority": "High",
					"variant":  []any{"Base", "Premium"},
					"trace":    []any{"upstream/UP-001~1", "MISSING-9"},
				},
			},
			{
				ID:     "Group",
				Title:  "A container",
				Kind:   model.KindContainer,
				Source: "/spec/down/dn.md",
			},
		},
	}
	synthetic := model.Document{
		Path:      "/spec/results",
		Title:     "Results",
		Synthetic: true,
		Nodes: []*model.Node{
			{
				ID:    "RESULT:UP-001",
				Kind:  model.KindRequirement,
				Attrs: map[string]any{"x-reqmd.outcome": "pass"},
			},
		},
	}
	return []model.Document{up, down, synthetic}
}

func encodeRows(t *testing.T, export func(w *bufio.Writer) error) []map[string]any {
	t.Helper()
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err := export(w); err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	var rows []map[string]any
	for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("line is not valid JSON: %v\n%s", err, line)
		}
		rows = append(rows, row)
	}
	return rows
}

func TestJSONExportRequirements(t *testing.T) {
	docs := jsonDocs()
	j := &JSON{}
	rows := encodeRows(t, func(w *bufio.Writer) error {
		return j.ExportRequirements(w, docs)
	})

	// The synthetic document must be skipped; 3 authored nodes remain
	// (UP-001, DN-001, and the Group container).
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (synthetic excluded), got %d", len(rows))
	}

	byID := map[string]map[string]any{}
	for _, r := range rows {
		byID[r["id"].(string)] = r
	}

	dn := byID["DN-001"]
	if dn == nil {
		t.Fatal("DN-001 row missing")
	}
	if dn["type"] != "req" {
		t.Errorf("DN-001 type = %v, want req", dn["type"])
	}
	if dn["status"] != "approved" {
		t.Errorf("DN-001 status = %v, want approved", dn["status"])
	}
	if dn["level"] != "downstream-level" {
		t.Errorf("DN-001 level = %v, want downstream-level", dn["level"])
	}
	// Built-ins are promoted; trace is a top-level array.
	trace, ok := dn["trace"].([]any)
	if !ok || len(trace) != 2 {
		t.Fatalf("DN-001 trace = %#v, want 2-element array", dn["trace"])
	}
	// attrs holds only the non-promoted declared attributes.
	attrs, ok := dn["attrs"].(map[string]any)
	if !ok {
		t.Fatalf("DN-001 attrs = %#v, want object", dn["attrs"])
	}
	if attrs["priority"] != "High" {
		t.Errorf("attrs.priority = %v, want High", attrs["priority"])
	}
	if _, promoted := attrs["trace"]; promoted {
		t.Error("attrs must not duplicate the promoted trace key")
	}
	if _, promoted := attrs["status"]; promoted {
		t.Error("attrs must not duplicate the promoted status key")
	}
	// Native JSON types survive.
	if _, isArray := attrs["variant"].([]any); !isArray {
		t.Errorf("attrs.variant = %#v, want array", attrs["variant"])
	}

	up := byID["UP-001"]
	if up["version"] != float64(3) {
		t.Errorf("UP-001 version = %#v, want 3", up["version"])
	}

	grp := byID["Group"]
	if grp["type"] != "container" {
		t.Errorf("Group type = %v, want container", grp["type"])
	}
}

func TestJSONExportTraces(t *testing.T) {
	docs := jsonDocs()
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("graph build failed: %v", err)
	}
	j := &JSON{}
	rows := encodeRows(t, func(w *bufio.Writer) error {
		return j.ExportTraces(w, docs, g)
	})

	if len(rows) != 2 {
		t.Fatalf("expected 2 trace rows, got %d", len(rows))
	}
	byRef := map[string]map[string]any{}
	for _, r := range rows {
		byRef[r["ref"].(string)] = r
	}

	resolved := byRef["upstream/UP-001~1"]
	if resolved == nil {
		t.Fatal("resolved trace row missing")
	}
	if resolved["resolved"] != true {
		t.Errorf("resolved = %v, want true", resolved["resolved"])
	}
	if resolved["to"] != "UP-001" {
		t.Errorf("to = %v, want UP-001", resolved["to"])
	}
	if resolved["doc_id"] != "upstream" {
		t.Errorf("doc_id = %v, want upstream", resolved["doc_id"])
	}
	if resolved["pin"] != float64(1) {
		t.Errorf("pin = %v, want 1", resolved["pin"])
	}

	dangling := byRef["MISSING-9"]
	if dangling == nil {
		t.Fatal("dangling trace row missing")
	}
	if dangling["resolved"] != false {
		t.Errorf("dangling resolved = %v, want false", dangling["resolved"])
	}
	if dangling["pin"] != nil {
		t.Errorf("dangling pin = %v, want null", dangling["pin"])
	}
}

func TestJSONExportFindings(t *testing.T) {
	docs := jsonDocs()
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("graph build failed: %v", err)
	}
	j := &JSON{}
	rows := encodeRows(t, func(w *bufio.Writer) error {
		return j.ExportFindings(w, g)
	})

	var broken, pin map[string]any
	for i := range rows {
		r := rows[i]
		switch {
		case r["code"] == "version-pin":
			pin = r
		case strings.Contains(r["message"].(string), "broken reference"):
			broken = r
		}
	}

	if broken == nil {
		t.Fatal("expected a broken-reference finding for MISSING-9")
	}
	if broken["level"] != "WARNING" {
		t.Errorf("broken-ref level = %v, want WARNING", broken["level"])
	}
	if broken["file"] == nil {
		t.Error("broken-ref finding must carry the source file")
	}

	if pin == nil {
		t.Fatal("expected a version-pin finding for the outdated ~1 pin")
	}
	if pin["level"] != "ERROR" {
		t.Errorf("version-pin level = %v, want ERROR", pin["level"])
	}
	if pin["direction"] != "outdated" {
		t.Errorf("version-pin direction = %v, want outdated", pin["direction"])
	}
}

func TestJSONExportDeterministic(t *testing.T) {
	docs := jsonDocs()
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("graph build failed: %v", err)
	}

	run := func() string {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		j := &JSON{}
		if err := j.ExportRequirements(w, docs); err != nil {
			t.Fatal(err)
		}
		if err := j.ExportTraces(w, docs, g); err != nil {
			t.Fatal(err)
		}
		if err := j.ExportFindings(w, g); err != nil {
			t.Fatal(err)
		}
		if err := w.Flush(); err != nil {
			t.Fatal(err)
		}
		return buf.String()
	}

	first := run()
	second := run()
	if first != second {
		t.Error("two exports of the same tree differ; output must be deterministic")
	}
}

func TestJSONExportStableColumns(t *testing.T) {
	// Every fixed key must be present in every row (null when unset), or
	// warehouse schema inference sees only the columns of whichever row it
	// samples first. Attrs is the one intentional exception.
	docs := jsonDocs()
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("graph build failed: %v", err)
	}

	keysOf := func(m map[string]any) map[string]bool {
		out := make(map[string]bool, len(m))
		for k := range m {
			out[k] = true
		}
		return out
	}

	// requirements: compare the sparse container row against the full one.
	reqRows := encodeRows(t, func(w *bufio.Writer) error {
		return (&JSON{}).ExportRequirements(w, docs)
	})
	var full, sparse map[string]any
	for _, r := range reqRows {
		switch r["id"] {
		case "DN-001":
			full = r
		case "Group":
			sparse = r
		}
	}
	if full == nil || sparse == nil {
		t.Fatal("expected both DN-001 and Group rows")
	}
	fullKeys, sparseKeys := keysOf(full), keysOf(sparse)
	for k := range fullKeys {
		if !sparseKeys[k] {
			t.Errorf("requirements: key %q present in a full row but missing from a sparse row", k)
		}
	}

	// traces: a dangling row and a resolved row must share columns.
	traceRows := encodeRows(t, func(w *bufio.Writer) error {
		return (&JSON{}).ExportTraces(w, docs, g)
	})
	var tFull, tSparse map[string]any
	for _, r := range traceRows {
		if r["doc_id"] == "upstream" {
			tFull = r
		} else {
			tSparse = r
		}
	}
	if tFull == nil || tSparse == nil {
		t.Fatal("expected both a resolved and a dangling trace row")
	}
	tFullKeys, tSparseKeys := keysOf(tFull), keysOf(tSparse)
	for k := range tFullKeys {
		if !tSparseKeys[k] {
			t.Errorf("traces: key %q present in a resolved row but missing from a dangling row", k)
		}
	}

	// findings: the version-pin row carries code/direction; a code-less
	// finding must still carry those keys as null.
	findRows := encodeRows(t, func(w *bufio.Writer) error {
		return (&JSON{}).ExportFindings(w, g)
	})
	findKeys := keysOf(findRows[0])
	for _, k := range []string{"level", "req_id", "code", "direction", "outcome", "file", "message"} {
		if !findKeys[k] {
			t.Errorf("findings: required column %q missing", k)
		}
	}
}

func TestJSONExportResults(t *testing.T) {
	docs := jsonDocs()
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("graph build failed: %v", err)
	}
	j := &JSON{}
	j.SetVerdicts(map[string]VerdictInfo{
		"UP-001": {Outcome: "pass", Source: "reviews/vr.md"},
	})
	rows := encodeRows(t, func(w *bufio.Writer) error {
		return j.ExportResults(w, g)
	})

	if len(rows) != 1 {
		t.Fatalf("expected 1 result row, got %d", len(rows))
	}
	if rows[0]["measure_id"] != "UP-001" {
		t.Errorf("measure_id = %v, want UP-001", rows[0]["measure_id"])
	}
	if rows[0]["outcome"] != "pass" {
		t.Errorf("outcome = %v, want pass", rows[0]["outcome"])
	}
	if rows[0]["source"] != "reviews/vr.md" {
		t.Errorf("source = %v, want reviews/vr.md", rows[0]["source"])
	}
}

func TestJSONExportResultsWithoutResults(t *testing.T) {
	docs := jsonDocs()
	g, err := graph.New(docs)
	if err != nil {
		t.Fatalf("graph build failed: %v", err)
	}
	var buf bytes.Buffer
	if err := (&JSON{}).ExportResults(&buf, g); err != nil {
		t.Fatalf("export without results must be a no-op, got %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output without --results, got %q", buf.String())
	}
}
