package exporter

import (
	"reqmd/internal/model"
	"strings"
	"testing"
)

func TestBuildRenderData(t *testing.T) {
	doc := model.Document{
		Nodes: []*model.Node{
			{
				ID:    "REQ-001",
				Title: "First",
				Attrs: map[string]any{
					"status": "approved",
					"asil":   "B",
				},
				Body: "First body.",
				Children: []*model.Node{
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
				Kind:  model.KindContainer,
				Title: "Section",
			},
		},
	}

	rd := buildRenderData(doc)
	if rd == nil {
		t.Fatal("buildRenderData returned nil")
	}

	// Root nodes: REQ-001 (with child), REQ-002, Section.
	if len(rd.Nodes) != 3 {
		t.Errorf("Nodes: got %d, want 3", len(rd.Nodes))
	}

	// Anchors: requirements anchor on their ID, items on info-N.
	if rd.Anchors[doc.Nodes[0]] != "REQ-001" {
		t.Errorf("anchor[REQ-001] = %q, want REQ-001", rd.Anchors[doc.Nodes[0]])
	}
	if rd.Anchors[doc.Nodes[2]] != "info-1" {
		t.Errorf("anchor[Section] = %q, want info-1", rd.Anchors[doc.Nodes[2]])
	}

	// Index: 4 entries (roots + child REQ-003), tree-linked.
	if len(rd.Index) != 4 {
		t.Fatalf("Index len = %d, want 4", len(rd.Index))
	}
	if len(rd.Index[0].Children) != 1 || rd.Index[0].Children[0] != "REQ-003" {
		t.Errorf("Index[0].Children = %v, want [REQ-003]", rd.Index[0].Children)
	}
	if rd.Index[1].ParentID != "REQ-001" {
		t.Errorf("Index[1].ParentID = %q, want REQ-001", rd.Index[1].ParentID)
	}
	if rd.Index[3].Type != "container" {
		t.Errorf("Index[3].Type = %q, want container", rd.Index[3].Type)
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
		Nodes: []*model.Node{
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
	kid1 := &model.Node{ID: "REQ-007-1"}
	kid2 := &model.Node{ID: "REQ-007-2"}
	req := &model.Node{
		ID:        "REQ-007",
		Title:     "A title",
		Attrs:     map[string]any{"status": "approved", "asil": "D", "trace": []string{"X-1"}},
		Body:      "body",
		Rationale: "why",
		Children:  []*model.Node{kid1, kid2},
	}
	anchors := map[*model.Node]string{
		req:  "REQ-007",
		kid1: "REQ-007-1",
		kid2: "REQ-007-2",
	}
	entry := buildIndexEntry(req, "REQ-007", "", anchors)

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

	// Children come from the tree, resolved via the anchor map.
	wantKids := []string{"REQ-007-1", "REQ-007-2"}
	if len(entry.Children) != len(wantKids) {
		t.Fatalf("Children len = %d, want %d", len(entry.Children), len(wantKids))
	}
	for i, k := range wantKids {
		if entry.Children[i] != k {
			t.Errorf("Children[%d] = %q, want %q", i, entry.Children[i], k)
		}
	}

	// Parent anchor is recorded.
	childEntry := buildIndexEntry(kid1, "REQ-007-1", "REQ-007", anchors)
	if childEntry.ParentID != "REQ-007" {
		t.Errorf("child ParentID = %q, want REQ-007", childEntry.ParentID)
	}

	// Body truncation at 80 runes.
	longBody := strings.Repeat("x", 100)
	truncReq := req
	truncReq.Body = longBody
	trunc := buildIndexEntry(truncReq, "REQ-007", "", anchors)
	wantBody := strings.Repeat("x", 80) + "..."
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

func TestItemHeadingTagNativeLevels(t *testing.T) {
	cases := []struct {
		want  string
		level int
	}{
		{level: 1, want: "h2"}, // defensive: h1 is the document title, never content
		{level: 2, want: "h2"},
		{level: 3, want: "h3"},
		{level: 4, want: "h4"},
		{level: 6, want: "h6"},
	}
	for _, tc := range cases {
		if got := itemHeadingTag(tc.level); got != tc.want {
			t.Errorf("itemHeadingTag(%d) = %q, want %q", tc.level, got, tc.want)
		}
	}
}

func TestDocTitleFallback(t *testing.T) {
	cases := []struct {
		name string
		want string
		doc  model.Document
	}{
		{name: "parsed h1 title wins", doc: model.Document{Title: "From h1", Schema: map[string]any{"title": "Schema"}}, want: "From h1"},
		{name: "schema title fallback", doc: model.Document{Schema: map[string]any{"title": "Schema"}}, want: "Schema"},
		{name: "empty", doc: model.Document{}, want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := docTitle(tc.doc); got != tc.want {
				t.Errorf("docTitle = %q, want %q", got, tc.want)
			}
		})
	}
}
