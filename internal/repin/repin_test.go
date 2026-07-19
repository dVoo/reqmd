package repin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"reqmd/internal/graph"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// writeFile creates a temp dir, writes content, and returns the dir path
// (caller is responsible for cleanup via t.TempDir()).
func writeFile(t *testing.T, dir, rel, content string) string {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return p
}

// minimalSchema is a permissive JSON Schema for attr blocks. Tests
// only need the schema to exist (parser requires it); they don't
// validate the attrs strictly.
const minimalSchema = `type: object
`

// makeTree creates a two-doc spec tree with one upstream (versioned
// at `upVersion`) and one downstream that traces to it with the
// supplied pin (or no pin when pin is ""). The downstream's trace
// line is rendered exactly as supplied so we can assert byte-level
// edits in Apply.
func makeTree(t *testing.T, upVersion int, downTrace string) string {
	t.Helper()
	root := t.TempDir()
	// Upstream doc
	writeFile(t, root, "upstream/schema.yaml", minimalSchema)
	writeFile(t, root, "upstream/up.md", "## UP-001\n```attr\ntitle: Up\nversion: "+
		strconv.Itoa(upVersion)+"\n```\nThe upstream.\n")
	// Downstream doc
	writeFile(t, root, "downstream/schema.yaml", minimalSchema)
	writeFile(t, root, "downstream/dn.md", "## DN-001\n```attr\ntitle: Down\ntrace: ["+downTrace+"]\n```\nThe downstream.\n")
	return root
}

// ---------------------------------------------------------------------------
// Build
// ---------------------------------------------------------------------------

func TestBuild_Outdated(t *testing.T) {
	root := makeTree(t, 3, "UP-001~1")
	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.Deltas) != 1 || res.Deltas[0].Kind != "outdated" {
		t.Fatalf("Build = %+v, want 1 outdated delta", res)
	}
	if res.Outdated != 1 || res.Unpinned != 0 || res.Predated != 0 {
		t.Errorf("counts = (outdated=%d, unpinned=%d, predated=%d), want (1,0,0)",
			res.Outdated, res.Unpinned, res.Predated)
	}
}

func TestBuild_PromoteUnpinned(t *testing.T) {
	root := makeTree(t, 4, "UP-001")
	res, err := Build(root, true)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.Deltas) != 1 || res.Deltas[0].Kind != "unpinned" {
		t.Fatalf("Build = %+v, want 1 unpinned delta", res)
	}
	if res.Unpinned != 1 {
		t.Errorf("Unpinned = %d, want 1", res.Unpinned)
	}
}

func TestBuild_NoPromoteUnpinned(t *testing.T) {
	root := makeTree(t, 4, "UP-001")
	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.Deltas) != 0 {
		t.Errorf("Build = %+v, want no deltas (unpinned skipped without promote)", res)
	}
}

func TestBuild_CurrentIsNoOp(t *testing.T) {
	root := makeTree(t, 3, "UP-001~3")
	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.Deltas) != 0 {
		t.Errorf("Build = %+v, want empty (pin == upstream.version)", res)
	}
}

// ---------------------------------------------------------------------------
// Apply: file rewrite semantics
// ---------------------------------------------------------------------------

func TestApply_RewritesAttrBlockOnly(t *testing.T) {
	root := makeTree(t, 3, "UP-001~1")
	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	rep, err := Apply(res.Deltas)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rep.Applied != 1 {
		t.Fatalf("Applied = %d, want 1", rep.Applied)
	}
	dnPath := filepath.Join(root, "downstream/dn.md")
	got, err := os.ReadFile(dnPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	gotStr := string(got)
	if !strings.Contains(gotStr, "trace: [UP-001~3]") {
		t.Errorf("downstream file missing updated trace, got:\n%s", gotStr)
	}
	if strings.Contains(gotStr, "UP-001~1") {
		t.Errorf("downstream file still contains old pin UP-001~1, got:\n%s", gotStr)
	}
	// Prose body must be preserved.
	if !strings.Contains(gotStr, "The downstream.") {
		t.Errorf("downstream file lost prose body, got:\n%s", gotStr)
	}
	// Upstream file must be untouched.
	upPath := filepath.Join(root, "upstream/up.md")
	upGot, _ := os.ReadFile(upPath)
	if !strings.Contains(string(upGot), "The upstream.") {
		t.Errorf("upstream file unexpectedly modified:\n%s", string(upGot))
	}
}

func TestApply_PromoteAddsPin(t *testing.T) {
	root := makeTree(t, 5, "UP-001")
	res, err := Build(root, true)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := Apply(res.Deltas); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "downstream/dn.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(got), "trace: [UP-001~5]") {
		t.Errorf("expected pin to be promoted to ~5, got:\n%s", string(got))
	}
}

func TestApply_PreservesProseWithTilde(t *testing.T) {
	// Defensive: ensure strings like "in case ~0 cases" in prose are
	// NOT touched. We put a tilde-bearing phrase in the prose body
	// outside the attr block; Apply must not corrupt it.
	root := t.TempDir()
	writeFile(t, root, "upstream/schema.yaml", minimalSchema)
	writeFile(t, root, "upstream/up.md", "## UP-001\n```attr\ntitle: Up\nversion: 3\n```\n")
	writeFile(t, root, "downstream/schema.yaml", minimalSchema)
	writeFile(t, root, "downstream/dn.md", strings.Join([]string{
		"## DN-001",
		"```attr",
		"title: Down",
		"trace: [UP-001~1]",
		"```",
		"Prose notes about upstream: in case ~0 cases the pin changes.",
		"",
		"*Rationale:* ~0 is the safe default.",
	}, "\n"))

	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := Apply(res.Deltas); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "downstream/dn.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	s := string(got)
	if !strings.Contains(s, "in case ~0 cases") {
		t.Errorf("prose 'in case ~0 cases' was modified; got:\n%s", s)
	}
	if !strings.Contains(s, "~0 is the safe default") {
		t.Errorf("prose '~0 is the safe default' was modified; got:\n%s", s)
	}
	if !strings.Contains(s, "trace: [UP-001~3]") {
		t.Errorf("attr block trace not updated, got:\n%s", s)
	}
}

func TestApply_Idempotent(t *testing.T) {
	root := makeTree(t, 3, "UP-001~1")
	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := Apply(res.Deltas); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	// Second Build should produce no deltas — pin now == upstream.
	res2, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build (post-apply): %v", err)
	}
	if len(res2.Deltas) != 0 {
		t.Errorf("second Build = %+v, want empty (idempotent)", res2)
	}
}

// TestApply_UnpinnedPromoteDoesNotCorruptPinnedSiblings guards
// against a real bug discovered in the smoke test: a naive substring
// match for SourceRef="UP-001" would rewrite the prefix of an
// already-pinned ref "UP-001~3" to "UP-001~3~3". The apply path
// must use whole-ref matching so an unpinned ref never corrupts a
// pinned sibling.
func TestApply_UnpinnedPromoteDoesNotCorruptPinnedSiblings(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "upstream/schema.yaml", minimalSchema)
	writeFile(t, root, "upstream/up.md", "## UP-001\n```attr\nversion: 3\n```\n")
	writeFile(t, root, "downstream/schema.yaml", minimalSchema)
	writeFile(t, root, "downstream/dn.md", strings.Join([]string{
		"## DN-001",
		"```attr",
		"title: Already pinned",
		"trace: [UP-001~3, UP-002~9]",
		"```",
		"## DN-002",
		"```attr",
		"title: Unpinned",
		"trace: [UP-001, UP-002]",
		"```",
	}, "\n"))
	// Add a second upstream to give the unpinned refs a real target.
	writeFile(t, root, "upstream/up2.md", "## UP-002\n```attr\nversion: 9\n```\n")

	// Without --promote, no changes (everything is current or pinned).
	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.Deltas) != 0 {
		t.Fatalf("Build = %+v, want empty", res)
	}

	// With --promote, the unpinned refs become ~N. The pinned refs
	// must remain untouched: no `~3~3` or `~5~9` corruption.
	res, err = Build(root, true)
	if err != nil {
		t.Fatalf("Build promote: %v", err)
	}
	if _, err := Apply(res.Deltas); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "downstream/dn.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	s := string(got)
	// Pinned refs must remain exactly as they were.
	if !strings.Contains(s, "trace: [UP-001~3, UP-002~9]") {
		t.Errorf("pinned refs corrupted; got:\n%s", s)
	}
	// Unpinned refs must now be pinned to upstream version.
	// DN-002's trace (originally UP-001, UP-002) must now be pinned.
	if !strings.Contains(s, "title: Unpinned\ntrace: [UP-001~3, UP-002~9]") {
		t.Errorf("unpinned refs not promoted correctly; got:\n%s", s)
	}
	// Defensive: no doubled `~`s anywhere.
	if strings.Contains(s, "~~") {
		t.Errorf("double-tilde corruption detected; got:\n%s", s)
	}
}

func TestApply_SkipsPredated(t *testing.T) {
	root := makeTree(t, 2, "UP-001~9")
	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.Deltas) != 1 || res.Deltas[0].Kind != "predated" {
		t.Fatalf("Build = %+v, want 1 predated delta", res)
	}
	rep, err := Apply(res.Deltas)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rep.Applied != 0 {
		t.Errorf("Applied = %d, want 0 (predated never auto-fixed)", rep.Applied)
	}
	if len(rep.Skipped) != 1 {
		t.Errorf("Skipped = %d, want 1", len(rep.Skipped))
	}
	// File must be unchanged.
	got, _ := os.ReadFile(filepath.Join(root, "downstream/dn.md"))
	if !strings.Contains(string(got), "UP-001~9") {
		t.Errorf("predated pin was rewritten, got:\n%s", string(got))
	}
}

func TestApply_MultipleDeltasSameFile(t *testing.T) {
	// Two requirements in the same file, both outdated against the
	// same upstream at version 5.
	root := t.TempDir()
	writeFile(t, root, "upstream/schema.yaml", minimalSchema)
	writeFile(t, root, "upstream/up.md", "## UP-001\n```attr\ntitle: Up\nversion: 5\n```\n")
	writeFile(t, root, "downstream/schema.yaml", minimalSchema)
	writeFile(t, root, "downstream/dn.md", strings.Join([]string{
		"## DN-001",
		"```attr",
		"title: First",
		"trace: [UP-001~1]",
		"```",
		"## DN-002",
		"```attr",
		"title: Second",
		"trace: [UP-001~2]",
		"```",
	}, "\n"))
	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.Deltas) != 2 {
		t.Fatalf("Build = %+v, want 2 deltas", res)
	}
	rep, err := Apply(res.Deltas)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if rep.Applied != 2 {
		t.Errorf("Applied = %d, want 2", rep.Applied)
	}
	if rep.Files != 1 {
		t.Errorf("Files = %d, want 1", rep.Files)
	}
	got, _ := os.ReadFile(filepath.Join(root, "downstream/dn.md"))
	s := string(got)
	if strings.Contains(s, "UP-001~1") || strings.Contains(s, "UP-001~2") {
		t.Errorf("old pins remain, got:\n%s", s)
	}
	if !strings.Contains(s, "UP-001~5") {
		t.Errorf("new pin ~5 not present, got:\n%s", s)
	}
}

// ---------------------------------------------------------------------------
// newRefFrom
// ---------------------------------------------------------------------------

func TestNewRefFrom(t *testing.T) {
	cases := []struct {
		ref       string
		newPin    int
		want      string
	}{
		{"UP-001~1", 3, "UP-001~3"},
		{"upstream/UP-001~1", 3, "upstream/UP-001~3"},
		{"UP-001", 5, "UP-001~5"},         // unpinned promote
		{"upstream/UP-001", 5, "upstream/UP-001~5"}, // qualified unpinned promote
		{"UP-001~0", 4, "UP-001~4"},       // ~0 → ~4
	}
	for _, c := range cases {
		d := graph.RepinDelta{SourceRef: c.ref, NewPin: c.newPin}
		got := newRefFrom(d)
		if got != c.want {
			t.Errorf("newRefFrom(%q, pin=%d) = %q, want %q", c.ref, c.newPin, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// locateAttrBlocks
// ---------------------------------------------------------------------------

func TestLocateAttrBlocks(t *testing.T) {
	src := []byte(strings.Join([]string{
		"# preamble",
		"```attr",
		"a: 1",
		"```",
		"prose",
		"## heading",
		"```attr",
		"b: 2",
		"```",
		"more prose",
	}, "\n"))
	ranges := locateAttrBlocks(src)
	if len(ranges) != 2 {
		t.Fatalf("locateAttrBlocks = %d, want 2", len(ranges))
	}
	// First block: bytes covering "a: 1"
	first := string(src[ranges[0].start:ranges[0].end])
	if !strings.Contains(first, "a: 1") {
		t.Errorf("first block content = %q, want contains 'a: 1'", first)
	}
	// Second block: bytes covering "b: 2"
	second := string(src[ranges[1].start:ranges[1].end])
	if !strings.Contains(second, "b: 2") {
		t.Errorf("second block content = %q, want contains 'b: 2'", second)
	}
}

func TestLocateAttrBlocks_NoBlocks(t *testing.T) {
	ranges := locateAttrBlocks([]byte("no fences here, just prose\n"))
	if len(ranges) != 0 {
		t.Errorf("locateAttrBlocks = %d, want 0", len(ranges))
	}
}

func TestLocateAttrBlocks_IgnoresFencedButNotAttr(t *testing.T) {
	// A ```python fence must not be picked up.
	src := []byte("```python\nprint('hi')\n```\n")
	ranges := locateAttrBlocks(src)
	if len(ranges) != 0 {
		t.Errorf("locateAttrBlocks picked up non-attr fence: %d ranges", len(ranges))
	}
}

// ---------------------------------------------------------------------------
// FormatText
// ---------------------------------------------------------------------------

func TestFormatText_Empty(t *testing.T) {
	got := FormatText(Result{ByFile: map[string]int{}})
	if got != "no version-pin changes needed\n" {
		t.Errorf("FormatText(empty) = %q", got)
	}
}

func TestFormatText_Outdated(t *testing.T) {
	res := Result{
		Deltas: []graph.RepinDelta{
			{File: "/x/a.md", ReqID: "A-1", SourceRef: "UP-001~1", NewPin: 3},
		},
		ByFile:   map[string]int{"/x/a.md": 1},
		Outdated: 1,
	}
	got := FormatText(res)
	want := "/x/a.md  A-1  UP-001 → UP-001~3\n1 outdated, 0 unpinned, 0 predated across 1 files.\n"
	if got != want {
		t.Errorf("FormatText = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// JSON output
// ---------------------------------------------------------------------------

func TestResultJSON_Roundtrip(t *testing.T) {
	res := Result{
		Deltas: []graph.RepinDelta{
			{Kind: "outdated", ReqID: "DN-001", File: "/x/dn.md", TargetID: "UP-001", SourceRef: "UP-001~1", OldPin: 1, NewPin: 3, NewVersion: 3},
		},
		ByFile:   map[string]int{"/x/dn.md": 1},
		Outdated: 1,
	}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// Stable field presence — protects the documented JSON schema.
	for _, key := range []string{`"kind":"outdated"`, `"req_id":"DN-001"`, `"target_id":"UP-001"`, `"source_ref":"UP-001~1"`, `"old_pin":1`, `"new_pin":3`, `"new_version":3`, `"outdated":1`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON missing key %q in: %s", key, string(b))
		}
	}
	// Roundtrip.
	var got Result
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(got.Deltas) != 1 || got.Deltas[0].NewPin != 3 {
		t.Errorf("roundtrip = %+v, want NewPin=3", got)
	}
	if got.Outdated != 1 {
		t.Errorf("roundtrip Outdated = %d, want 1", got.Outdated)
	}
}

// ---------------------------------------------------------------------------
// Sort order — repin reuses graph's sort; this is a regression test
// to ensure Build preserves it.
// ---------------------------------------------------------------------------

func TestBuild_SortedByFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "upstream/schema.yaml", minimalSchema)
	writeFile(t, root, "upstream/up.md", "## UP-001\n```attr\nversion: 5\n```\n")
	// Two downstream docs in non-alphabetical order to verify sort.
	writeFile(t, root, "z/schema.yaml", minimalSchema)
	writeFile(t, root, "z/zz.md", "## Z-1\n```attr\ntrace: [UP-001~1]\n```\n")
	writeFile(t, root, "a/schema.yaml", minimalSchema)
	writeFile(t, root, "a/aa.md", "## A-1\n```attr\ntrace: [UP-001~2]\n```\n")

	res, err := Build(root, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.Deltas) != 2 {
		t.Fatalf("Build = %+v, want 2 deltas", res)
	}
	if !sort.StringsAreSorted([]string{filepath.Base(res.Deltas[0].File), filepath.Base(res.Deltas[1].File)}) {
		// Just confirm /a/ comes before /z/ via a path comparison.
		if res.Deltas[0].File > res.Deltas[1].File {
			t.Errorf("deltas not sorted by file: %v", []string{res.Deltas[0].File, res.Deltas[1].File})
		}
	}
}
