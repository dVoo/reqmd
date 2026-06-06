package exporter

import (
	"testing"

	"reqmd/internal/model"
)

func TestBuildRenderData(t *testing.T) {
	doc := model.Document{
		Requirements: []model.Requirement{
			{
				ID:    "REQ-001",
				Title: "First",
				Attrs: map[string]any{
					"status": "approved",
					"asil":   "B",
				},
				Body: "First body.",
			},
			{
				ID:    "REQ-002",
				Title: "Second",
				Attrs: map[string]any{
					"status": "draft",
					"asil":   "C",
				},
				Body: "Second body.",
			},
			{
				// Single-value attr (same on both) must be dropped from filter set.
				ID:       "REQ-003",
				ParentID: "REQ-001",
				Attrs: map[string]any{
					"asil":      "B",
					"maturity":  "released",
					"trace":     []string{"REQ-002"},
					"req-only":  "same",
					"req-only2": "same",
				},
				Body: "Third body.",
			},
		},
	}

	rd := buildRenderData(doc)
	if rd == nil {
		t.Fatal("buildRenderData returned nil")
	}

	// TopLevel: requirements without ParentID.
	if len(rd.TopLevel) != 2 {
		t.Errorf("TopLevel: got %d, want 2", len(rd.TopLevel))
	}

	// ChildrenOf: REQ-003 → REQ-001.
	kids, ok := rd.ChildrenOf["REQ-001"]
	if !ok || len(kids) != 1 || kids[0].ID != "REQ-003" {
		t.Errorf("ChildrenOf[REQ-001] = %v, want [REQ-003]", kids)
	}

	// StatusCounts: 1 approved (REQ-001) + 1 default-approved (REQ-003, no status)
	// = 2, and 1 draft (REQ-002).
	if rd.StatusCnt["approved"] != 2 {
		t.Errorf("StatusCnt[approved] = %d, want 2", rd.StatusCnt["approved"])
	}
	if rd.StatusCnt["draft"] != 1 {
		t.Errorf("StatusCnt[draft] = %d, want 1", rd.StatusCnt["draft"])
	}

	// Single-value attrs must be omitted from the filter set.
	for _, k := range []string{"maturity", "req-only", "req-only2"} {
		if _, found := rd.Filters[k]; found {
			t.Errorf("filter %q should be omitted (single value)", k)
		}
	}

	// status filter has 2 values, should be present and ordered (draft, approved).
	stat, ok := rd.Filters["status"]
	if !ok {
		t.Fatal("status filter missing")
	}
	wantStat := []string{"draft", "approved"}
	if len(stat) != len(wantStat) {
		t.Fatalf("status filter len = %d, want %d (%v)", len(stat), len(wantStat), stat)
	}
	for i, v := range wantStat {
		if stat[i] != v {
			t.Errorf("status filter[%d] = %q, want %q", i, stat[i], v)
		}
	}
}

func TestNormalizeAttrValue(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want []string
	}{
		{
			name: "[]any with mixed scalars",
			in:   []any{"a", 1, nil, "b"},
			want: []string{"a", "1", "b"},
		},
		{
			name: "[]any with nested map dropped",
			in:   []any{"x", map[string]any{"k": "v"}, "y"},
			want: []string{"x", "y"},
		},
		{
			name: "[]string copy",
			in:   []string{"a", "b"},
			want: []string{"a", "b"},
		},
		{
			name: "nil input",
			in:   nil,
			want: nil,
		},
		{
			name: "[]any(nil)",
			in:   []any(nil),
			want: nil,
		},
		{
			name: "[]string(nil)",
			in:   []string(nil),
			want: nil,
		},
		{
			name: "scalar string",
			in:   "hello",
			want: []string{"hello"},
		},
		{
			name: "scalar int",
			in:   42,
			want: []string{"42"},
		},
		{
			name: "map[string]any dropped",
			in:   map[string]any{"k": "v"},
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeAttrValue(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v (%d), want %v (%d)", got, len(got), tc.want, len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("[%d] got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestBuildTraceCache(t *testing.T) {
	doc := model.Document{
		Requirements: []model.Requirement{
			{ID: "REQ-001"},
			{ID: "REQ-002"},
		},
	}

	// Nil trace resolver → nil cache.
	if got := buildTraceCache(doc, nil); got != nil {
		t.Errorf("nil tc: got %v, want nil", got)
	}

	// Populated resolver populates both maps.
	tc := &TraceResolver{
		Upstream:    func(id string) []string { return []string{"U-" + id} },
		Downstream:  func(id string) []string { return []string{"D-" + id} },
		ResolveLink: func(id string) string { return "link-" + id },
		TitleOf:     func(id string) string { return "Title " + id },
	}
	cache := buildTraceCache(doc, tc)
	if len(cache) != 2 {
		t.Fatalf("cache len = %d, want 2", len(cache))
	}
	pair, ok := cache["REQ-001"]
	if !ok {
		t.Fatal("cache missing REQ-001")
	}
	if len(pair.up) != 1 || pair.up[0].ID != "U-REQ-001" {
		t.Errorf("pair.up = %+v, want [U-REQ-001]", pair.up)
	}
	if len(pair.down) != 1 || pair.down[0].ID != "D-REQ-001" {
		t.Errorf("pair.down = %+v, want [D-REQ-001]", pair.down)
	}
}

func TestBuildIndexEntry(t *testing.T) {
	req := model.Requirement{
		ID:    "REQ-007",
		Title: "A title",
		Attrs: map[string]any{
			"status": "approved",
			"asil":   "D",
			"trace":  []string{"X-1"}, // should be dropped from entry attrs
		},
		Body:      "body",
		Rationale: "why",
	}
	entry := buildIndexEntry(req, nil)

	if entry.ID != "REQ-007" {
		t.Errorf("ID = %q, want REQ-007", entry.ID)
	}
	if entry.Title != "A title" {
		t.Errorf("Title = %q, want 'A title'", entry.Title)
	}
	if entry.Body != "body" {
		t.Errorf("Body = %q, want 'body'", entry.Body)
	}
	if entry.Rationale != "why" {
		t.Errorf("Rationale = %q, want 'why'", entry.Rationale)
	}
	if entry.Attrs["status"] != "approved" {
		t.Errorf("Attrs[status] = %v, want 'approved'", entry.Attrs["status"])
	}
	if entry.Attrs["asil"] != "D" {
		t.Errorf("Attrs[asil] = %v, want 'D'", entry.Attrs["asil"])
	}
	if _, found := entry.Attrs["trace"]; found {
		t.Error("trace attr should be omitted from index entry attrs")
	}

	// With children.
	childEntry := buildIndexEntry(req, map[string][]model.Requirement{
		"REQ-007": {{ID: "REQ-007-1"}, {ID: "REQ-007-2"}},
	})
	wantKids := []string{"REQ-007-1", "REQ-007-2"}
	if len(childEntry.Children) != len(wantKids) {
		t.Fatalf("Children len = %d, want %d", len(childEntry.Children), len(wantKids))
	}
	for i, k := range wantKids {
		if childEntry.Children[i] != k {
			t.Errorf("Children[%d] = %q, want %q", i, childEntry.Children[i], k)
		}
	}

	// Body truncation at 80 runes.
	longBody := ""
	for i := 0; i < 100; i++ {
		longBody += "x"
	}
	truncReq := req
	truncReq.Body = longBody
	trunc := buildIndexEntry(truncReq, nil)
	wantBody := ""
	for i := 0; i < 80; i++ {
		wantBody += "x"
	}
	wantBody += "..."
	if trunc.Body != wantBody {
		t.Errorf("truncated Body = %q (len %d), want len %d", trunc.Body, len(trunc.Body), len(wantBody))
	}
}

func TestSortedStatusValues(t *testing.T) {
	cases := []struct {
		name      string
		values    []string
		additions []string
		want      []string
	}{
		{
			name:   "only builtins present",
			values: []string{"approved", "draft"},
			want:   []string{"draft", "approved"},
		},
		{
			name:      "builtins + additions",
			values:    []string{"approved", "review", "draft"},
			additions: []string{"review"},
			want:      []string{"draft", "approved", "review"},
		},
		{
			name:   "builtins + unrecognized, sorted A→Z",
			values: []string{"approved", "zeta", "draft", "alpha"},
			want:   []string{"draft", "approved", "alpha", "zeta"},
		},
		{
			name:   "empty input",
			values: nil,
			want:   []string{},
		},
		{
			name:      "additions with no matching values",
			values:    []string{"draft"},
			additions: []string{"review", "pending"},
			want:      []string{"draft"},
		},
		{
			name:   "unrecognized only",
			values: []string{"zeta", "alpha"},
			want:   []string{"alpha", "zeta"},
		},
		{
			name:      "additions come before unrecognized",
			values:    []string{"zeta", "review", "draft", "alpha"},
			additions: []string{"review"},
			want:      []string{"draft", "review", "alpha", "zeta"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sortedStatusValues(tc.values, tc.additions)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d (got %v, want %v)", len(got), len(tc.want), got, tc.want)
			}
			for i, v := range tc.want {
				if got[i] != v {
					t.Errorf("[%d] got %q, want %q (full: %v)", i, got[i], v, got)
				}
			}
		})
	}
}
