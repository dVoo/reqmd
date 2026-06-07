package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseMD_Full verifies a normal requirement with heading, attr block,
// prose body, and *Rationale: line.
func TestParseMD_Full(t *testing.T) {
	src := []byte("## REQ-001\n```attr\nid: REQ-001\nasil: ASIL D\n```\nThis requirement shall do something important.\n\n*Rationale: safety critical function\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	if r.ID != "REQ-001" {
		t.Errorf("ID = %q, want REQ-001", r.ID)
	}
	if r.Attrs["id"] != "REQ-001" {
		t.Errorf("attrs[id] = %v, want REQ-001", r.Attrs["id"])
	}
	if r.Attrs["asil"] != "ASIL D" {
		t.Errorf("attrs[asil] = %v, want ASIL D", r.Attrs["asil"])
	}
	if r.Body != "This requirement shall do something important." {
		t.Errorf("Body = %q, want %q", r.Body, "This requirement shall do something important.")
	}
	if r.Rationale != "safety critical function" {
		t.Errorf("Rational = %q, want %q", r.Rationale, "safety critical function")
	}
	if r.Source != "test.md" {
		t.Errorf("Source = %q, want test.md", r.Source)
	}
}

// TestParseMD_NoRationale verifies that a requirement without *Rationale:*
// has an empty Rational field.
func TestParseMD_NoRationale(t *testing.T) {
	src := []byte("## REQ-002\n```attr\nid: REQ-002\nasil: ASIL B\n```\nThis requirement has no rationale.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	if r.ID != "REQ-002" {
		t.Errorf("ID = %q, want REQ-002", r.ID)
	}
	if r.Rationale != "" {
		t.Errorf("Rational = %q, want empty string", r.Rationale)
	}
}

// TestParseMD_MultipleRequirements verifies that multiple requirements
// in one document are each parsed correctly.
func TestParseMD_MultipleRequirements(t *testing.T) {
	src := []byte("## REQ-A\n```attr\nid: REQ-A\n```\nBody A.\n\n## REQ-B\n```attr\nid: REQ-B\n```\nBody B.\n\n## REQ-C\n```attr\nid: REQ-C\n```\nBody C.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 3 {
		t.Fatalf("expected 3 requirements, got %d", len(reqs))
	}
	expectIDs := []string{"REQ-A", "REQ-B", "REQ-C"}
	expectBodies := []string{"Body A.", "Body B.", "Body C."}
	for i, r := range reqs {
		if r.ID != expectIDs[i] {
			t.Errorf("req[%d] ID = %q, want %q", i, r.ID, expectIDs[i])
		}
		if r.Body != expectBodies[i] {
			t.Errorf("req[%d] Body = %q, want %q", i, r.Body, expectBodies[i])
		}
		if r.Attrs["id"] != expectIDs[i] {
			t.Errorf("req[%d] attrs[id] = %v, want %q", i, r.Attrs["id"], expectIDs[i])
		}
	}
}

// TestParseMD_MissingAttrBlock verifies that a ## heading without an attr
// block produces no requirements (treated as section divider / body text).
func TestParseMD_MissingAttrBlock(t *testing.T) {
	src := []byte("## REQ-NOATTR\nThis has no attr block.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatalf("expected no error for non-attr heading (section divider), got: %v", err)
	}
	if len(reqs) != 0 {
		t.Fatalf("expected 0 requirements, got %d", len(reqs))
	}
}

// TestParseMD_BadYAML verifies that invalid YAML inside an ```attr``` block
// returns an error.
func TestParseMD_BadYAML(t *testing.T) {
	src := []byte("## REQ-BAD\n```attr\n: : invalid yaml\n```\n")
	reqs, _, err := parseMD(src, "test.md")
	if err == nil {
		t.Fatal("expected error for invalid attr YAML, got nil")
	}
	if reqs != nil {
		t.Fatalf("expected nil reqs on error, got %d reqs", len(reqs))
	}
}

// TestParseMD_NonAttrCodeBlock verifies that a fenced code block with a
// non-attr info string (e.g. ```json) inside a requirement body is preserved.
func TestParseMD_NonAttrCodeBlock(t *testing.T) {
	src := []byte("## REQ-001\n```attr\nid: REQ-001\n```\nSome text.\n\n```json\n{\"key\": \"value\"}\n```\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	expected := "Some text.```json\n{\"key\": \"value\"}\n```\n\n"
	if r.Body != expected {
		t.Errorf("Body = %q, want %q", r.Body, expected)
	}
}

// TestParseMD_EmptyFile verifies that parsing an empty source returns
// nil requirements with no error.
func TestParseMD_EmptyFile(t *testing.T) {
	reqs, _, err := parseMD([]byte{}, "empty.md")
	if err != nil {
		t.Fatal(err)
	}
	if reqs != nil {
		t.Fatalf("expected nil reqs for empty file, got %d", len(reqs))
	}
}

// TestParseMD_ThematicBreak verifies that thematic breaks (---) inside
// a requirement body are ignored and don't disrupt body accumulation.
func TestParseMD_ThematicBreak(t *testing.T) {
	src := []byte("## REQ-001\n```attr\nid: REQ-001\n```\nBefore break.\n\n---\nAfter break.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	expected := "Before break.\n\nAfter break."
	if r.Body != expected {
		t.Errorf("Body = %q, want %q", r.Body, expected)
	}
}

// TestParseMD_ArrayAttr verifies that an attribute with a YAML array value
// is parsed into []interface{} with the correct elements.
func TestParseMD_ArrayAttr(t *testing.T) {
	src := []byte("## REQ-001\n```attr\ntrace: [SYS-001, SAFE-003]\n```\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	trace, ok := r.Attrs["trace"].([]interface{})
	if !ok {
		t.Fatalf("attrs[trace] is %T, want []interface{}", r.Attrs["trace"])
	}
	if len(trace) != 2 {
		t.Fatalf("attrs[trace] has length %d, want 2", len(trace))
	}
	if trace[0] != "SYS-001" {
		t.Errorf("attrs[trace][0] = %v, want SYS-001", trace[0])
	}
	if trace[1] != "SAFE-003" {
		t.Errorf("attrs[trace][1] = %v, want SAFE-003", trace[1])
	}
}

// TestParseMD_SubReqUnderParent verifies that a ### heading with an attr block
// immediately following a ## heading creates a sub-requirement with ParentID set.
func TestParseMD_SubReqUnderParent(t *testing.T) {
	src := []byte("## PARENT-001\n```attr\nid: PARENT-001\n```\nParent body.\n\n### CHILD-001\n```attr\nid: CHILD-001\n```\nChild body.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requirements (parent + child), got %d", len(reqs))
	}
	r0 := reqs[0]
	if r0.ID != "PARENT-001" {
		t.Errorf("req[0] ID = %q, want PARENT-001", r0.ID)
	}
	if r0.ParentID != "" {
		t.Errorf("req[0] ParentID = %q, want empty", r0.ParentID)
	}
	if r0.Body != "Parent body." {
		t.Errorf("req[0] Body = %q, want %q", r0.Body, "Parent body.")
	}

	r1 := reqs[1]
	if r1.ID != "CHILD-001" {
		t.Errorf("req[1] ID = %q, want CHILD-001", r1.ID)
	}
	if r1.ParentID != "PARENT-001" {
		t.Errorf("req[1] ParentID = %q, want PARENT-001", r1.ParentID)
	}
	if r1.Body != "Child body." {
		t.Errorf("req[1] Body = %q, want %q", r1.Body, "Child body.")
	}
}

// TestParseMD_MultipleSubReqs verifies that multiple ### sub-requirements
// under the same ## parent all get the correct ParentID.
func TestParseMD_MultipleSubReqs(t *testing.T) {
	src := []byte("## PARENT\n```attr\nid: PARENT\n```\nP.\n\n### CHILD-A\n```attr\nid: CHILD-A\n```\nA.\n\n### CHILD-B\n```attr\nid: CHILD-B\n```\nB.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 3 {
		t.Fatalf("expected 3 requirements, got %d", len(reqs))
	}
	if reqs[0].ParentID != "" {
		t.Errorf("req[0] ParentID = %q, want empty", reqs[0].ParentID)
	}
	if reqs[1].ParentID != "PARENT" {
		t.Errorf("req[1] ParentID = %q, want PARENT", reqs[1].ParentID)
	}
	if reqs[2].ParentID != "PARENT" {
		t.Errorf("req[2] ParentID = %q, want PARENT", reqs[2].ParentID)
	}
}

// TestParseMD_SubReqDynamicLevel verifies that a ### heading with attr block
// as the FIRST requirement sets reqLevel to 3 (dynamic discovery), making it
// a top-level requirement, not an orphan child.
func TestParseMD_SubReqDynamicLevel(t *testing.T) {
	src := []byte("### ORPHAN\n```attr\nid: ORPHAN\n```\nNo parent.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	if reqs[0].ID != "ORPHAN" {
		t.Errorf("req[0] ID = %q, want ORPHAN", reqs[0].ID)
	}
	if reqs[0].ParentID != "" {
		t.Errorf("req[0] ParentID = %q, want empty (dynamic reqLevel = 3)", reqs[0].ParentID)
	}
}

// TestParseMD_SubHeadingInBody is preserved: a ### heading without an attr
// block remains body text (backward compatible).
func TestParseMD_SubHeadingInBody(t *testing.T) {
	// This test is the same as the existing one but confirms backward compat
	// with the new parser: ### without attr → body text, not a sub-req.
	src := []byte("## REQ-001\n```attr\nid: REQ-001\n```\n### Sub Heading\nSome text.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	if r.ParentID != "" {
		t.Errorf("req ParentID = %q, want empty (should not be a sub-req)", r.ParentID)
	}
	expected := "### Sub Heading\n\n\n\nSome text."
	if r.Body != expected {
		t.Errorf("Body = %q, want %q", r.Body, expected)
	}
}

// TestParseMD_DynamicReqLevel verifies that the first heading level with an
// attr block defines the requirement level (e.g., # instead of ##).
func TestParseMD_DynamicReqLevel(t *testing.T) {
	src := []byte("# TOP\n```attr\nid: TOP\n```\nTop body.\n\n## CHILD\n```attr\nid: CHILD\n```\nChild body.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(reqs))
	}
	if reqs[0].ID != "TOP" {
		t.Errorf("req[0] ID = %q, want TOP", reqs[0].ID)
	}
	if reqs[0].ParentID != "" {
		t.Errorf("req[0] ParentID = %q, want empty", reqs[0].ParentID)
	}
	if reqs[1].ID != "CHILD" {
		t.Errorf("req[1] ID = %q, want CHILD", reqs[1].ID)
	}
	if reqs[1].ParentID != "TOP" {
		t.Errorf("req[1] ParentID = %q, want TOP", reqs[1].ParentID)
	}
}

// TestParseSingleFile verifies ParseSingleFile reads a real file from disk
// and produces the correct results. Uses t.TempDir() for the temp file.
func TestParseSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_req.md")
	content := []byte("## REQ-FILE\n```attr\nid: REQ-FILE\nasil: ASIL C\n```\nThis requirement was loaded from a file.\n\n*Rationale: file-based parsing\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	reqs, err := ParseSingleFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	if r.ID != "REQ-FILE" {
		t.Errorf("ID = %q, want REQ-FILE", r.ID)
	}
	if r.Attrs["id"] != "REQ-FILE" {
		t.Errorf("attrs[id] = %v, want REQ-FILE", r.Attrs["id"])
	}
	if r.Attrs["asil"] != "ASIL C" {
		t.Errorf("attrs[asil] = %v, want ASIL C", r.Attrs["asil"])
	}
	if r.Body != "This requirement was loaded from a file." {
		t.Errorf("Body = %q, want %q", r.Body, "This requirement was loaded from a file.")
	}
	if r.Rationale != "file-based parsing" {
		t.Errorf("Rational = %q, want %q", r.Rationale, "file-based parsing")
	}
	if r.Source != path {
		t.Errorf("Source = %q, want %q", r.Source, path)
	}
}

// TestParseMD_Title verifies that a heading with "ID: Title" format
// splits the ID and Title correctly.
func TestParseMD_Title(t *testing.T) {
	src := []byte("## REQ-001: Login requirement\n```attr\nstatus: Approved\n```\nBody text.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	if r.ID != "REQ-001" {
		t.Errorf("ID = %q, want REQ-001", r.ID)
	}
	if r.Title != "Login requirement" {
		t.Errorf("Title = %q, want %q", r.Title, "Login requirement")
	}
}

// TestParseMD_NoTitle verifies that a heading without ": " has empty Title.
func TestParseMD_NoTitle(t *testing.T) {
	src := []byte("## REQ-002\n```attr\nstatus: Draft\n```\nNo title.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	if r.ID != "REQ-002" {
		t.Errorf("ID = %q, want REQ-002", r.ID)
	}
	if r.Title != "" {
		t.Errorf("Title = %q, want empty string", r.Title)
	}
}

// TestParseMD_TitleWithColonInID verifies that only the first ": " splits
// (IDs can't contain colons per the ID pattern, but titles may).
func TestParseMD_TitleWithColonInTitle(t *testing.T) {
	src := []byte("## REQ-003: Do this: then that\n```attr\nstatus: Approved\n```\nBody.\n")
	reqs, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	r := reqs[0]
	if r.ID != "REQ-003" {
		t.Errorf("ID = %q, want REQ-003", r.ID)
	}
	if r.Title != "Do this: then that" {
		t.Errorf("Title = %q, want %q", r.Title, "Do this: then that")
	}
}

func TestParseLsTreeOutput_Empty(t *testing.T) {
	result := parseLsTreeOutput("")
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestParseLsTreeOutput_NoSubmodules(t *testing.T) {
	output := `100644 blob abc123def4567890123456789012345678901234	README.md
100644 blob def4567890123456789012345678901234567890	go.mod`
	result := parseLsTreeOutput(output)
	if len(result) != 0 {
		t.Errorf("expected empty map (no submodules), got %d entries", len(result))
	}
}

func TestParseLsTreeOutput_OneSubmodule(t *testing.T) {
	output := `100644 blob abc123def4567890123456789012345678901234	README.md
160000 commit def4567890123456789012345678901234567890	vendor/spec-a
100644 blob 789012345678901234567890123456789012345a	go.mod`
	result := parseLsTreeOutput(output)
	if len(result) != 1 {
		t.Fatalf("expected 1 submodule, got %d", len(result))
	}
	if result["vendor/spec-a"] != "def4567890123456789012345678901234567890" {
		t.Errorf("unexpected SHA: %s", result["vendor/spec-a"])
	}
}

func TestParseLsTreeOutput_MultipleSubmodules(t *testing.T) {
	output := `160000 commit aaa1111111111111111111111111111111111111	vendor/a
160000 commit bbb2222222222222222222222222222222222222	vendor/b
160000 commit ccc3333333333333333333333333333333333333	vendor/c
100644 blob ddd4444444444444444444444444444444444444	README.md`
	result := parseLsTreeOutput(output)
	if len(result) != 3 {
		t.Fatalf("expected 3 submodules, got %d", len(result))
	}
	if result["vendor/a"] != "aaa1111111111111111111111111111111111111" {
		t.Errorf("vendor/a: unexpected SHA")
	}
	if result["vendor/b"] != "bbb2222222222222222222222222222222222222" {
		t.Errorf("vendor/b: unexpected SHA")
	}
	if result["vendor/c"] != "ccc3333333333333333333333333333333333333" {
		t.Errorf("vendor/c: unexpected SHA")
	}
}

func TestParseLsTreeOutput_Malformed(t *testing.T) {
	output := `garbage line without tab
160000 commit abc123	invalid
not enough fields here
`
	result := parseLsTreeOutput(output)
	// "garbage line without tab" has no tab → skipped
	// "160000 commit abc123\tinvalid" → valid submodule entry
	// "not enough fields here" → no tab → skipped
	if len(result) != 1 {
		t.Errorf("expected 1 valid entry, got %d", len(result))
	}
	if result["invalid"] != "abc123" {
		t.Errorf("unexpected SHA: %s", result["invalid"])
	}
}
