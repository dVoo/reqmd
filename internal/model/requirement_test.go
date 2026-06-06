package model

import "testing"

func TestRequirementAllFields(t *testing.T) {
	r := Requirement{
		ID: "REQ-001",
		Title: "Login requirement",
		Attrs: map[string]any{
			"title": "Login",
			"asil":  "B",
		},
		Body:     "The system shall support login.",
		Rationale: "Required by safety standard XYZ.",
		Source:   "/test/doc.md",
	}
	if r.ID != "REQ-001" {
		t.Errorf("ID = %q, want %q", r.ID, "REQ-001")
	}
	if r.Title != "Login requirement" {
		t.Errorf("Title = %q, want %q", r.Title, "Login requirement")
	}
	if r.Attrs["title"] != "Login" {
		t.Errorf(`Attrs["title"] = %v, want %v`, r.Attrs["title"], "Login")
	}
	if r.Attrs["asil"] != "B" {
		t.Errorf(`Attrs["asil"] = %v, want %v`, r.Attrs["asil"], "B")
	}
	if r.Body != "The system shall support login." {
		t.Errorf("Body = %q, want %q", r.Body, "The system shall support login.")
	}
	if r.Rationale != "Required by safety standard XYZ." {
		t.Errorf("Rational = %q, want %q", r.Rationale, "Required by safety standard XYZ.")
	}
	if r.Source != "/test/doc.md" {
		t.Errorf("Source = %q, want %q", r.Source, "/test/doc.md")
	}
}

func TestRequirementEdgeCases(t *testing.T) {
	r := Requirement{
		ID:       "",
		Attrs:    nil,
		Body:     "",
		Rationale: "",
		Source:   "",
	}
	if r.ID != "" {
		t.Errorf("expected empty ID, got %q", r.ID)
	}
	if r.Attrs != nil {
		t.Errorf("expected nil Attrs, got %v", r.Attrs)
	}
	if r.Body != "" {
		t.Errorf("expected empty Body, got %q", r.Body)
	}
	if r.Rationale != "" {
		t.Errorf("expected empty Rational, got %q", r.Rationale)
	}
	if r.Source != "" {
		t.Errorf("expected empty Source, got %q", r.Source)
	}
}

func TestDocumentZeroRequirements(t *testing.T) {
	d := Document{
		Path:         "/test",
		Schema:       map[string]any{"type": "object"},
		Requirements: []Requirement{},
	}
	if d.Path != "/test" {
		t.Errorf("Path = %q, want %q", d.Path, "/test")
	}
	if len(d.Requirements) != 0 {
		t.Errorf("expected 0 requirements, got %d", len(d.Requirements))
	}
}

func TestDocumentMultipleRequirements(t *testing.T) {
	d := Document{
		Path:   "/test",
		Schema: map[string]any{"type": "object"},
		Requirements: []Requirement{
			{ID: "REQ-001", Body: "First requirement."},
			{ID: "REQ-002", Body: "Second requirement."},
			{ID: "REQ-003", Body: "Third requirement."},
		},
	}
	if len(d.Requirements) != 3 {
		t.Fatalf("expected 3 requirements, got %d", len(d.Requirements))
	}
	for i, id := range []string{"REQ-001", "REQ-002", "REQ-003"} {
		if d.Requirements[i].ID != id {
			t.Errorf("Requirements[%d].ID = %q, want %q", i, d.Requirements[i].ID, id)
		}
	}
}
