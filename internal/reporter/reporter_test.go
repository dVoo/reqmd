package reporter

import (
	"encoding/json"
	"reqmd/internal/graph"
	"strings"
	"testing"
)

// mustUnmarshal decodes JSON string into a map for assertion checks.
func mustUnmarshal(t *testing.T, raw string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("invalid JSON: %v\nraw:\n%s", err, raw)
	}
	return m
}

// --- ExitCode ---

func TestExitCode(t *testing.T) {
	tests := []struct {
		name   string
		report Report
		want   int
	}{
		{
			name:   "empty report returns 0",
			report: Report{},
			want:   0,
		},
		{
			name: "only validation errors returns 1",
			report: Report{
				ValErrors: []ValidationError{
					{File: "a.md", ReqID: "R1", Message: "bad"},
				},
			},
			want: 1,
		},
		{
			name: "only parse errors returns 2",
			report: Report{
				ParseErrors: []ParseError{
					{File: "a.md", Message: "unreadable"},
				},
			},
			want: 2,
		},
		{
			name: "both validation and parse errors returns 2",
			report: Report{
				ValErrors: []ValidationError{
					{File: "a.md", ReqID: "R1", Message: "bad"},
				},
				ParseErrors: []ParseError{
					{File: "b.md", Message: "unreadable"},
				},
			},
			want: 2,
		},
		{
			name: "all valid returns 0",
			report: Report{
				ValidReqs: 5,
				TotalReqs: 5,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.report.ExitCode(); got != tt.want {
				t.Errorf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

// --- Format (empty) ---

func TestFormat_Empty(t *testing.T) {
	r := Report{}
	got := r.Format()
	if !strings.Contains(got, "Summary: 0 total, 0 valid, 0 invalid, 0 parse errors") {
		t.Errorf("empty report format should contain summary line\ngot:\n%s", got)
	}
}

// --- Format (validation errors) ---

func TestFormat_WithValidationErrors(t *testing.T) {
	r := Report{
		TotalReqs: 5,
		ValidReqs: 3,
		DocHeaders: []DocHeader{
			{Path: "docs", Title: "test", ReqCount: 5, ReqIDs: []string{"IVI-FUN-001", "IVI-FUN-002", "IVI-FUN-003", "IVI-FUN-004", "IVI-FUN-005"}},
		},
		ValErrors: []ValidationError{
			{File: "docs/ivi.md", ReqID: "IVI-FUN-001", Message: "missing required \"title\""},
			{File: "docs/ivi.md", ReqID: "IVI-FUN-003", Message: "missing required \"asil\""},
		},
	}
	r.NewIndex()
	got := r.Format()

	// Should show the doc header
	if !strings.Contains(got, "Schema : test") {
		t.Errorf("format should include Schema title\ngot:\n%s", got)
	}
	// Should contain per-req error lines with req IDs
	if !strings.Contains(got, "❌  IVI-FUN-001  missing required \"title\"") {
		t.Errorf("format should include first error\ngot:\n%s", got)
	}
	if !strings.Contains(got, "❌  IVI-FUN-003  missing required \"asil\"") {
		t.Errorf("format should include second error\ngot:\n%s", got)
	}
	// Should have ✅ for clean reqs
	if !strings.Contains(got, "✅  IVI-FUN-002") {
		t.Errorf("format should show clean reqs\ngot:\n%s", got)
	}
	// Should have correct summary counts
	if !strings.Contains(got, "Summary: 5 total, 3 valid, 2 invalid, 0 parse errors") {
		t.Errorf("format should show correct summary\ngot:\n%s", got)
	}
}

// --- Format (new format with DocHeaders) ---

func TestFormat_WithNewFormat(t *testing.T) {
	r := Report{
		TotalReqs: 2,
		ValidReqs: 1,
		DocHeaders: []DocHeader{
			{Path: "spec/example", Title: "ivi-srs-001 — IVI Attributes", ReqCount: 2, ReqIDs: []string{"R1", "R2"}},
		},
		ValErrors: []ValidationError{
			{File: "spec/example/ivi.md", ReqID: "R2", Message: "missing required: asil"},
		},
		GraphChecks: []graph.CheckResult{
			{Level: "WARNING", ReqID: "R1", File: "spec/example/ivi.md", Message: "reference SYS-001 not found"},
			{Level: "WARNING", ReqID: "R2", File: "spec/example/ivi.md", Message: "no upstream reference"},
		},
	}
	r.NewIndex()
	got := r.Format()

	// Should contain doc header fields
	if !strings.Contains(got, "Schema : ivi-srs-001") {
		t.Errorf("format should include Schema title\ngot:\n%s", got)
	}
	if !strings.Contains(got, "File   : spec/example") {
		t.Errorf("format should include File path\ngot:\n%s", got)
	}

	// R1 has trace warnings → ⚠ not ✅
	if !strings.Contains(got, "⚠  R1  attributes valid (see trace checks below)") {
		t.Errorf("format should show R1 with warning status\ngot:\n%s", got)
	}

	// R2 has a validation error → ❌
	if !strings.Contains(got, "❌  R2  missing required: asil") {
		t.Errorf("format should show R2 with error\ngot:\n%s", got)
	}

	// Graph warning message for R1
	if !strings.Contains(got, "reference SYS-001 not found") {
		t.Errorf("format should include graph warning message\ngot:\n%s", got)
	}

	// Graph warning message for R2 (missing-downstream WARNING)
	if !strings.Contains(got, "no upstream reference") {
		t.Errorf("format should include graph warning message\ngot:\n%s", got)
	}

	// Summary line
	if !strings.Contains(got, "Summary: 2 total, 1 valid, 1 invalid, 0 parse errors, 2 warnings") {
		t.Errorf("format should show correct summary\ngot:\n%s", got)
	}
}

// --- Format (empty with DocHeaders) ---

func TestFormat_EmptyWithDocHeaders(t *testing.T) {
	r := Report{
		DocHeaders: []DocHeader{
			{Path: "spec/example", Title: "ivi-srs-001", ReqCount: 0},
		},
	}
	r.NewIndex()
	got := r.Format()

	// Should still show the doc header
	if !strings.Contains(got, "Schema : ivi-srs-001") {
		t.Errorf("format should include Schema title\ngot:\n%s", got)
	}
	if !strings.Contains(got, "File   : spec/example  (0 requirements)") {
		t.Errorf("format should include File path\ngot:\n%s", got)
	}

	// Should show empty summary
	if !strings.Contains(got, "Summary: 0 total, 0 valid, 0 invalid, 0 parse errors") {
		t.Errorf("empty report with doc headers should show zero summary\ngot:\n%s", got)
	}
}

// --- ExitCode with graph errors ---

func TestExitCode_GraphErrorChangesExitCode(t *testing.T) {
	r := Report{
		TotalReqs: 2,
		ValidReqs: 2,
		GraphChecks: []graph.CheckResult{
			{Level: "ERROR", ReqID: "R1", File: "spec/example/ivi.md", Message: "cycle detected: R1 → R2 → R1"},
		},
	}
	if got := r.ExitCode(); got != 1 {
		t.Errorf("ExitCode() with graph ERROR = %d, want 1", got)
	}
}

// --- Format (parse errors) ---

func TestFormat_WithParseErrors(t *testing.T) {
	r := Report{
		TotalReqs: 1,
		ValidReqs: 0,
		ParseErrors: []ParseError{
			{File: "docs/broken.md", Message: "yaml: line 3: unmarshal errors"},
		},
	}
	got := r.Format()

	if !strings.Contains(got, "PARSE ERROR") {
		t.Errorf("format should contain PARSE ERROR heading\ngot:\n%s", got)
	}
	if !strings.Contains(got, "docs/broken.md") {
		t.Errorf("format should contain the file path\ngot:\n%s", got)
	}
	if !strings.Contains(got, "yaml: line 3: unmarshal errors") {
		t.Errorf("format should contain the parse error message\ngot:\n%s", got)
	}
	if !strings.Contains(got, "Summary: 1 total, 0 valid, 0 invalid, 1 parse errors") {
		t.Errorf("format should show correct summary\ngot:\n%s", got)
	}
}

// --- FormatList ---

func TestFormatList(t *testing.T) {
	docs := []DocumentSummary{
		{
			Path:       "spec/example",
			Properties: []string{"status", "asil"},
			Rows: []NodeRow{
				{ID: "IVI-FUN-001", Attrs: map[string]any{"status": "approved", "asil": "B"}},
				{ID: "IVI-FUN-002", Attrs: map[string]any{"status": "draft", "asil": "A"}},
			},
		},
	}
	got := FormatList(docs)

	// Should contain document heading
	if !strings.Contains(got, "=== spec/example ===") {
		t.Errorf("FormatList should include document path heading\ngot:\n%s", got)
	}
	// Should contain header columns
	if !strings.Contains(got, "ID") {
		t.Errorf("FormatList should have ID header\ngot:\n%s", got)
	}
	if !strings.Contains(got, "status") {
		t.Errorf("FormatList should have status header\ngot:\n%s", got)
	}
	if !strings.Contains(got, "asil") {
		t.Errorf("FormatList should have asil header\ngot:\n%s", got)
	}
	// Should contain requirement IDs
	if !strings.Contains(got, "IVI-FUN-001") {
		t.Errorf("FormatList should contain IVI-FUN-001\ngot:\n%s", got)
	}
	if !strings.Contains(got, "IVI-FUN-002") {
		t.Errorf("FormatList should contain IVI-FUN-002\ngot:\n%s", got)
	}
	// Should contain attribute values
	if !strings.Contains(got, "approved") {
		t.Errorf("FormatList should contain 'approved'\ngot:\n%s", got)
	}
	if !strings.Contains(got, "draft") {
		t.Errorf("FormatList should contain 'draft'\ngot:\n%s", got)
	}
	if !strings.Contains(got, "B") {
		t.Errorf("FormatList should contain asil value B\ngot:\n%s", got)
	}
	if !strings.Contains(got, "A") {
		t.Errorf("FormatList should contain asil value A\ngot:\n%s", got)
	}
}

func TestFormatList_WithTitle(t *testing.T) {
	docs := []DocumentSummary{
		{
			Path:       "spec/example",
			Properties: []string{"status"},
			Rows: []NodeRow{
				{ID: "IVI-FUN-001", Title: "Login req", Attrs: map[string]any{"status": "approved"}},
				{ID: "IVI-FUN-002", Title: "", Attrs: map[string]any{"status": "draft"}},
			},
		},
	}
	got := FormatList(docs)

	// Row with title should show "ID: Title" format
	if !strings.Contains(got, "IVI-FUN-001: Login req") {
		t.Errorf("FormatList should show ID: Title when Title is present\ngot:\n%s", got)
	}
	// Row without title should show just ID
	if !strings.Contains(got, "IVI-FUN-002") {
		t.Errorf("FormatList should contain IVI-FUN-002\ngot:\n%s", got)
	}
}

func TestFormatListJSON_WithTitle(t *testing.T) {
	docs := []DocumentSummary{
		{
			Path:       "spec/example",
			Properties: []string{"status"},
			Rows: []NodeRow{
				{ID: "R1", Title: "Login req", Attrs: map[string]any{"status": "approved"}},
				{ID: "R2", Title: "", Attrs: map[string]any{"status": "draft"}},
			},
		},
	}
	got := FormatListJSON(docs)
	m := mustUnmarshal(t, got)

	docsArr := m["documents"].([]any)
	doc := docsArr[0].(map[string]any)
	rows := doc["rows"].([]any)

	// R1 should have title field
	r1 := rows[0].(map[string]any)
	if r1["id"] != "R1" {
		t.Errorf("req[0] id = %v", r1["id"])
	}
	if r1["title"] != "Login req" {
		t.Errorf("req[0] title = %v, want 'Login req'", r1["title"])
	}
	// R2 should omit title when empty (omitempty)
	r2 := rows[1].(map[string]any)
	if r2["id"] != "R2" {
		t.Errorf("req[1] id = %v", r2["id"])
	}
	if _, hasTitle := r2["title"]; hasTitle {
		t.Errorf("req[1] should not have title field when empty, got title = %v", r2["title"])
	}
}

// --- FormatStats ---

func TestFormatStats(t *testing.T) {
	docs := []DocumentSummary{
		{
			Path:       "spec/example",
			Properties: []string{"status"},
			ReqCount:   3,
			Rows: []NodeRow{
				{ID: "R1", Attrs: map[string]any{"status": "approved"}},
				{ID: "R2", Attrs: map[string]any{"status": "draft"}},
				{ID: "R3", Attrs: map[string]any{"status": "approved"}},
			},
		},
	}
	got := FormatStats(docs, 3)

	// Should contain total requirements line
	if !strings.Contains(got, "Requirements: 3") {
		t.Errorf("FormatStats should show Requirements: 3\ngot:\n%s", got)
	}
	// Should contain document count
	if !strings.Contains(got, "Documents:    1") {
		t.Errorf("FormatStats should show Documents: 1\ngot:\n%s", got)
	}
	// Should contain document heading with req count
	if !strings.Contains(got, "=== spec/example (3 reqs) ===") {
		t.Errorf("FormatStats should include doc heading\ngot:\n%s", got)
	}
	// Should contain property name
	if !strings.Contains(got, "status:") {
		t.Errorf("FormatStats should include property name 'status'\ngot:\n%s", got)
	}
	// Should contain grouped counts (approved: 2, draft: 1)
	if !strings.Contains(got, "approved") || !strings.Contains(got, "2") {
		t.Errorf("FormatStats should show approved count of 2\ngot:\n%s", got)
	}
	if !strings.Contains(got, "draft") || !strings.Contains(got, "1") {
		t.Errorf("FormatStats should show draft count of 1\ngot:\n%s", got)
	}
}

// --- JSON output ---

func TestFormatJSON_Simple(t *testing.T) {
	r := Report{
		TotalReqs: 5,
		ValidReqs: 3,
		DocHeaders: []DocHeader{
			{Path: "spec/example", Title: "ivi-srs-001", ReqCount: 5, ReqIDs: []string{"R1", "R2", "R3", "R4", "R5"}},
		},
		ValErrors: []ValidationError{
			{File: "spec/example/ivi.md", ReqID: "R1", Message: "missing required \"title\""},
		},
		GraphChecks: []graph.CheckResult{
			{Level: "WARNING", ReqID: "R2", File: "spec/example/ivi.md", Message: "broken reference: SYS-001"},
			{Level: "WARNING", ReqID: "R3", File: "spec/example/ivi.md", Message: "no upstream reference"},
		},
	}
	r.NewIndex()
	got := r.FormatJSON()
	m := mustUnmarshal(t, got)

	// Version
	if v := m["version"].(float64); v != 1 {
		t.Errorf("version = %v, want 1", v)
	}
	// Exit code
	if ec := m["exit_code"].(float64); ec != 1 {
		t.Errorf("exit_code = %v, want 1", ec)
	}
	// Summary
	sm := m["summary"].(map[string]any)
	if sm["total"].(float64) != 5 {
		t.Errorf("summary.total = %v, want 5", sm["total"])
	}
	if sm["invalid"].(float64) != 1 {
		t.Errorf("summary.invalid = %v, want 1", sm["invalid"])
	}
	if sm["warnings"].(float64) != 2 {
		t.Errorf("summary.warnings = %v, want 2", sm["warnings"])
	}
	// Documents
	docs := m["documents"].([]any)
	if len(docs) != 1 {
		t.Fatalf("documents len = %d, want 1", len(docs))
	}
	doc := docs[0].(map[string]any)
	if doc["path"] != "spec/example" {
		t.Errorf("doc path = %v", doc["path"])
	}
	// Requirements
	reqs := doc["requirements"].([]any)
	if len(reqs) != 5 {
		t.Fatalf("requirements len = %d, want 5", len(reqs))
	}
	// R1 = has schema validation error
	r1 := reqs[0].(map[string]any)
	if r1["valid"] != false {
		t.Errorf("R1 valid = %v, want false", r1["valid"])
	}
	r1checks := r1["checks"].([]any)
	if len(r1checks) < 1 {
		t.Fatal("R1 should have at least 1 check")
	}
	c0 := r1checks[0].(map[string]any)
	if c0["level"] != "ERROR" {
		t.Errorf("R1 check[0] level = %v, want ERROR", c0["level"])
	}
	// R2 = valid with trace warning
	r2 := reqs[1].(map[string]any)
	if r2["valid"] != true {
		t.Errorf("R2 valid = %v, want true", r2["valid"])
	}
	r2checks := r2["checks"].([]any)
	if len(r2checks) != 1 {
		t.Fatalf("R2 checks len = %d, want 1", len(r2checks))
	}
	chk := r2checks[0].(map[string]any)
	if chk["level"] != "WARNING" {
		t.Errorf("R2 check level = %v", chk["level"])
	}
	// Parse errors should be empty
	pes := m["parse_errors"].([]any)
	if len(pes) != 0 {
		t.Errorf("parse_errors len = %d, want 0", len(pes))
	}
}

func TestFormatJSON_Empty(t *testing.T) {
	r := Report{}
	r.NewIndex()
	got := r.FormatJSON()
	m := mustUnmarshal(t, got)

	if v := m["version"].(float64); v != 1 {
		t.Errorf("version = %v, want 1", v)
	}
	if ec := m["exit_code"].(float64); ec != 0 {
		t.Errorf("exit_code = %v, want 0", ec)
	}
	sm := m["summary"].(map[string]any)
	if sm["total"].(float64) != 0 {
		t.Errorf("summary.total = %v, want 0", sm["total"])
	}
}

func TestFormatJSON_ParseErrors(t *testing.T) {
	r := Report{
		ParseErrors: []ParseError{
			{File: "broken.md", Message: "yaml: line 3: unmarshal errors"},
		},
	}
	got := r.FormatJSON()
	m := mustUnmarshal(t, got)

	if ec := m["exit_code"].(float64); ec != 2 {
		t.Errorf("exit_code = %v, want 2", ec)
	}
	pes := m["parse_errors"].([]any)
	if len(pes) != 1 {
		t.Fatalf("parse_errors len = %d, want 1", len(pes))
	}
	pe := pes[0].(map[string]any)
	if pe["file"] != "broken.md" {
		t.Errorf("parse_error file = %v", pe["file"])
	}
}

func TestFormatJSON_GraphErrors(t *testing.T) {
	r := Report{
		TotalReqs: 2,
		ValidReqs: 2,
		DocHeaders: []DocHeader{
			{Path: "docs", Title: "test", ReqCount: 2, ReqIDs: []string{"A", "B"}},
		},
		GraphChecks: []graph.CheckResult{
			{Level: "ERROR", ReqID: "A", File: "docs/a.md", Message: "circular dependency detected: A->B->A"},
		},
	}
	r.NewIndex()
	got := r.FormatJSON()
	m := mustUnmarshal(t, got)

	if ec := m["exit_code"].(float64); ec != 1 {
		t.Errorf("exit_code with graph error = %v, want 1", ec)
	}
	sm := m["summary"].(map[string]any)
	if sm["invalid"].(float64) != 1 {
		t.Errorf("summary.invalid = %v, want 1", sm["invalid"])
	}
}

func TestFormatListJSON(t *testing.T) {
	docs := []DocumentSummary{
		{
			Path:       "spec/example",
			Properties: []string{"status", "asil"},
			Rows: []NodeRow{
				{ID: "R1", Attrs: map[string]any{"status": "approved", "asil": "B"}},
				{ID: "R2", Attrs: map[string]any{"status": "draft"}},
			},
		},
	}
	got := FormatListJSON(docs)
	m := mustUnmarshal(t, got)

	docsArr := m["documents"].([]any)
	if len(docsArr) != 1 {
		t.Fatalf("documents len = %d, want 1", len(docsArr))
	}
	doc := docsArr[0].(map[string]any)
	if doc["path"] != "spec/example" {
		t.Errorf("path = %v", doc["path"])
	}
	rows := doc["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("requirements len = %d, want 2", len(rows))
	}
	r1 := rows[0].(map[string]any)
	if r1["id"] != "R1" {
		t.Errorf("req[0] id = %v", r1["id"])
	}
}

func TestFormatStatsJSON(t *testing.T) {
	docs := []DocumentSummary{
		{
			Path:       "spec/example",
			Properties: []string{"status"},
			ReqCount:   3,
			Rows: []NodeRow{
				{ID: "R1", Attrs: map[string]any{"status": "approved"}},
				{ID: "R2", Attrs: map[string]any{"status": "draft"}},
				{ID: "R3", Attrs: map[string]any{"status": "approved"}},
			},
		},
	}
	got := FormatStatsJSON(docs, 3)
	m := mustUnmarshal(t, got)

	if m["total_requirements"].(float64) != 3 {
		t.Errorf("total_requirements = %v, want 3", m["total_requirements"])
	}
	if m["total_documents"].(float64) != 1 {
		t.Errorf("total_documents = %v, want 1", m["total_documents"])
	}
	docsArr := m["documents"].([]any)
	doc := docsArr[0].(map[string]any)
	astats := doc["attribute_stats"].(map[string]any)
	statusCounts := astats["status"].(map[string]any)
	if statusCounts["approved"].(float64) != 2 {
		t.Errorf("approved count = %v, want 2", statusCounts["approved"])
	}
	if statusCounts["draft"].(float64) != 1 {
		t.Errorf("draft count = %v, want 1", statusCounts["draft"])
	}
}
