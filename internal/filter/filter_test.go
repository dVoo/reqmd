package filter

import (
	"reqmd/internal/model"
	"testing"
)

func testReq(id, title string, attrs map[string]any) *model.Node {
	return &model.Node{
		Kind:  model.KindRequirement,
		ID:    id,
		Title: title,
		Attrs: attrs,
	}
}

func testDocs() []model.Document {
	return []model.Document{
		{
			Path:       "spec/sw",
			Properties: []string{"status", "trace", "version", "variant", "priority"},
			Nodes: []*model.Node{
				testReq("SW-001", "Parser", map[string]any{
					"status":   "approved",
					"version":  3,
					"variant":  []any{"Base", "Premium"},
					"priority": "High",
				}),
				testReq("SW-002", "Schema", map[string]any{
					"status":   "draft",
					"variant":  []any{"Sport"},
					"priority": "Critical",
				}),
				testReq("SW-003", "Common", map[string]any{
					"status":   "approved",
					"priority": "Medium",
				}),
			},
		},
		{
			Path:       "spec/stk",
			Properties: []string{"status", "source", "category"},
			Nodes: []*model.Node{
				testReq("STK-001", "Safety goal", map[string]any{
					"status":   "approved",
					"source":   "Safety",
					"category": "Validation",
				}),
			},
		},
	}
}

func validAttrsFor(docs []model.Document) map[string]struct{} {
	return BuildValidAttrs(docs)
}

func TestCompileAndMatch_Equality(t *testing.T) {
	docs := testDocs()
	f, err := Compile(`status == "approved"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	match, err := f.Match(testReq("X", "", map[string]any{"status": "approved"}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !match {
		t.Fatal("approved req should match")
	}

	match, err = f.Match(testReq("Y", "", map[string]any{"status": "draft"}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if match {
		t.Fatal("draft req should not match")
	}
}

func TestCompileAndMatch_ArrayIn(t *testing.T) {
	docs := testDocs()
	f, err := Compile(`"Premium" in variant`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	match, err := f.Match(testReq("SW-001", "", map[string]any{"variant": []any{"Base", "Premium"}}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !match {
		t.Fatal("req with [Base, Premium] should match \"Premium\" in variant")
	}

	match, err = f.Match(testReq("SW-002", "", map[string]any{"variant": []any{"Sport"}}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if match {
		t.Fatal("req with [Sport] should not match \"Premium\" in variant")
	}
}

func TestCompileAndMatch_NilAttribute(t *testing.T) {
	docs := testDocs()
	// SW-003 has no variant attr — should be nil, not match.
	f, err := Compile(`"Premium" in variant`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	match, err := f.Match(testReq("SW-003", "", map[string]any{"status": "approved"}))
	if err != nil {
		t.Fatalf("Match on nil variant: %v", err)
	}
	if match {
		t.Fatal("req without variant attr should not match \"Premium\" in variant")
	}
}

func TestCompileAndMatch_NilCoalescing(t *testing.T) {
	docs := testDocs()
	// Expression handles absence: variant == nil or "Base" in variant
	f, err := Compile(`variant == nil or "Base" in variant`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	// Req with no variant attr should match (variant == nil).
	match, err := f.Match(testReq("SW-003", "", map[string]any{"status": "approved"}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !match {
		t.Fatal("req without variant should match `variant == nil or \"Base\" in variant`")
	}

	// Req with [Base, Premium] should match.
	match, err = f.Match(testReq("SW-001", "", map[string]any{"variant": []any{"Base", "Premium"}}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !match {
		t.Fatal("req with [Base, Premium] should match")
	}

	// Req with [Sport] should not match.
	match, err = f.Match(testReq("SW-002", "", map[string]any{"variant": []any{"Sport"}}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if match {
		t.Fatal("req with [Sport] should not match")
	}
}

func TestCompileAndMatch_NumericComparison(t *testing.T) {
	docs := testDocs()
	f, err := Compile(`version > 2`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	match, err := f.Match(testReq("X", "", map[string]any{"version": 3}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !match {
		t.Fatal("version 3 > 2 should match")
	}

	match, err = f.Match(testReq("Y", "", map[string]any{"version": 1}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if match {
		t.Fatal("version 1 > 2 should not match")
	}
}

func TestCompileAndMatch_BuiltInIDAndTitle(t *testing.T) {
	docs := testDocs()
	f, err := Compile(`id startsWith "SW-"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	match, err := f.Match(testReq("SW-001", "Parser", nil))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !match {
		t.Fatal("SW-001 should match id startsWith \"SW-\"")
	}

	match, err = f.Match(testReq("STK-001", "Safety", nil))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if match {
		t.Fatal("STK-001 should not match id startsWith \"SW-\"")
	}
}

func TestCompileAndMatch_BooleanComposition(t *testing.T) {
	docs := testDocs()
	f, err := Compile(`status == "approved" and priority == "High"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	match, err := f.Match(testReq("X", "", map[string]any{"status": "approved", "priority": "High"}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if !match {
		t.Fatal("approved + High should match")
	}

	match, err = f.Match(testReq("Y", "", map[string]any{"status": "approved", "priority": "Medium"}))
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	if match {
		t.Fatal("approved + Medium should not match")
	}
}

func TestCompile_SyntaxError(t *testing.T) {
	docs := testDocs()
	_, err := Compile(`status ==`, validAttrsFor(docs))
	if err == nil {
		t.Fatal("expected syntax error for incomplete expression")
	}
}

func TestCompile_UnknownAttribute(t *testing.T) {
	docs := testDocs()
	_, err := Compile(`nonexistent == "foo"`, validAttrsFor(docs))
	if err == nil {
		t.Fatal("expected error for unknown attribute")
	}
}

func TestCompile_EmptyExpression(t *testing.T) {
	_, err := Compile("", validAttrsFor(testDocs()))
	if err == nil {
		t.Fatal("expected error for empty expression")
	}
}

func TestCompile_AllowsExprBuiltins(t *testing.T) {
	docs := testDocs()
	// len is an expr builtin, not an attribute — should not trigger
	// the unknown-attribute check.
	_, err := Compile(`len(variant) > 0`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("len builtin should not be treated as unknown attribute: %v", err)
	}
}

func TestCompile_AllowsCrossSchemaAttribute(t *testing.T) {
	docs := testDocs()
	// `source` is only in spec/stk schema, not in spec/sw.
	// Global typo check: valid because it exists in at least one schema.
	_, err := Compile(`source == "Safety"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("attribute from a different schema should be valid: %v", err)
	}
}

func TestFilterDocs(t *testing.T) {
	docs := testDocs()
	f, err := Compile(`status == "approved"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	filtered, err := f.FilterDocs(docs)
	if err != nil {
		t.Fatalf("FilterDocs: %v", err)
	}

	// spec/sw: SW-001 (approved), SW-002 (draft), SW-003 (approved) → 2 match
	if got := len(filtered[0].Requirements()); got != 2 {
		t.Fatalf("spec/sw: expected 2 reqs, got %d", got)
	}
	// spec/stk: STK-001 (approved) → 1 match
	if got := len(filtered[1].Requirements()); got != 1 {
		t.Fatalf("spec/stk: expected 1 req, got %d", got)
	}

	// Verify IDs
	if filtered[0].Requirements()[0].ID != "SW-001" {
		t.Fatalf("expected SW-001, got %s", filtered[0].Requirements()[0].ID)
	}
	if filtered[0].Requirements()[1].ID != "SW-003" {
		t.Fatalf("expected SW-003, got %s", filtered[0].Requirements()[1].ID)
	}
}

func TestFilterDocs_PreservesDocStructure(t *testing.T) {
	docs := testDocs()
	// Filter that matches nothing in spec/sw
	f, err := Compile(`source == "Safety"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	filtered, err := f.FilterDocs(docs)
	if err != nil {
		t.Fatalf("FilterDocs: %v", err)
	}

	// spec/sw should still be present (with 0 reqs)
	if filtered[0].Path != "spec/sw" {
		t.Fatalf("spec/sw should be preserved, got path %s", filtered[0].Path)
	}
	if len(filtered[0].Requirements()) != 0 {
		t.Fatalf("spec/sw should have 0 reqs, got %d", len(filtered[0].Requirements()))
	}
	// spec/stk should have 1 req
	if len(filtered[1].Requirements()) != 1 {
		t.Fatalf("spec/stk should have 1 req, got %d", len(filtered[1].Requirements()))
	}
}

func TestFilterDocs_DoesNotMutateOriginal(t *testing.T) {
	docs := testDocs()
	originalCount := len(docs[0].Requirements())

	f, err := Compile(`status == "approved"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	_, err = f.FilterDocs(docs)
	if err != nil {
		t.Fatalf("FilterDocs: %v", err)
	}

	if len(docs[0].Requirements()) != originalCount {
		t.Fatalf("original docs mutated: expected %d reqs, got %d", originalCount, len(docs[0].Requirements()))
	}
}

func TestFilterDocs_KeepsItems(t *testing.T) {
	docs := testDocs()
	docs[0].Nodes = []*model.Node{
		{Kind: model.KindContainer, Title: "Section", Children: []*model.Node{
			testReq("SW-001", "Parser", map[string]any{"status": "approved"}),
			testReq("SW-002", "Schema", map[string]any{"status": "draft"}),
		}},
		{Kind: model.KindInfo, Title: "Note"},
	}
	f, err := Compile(`status == "approved"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	filtered, err := f.FilterDocs(docs)
	if err != nil {
		t.Fatalf("FilterDocs: %v", err)
	}

	nodes := filtered[0].Nodes
	if len(nodes) != 2 {
		t.Fatalf("expected 2 top-level nodes (container + info), got %d", len(nodes))
	}
	container := nodes[0]
	if container.Kind != model.KindContainer || container.Title != "Section" {
		t.Fatalf("nodes[0] = %v %q, want container Section", container.Kind, container.Title)
	}
	if len(container.Children) != 1 || container.Children[0].ID != "SW-001" {
		t.Fatalf("container children = %+v, want [SW-001] (SW-002 pruned)", container.Children)
	}
	if nodes[1].Kind != model.KindInfo || nodes[1].Title != "Note" {
		t.Fatalf("nodes[1] = %v %q, want info Note", nodes[1].Kind, nodes[1].Title)
	}
	if reqs := filtered[0].Requirements(); len(reqs) != 1 || reqs[0].ID != "SW-001" {
		t.Fatalf("Requirements() = %+v, want [SW-001]", reqs)
	}
}

func TestFilterDocs_ReattachesChildrenOfPrunedReq(t *testing.T) {
	docs := testDocs()
	child := testReq("SW-CHILD", "Child", map[string]any{"status": "approved"})
	docs[0].Nodes = []*model.Node{
		{Kind: model.KindRequirement, ID: "SW-PARENT", Title: "Parent", Attrs: map[string]any{"status": "draft"}, Children: []*model.Node{child}},
	}
	f, err := Compile(`status == "approved"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	filtered, err := f.FilterDocs(docs)
	if err != nil {
		t.Fatalf("FilterDocs: %v", err)
	}

	nodes := filtered[0].Nodes
	if len(nodes) != 1 || nodes[0].ID != "SW-CHILD" {
		t.Fatalf("pruned parent's child should be re-attached: %+v", nodes)
	}
}

func TestMatchingIDs(t *testing.T) {
	docs := testDocs()
	f, err := Compile(`status == "approved"`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	ids, err := f.MatchingIDs(docs)
	if err != nil {
		t.Fatalf("MatchingIDs: %v", err)
	}

	expected := map[string]struct{}{
		"SW-001":  {},
		"SW-003":  {},
		"STK-001": {},
	}
	if len(ids) != len(expected) {
		t.Fatalf("expected %d IDs, got %d", len(expected), len(ids))
	}
	for id := range expected {
		if _, ok := ids[id]; !ok {
			t.Fatalf("expected ID %s in matching set", id)
		}
	}
}

func TestExprAndRefs(t *testing.T) {
	docs := testDocs()
	f, err := Compile(`status == "approved" and "Base" in variant`, validAttrsFor(docs))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	if f.Expr() != `status == "approved" and "Base" in variant` {
		t.Fatalf("Expr(): got %q", f.Expr())
	}

	// refs should be deduplicated and sorted: status, variant
	refs := f.Refs()
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d: %v", len(refs), refs)
	}
	if refs[0] != "status" || refs[1] != "variant" {
		t.Fatalf("expected [status variant], got %v", refs)
	}
}

func TestBuildValidAttrs(t *testing.T) {
	docs := testDocs()
	valid := BuildValidAttrs(docs)

	// Should include schema properties from both docs
	for _, attr := range []string{"status", "trace", "version", "variant", "priority", "source", "category"} {
		if _, ok := valid[attr]; !ok {
			t.Errorf("expected %q in valid attrs", attr)
		}
	}
	// Should include built-in vars
	for _, attr := range []string{"id", "title"} {
		if _, ok := valid[attr]; !ok {
			t.Errorf("expected built-in %q in valid attrs", attr)
		}
	}
}
