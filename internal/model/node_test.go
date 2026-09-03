package model

import "testing"

func TestNodeAllFields(t *testing.T) {
	n := Node{
		Kind:      KindRequirement,
		ID:        "REQ-001",
		Title:     "Login requirement",
		Level:     2,
		Attrs:     map[string]any{"title": "Login", "asil": "B"},
		Body:      "The system shall support login.",
		Rationale: "Required by safety standard XYZ.",
		Source:    "/test/doc.md",
	}
	if n.Kind != KindRequirement {
		t.Errorf("Kind = %v, want KindRequirement", n.Kind)
	}
	if n.ID != "REQ-001" {
		t.Errorf("ID = %q, want %q", n.ID, "REQ-001")
	}
	if n.Title != "Login requirement" {
		t.Errorf("Title = %q, want %q", n.Title, "Login requirement")
	}
	if n.Level != 2 {
		t.Errorf("Level = %d, want 2", n.Level)
	}
	if n.Attrs["title"] != "Login" {
		t.Errorf(`Attrs["title"] = %v, want %v`, n.Attrs["title"], "Login")
	}
	if n.Attrs["asil"] != "B" {
		t.Errorf(`Attrs["asil"] = %v, want %v`, n.Attrs["asil"], "B")
	}
	if n.Body != "The system shall support login." {
		t.Errorf("Body = %q, want %q", n.Body, "The system shall support login.")
	}
	if n.Rationale != "Required by safety standard XYZ." {
		t.Errorf("Rationale = %q, want %q", n.Rationale, "Required by safety standard XYZ.")
	}
	if n.Source != "/test/doc.md" {
		t.Errorf("Source = %q, want %q", n.Source, "/test/doc.md")
	}
	if !n.IsRequirement() {
		t.Error("IsRequirement() = false, want true")
	}
}

func TestNodeEdgeCases(t *testing.T) {
	n := Node{}
	if n.ID != "" {
		t.Errorf("expected empty ID, got %q", n.ID)
	}
	if n.Attrs != nil {
		t.Errorf("expected nil Attrs, got %v", n.Attrs)
	}
	if n.Body != "" {
		t.Errorf("expected empty Body, got %q", n.Body)
	}
	if n.Rationale != "" {
		t.Errorf("expected empty Rationale, got %q", n.Rationale)
	}
	if n.Source != "" {
		t.Errorf("expected empty Source, got %q", n.Source)
	}
	if n.Kind != KindRequirement {
		t.Errorf("zero-value Kind = %v, want KindRequirement", n.Kind)
	}
	if !n.IsRequirement() {
		t.Error("zero-value node should be a requirement")
	}
}

func TestKindString(t *testing.T) {
	cases := []struct {
		want string
		kind Kind
	}{
		{want: "req", kind: KindRequirement},
		{want: "container", kind: KindContainer},
		{want: "info", kind: KindInfo},
	}
	for _, c := range cases {
		if got := c.kind.String(); got != c.want {
			t.Errorf("Kind(%d).String() = %q, want %q", c.kind, got, c.want)
		}
	}
}

func TestCollectRequirements(t *testing.T) {
	// tree: container -> [req A, info -> req B], req C
	reqA := &Node{Kind: KindRequirement, ID: "A"}
	reqB := &Node{Kind: KindRequirement, ID: "B"}
	reqC := &Node{Kind: KindRequirement, ID: "C"}
	info := &Node{Kind: KindInfo, Title: "Notes", Children: []*Node{reqB}}
	container := &Node{Kind: KindContainer, Title: "Section", Children: []*Node{reqA, info}}
	nodes := []*Node{container, reqC}

	got := CollectRequirements(nodes)
	if len(got) != 3 {
		t.Fatalf("CollectRequirements = %d nodes, want 3", len(got))
	}
	for i, want := range []string{"A", "B", "C"} {
		if got[i].ID != want {
			t.Errorf("CollectRequirements[%d].ID = %q, want %q", i, got[i].ID, want)
		}
	}
}

func TestDocumentRequirements(t *testing.T) {
	d := Document{Nodes: []*Node{
		{Kind: KindContainer, Title: "Section", Children: []*Node{
			{Kind: KindRequirement, ID: "REQ-001"},
			{Kind: KindInfo, Title: "Note", Children: []*Node{
				{Kind: KindRequirement, ID: "REQ-002"},
			}},
		}},
		{Kind: KindRequirement, ID: "REQ-003"},
	}}
	reqs := d.Requirements()
	if len(reqs) != 3 {
		t.Fatalf("Requirements() = %d, want 3", len(reqs))
	}
	for i, id := range []string{"REQ-001", "REQ-002", "REQ-003"} {
		if reqs[i].ID != id {
			t.Errorf("Requirements()[%d].ID = %q, want %q", i, reqs[i].ID, id)
		}
	}
	// In-place mutation must update the tree.
	reqs[1].Body = "mutated"
	if d.Nodes[0].Children[1].Children[0].Body != "mutated" {
		t.Error("mutation through Requirements() did not reach the tree")
	}
}

func TestDocumentZeroNodes(t *testing.T) {
	d := Document{
		Path:   "/test",
		Schema: map[string]any{"type": "object"},
		Nodes:  []*Node{},
	}
	if d.Path != "/test" {
		t.Errorf("Path = %q, want %q", d.Path, "/test")
	}
	if len(d.Requirements()) != 0 {
		t.Errorf("expected 0 requirements, got %d", len(d.Requirements()))
	}
}
