package parser

import (
	"os"
	"path/filepath"
	"reqmd/internal/model"
	"strings"
	"testing"
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
	if n.Kind != model.KindInfo {
		t.Errorf("Kind = %v, want KindInfo", n.Kind)
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
	expected := "Some text.```json\n{\"key\": \"value\"}\n```\n\n"
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
	if info.Kind != model.KindInfo {
		t.Errorf("child Kind = %v, want KindInfo", info.Kind)
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
	if nodes[0].Kind != model.KindInfo || nodes[0].Title != "" || nodes[0].Body != "Intro prose." {
		t.Errorf("nodes[0] = %+v, want leading info node with prose", nodes[0])
	}
	if nodes[1].Kind != model.KindRequirement || nodes[1].ID != "REQ-001" || nodes[1].ParentID != "" {
		t.Errorf("nodes[1] = %v %q, want top-level requirement REQ-001", nodes[1].Kind, nodes[1].ID)
	}
}

// TestParseSingleFile verifies ParseSingleFile reads a real file from disk
// and produces the correct results. Uses t.TempDir() for the temp file.
func TestParseSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_req.md")
	content := []byte("## REQ-FILE\n```attr\nid: REQ-FILE\nasil: ASIL C\n```\nThis requirement was loaded from a file.\n\n*Rationale: file-based parsing\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
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
// (no attr block) becomes a KindContainer node that holds the requirements
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
	if container.Kind != model.KindContainer {
		t.Errorf("nodes[1] Kind = %v, want KindContainer", container.Kind)
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
			if n.Kind == model.KindContainer && n.Title == title {
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
	if len(r1.Children) != 1 || r1.Children[0].Kind != model.KindContainer || r1.Children[0].Title != "Section" {
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
	if r2.Children[0].Kind != model.KindInfo || r2.Children[0].Title != "Info" {
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
	if preamble.Kind != model.KindInfo {
		t.Errorf("preamble Kind = %v, want KindInfo", preamble.Kind)
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
	if intro.Kind != model.KindInfo || intro.Title != "" || intro.Body != "Just some background." {
		t.Errorf("nodes[0] = %v %q/%q, want leading info/background", intro.Kind, intro.Title, intro.Body)
	}
	details := nodes[1]
	if details.Kind != model.KindInfo || details.Title != "Details" || details.Body != "More prose." {
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
	if len(nodes) != 4 || nodes[0].Kind != model.KindInfo || nodes[0].Body != "Intro prose." {
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
