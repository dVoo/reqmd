package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/internal/model"
)

// parseReqTree is a test helper: parses src and returns the assembled
// content tree roots.
func parseReqTree(t *testing.T, src []byte) []*model.Node {
	t.Helper()
	nodes, _, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	return nodes
}

// reqsOf returns the flat requirement list of an assembled tree.
func reqsOf(nodes []*model.Node) []*model.Node {
	return model.CollectRequirements(nodes)
}

// TestParseMD_Full verifies a normal requirement with heading, attr block,
// prose body, and *Rationale: line.
func TestParseMD_Full(t *testing.T) {
	src := []byte("## REQ-001\n```attr\nid: REQ-001\nasil: ASIL D\n```\nThis requirement shall do something important.\n\n*Rationale: safety critical function\n")
	nodes := parseReqTree(t, src)
	reqs := reqsOf(nodes)
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
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	if reqs[0].Rationale != "" {
		t.Errorf("Rational = %q, want empty string", reqs[0].Rationale)
	}
}

// TestParseMD_MultipleRequirements verifies that multiple requirements
// in one document are each parsed correctly.
func TestParseMD_MultipleRequirements(t *testing.T) {
	src := []byte("## REQ-A\n```attr\nid: REQ-A\n```\nBody A.\n\n## REQ-B\n```attr\nid: REQ-B\n```\nBody B.\n\n## REQ-C\n```attr\nid: REQ-C\n```\nBody C.\n")
	reqs := reqsOf(parseReqTree(t, src))
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

// TestParseMD_HeadingWithoutAttrIsInfo verifies that a heading without an
// attr block produces an info item instead of a requirement.
func TestParseMD_HeadingWithoutAttrIsInfo(t *testing.T) {
	src := []byte("## Section Title\nThis has no attr block.\n")
	nodes := parseReqTree(t, src)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	n := nodes[0]
	if n.Kind != model.KindSection {
		t.Errorf("Kind = %v, want KindSection", n.Kind)
	}
	if n.Title != "Section Title" {
		t.Errorf("Title = %q, want %q", n.Title, "Section Title")
	}
	if n.Body != "This has no attr block." {
		t.Errorf("Body = %q, want %q", n.Body, "This has no attr block.")
	}
	if reqs := reqsOf(nodes); len(reqs) != 0 {
		t.Fatalf("expected 0 requirements, got %d", len(reqs))
	}
}

// TestParseMD_BadYAML verifies that invalid YAML inside an ```attr``` block
// returns an error.
func TestParseMD_BadYAML(t *testing.T) {
	src := []byte("## REQ-BAD\n```attr\n: : invalid yaml\n```\n")
	nodes, _, _, err := parseMD(src, "test.md")
	if err == nil {
		t.Fatal("expected error for invalid attr YAML, got nil")
	}
	if nodes != nil {
		t.Fatalf("expected nil nodes on error, got %d nodes", len(nodes))
	}
}

// TestParseMD_NonAttrCodeBlock verifies that a fenced code block with a
// non-attr info string (e.g. ```json) inside a requirement body is preserved.
func TestParseMD_NonAttrCodeBlock(t *testing.T) {
	src := []byte("## REQ-001\n```attr\nid: REQ-001\n```\nSome text.\n\n```json\n{\"key\": \"value\"}\n```\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	// The blank line matters: without it the fence glues onto the paragraph and
	// stops being a fenced block when the body is rendered again.
	expected := "Some text.\n\n```json\n{\"key\": \"value\"}\n```\n\n"
	if reqs[0].Body != expected {
		t.Errorf("Body = %q, want %q", reqs[0].Body, expected)
	}
}

// TestParseMD_EmptyFile verifies that parsing an empty source returns
// no nodes and no error.
func TestParseMD_EmptyFile(t *testing.T) {
	nodes, _, _, err := parseMD([]byte{}, "empty.md")
	if err != nil {
		t.Fatal(err)
	}
	if nodes != nil {
		t.Fatalf("expected nil nodes for empty file, got %d", len(nodes))
	}
}

// TestParseMD_ThematicBreak verifies that thematic breaks (---) inside
// a requirement body are preserved and don't disrupt body accumulation.
func TestParseMD_ThematicBreak(t *testing.T) {
	src := []byte("## REQ-001\n```attr\nid: REQ-001\n```\nBefore break.\n\n---\nAfter break.\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	expected := "Before break.\n\n---\n\nAfter break."
	if reqs[0].Body != expected {
		t.Errorf("Body = %q, want %q", reqs[0].Body, expected)
	}
}

// TestParseMD_ArrayAttr verifies that an attribute with a YAML array value
// is parsed into []interface{} with the correct elements.
func TestParseMD_ArrayAttr(t *testing.T) {
	src := []byte("## REQ-001\n```attr\ntrace: [SYS-001, SAFE-003]\n```\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	trace, ok := reqs[0].Attrs["trace"].([]any)
	if !ok {
		t.Fatalf("attrs[trace] is %T, want []any", reqs[0].Attrs["trace"])
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
	reqs := reqsOf(parseReqTree(t, src))
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
	reqs := reqsOf(parseReqTree(t, src))
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
// as the FIRST requirement is a top-level requirement, not an orphan child.
func TestParseMD_SubReqDynamicLevel(t *testing.T) {
	src := []byte("### ORPHAN\n```attr\nid: ORPHAN\n```\nNo parent.\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	if reqs[0].ID != "ORPHAN" {
		t.Errorf("req[0] ID = %q, want ORPHAN", reqs[0].ID)
	}
	if reqs[0].ParentID != "" {
		t.Errorf("req[0] ParentID = %q, want empty (first requirement is top-level)", reqs[0].ParentID)
	}
}

// TestParseMD_SubHeadingBecomesInfo verifies that a ### heading without an
// attr block under a requirement becomes an info item child, not body text.
func TestParseMD_SubHeadingBecomesInfo(t *testing.T) {
	src := []byte("## REQ-001\n```attr\nid: REQ-001\n```\n### Sub Heading\nSome text.\n")
	nodes := parseReqTree(t, src)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(nodes))
	}
	req := nodes[0]
	if req.Kind != model.KindRequirement || req.ID != "REQ-001" {
		t.Fatalf("root = %v %q, want requirement REQ-001", req.Kind, req.ID)
	}
	if len(req.Children) != 1 {
		t.Fatalf("expected 1 child info node, got %d", len(req.Children))
	}
	info := req.Children[0]
	if info.Kind != model.KindSection {
		t.Errorf("child Kind = %v, want KindSection", info.Kind)
	}
	if info.Title != "Sub Heading" {
		t.Errorf("child Title = %q, want %q", info.Title, "Sub Heading")
	}
	if info.Body != "Some text." {
		t.Errorf("child Body = %q, want %q", info.Body, "Some text.")
	}
	if req.Body != "" {
		t.Errorf("req Body = %q, want empty (sub-heading no longer body text)", req.Body)
	}
}

// TestParseMD_DynamicReqLevel verifies that requirements chain by heading
// level when the content tree starts at level 2 (h1 is reserved for the
// document title).
func TestParseMD_DynamicReqLevel(t *testing.T) {
	src := []byte("## TOP\n```attr\nid: TOP\n```\nTop body.\n\n### CHILD\n```attr\nid: CHILD\n```\nChild body.\n")
	reqs := reqsOf(parseReqTree(t, src))
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

// TestParseMD_H1IsDocumentTitle verifies that a level-1 heading is treated
// as the document title: an h1 with an attr block is rejected (requirements
// must be level 2+), and a plain h1 is captured as the title and never
// becomes a content node.
func TestParseMD_H1IsDocumentTitle(t *testing.T) {
	// h1 + attr block → error.
	_, _, _, err := parseMD([]byte("# TOP\n```attr\nid: TOP\n```\nBody.\n"), "test.md")
	if err == nil {
		t.Fatal("expected error for level-1 requirement, got nil")
	}
	if !strings.Contains(err.Error(), "level 2 or higher") {
		t.Errorf("error = %q, want mention of level 2+", err)
	}

	// Plain h1 → title; no node is created for it.
	src := []byte("# Widget Spec\n\nIntro prose.\n\n## REQ-001\n```attr\nid: REQ-001\n```\nBody.\n")
	nodes, title, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Widget Spec" {
		t.Errorf("title = %q, want %q", title, "Widget Spec")
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 roots (intro info + req), got %d", len(nodes))
	}
	if nodes[0].Kind != model.KindSection || nodes[0].Title != "" || nodes[0].Body != "Intro prose." {
		t.Errorf("nodes[0] = %+v, want leading info node with prose", nodes[0])
	}
	if nodes[1].Kind != model.KindRequirement || nodes[1].ID != "REQ-001" || nodes[1].ParentID != "" {
		t.Errorf("nodes[1] = %v %q, want top-level requirement REQ-001", nodes[1].Kind, nodes[1].ID)
	}
}

// TestParseMD_SecondH1Rejected verifies that a second level-1 heading is a
// parse error. Before this rule the heading text was silently dropped: only
// the first h1 became the title, and the second one's prose was re-attached
// to a fresh headingless node.
func TestParseMD_SecondH1Rejected(t *testing.T) {
	src := []byte("# First\n\nIntro.\n\n## REQ-001\n```attr\nid: REQ-001\n```\nBody.\n\n# Second\n\nMore.\n")
	_, _, _, err := parseMD(src, "test.md")
	if err == nil {
		t.Fatal("expected error for a second level-1 heading, got nil")
	}
	if !strings.Contains(err.Error(), "second level-1 heading") {
		t.Errorf("error = %q, want mention of a second level-1 heading", err)
	}
	// An h1 that is not the first heading but the only one is rejected too.
	src = []byte("## REQ-001\n```attr\nid: REQ-001\n```\n\n# Late title\n")
	if _, _, _, err := parseMD(src, "test.md"); err == nil {
		t.Error("expected error for an h1 after content, got nil")
	}
}

// TestParseSingleFile verifies ParseSingleFile reads a real file from disk
// and produces the correct results. Uses t.TempDir() for the temp file.
func TestParseSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_req.md")
	content := []byte("## REQ-FILE\n```attr\nid: REQ-FILE\nasil: ASIL C\n```\nThis requirement was loaded from a file.\n\n*Rationale: file-based parsing\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	nodes, err := ParseSingleFile(path)
	if err != nil {
		t.Fatal(err)
	}
	reqs := reqsOf(nodes)
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
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	if reqs[0].ID != "REQ-001" {
		t.Errorf("ID = %q, want REQ-001", reqs[0].ID)
	}
	if reqs[0].Title != "Login requirement" {
		t.Errorf("Title = %q, want %q", reqs[0].Title, "Login requirement")
	}
}

// TestParseMD_NoTitle verifies that a heading without ": " has empty Title.
func TestParseMD_NoTitle(t *testing.T) {
	src := []byte("## REQ-002\n```attr\nstatus: Draft\n```\nNo title.\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	if reqs[0].ID != "REQ-002" {
		t.Errorf("ID = %q, want REQ-002", reqs[0].ID)
	}
	if reqs[0].Title != "" {
		t.Errorf("Title = %q, want empty string", reqs[0].Title)
	}
}

// TestParseMD_TitleWithColonInTitle verifies that only the first ": " splits
// (IDs can't contain colons per the ID pattern, but titles may).
func TestParseMD_TitleWithColonInTitle(t *testing.T) {
	src := []byte("## REQ-003: Do this: then that\n```attr\nstatus: Approved\n```\nBody.\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	if reqs[0].ID != "REQ-003" {
		t.Errorf("ID = %q, want REQ-003", reqs[0].ID)
	}
	if reqs[0].Title != "Do this: then that" {
		t.Errorf("Title = %q, want %q", reqs[0].Title, "Do this: then that")
	}
}

// TestParseMD_ContainerHoldsRequirements verifies that a container heading
// (no attr block) becomes a KindSection node that holds the requirements
// nested under it, which stay top-level (no requirement parent).
func TestParseMD_ContainerHoldsRequirements(t *testing.T) {
	src := []byte("## REQ-1\n```attr\nid: REQ-1\n```\nBody 1.\n\n## Container\n\n### REQ-2\n```attr\nid: REQ-2\n```\nBody 2.\n")
	nodes := parseReqTree(t, src)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 root nodes (REQ-1 + Container), got %d", len(nodes))
	}
	if nodes[0].Kind != model.KindRequirement || nodes[0].ID != "REQ-1" {
		t.Errorf("nodes[0] = %v %q, want requirement REQ-1", nodes[0].Kind, nodes[0].ID)
	}
	container := nodes[1]
	if container.Kind != model.KindSection {
		t.Errorf("nodes[1] Kind = %v, want KindSection", container.Kind)
	}
	if container.Title != "Container" {
		t.Errorf("container Title = %q, want %q", container.Title, "Container")
	}
	if len(container.Children) != 1 {
		t.Fatalf("container children = %d, want 1", len(container.Children))
	}
	child := container.Children[0]
	if child.ID != "REQ-2" {
		t.Errorf("container child ID = %q, want REQ-2", child.ID)
	}
	if child.ParentID != "" {
		t.Errorf("REQ-2 ParentID = %q, want empty (container is not a requirement)", child.ParentID)
	}
	if child.Body != "Body 2." {
		t.Errorf("REQ-2 Body = %q, want %q", child.Body, "Body 2.")
	}
}

// TestParseMD_UserExample1 reproduces the scenario the user described:
// a container at the same level breaks the requirement chain; requirements
// nest under containers and stay top-level.
func TestParseMD_UserExample1(t *testing.T) {
	src := []byte("## REQ-1\n```attr\nid: REQ-1\n```\n## Container\n\n### REQ-2\n```attr\nid: REQ-2\n```\n## REQ-3\n```attr\nid: REQ-3\n```\n## Another Container\n\n### Another Container 2\n\n#### REQ-4\n```attr\nid: REQ-4\n```\n")
	nodes := parseReqTree(t, src)
	reqs := reqsOf(nodes)
	if len(reqs) != 4 {
		t.Fatalf("expected 4 requirements, got %d", len(reqs))
	}
	for i, want := range []string{"REQ-1", "REQ-2", "REQ-3", "REQ-4"} {
		if reqs[i].ID != want {
			t.Errorf("req[%d]: ID=%q, want %q", i, reqs[i].ID, want)
		}
		if reqs[i].ParentID != "" {
			t.Errorf("req[%d]: ParentID=%q, want empty (all top-level)", i, reqs[i].ParentID)
		}
	}
	// Containers: "Container" holds REQ-2; "Another Container" holds
	// "Another Container 2" which holds REQ-4.
	var findContainer func(nodes []*model.Node, title string) *model.Node
	findContainer = func(nodes []*model.Node, title string) *model.Node {
		for _, n := range nodes {
			if n.Kind == model.KindSection && n.Title == title {
				return n
			}
			if c := findContainer(n.Children, title); c != nil {
				return c
			}
		}
		return nil
	}
	outer := findContainer(nodes, "Container")
	if outer == nil || len(outer.Children) != 1 || outer.Children[0].ID != "REQ-2" {
		t.Error("Container should hold REQ-2")
	}
	mid := findContainer(nodes, "Another Container 2")
	if mid == nil || len(mid.Children) != 1 || mid.Children[0].ID != "REQ-4" {
		t.Error("Another Container 2 should hold REQ-4")
	}
	if c := findContainer(nodes, "Another Container"); c == nil || len(c.Children) != 1 || c.Children[0] != mid {
		t.Error("Another Container should hold Another Container 2")
	}
}

// TestParseMD_DeepNesting verifies requirements at arbitrary depth
// (3+ levels) chain correctly when no containers intervene.
func TestParseMD_DeepNesting(t *testing.T) {
	src := []byte("## R1\n```attr\nid: R1\n```\n### R1.1\n```attr\nid: R1.1\n```\n#### R1.1.1\n```attr\nid: R1.1.1\n```\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 3 {
		t.Fatalf("expected 3 requirements, got %d", len(reqs))
	}
	if reqs[0].ID != "R1" || reqs[0].ParentID != "" {
		t.Errorf("req[0]: ID=%q ParentID=%q, want R1/empty", reqs[0].ID, reqs[0].ParentID)
	}
	if reqs[1].ID != "R1.1" || reqs[1].ParentID != "R1" {
		t.Errorf("req[1]: ID=%q ParentID=%q, want R1.1/R1", reqs[1].ID, reqs[1].ParentID)
	}
	if reqs[2].ID != "R1.1.1" || reqs[2].ParentID != "R1.1" {
		t.Errorf("req[2]: ID=%q ParentID=%q, want R1.1.1/R1.1", reqs[2].ID, reqs[2].ParentID)
	}
}

// TestParseMD_ContainerDoesNotBreakDeeperChain verifies that a container
// heading deeper than the current requirement does NOT break the chain:
// requirements after it still link to the nearest shallower requirement.
func TestParseMD_ContainerDoesNotBreakDeeperChain(t *testing.T) {
	// ### Section is deeper than ## R1 → R1 keeps it as an info child.
	// #### R2 is deeper than R1 → parent=R1 (chain held).
	// ##### Info is deeper than R2 → info child of R2, chain holds.
	// ##### R2.1 is deeper than R2 → parent=R2 (chain held).
	src := []byte("## R1\n```attr\nid: R1\n```\n### Section\n\n#### R2\n```attr\nid: R2\n```\n##### Info\n\n##### R2.1\n```attr\nid: R2.1\n```\n")
	nodes := parseReqTree(t, src)
	reqs := reqsOf(nodes)
	if len(reqs) != 3 {
		t.Fatalf("expected 3 requirements, got %d", len(reqs))
	}
	if reqs[0].ID != "R1" || reqs[0].ParentID != "" {
		t.Errorf("req[0]: ID=%q ParentID=%q, want R1/empty", reqs[0].ID, reqs[0].ParentID)
	}
	if reqs[1].ID != "R2" || reqs[1].ParentID != "R1" {
		t.Errorf("req[1]: ID=%q ParentID=%q, want R2/R1 (## Section is deeper, doesn't break)", reqs[1].ID, reqs[1].ParentID)
	}
	if reqs[2].ID != "R2.1" || reqs[2].ParentID != "R2" {
		t.Errorf("req[2]: ID=%q ParentID=%q, want R2.1/R2 (#### Info is deeper, doesn't break)", reqs[2].ID, reqs[2].ParentID)
	}
	// Tree: R1 → Section(container) → R2 → [Info, R2.1]
	if len(nodes) != 1 {
		t.Fatalf("expected 1 root, got %d", len(nodes))
	}
	r1 := nodes[0]
	if len(r1.Children) != 1 || r1.Children[0].Kind != model.KindSection || r1.Children[0].Title != "Section" {
		t.Fatalf("R1 children = %+v, want [container Section]", r1.Children)
	}
	section := r1.Children[0]
	if len(section.Children) != 1 || section.Children[0].ID != "R2" {
		t.Fatalf("Section children = %+v, want [req R2]", section.Children)
	}
	r2 := section.Children[0]
	if len(r2.Children) != 2 {
		t.Fatalf("R2 children = %d, want 2", len(r2.Children))
	}
	if r2.Children[0].Kind != model.KindSection || r2.Children[0].Title != "Info" {
		t.Errorf("R2 children[0] = %v %q, want info Info", r2.Children[0].Kind, r2.Children[0].Title)
	}
	if r2.Children[1].Kind != model.KindRequirement || r2.Children[1].ID != "R2.1" {
		t.Errorf("R2 children[1] = %v %q, want req R2.1", r2.Children[1].Kind, r2.Children[1].ID)
	}
}

// TestParseMD_LevelGapNoContainer verifies that a heading gap without a
// container still links children to the nearest shallower requirement.
func TestParseMD_LevelGapNoContainer(t *testing.T) {
	src := []byte("## R-A\n```attr\nid: R-A\n```\n#### R-B\n```attr\nid: R-B\n```\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(reqs))
	}
	if reqs[0].ID != "R-A" || reqs[0].ParentID != "" {
		t.Errorf("req[0]: ID=%q ParentID=%q, want R-A/empty", reqs[0].ID, reqs[0].ParentID)
	}
	if reqs[1].ID != "R-B" || reqs[1].ParentID != "R-A" {
		t.Errorf("req[1]: ID=%q ParentID=%q, want R-B/R-A (nearest shallower, no container)", reqs[1].ID, reqs[1].ParentID)
	}
}

// TestParseMD_SiblingContainersBreakChain verifies that multiple containers
// at successive levels produce only top-level requirements.
func TestParseMD_SiblingContainersBreakChain(t *testing.T) {
	src := []byte("## R1\n```attr\nid: R1\n```\n## C1\n\n## C2\n\n#### R2\n```attr\nid: R2\n```\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(reqs))
	}
	if reqs[0].ID != "R1" || reqs[0].ParentID != "" {
		t.Errorf("req[0]: ID=%q ParentID=%q, want R1/empty", reqs[0].ID, reqs[0].ParentID)
	}
	if reqs[1].ID != "R2" || reqs[1].ParentID != "" {
		t.Errorf("req[1]: ID=%q ParentID=%q, want R2/empty (all containers broke chain)", reqs[1].ID, reqs[1].ParentID)
	}
}

// TestParseMD_Preamble verifies that leading prose before the first heading
// is captured as an info node and not dropped, and that an h1 title does not
// become a wrapping container.
func TestParseMD_Preamble(t *testing.T) {
	src := []byte("This document describes the widget.\n\n# Widget Spec\n\n## REQ-001\n```attr\nid: REQ-001\n```\nBody.\n")
	nodes, title, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Widget Spec" {
		t.Errorf("title = %q, want %q", title, "Widget Spec")
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 root nodes (preamble + requirement), got %d", len(nodes))
	}
	preamble := nodes[0]
	if preamble.Kind != model.KindSection {
		t.Errorf("preamble Kind = %v, want KindSection", preamble.Kind)
	}
	if preamble.Title != "" {
		t.Errorf("preamble Title = %q, want empty", preamble.Title)
	}
	if preamble.Level != 0 {
		t.Errorf("preamble Level = %d, want 0", preamble.Level)
	}
	if preamble.Body != "This document describes the widget." {
		t.Errorf("preamble Body = %q, want %q", preamble.Body, "This document describes the widget.")
	}
	req := nodes[1]
	if req.Kind != model.KindRequirement || req.ID != "REQ-001" || req.ParentID != "" {
		t.Errorf("nodes[1] = %v %q, want top-level requirement REQ-001", req.Kind, req.ID)
	}
	if reqs := reqsOf(nodes); len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
}

// TestParseMD_RequirementlessFile verifies that a file with no requirement
// headings at all is captured entirely as info nodes, with the h1 title kept
// out of the content tree.
func TestParseMD_RequirementlessFile(t *testing.T) {
	src := []byte("# Overview\n\nJust some background.\n\n## Details\n\nMore prose.\n")
	nodes, title, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Overview" {
		t.Errorf("title = %q, want %q", title, "Overview")
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(nodes))
	}
	intro := nodes[0]
	if intro.Kind != model.KindSection || intro.Title != "" || intro.Body != "Just some background." {
		t.Errorf("nodes[0] = %v %q/%q, want leading info/background", intro.Kind, intro.Title, intro.Body)
	}
	details := nodes[1]
	if details.Kind != model.KindSection || details.Title != "Details" || details.Body != "More prose." {
		t.Errorf("nodes[1] = %v %q/%q, want info Details/More prose.", details.Kind, details.Title, details.Body)
	}
	if reqs := reqsOf(nodes); len(reqs) != 0 {
		t.Fatalf("expected 0 requirements, got %d", len(reqs))
	}
}

// TestParseMD_ChunkedPreamble verifies that the parallel chunked path
// (multiple root requirements) still captures the file preamble and title.
func TestParseMD_ChunkedPreamble(t *testing.T) {
	src := []byte("Intro prose.\n\n# Widget Spec\n\n## REQ-001\n```attr\nid: REQ-001\n```\nBody 1.\n\n## REQ-002\n```attr\nid: REQ-002\n```\nBody 2.\n\n## REQ-003\n```attr\nid: REQ-003\n```\nBody 3.\n")
	nodes, title, _, err := parseMD(src, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Widget Spec" {
		t.Errorf("title = %q, want %q", title, "Widget Spec")
	}
	reqs := reqsOf(nodes)
	if len(reqs) != 3 {
		t.Fatalf("expected 3 requirements, got %d", len(reqs))
	}
	for i, want := range []string{"REQ-001", "REQ-002", "REQ-003"} {
		if reqs[i].ID != want {
			t.Errorf("req[%d] ID = %q, want %q", i, reqs[i].ID, want)
		}
	}
	// Preamble prose survives the chunked path; the h1 title is not a node.
	if len(nodes) != 4 || nodes[0].Kind != model.KindSection || nodes[0].Body != "Intro prose." {
		t.Errorf("preamble not preserved: %+v", nodes)
	}
}

// TestParseMD_LevelGapUnderContainer verifies that a requirement inside a
// container under another requirement gets the right ParentID (nearest
// shallower requirement) and container nesting.
func TestParseMD_LevelGapUnderContainer(t *testing.T) {
	src := []byte("## REQ-A\n```attr\nid: REQ-A\n```\n### Section\n\n#### REQ-B\n```attr\nid: REQ-B\n```\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(reqs))
	}
	if reqs[0].ID != "REQ-A" || reqs[0].ParentID != "" {
		t.Errorf("req[0]: ID=%q ParentID=%q, want REQ-A/empty", reqs[0].ID, reqs[0].ParentID)
	}
	if reqs[1].ID != "REQ-B" || reqs[1].ParentID != "REQ-A" {
		t.Errorf("req[1]: ID=%q ParentID=%q, want REQ-B/REQ-A", reqs[1].ID, reqs[1].ParentID)
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

// bigDoc builds a document large enough to take the chunked parse path: a
// title, preamble prose, a root container holding requirements, then many root
// requirements. Its size matters — parseMD only splits above
// parseParallelThreshold.
func bigDoc(t *testing.T, roots int) []byte {
	t.Helper()
	var b strings.Builder
	b.WriteString("# Big Spec\n\nIntro prose that must survive chunking.\n\n")
	b.WriteString("## Chapter\n\nContainer prose.\n\n")
	for _, id := range []string{"BIG-0001", "BIG-0002"} {
		b.WriteString("### " + id + ": In chapter\n\n```attr\nstatus: approved\n```\n\nBody of " + id + ".\n\n")
	}
	for i := 0; i < roots; i++ {
		id := fmt.Sprintf("BIG-%04d", i+100)
		b.WriteString("## " + id + ": Root requirement\n\n```attr\nstatus: approved\n```\n\n")
		b.WriteString(strings.Repeat("Filler prose that pads the file past the parallel threshold. ", 12))
		b.WriteString("\n\n")
	}
	src := []byte(b.String())
	if len(src) < parseParallelThreshold {
		t.Fatalf("fixture is %d bytes, below the %d-byte parallel threshold", len(src), parseParallelThreshold)
	}
	return src
}

// nodeSummary renders a tree into a comparable string so the chunked result can
// be compared with the serial one without walking structs field by field.
func nodeSummary(nodes []*model.Node, depth int) string {
	var b strings.Builder
	for _, n := range nodes {
		fmt.Fprintf(&b, "%*s%s|%s|%s|L%d\n", depth*2, "", n.Kind.String(), n.ID, n.Title, n.Level)
		b.WriteString(nodeSummary(n.Children, depth+1))
	}
	return b.String()
}

// TestParseMD_ChunkedMatchesSerial is the invariant behind the parallel path:
// batching roots into chunks must produce exactly the tree the single-pass
// parser produces, including the preamble (chunk 0 covers [0, first root)).
func TestParseMD_ChunkedMatchesSerial(t *testing.T) {
	src := bigDoc(t, 400)

	chunked, title, _, err := parseMD(src, "big.md")
	if err != nil {
		t.Fatalf("parseMD (chunked): %v", err)
	}
	// The serial reference: one parse of the whole file, assembled the same way.
	flat, serialTitle, _, err := parseMDChunk(src, "big.md")
	if err != nil {
		t.Fatalf("parseMDChunk (serial reference): %v", err)
	}
	serial := assembleTree(flat)

	if title != serialTitle {
		t.Errorf("title = %q, want %q", title, serialTitle)
	}
	if got, want := nodeSummary(chunked, 0), nodeSummary(serial, 0); got != want {
		t.Fatalf("chunked tree differs from serial tree\n--- chunked ---\n%s\n--- serial ---\n%s", got, want)
	}
}

// TestParseMD_ChunkedKeepsPreambleAndBoundaries checks the two properties the
// chunking exists for: leading prose and the container it opens must not be
// dropped, and no requirement may be lost or duplicated at a chunk boundary.
func TestParseMD_ChunkedKeepsPreambleAndBoundaries(t *testing.T) {
	const roots = 400
	src := bigDoc(t, roots)

	nodes, title, _, err := parseMD(src, "big.md")
	if err != nil {
		t.Fatalf("parseMD: %v", err)
	}
	if title != "Big Spec" {
		t.Errorf("document title = %q, want %q (chunk 0 must include the preamble)", title, "Big Spec")
	}

	// Root 0 is the preamble info node (leading prose), root 1 the container.
	if len(nodes) < 2 {
		t.Fatalf("roots = %d, want at least preamble + chapter", len(nodes))
	}
	if nodes[0].Kind != model.KindSection || nodes[0].Level != 0 || !strings.Contains(nodes[0].Body, "Intro prose") {
		t.Errorf("root 0 = %+v, want the headingless preamble", nodes[0])
	}
	chapter := nodes[1]
	if chapter.Title != "Chapter" || len(chapter.Children) != 2 {
		t.Fatalf("root 1 = %+v, want Chapter holding 2 requirements", chapter)
	}

	// Every root requirement must be present exactly once, in document order.
	var want []string
	for i := 0; i < roots; i++ {
		want = append(want, fmt.Sprintf("BIG-%04d", i+100))
	}
	var got []string
	for _, r := range model.CollectRequirements(nodes) {
		got = append(got, r.ID)
	}
	if len(got) != roots+2 {
		t.Errorf("requirements = %d, want %d", len(got), roots+2)
	}
	tail := got[2:] // skip the two chapter children
	if strings.Join(tail, ",") != strings.Join(want, ",") {
		t.Errorf("root requirement order/content changed by chunking:\n got %v\nwant %v", tail, want)
	}
}

// TestParseMD_WrappedParagraphKeepsLineBreaks pins that a source-wrapped
// paragraph keeps its line breaks in the stored body. Dropping them glued the
// last word of one line onto the first word of the next in the HTML export.
func TestParseMD_WrappedParagraphKeepsLineBreaks(t *testing.T) {
	src := []byte("# T\n\n## X-1: t\n\n```attr\nstatus: approved\n```\n\n" +
		"A flat list\nof top-level symbols.\n")
	nodes, _, _, err := parseMD(src, "wrap.md")
	if err != nil {
		t.Fatal(err)
	}
	reqs := model.CollectRequirements(assembleTree(nodes))
	if len(reqs) != 1 {
		t.Fatalf("requirements = %d, want 1", len(reqs))
	}
	if got, want := reqs[0].Body, "A flat list\nof top-level symbols."; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

// TestParseMD_ListBodyPreserved pins the body-capture contract for blocks that
// have no dedicated case: a list in a requirement body must keep its markers and
// its line breaks, and must be separated from the prose above it. Concatenating
// the AST line segments dropped both, which turned every bullet list into one
// run-on paragraph in the HTML export.
func TestParseMD_ListBodyPreserved(t *testing.T) {
	src := []byte("# T\n\n## X-1: t\n\n```attr\nstatus: approved\n```\n\n" +
		"Para intro:\n\n- one\n- two\n- three\n\nAfter.\n")
	nodes, _, _, err := parseMD(src, "list.md")
	if err != nil {
		t.Fatal(err)
	}
	reqs := model.CollectRequirements(assembleTree(nodes))
	if len(reqs) != 1 {
		t.Fatalf("requirements = %d, want 1", len(reqs))
	}
	want := "Para intro:\n\n- one\n- two\n- three\n\nAfter."
	if got := reqs[0].Body; got != want {
		t.Errorf("body =\n%q\nwant\n%q", got, want)
	}
}

// TestParseMD_TableBodyPreserved covers the other block that used to be
// flattened: a GFM table must survive as markdown, rows intact.
func TestParseMD_TableBodyPreserved(t *testing.T) {
	src := []byte("# T\n\n## X-1: t\n\n```attr\nstatus: approved\n```\n\n" +
		"| a | b |\n| --- | --- |\n| 1 | 2 |\n")
	nodes, _, _, err := parseMD(src, "table.md")
	if err != nil {
		t.Fatal(err)
	}
	reqs := model.CollectRequirements(assembleTree(nodes))
	body := reqs[0].Body
	for _, want := range []string{"| a | b |", "| --- | --- |", "| 1 | 2 |"} {
		if !strings.Contains(body, want) {
			t.Errorf("table row %q missing from body %q", want, body)
		}
	}
}

// TestParseMD_MermaidFencePreserved verifies that a mermaid block survives the
// parser with its delimiters intact. Without them the exporter cannot turn it
// into a diagram and renders the diagram source as prose.
func TestParseMD_MermaidFencePreserved(t *testing.T) {
	src := []byte("# T\n\n## X-1: t\n\n```attr\nstatus: approved\n```\n\n" +
		"```mermaid\ngraph TD\n  A --> B\n```\n")
	nodes, _, _, err := parseMD(src, "diagram.md")
	if err != nil {
		t.Fatal(err)
	}
	reqs := model.CollectRequirements(assembleTree(nodes))
	want := "```mermaid\ngraph TD\n  A --> B\n```\n\n"
	if got := reqs[0].Body; got != want {
		t.Errorf("body =\n%q\nwant\n%q", got, want)
	}
}

// TestParseMD_FrontmatterParsed verifies that a leading YAML frontmatter block
// is parsed into the document's metadata and contributes no content: no node,
// no body prose.
func TestParseMD_FrontmatterParsed(t *testing.T) {
	src := []byte("---\ndescription: Abstract with **markdown**\ntags: [a, b]\n---\n\n" +
		"# Title\n\n## X-1: first\n\n```attr\nstatus: approved\n```\n\nBody.\n")
	nodes, title, meta, err := parseMD(src, "fm.md")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Title" {
		t.Errorf("title = %q, want %q", title, "Title")
	}
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d, want 1 (frontmatter must not open a node)", len(nodes))
	}
	if got, _ := meta["description"].(string); got != "Abstract with **markdown**" {
		t.Errorf("description = %q", got)
	}
	reqs := model.CollectRequirements(assembleTree(nodes))
	if len(reqs) != 1 {
		t.Fatalf("requirements = %d, want 1", len(reqs))
	}
	if got := reqs[0].Body; got != "Body." {
		t.Errorf("body = %q, want %q (metadata must not leak into the body)", got, "Body.")
	}
}

// TestParseFiles_FrontmatterFirstFileWins pins the merge rule across the files
// of one document directory.
func TestParseFiles_FrontmatterFirstFileWins(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	a := write("a.md", "---\ndescription: first\nowner: alice\n---\n\n# A\n")
	b := write("b.md", "---\ndescription: second\nowner: bob\n---\n\n# B\n")

	_, _, meta, errs := parseFiles([]string{a, b})
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if got, _ := meta["description"].(string); got != "first" {
		t.Errorf("description = %q, want the first file's value", got)
	}
	if got, _ := meta["owner"].(string); got != "alice" {
		t.Errorf("owner = %q, want the first file's value", got)
	}
}

// TestParseMD_ChunkedFrontmatter covers the same rule on the parallel path:
// chunk 0 carries the file preamble, so it is the chunk that yields the
// metadata, and the block must not reach any node's body.
func TestParseMD_ChunkedFrontmatter(t *testing.T) {
	src := append([]byte("---\ndescription: chunked\n---\n\n"), bigDoc(t, 400)...)
	if len(src) < parseParallelThreshold {
		t.Fatalf("fixture is %d bytes, below the threshold", len(src))
	}
	nodes, title, meta, err := parseMD(src, "big.md")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Big Spec" {
		t.Errorf("title = %q, want %q", title, "Big Spec")
	}
	if got, _ := meta["description"].(string); got != "chunked" {
		t.Errorf("description = %q, want %q", got, "chunked")
	}
	for _, req := range model.CollectRequirements(assembleTree(nodes)) {
		if strings.Contains(req.Body, "description:") {
			t.Errorf("frontmatter leaked into %s body: %q", req.ID, req.Body)
		}
	}
}

// TestParseMD_FrontmatterUnterminatedIsError pins the one failure shape where
// content would otherwise vanish without a diagnostic: a file whose first line
// opens a metadata block that is never closed. The extension accepts any run of
// dashes as a separator, so the guard has to as well.
func TestParseMD_FrontmatterUnterminatedIsError(t *testing.T) {
	body := "description: never closed\n\n## REQ-1: lost\n\n```attr\nid: REQ-1\ntitle: lost\n```\n"
	for _, opener := range []string{"--", "---", "----"} {
		src := []byte(opener + "\n" + body)
		if _, _, _, err := parseMD(src, "unterminated.md"); err == nil {
			t.Errorf("opener %q: expected an error for an unterminated frontmatter block", opener)
		}
	}

	// A properly closed block is metadata, not an error.
	closed := []byte("---\ndescription: fine\n---\n\n## REQ-1: found\n\n```attr\nid: REQ-1\ntitle: found\n```\n")
	if _, _, meta, err := parseMD(closed, "closed.md"); err != nil {
		t.Fatalf("closed frontmatter must parse: %v", err)
	} else if meta["description"] != "fine" {
		t.Errorf("description = %v, want %q", meta["description"], "fine")
	}
}

// TestParseMD_IndentedCodeBlockBecomesFence verifies the v2 code-block merge:
// an indented block is re-emitted as a fence so the body keeps code semantics
// instead of losing its indentation markers.
func TestParseMD_IndentedCodeBlockBecomesFence(t *testing.T) {
	src := []byte("# T\n\n## X-1: t\n\n```attr\nstatus: approved\n```\n\nProse.\n\n" +
		"    indented one\n    indented two\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("requirements = %d, want 1", len(reqs))
	}
	want := "Prose.\n\n```\nindented one\nindented two\n```\n\n"
	if got := reqs[0].Body; got != want {
		t.Errorf("body =\n%q\nwant\n%q", got, want)
	}
}

// TestParseMD_FenceInfoPreserved verifies that everything after the language in
// a fence's info string survives the body round-trip, not just the language.
func TestParseMD_FenceInfoPreserved(t *testing.T) {
	src := []byte("# T\n\n## X-1: t\n\n```attr\nstatus: approved\n```\n\n" +
		"```go title=\"main\"\nfunc main() {}\n```\n")
	reqs := reqsOf(parseReqTree(t, src))
	if len(reqs) != 1 {
		t.Fatalf("requirements = %d, want 1", len(reqs))
	}
	want := "```go title=\"main\"\nfunc main() {}\n```\n\n"
	if got := reqs[0].Body; got != want {
		t.Errorf("body =\n%q\nwant\n%q", got, want)
	}
}

// TestParseMD_ChunkedFrontmatterSeparatorRule keeps the split scan and the
// extension in agreement: the extension opens a metadata block on any run of
// dashes, so the scan must skip the same block. A heading-looking line inside
// the block would otherwise be taken for the document title.
func TestParseMD_ChunkedFrontmatterSeparatorRule(t *testing.T) {
	fm := "----\ndescription: dashes\nnot-a-heading: ## impostor: title\n----\n\n"
	src := append([]byte(fm), bigDoc(t, 400)...)
	if len(src) < parseParallelThreshold {
		t.Fatalf("fixture is %d bytes, below the threshold", len(src))
	}
	nodes, title, meta, err := parseMD(src, "dashes.md")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Big Spec" {
		t.Errorf("title = %q, want %q (a heading inside the metadata block must not win)", title, "Big Spec")
	}
	if got, _ := meta["description"].(string); got != "dashes" {
		t.Errorf("description = %q, want %q", got, "dashes")
	}
	for _, req := range model.CollectRequirements(assembleTree(nodes)) {
		if strings.Contains(req.Body, "impostor") {
			t.Errorf("metadata leaked into %s body", req.ID)
		}
	}
}
