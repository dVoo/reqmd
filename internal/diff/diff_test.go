package diff

import (
	"reqmd/internal/model"
	"testing"
)

func TestDiff_Empty(t *testing.T) {
	result := Diff(nil, nil, nil, nil, "v1", "v2")

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Tag1 != "v1" || result.Tag2 != "v2" {
		t.Errorf("tags: want v1/v2, got %s/%s", result.Tag1, result.Tag2)
	}
	if len(result.Documents) != 0 {
		t.Errorf("expected no documents, got %d", len(result.Documents))
	}
	if result.Added != 0 || result.Removed != 0 || result.Modified != 0 || result.Unchanged != 0 {
		t.Errorf("expected zero totals, got added=%d removed=%d modified=%d unchanged=%d",
			result.Added, result.Removed, result.Modified, result.Unchanged)
	}
}

func TestDiff_Added(t *testing.T) {
	docs1 := []model.Document{{Path: "doc"}}
	docs2 := []model.Document{{
		Path: "doc",
		Nodes: []*model.Node{{
			ID:    "REQ-001",
			Title: "New Requirement",
			Attrs: map[string]any{"status": "draft"},
		}},
	}}

	result := Diff(docs1, docs2, nil, nil, "v1", "v2")

	if result.Added != 1 {
		t.Fatalf("expected 1 added, got %d", result.Added)
	}
	if len(result.Documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(result.Documents))
	}

	rd := result.Documents[0].Reqs[0]
	if rd.Category != Added {
		t.Errorf("expected category %s, got %s", Added, rd.Category)
	}
	if rd.ID != "REQ-001" {
		t.Errorf("expected ID REQ-001, got %s", rd.ID)
	}
}

func TestDiff_Removed(t *testing.T) {
	docs1 := []model.Document{{
		Path: "doc",
		Nodes: []*model.Node{{
			ID:    "REQ-001",
			Title: "Old Requirement",
			Attrs: map[string]any{"status": "approved"},
		}},
	}}
	docs2 := []model.Document{{Path: "doc"}}

	result := Diff(docs1, docs2, nil, nil, "v1", "v2")

	if result.Removed != 1 {
		t.Fatalf("expected 1 removed, got %d", result.Removed)
	}

	rd := result.Documents[0].Reqs[0]
	if rd.Category != Removed {
		t.Errorf("expected category %s, got %s", Removed, rd.Category)
	}
	if rd.Title != "Old Requirement" {
		t.Errorf("expected title 'Old Requirement', got %s", rd.Title)
	}
}

func TestDiff_Modified(t *testing.T) {
	docs1 := []model.Document{{
		Path: "doc",
		Nodes: []*model.Node{{
			ID:    "REQ-001",
			Title: "Requirement",
			Attrs: map[string]any{"status": "draft"},
		}},
	}}
	docs2 := []model.Document{{
		Path: "doc",
		Nodes: []*model.Node{{
			ID:    "REQ-001",
			Title: "Requirement",
			Attrs: map[string]any{"status": "approved"},
		}},
	}}

	result := Diff(docs1, docs2, nil, nil, "v1", "v2")

	if result.Modified != 1 {
		t.Fatalf("expected 1 modified, got %d", result.Modified)
	}
	if result.Unchanged != 0 {
		t.Fatalf("expected 0 unchanged, got %d", result.Unchanged)
	}

	rd := result.Documents[0].Reqs[0]
	if rd.Category != Modified {
		t.Errorf("expected category %s, got %s", Modified, rd.Category)
	}
	if len(rd.AttrChanges) != 1 {
		t.Fatalf("expected 1 attr change, got %d", len(rd.AttrChanges))
	}
	ac := rd.AttrChanges[0]
	if ac.Key != "status" {
		t.Errorf("expected key status, got %s", ac.Key)
	}
	if ac.OldVal != "draft" {
		t.Errorf("expected old value draft, got %v", ac.OldVal)
	}
	if ac.NewVal != "approved" {
		t.Errorf("expected new value approved, got %v", ac.NewVal)
	}
}

func TestDiff_Unchanged(t *testing.T) {
	docs := []model.Document{{
		Path: "doc",
		Nodes: []*model.Node{{
			ID:    "REQ-001",
			Title: "Requirement",
			Attrs: map[string]any{"status": "approved"},
		}},
	}}

	result := Diff(docs, docs, nil, nil, "v1", "v2")

	if result.Unchanged != 1 {
		t.Fatalf("expected 1 unchanged, got %d", result.Unchanged)
	}
	if result.Modified != 0 || result.Added != 0 || result.Removed != 0 {
		t.Errorf("expected only unchanged, got added=%d removed=%d modified=%d",
			result.Added, result.Removed, result.Modified)
	}

	rd := result.Documents[0].Reqs[0]
	if rd.Category != Unchanged {
		t.Errorf("expected category %s, got %s", Unchanged, rd.Category)
	}
}

func TestDiff_Mixed(t *testing.T) {
	docs1 := []model.Document{
		{
			Path: "a",
			Nodes: []*model.Node{
				{ID: "REQ-A1", Title: "A1", Attrs: map[string]any{"status": "approved"}},
				{ID: "REQ-A2", Title: "A2", Attrs: map[string]any{"status": "approved"}},
			},
		},
		{
			Path: "b",
			Nodes: []*model.Node{
				{ID: "REQ-B1", Title: "B1", Attrs: map[string]any{"status": "approved"}},
			},
		},
	}

	docs2 := []model.Document{
		{
			Path: "a",
			Nodes: []*model.Node{
				{ID: "REQ-A1", Title: "A1", Attrs: map[string]any{"status": "approved"}},
				{ID: "REQ-A2", Title: "A2", Attrs: map[string]any{"status": "draft"}},
			},
		},
		{
			Path: "c",
			Nodes: []*model.Node{
				{ID: "REQ-C1", Title: "C1", Attrs: map[string]any{"status": "draft"}},
			},
		},
	}

	result := Diff(docs1, docs2, nil, nil, "v1", "v2")

	if result.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Added)
	}
	if result.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Removed)
	}
	if result.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", result.Modified)
	}
	if result.Unchanged != 1 {
		t.Errorf("expected 1 unchanged, got %d", result.Unchanged)
	}

	if len(result.Documents) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(result.Documents))
	}
}

func TestSchemas_AddedProperty(t *testing.T) {
	schemas1 := map[string]map[string]any{
		"doc1": {"properties": map[string]any{"a": "string"}},
	}
	schemas2 := map[string]map[string]any{
		"doc1": {"properties": map[string]any{"a": "string", "b": "string"}},
	}
	result := Schemas(schemas1, schemas2)
	if len(result) != 1 {
		t.Fatalf("expected 1 schema diff, got %d", len(result))
	}
	if len(result[0].Changes) == 0 {
		t.Fatal("expected at least one change")
	}
}

func TestSchemas_RemovedProperty(t *testing.T) {
	schemas1 := map[string]map[string]any{
		"doc1": {"properties": map[string]any{"a": "string", "b": "string"}},
	}
	schemas2 := map[string]map[string]any{
		"doc1": {"properties": map[string]any{"a": "string"}},
	}
	result := Schemas(schemas1, schemas2)
	if len(result) != 1 {
		t.Fatalf("expected 1 schema diff, got %d", len(result))
	}
}

func TestSchemas_ChangedRequired(t *testing.T) {
	schemas1 := map[string]map[string]any{
		"doc1": {"required": []any{"a"}},
	}
	schemas2 := map[string]map[string]any{
		"doc1": {"required": []any{"a", "b"}},
	}
	result := Schemas(schemas1, schemas2)
	if len(result) != 1 {
		t.Fatalf("expected 1 schema diff, got %d", len(result))
	}
}

func TestSchemas_NoChange(t *testing.T) {
	schemas1 := map[string]map[string]any{
		"doc1": {"properties": map[string]any{"a": "string"}},
	}
	schemas2 := map[string]map[string]any{
		"doc1": {"properties": map[string]any{"a": "string"}},
	}
	result := Schemas(schemas1, schemas2)
	if len(result) != 0 {
		t.Fatalf("expected 0 schema diffs, got %d", len(result))
	}
}

func TestSubmodules_Empty(t *testing.T) {
	result := Submodules(map[string]string{}, map[string]string{})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}

func TestSubmodules_NilMaps(t *testing.T) {
	result := Submodules(nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty result from nil maps, got %d entries", len(result))
	}
}

func TestSubmodules_NoChange(t *testing.T) {
	subs := map[string]string{
		"vendor/a": "aaa1111111111111111111111111111111111111",
		"vendor/b": "bbb2222222222222222222222222222222222222",
	}
	result := Submodules(subs, subs)
	if len(result) != 0 {
		t.Errorf("expected empty result (no changes), got %d entries", len(result))
	}
}

func TestSubmodules_Added(t *testing.T) {
	subs1 := map[string]string{
		"vendor/a": "aaa1111111111111111111111111111111111111",
	}
	subs2 := map[string]string{
		"vendor/a": "aaa1111111111111111111111111111111111111",
		"vendor/b": "bbb2222222222222222222222222222222222222",
	}
	result := Submodules(subs1, subs2)
	if len(result) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result))
	}
	if result[0].Path != "vendor/b" {
		t.Errorf("expected vendor/b, got %s", result[0].Path)
	}
	if result[0].Status != "added" {
		t.Errorf("expected status=added, got %s", result[0].Status)
	}
	if result[0].OldSHA != "" {
		t.Errorf("expected empty OldSHA, got %s", result[0].OldSHA)
	}
	if result[0].NewSHA != "bbb2222222222222222222222222222222222222" {
		t.Errorf("unexpected NewSHA: %s", result[0].NewSHA)
	}
}

func TestSubmodules_Removed(t *testing.T) {
	subs1 := map[string]string{
		"vendor/a": "aaa1111111111111111111111111111111111111",
		"vendor/b": "bbb2222222222222222222222222222222222222",
	}
	subs2 := map[string]string{
		"vendor/a": "aaa1111111111111111111111111111111111111",
	}
	result := Submodules(subs1, subs2)
	if len(result) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result))
	}
	if result[0].Path != "vendor/b" {
		t.Errorf("expected vendor/b, got %s", result[0].Path)
	}
	if result[0].Status != "removed" {
		t.Errorf("expected status=removed, got %s", result[0].Status)
	}
	if result[0].OldSHA != "bbb2222222222222222222222222222222222222" {
		t.Errorf("unexpected OldSHA: %s", result[0].OldSHA)
	}
	if result[0].NewSHA != "" {
		t.Errorf("expected empty NewSHA, got %s", result[0].NewSHA)
	}
}

func TestSubmodules_Updated(t *testing.T) {
	subs1 := map[string]string{
		"vendor/a": "aaa1111111111111111111111111111111111111",
	}
	subs2 := map[string]string{
		"vendor/a": "ccc3333333333333333333333333333333333333",
	}
	result := Submodules(subs1, subs2)
	if len(result) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result))
	}
	if result[0].Path != "vendor/a" {
		t.Errorf("expected vendor/a, got %s", result[0].Path)
	}
	if result[0].Status != "updated" {
		t.Errorf("expected status=updated, got %s", result[0].Status)
	}
	if result[0].OldSHA != "aaa1111111111111111111111111111111111111" {
		t.Errorf("unexpected OldSHA: %s", result[0].OldSHA)
	}
	if result[0].NewSHA != "ccc3333333333333333333333333333333333333" {
		t.Errorf("unexpected NewSHA: %s", result[0].NewSHA)
	}
}

func TestSubmodules_Mixed(t *testing.T) {
	subs1 := map[string]string{
		"vendor/a": "aaa1111111111111111111111111111111111111",
		"vendor/b": "bbb2222222222222222222222222222222222222",
		"vendor/c": "ccc3333333333333333333333333333333333333",
	}
	subs2 := map[string]string{
		"vendor/a": "aaa1111111111111111111111111111111111111", // unchanged
		"vendor/b": "ddd4444444444444444444444444444444444444", // updated
		// vendor/c removed
		"vendor/d": "eee5555555555555555555555555555555555555", // added
	}
	result := Submodules(subs1, subs2)
	if len(result) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(result))
	}
	// Verify sort order: b, c, d (alphabetical)
	if result[0].Path != "vendor/b" || result[0].Status != "updated" {
		t.Errorf("result[0]: expected vendor/b updated, got %s %s", result[0].Path, result[0].Status)
	}
	if result[1].Path != "vendor/c" || result[1].Status != "removed" {
		t.Errorf("result[1]: expected vendor/c removed, got %s %s", result[1].Path, result[1].Status)
	}
	if result[2].Path != "vendor/d" || result[2].Status != "added" {
		t.Errorf("result[2]: expected vendor/d added, got %s %s", result[2].Path, result[2].Status)
	}
}

func TestSubmodules_SortedOutput(t *testing.T) {
	subs1 := map[string]string{
		"z/vendor": "aaa1111111111111111111111111111111111111",
		"a/vendor": "bbb2222222222222222222222222222222222222",
		"m/vendor": "ccc3333333333333333333333333333333333333",
	}
	subs2 := map[string]string{
		"z/vendor": "ddd4444444444444444444444444444444444444",
		"a/vendor": "eee5555555555555555555555555555555555555",
		"m/vendor": "fff6666666666666666666666666666666666666",
	}
	result := Submodules(subs1, subs2)
	if len(result) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(result))
	}
	if result[0].Path != "a/vendor" {
		t.Errorf("expected first result a/vendor, got %s", result[0].Path)
	}
	if result[1].Path != "m/vendor" {
		t.Errorf("expected second result m/vendor, got %s", result[1].Path)
	}
	if result[2].Path != "z/vendor" {
		t.Errorf("expected third result z/vendor, got %s", result[2].Path)
	}
}
