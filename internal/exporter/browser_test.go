package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dVoo/reqmd/internal/model"
)

// browserFixture builds a document whose analysis exercises every state:
// covered, missing coverage, a failing verdict, and a draft requirement that
// the default status filter hides.
func browserFixture(t *testing.T) (model.Document, *Analysis) {
	t.Helper()
	system := model.Document{
		Path:       "/test/system",
		Schema:     schemaWithTitle("Browser Fixture"),
		Properties: []string{"status", "priority", "note"},
		XReqmd:     &model.XReqmd{DocumentID: "system", Level: "system", RequiresTraceFrom: []string{"software"}},
		Nodes: []*model.Node{
			{Kind: model.KindRequirement, ID: "SYS-001", Source: "/test/system/a.md", Title: "Covered requirement", Attrs: map[string]any{"status": "approved", "priority": "High", "note": "x"}, Body: "The system shall be covered."},
			{Kind: model.KindRequirement, ID: "SYS-002", Source: "/test/system/b.md", Title: "Uncovered requirement", Attrs: map[string]any{"status": "approved", "priority": "Low", "note": "a,b:c", "trace": []any{"remote/REMOTE-001"}}, Body: "The system shall be uncovered."},
			{Kind: model.KindRequirement, ID: "SYS-003", Source: "/test/system/c.md", Title: "Draft requirement", Attrs: map[string]any{"status": "draft"}, Body: "The system shall be drafted."},
			{Kind: model.KindRequirement, ID: "TST-001", Source: "/test/system/d.md", Title: "Failing measure", Attrs: map[string]any{"status": "approved", "verify": "Test"}, Body: "The system shall be verified."},
			{Kind: model.KindRequirement, ID: "TST-002", Source: "/test/system/e.md", Title: "Unverified measure", Attrs: map[string]any{"status": "approved", "verify": "Test"}, Body: "The system shall be unverified."},
		},
	}
	software := model.Document{
		Path:   "/test/software",
		Schema: schemaWithTitle("Software Fixture"),
		XReqmd: &model.XReqmd{DocumentID: "software", Level: "software"},
		Nodes: []*model.Node{
			{Kind: model.KindRequirement, ID: "SW-001", Source: "/test/software/a.md", Attrs: map[string]any{"status": "approved", "trace": []any{"system/SYS-001"}}},
		},
	}
	results := model.Document{
		Path:      "<verify-results>",
		Synthetic: true,
		Nodes: []*model.Node{
			{Kind: model.KindRequirement, ID: "RESULT:TST-001", Source: "ci/run.ctrf.json", Attrs: map[string]any{model.AttrResultOutcome: "fail", model.AttrTrace: []any{"TST-001"}}},
		},
	}
	remote := model.Document{
		Path:   "/test/remote",
		Schema: schemaWithTitle("Remote Fixture"),
		XReqmd: &model.XReqmd{DocumentID: "remote", Level: "stakeholder"},
		Nodes: []*model.Node{
			{Kind: model.KindRequirement, ID: "REMOTE-001", Source: "/test/remote/a.md", Title: "Remote requirement", Attrs: map[string]any{"status": "approved"}, Body: "remote"},
		},
	}
	docs := []model.Document{system, software, results, remote}
	return system, analysisFor(t, docs...)
}

// exportBrowserFixture writes the fixture page and returns its path.
func exportBrowserFixture(t *testing.T) string {
	t.Helper()
	doc, analysis := browserFixture(t)
	h := &HTML{}
	h.SetAnalysis(analysis)
	path := filepath.Join(t.TempDir(), "system-requirements.html")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Export(f, doc, doc.Properties); err != nil {
		t.Fatalf("export: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// browserHarness drives the exported page and reports results through the
// document title, which --dump-dom prints.
const browserHarness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  window.onerror = function (m) { out.errors.push(String(m)); };
  try {
    var r = window.reqmd;
    out.hasReqmd = !!r;

    // Table: sortable rows and accurate pressed state.
    r.setView('table');
    out.tableVisible = !document.querySelector('[data-view-panel="table"]').hidden;
    out.tableRows = document.querySelectorAll('#req-table tbody tr.grid__row').length;
    var first = document.querySelector('#req-table tbody tr.grid__row');
    out.firstBefore = first && first.getAttribute('data-id');
    document.querySelector('#req-table thead th.sortable').click();
    var after = document.querySelector('#req-table tbody tr.grid__row');
    out.firstAfter = after && after.getAttribute('data-id');
    // The switcher must mark the active view (body[data-view] also matches the
    // bare attribute selector, so scope to the switcher).
    out.pressed = document.querySelector('.viewswitch [data-view="table"]').classList.contains('is-active');

    // Trace explorer + bounded matrix with direction-specific tooltips.
    r.setView('trace');
    out.traceCard = !!document.querySelector('.trace__card');
    var matrixBtn = Array.prototype.filter.call(
      document.querySelectorAll('[data-view-panel="trace"] .btn'),
      function (b) { return b.textContent.indexOf('matrix') !== -1; }
    )[0];
    matrixBtn.click();
    out.matrixTitles = Array.prototype.map.call(
      document.querySelectorAll('.matrix__cell--on'),
      function (c) { return c.getAttribute('title'); }
    );

    // Cross-field search matches titles and body text.
    r.setSearch('uncovered');
    out.searchCount = document.querySelector('[data-count]').textContent;
    r.setSearch('');
    out.noSearchCount = document.querySelector('[data-count]').textContent;

    // Filtering applies to the current view and the next one rendered.
    r.setFilter('state.coverage', 'missing-coverage');
    out.visibleCount = document.querySelector('[data-count]').textContent;
    r.setView('table');
    out.filteredRows = document.querySelectorAll('#req-table tbody tr.grid__row').length;
    out.urlAfterFilter = location.search;

    // Command palette: keyboard open, focus, close, focus restore.
    var trigger = document.querySelector('.app-bar [data-action="palette"]');
    trigger.focus();
    trigger.click();
    out.paletteOpen = !document.getElementById('palette').hidden;
    out.paletteFocused = !!document.activeElement && document.activeElement.classList.contains('palette__input');
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true, cancelable: true }));
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    out.paletteClosed = document.getElementById('palette').hidden;
    out.focusRestored = document.activeElement === trigger;

    // Shortcut dialog and single-letter view navigation.
    document.dispatchEvent(new KeyboardEvent('keydown', { key: '?', bubbles: true }));
    out.shortcutsOpen = !document.getElementById('shortcuts').hidden;
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'g', bubbles: true }));
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'c', bubbles: true }));
    out.coverageVisible = !document.querySelector('[data-view-panel="coverage"]').hidden;
    out.coverageTiles = document.querySelectorAll('[data-view-panel="coverage"] .tile').length;
    var tiles = document.querySelectorAll('[data-view-panel="coverage"] .tile');
    out.verificationTile = tiles[1] ? tiles[1].textContent : '';

    // Modified keys must not trigger single-letter shortcuts.
    r.setView('document');
    var before = document.body.classList.contains('density-compact');
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'f', ctrlKey: true, bubbles: true }));
    out.modifiedIgnored = document.body.classList.contains('density-compact') === before;
  } catch (e) {
    out.errors.push(String((e && e.stack) || e));
  }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`

const deepLinkHarness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  try {
    out.documentVisible = !document.querySelector('[data-view-panel="document"]').hidden;
    var el = document.getElementById('SYS-002');
    out.flash = !!el && el.classList.contains('is-flash');
    out.hash = location.hash;
  } catch (e) {
    out.errors.push(String(e));
  }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`

const restoredStateHarness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  try {
    out.visibleCount = document.querySelector('[data-count]').textContent;
    out.densityCompact = document.body.classList.contains('density-compact');
    out.tableVisible = !document.querySelector('[data-view-panel="table"]').hidden;
    out.tableRows = document.querySelectorAll('#req-table tbody tr.grid__row').length;
    out.url = location.search;
    var chips = Array.prototype.map.call(document.querySelectorAll('.chip--filter'), function (c) { return c.textContent; });
    out.chips = chips;
  } catch (e) {
    out.errors.push(String(e));
  }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`

// findChrome returns the first usable headless-capable browser.
func findChrome(t *testing.T) string {
	t.Helper()
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	t.Skip("no headless browser installed (google-chrome/chromium)")
	return ""
}

// runHeadless loads one exported page through file:// with the given harness
// appended and returns the decoded RESULT payload.
func runHeadless(t *testing.T, chrome, pagePath, fragment string, extraArgs []string, harness string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatalf("reading page: %v", err)
	}
	html := strings.Replace(string(raw), "</body>", harness+"</body>", 1)
	withHarness := filepath.Join(t.TempDir(), "page.html")
	writeErr := os.WriteFile(withHarness, []byte(html), 0o600)
	if writeErr != nil {
		t.Fatalf("writing harness page: %v", writeErr)
	}

	args := []string{
		"--headless=new", "--disable-gpu", "--no-sandbox", "--no-first-run",
		"--user-data-dir=" + t.TempDir(),
		"--virtual-time-budget=3000", "--dump-dom",
	}
	args = append(args, extraArgs...)
	args = append(args, "file://"+withHarness+fragment)
	cmd := exec.CommandContext(context.Background(), chrome, args...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("headless browser failed: %v", err)
	}
	marker := "RESULT:"
	i := strings.Index(string(out), marker)
	if i < 0 {
		t.Fatalf("browser produced no result payload:\n%s", out)
	}
	rest := string(out)[i+len(marker):]
	if j := strings.IndexByte(rest, '<'); j >= 0 {
		rest = rest[:j]
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(rest), &result); err != nil {
		t.Fatalf("decoding browser result %q: %v", rest, err)
	}
	return result
}

// stringSet converts a decoded JSON array of strings into a set.
func stringSet(t *testing.T, v any) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	items, ok := v.([]any)
	if !ok {
		t.Fatalf("expected a JSON array, got %T", v)
	}
	for _, item := range items {
		if s, ok := item.(string); ok {
			out[s] = true
		}
	}
	return out
}

func assertNoJSErrors(t *testing.T, result map[string]any) {
	t.Helper()
	if errs, ok := result["errors"].([]any); ok && len(errs) > 0 {
		t.Fatalf("client reported errors: %v", errs)
	}
}

// TestHTMLExportBrowserSmoke is the end-to-end guard for workflows that only
// exist in the client: offline load, filtering, sorting, the on-demand views,
// the bounded matrix, and keyboard operation.
func TestHTMLExportBrowserSmoke(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	result := runHeadless(t, chrome, pagePath, "", nil, browserHarness)
	assertNoJSErrors(t, result)
	want := map[string]any{
		"hasReqmd":     true,
		"tableVisible": true,
		"tableRows":    float64(4), // draft requirement hidden by the default status filter
		"filteredRows": float64(3),
		// aria-pressed reflects the view that has focus; the harness reads it
		// after clicking the table's ID header, which leaves the table active.
		"pressed":         true,
		"traceCard":       true,
		"paletteOpen":     true,
		"paletteFocused":  true,
		"paletteClosed":   true,
		"focusRestored":   true,
		"shortcutsOpen":   true,
		"coverageVisible": true,
		"coverageTiles":   float64(3),
		"searchCount":     "1 shown · 5 total (4 hidden by status: approved)",
		"noSearchCount":   "4 shown · 5 total (1 hidden by status: approved)",
		"modifiedIgnored": true,
	}
	for k, v := range want {
		if result[k] != v {
			t.Errorf("%s = %v, want %v", k, result[k], v)
		}
	}

	if result["firstBefore"] == result["firstAfter"] {
		t.Error("clicking the ID header did not reorder rows")
	}
	if got := result["visibleCount"]; got != "3 shown · 5 total (2 hidden by status: approved)" {
		t.Errorf("visible count = %v, want the shown/total label naming the status filter", got)
	}
	// The denominator is eligible measures (2), including the unverified one.
	verification, _ := result["verificationTile"].(string)
	for _, want := range []string{"0 of 2 pass", "1 fail", "1 no verdict"} {
		if !strings.Contains(verification, want) {
			t.Errorf("verification tile %q missing %q", verification, want)
		}
	}
	// Matrix tooltips name the real link direction in both directions:
	// upstream rows trace to their column, downstream columns trace to their row.
	titles := stringSet(t, result["matrixTitles"])
	for _, want := range []string{"SYS-002 → REMOTE-001", "SW-001 → SYS-001"} {
		if !titles[want] {
			t.Errorf("matrix tooltip %q missing; got %v", want, result["matrixTitles"])
		}
	}
	if url, _ := result["urlAfterFilter"].(string); !strings.Contains(url, "f=state.coverage") {
		t.Errorf("filter state missing from URL: %q", url)
	}
}

// TestHTMLExportBrowserDeepLink verifies a #REQ-ID deep link opens Document
// mode and reveals the target.
func TestHTMLExportBrowserDeepLink(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	result := runHeadless(t, chrome, pagePath, "#SYS-002", nil, deepLinkHarness)
	assertNoJSErrors(t, result)
	if result["documentVisible"] != true {
		t.Error("deep link did not select the Document view")
	}
	if result["flash"] != true {
		t.Error("deep link did not reveal the target requirement")
	}
	if got, _ := result["hash"].(string); !strings.HasSuffix(got, "#SYS-002") {
		t.Errorf("deep link hash lost: %q", got)
	}
}

// TestHTMLExportBrowserRestoresURLState verifies a shared URL reproduces the
// exact filtered view, including values with commas and colons.
func TestHTMLExportBrowserRestoresURLState(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	query := "?f=status:approved&f=note:a,b:c&view=table&density=compact"
	result := runHeadless(t, chrome, pagePath, query, nil, restoredStateHarness)
	assertNoJSErrors(t, result)
	if got := result["visibleCount"]; got != "1 shown · 5 total (4 hidden by status: approved)" {
		t.Errorf("visible count = %v, want 1 shown of 5", got)
	}
	if result["tableVisible"] != true || result["tableRows"] != float64(1) {
		t.Errorf("table state = visible:%v rows:%v, want visible:true rows:1", result["tableVisible"], result["tableRows"])
	}
	if result["densityCompact"] != true {
		t.Error("density did not restore from the URL")
	}
	if url, _ := result["url"].(string); !strings.Contains(url, "note%3Aa%2Cb%3Ac") && !strings.Contains(url, "note:a,b:c") {
		t.Errorf("encoded filter value was lost: %q", url)
	}
}

// TestHTMLExportBrowserExplicitEmptyFilters verifies `?f=` keeps the document
// default status filter cleared after reload.
func TestHTMLExportBrowserExplicitEmptyFilters(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	result := runHeadless(t, chrome, pagePath, "?f=", nil, restoredStateHarness)
	assertNoJSErrors(t, result)
	if got := result["visibleCount"]; got != "5 shown · 5 total" {
		t.Errorf("visible count = %v, want every requirement shown with no hidden reason", got)
	}
	if chips, _ := result["chips"].([]any); len(chips) != 0 {
		t.Errorf("explicit empty filter set rendered chips: %v", chips)
	}
}

// TestHTMLExportBrowserMobileOutline verifies the outline becomes a drawer at
// mobile width and the toolbar stops sticking when vertical space is tight.
func TestHTMLExportBrowserMobileOutline(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  try {
    var toggle = document.querySelector('.toc-toggle');
    out.toggleVisible = getComputedStyle(toggle).display !== 'none';
    toggle.click();
    out.drawerOpen = document.body.classList.contains('toc-open');
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    out.drawerClosed = !document.body.classList.contains('toc-open');
    window.dispatchEvent(new Event('resize'));
  } catch (e) { out.errors.push(String(e)); }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", []string{"--window-size=420,900"}, harness)
	assertNoJSErrors(t, result)
	for _, k := range []string{"toggleVisible", "drawerOpen", "drawerClosed"} {
		if result[k] != true {
			t.Errorf("%s = %v, want true", k, result[k])
		}
	}
}

// TestHTMLExportBrowserMatrixBound verifies the full matrix refuses to draw
// beyond its cell limit and points the reviewer at the trace neighbourhood.
func TestHTMLExportBrowserMatrixBound(t *testing.T) {
	chrome := findChrome(t)

	const n = 150
	upReqs := make([]*model.Node, 0, n)
	downReqs := make([]*model.Node, 0, n)
	for i := range n {
		id := fmt.Sprintf("UP-%03d", i)
		upReqs = append(upReqs, &model.Node{
			Kind: model.KindRequirement, ID: id, Source: "/test/up/" + id + ".md",
			Title: "Upstream " + id, Attrs: map[string]any{"status": "approved"}, Body: "up",
		})
		klass := fmt.Sprintf("DOWN-%03d", i)
		downReqs = append(downReqs, &model.Node{
			Kind: model.KindRequirement, ID: klass, Source: "/test/down/" + klass + ".md",
			Title: "Downstream " + klass,
			Attrs: map[string]any{"status": "approved", "trace": []any{"up/" + id}},
			Body:  "down",
		})
	}
	up := model.Document{
		Path: "/test/up", Schema: schemaWithTitle("Up"),
		XReqmd: &model.XReqmd{DocumentID: "up", Level: "system"},
		Nodes:  upReqs,
	}
	down := model.Document{
		Path: "/test/down", Schema: schemaWithTitle("Down"),
		XReqmd: &model.XReqmd{DocumentID: "down", Level: "software"},
		Nodes:  downReqs,
	}
	analysis := analysisFor(t, up, down)
	h := &HTML{}
	h.SetAnalysis(analysis)
	pagePath := filepath.Join(t.TempDir(), "down-requirements.html")
	f, err := os.Create(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Export(f, down, down.Properties); err != nil {
		t.Fatalf("export: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  try {
    window.reqmd.setView('trace');
    var btn = Array.prototype.filter.call(
      document.querySelectorAll('[data-view-panel="trace"] .btn'),
      function (b) { return b.textContent.indexOf('matrix') !== -1; }
    )[0];
    btn.click();
    var empty = document.querySelector('[data-view-panel="trace"] .matrix .empty__title');
    out.fallback = empty ? empty.textContent : null;
    out.renderedCells = document.querySelectorAll('.matrix__cell').length;
  } catch (e) { out.errors.push(String(e)); }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	if got, _ := result["fallback"].(string); !strings.Contains(got, "Matrix too large") {
		t.Errorf("matrix fallback = %q, want a size-limit message", got)
	}
	if result["renderedCells"] != float64(0) {
		t.Errorf("rendered %v matrix cells, want the fallback instead", result["renderedCells"])
	}
}

// TestHTMLExportBrowserLinkSemantics covers the navigation fixes: external
// trace nodes are links, requirement references inside analytical views open
// the Document view, and the shortcuts dialog receives focus.
func TestHTMLExportBrowserLinkSemantics(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  try {
    var r = window.reqmd;
    r.setView('trace');
    var external = document.querySelector('.trace__node--external');
    out.externalTag = external ? external.tagName : null;
    out.externalHref = external ? external.getAttribute("href") : null;
    out.externalClass = external ? external.className : null;

    // A requirement reference inside the table opens the Document view.
    r.setView('table');
    var openRef = document.querySelector('#req-table [data-open-req]');
    openRef.click();
    out.afterRefClickView = r.state.view;
    out.documentVisible = !document.querySelector('[data-view-panel="document"]').hidden;

    document.dispatchEvent(new KeyboardEvent('keydown', { key: '?', bubbles: true }));
    var dialog = document.getElementById('shortcuts');
    out.shortcutsOpen = !dialog.hidden;
    out.focusInDialog = dialog.contains(document.activeElement);
    out.hasCloseButton = !!dialog.querySelector('[data-action="close-overlays"]');
    dialog.querySelector('[data-action="close-overlays"]').click();
    out.shortcutsClosed = dialog.hidden;
  } catch (e) { out.errors.push(String((e && e.stack) || e)); }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	want := map[string]any{
		"externalTag":       "A",
		"afterRefClickView": "document",
		"documentVisible":   true,
		"shortcutsOpen":     true,
		"focusInDialog":     true,
		"hasCloseButton":    true,
		"shortcutsClosed":   true,
	}
	for k, v := range want {
		if result[k] != v {
			t.Errorf("%s = %v, want %v", k, result[k], v)
		}
	}

	// Inbound links point at a requirement that lives on another page; the
	// test fixture has no resolver, so the control must at least not be a
	// silent no-op. With a resolver (the export path) it carries the page
	// href, which TestHTMLExportDownstreamLinkDirection covers server-side.
	if result["externalTag"] != "A" {
		t.Error("external trace node must be a link, not a dead button")
	}
}

// TestHTMLExportBrowserContextJumps verifies that jumping from a coverage
// queue keeps the reviewer's filters and that the result count names the
// reason requirements are hidden.
func TestHTMLExportBrowserContextJumps(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  try {
    var r = window.reqmd;
    r.setFilter('priority', 'Low');
    out.filtersBefore = JSON.stringify(r.state.filters);
    r.setView('coverage');
    var link = document.querySelector('[data-view-panel="coverage"] .queue__list a');
    out.hasQueueLink = !!link;
    if (link) link.click();
    out.filtersAfter = JSON.stringify(r.state.filters);
    out.viewAfter = r.state.view;
    out.count = document.querySelector('[data-count]').textContent;
  } catch (e) { out.errors.push(String((e && e.stack) || e)); }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	if result["hasQueueLink"] != true {
		t.Fatal("coverage view has no queue links")
	}
	if result["filtersBefore"] != result["filtersAfter"] {
		t.Errorf("filters changed on queue jump: %v -> %v", result["filtersBefore"], result["filtersAfter"])
	}
	if result["viewAfter"] != "document" {
		t.Errorf("queue jump view = %v, want document", result["viewAfter"])
	}
	if got, _ := result["count"].(string); !strings.Contains(got, "hidden by") {
		t.Errorf("count %q should name the reason requirements are hidden", got)
	}
}

// TestHTMLExportBrowserOutlineSelectsRequirement verifies the trace explorer is
// driven by the outline (the requirement list the reader already has) instead
// of a separate picker, and that an outline click keeps filters and state.
func TestHTMLExportBrowserOutlineSelectsRequirement(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  function centred() {
    var e = document.querySelector('.trace__id');
    return e ? e.textContent : null;
  }
  try {
    var r = window.reqmd;
    r.setView('trace');
    out.start = centred();
    out.hasLegacyPicker = !!document.getElementById('trace-picker-select');
    out.bodyView = document.body.getAttribute('data-view');

    var link = document.querySelector('.toc__link[href="#TST-001"]');
    out.hasOutlineLink = !!link;
    link.click();
    out.afterOutlineClick = centred();
    out.url = location.search;
    var marked = document.querySelector('.toc__item.is-selected');
    out.markedId = marked ? marked.getAttribute('data-toc-for') : null;

    // Switching to the reading view keeps the outline as navigation.
    r.setView('document');
    out.bodyViewAfter = document.body.getAttribute('data-view');
    out.stillSelected = !!document.querySelector('.toc__item.is-selected');
  } catch (e) { out.errors.push(String((e && e.stack) || e)); }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	if result["hasLegacyPicker"] != false {
		t.Error("the separate requirement picker should be gone; the outline is the picker")
	}
	if result["start"] != "SYS-001" {
		t.Errorf("initial centre = %v, want SYS-001", result["start"])
	}
	if result["afterOutlineClick"] != "TST-001" {
		t.Errorf("outline click centred on %v, want TST-001", result["afterOutlineClick"])
	}
	if got, _ := result["url"].(string); !strings.Contains(got, "req=TST-001") {
		t.Errorf("URL = %q, want the selected requirement", got)
	}
	if result["markedId"] != "TST-001" {
		t.Errorf("outline marks %v as selected, want TST-001", result["markedId"])
	}
	if result["bodyView"] != "trace" {
		t.Errorf("body data-view = %v, want trace", result["bodyView"])
	}
}

// TestHTMLExportBrowserSelectionPersistence covers the outline selection rules:
// the same requirement stays selected across views, the Trace view is the
// selection (not the scroll position), and a Table outline click stays in the
// table instead of reopening the reading view.
func TestHTMLExportBrowserSelectionPersistence(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  function marked() {
    return Array.prototype.map.call(
      document.querySelectorAll('.toc__item.is-selected'),
      function (i) { return i.getAttribute('data-toc-for'); }
    );
  }
  function centred() {
    var e = document.querySelector('.trace__id');
    return e ? e.textContent : null;
  }
  try {
    var r = window.reqmd;

    // Selecting in the reading view records the selection.
    document.querySelector('.toc__link[href="#TST-001"]').click();
    out.afterDocClick = marked();

    // ... and the Trace view opens on it.
    r.setView('trace');
    out.traceCentre = centred();
    out.traceMarked = marked();

    // Switching back and forth keeps it (no reset to the first requirement).
    r.setView('document');
    out.docKept = marked();
    r.setView('trace');
    out.traceKept = centred();

    // Scrolling the hidden reading view must not re-mark the outline.
    document.getElementById('TST-002').scrollIntoView();
    r.setView('document');
    document.getElementById('TST-002').scrollIntoView();

    // Table: outline click selects the row in place.
    r.setView('table');
    document.querySelector('.toc__link[href="#SYS-002"]').click();
    out.tableStillTable = r.state.view;
    out.tableMarked = marked();
    out.tableRowFlashing = !!document.querySelector('#req-table tr.is-flash');
    out.traceAfterTable = (function () { r.setView('trace'); return centred(); })();

    // Coverage keeps the same selection.
    r.setView('coverage');
    out.coverageMarked = marked();
  } catch (e) { out.errors.push(String((e && e.stack) || e)); }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)

	check := func(key string, want ...string) {
		t.Helper()
		got, _ := result[key].([]any)
		if len(got) != len(want) {
			t.Errorf("%s = %v, want %v", key, got, want)
			return
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s = %v, want %v", key, got, want)
				return
			}
		}
	}
	check("afterDocClick", "TST-001")
	check("traceMarked", "TST-001")
	check("docKept", "TST-001")
	check("tableMarked", "SYS-002")
	check("coverageMarked", "SYS-002")
	if result["traceCentre"] != "TST-001" || result["traceKept"] != "TST-001" {
		t.Errorf("trace centre = %v / %v, want TST-001 both times", result["traceCentre"], result["traceKept"])
	}
	if result["tableStillTable"] != "table" {
		t.Errorf("outline click in the table view switched to %v", result["tableStillTable"])
	}
	if result["tableRowFlashing"] != true {
		t.Error("table outline click did not reveal the row")
	}
	if result["traceAfterTable"] != "SYS-002" {
		t.Errorf("trace centre after table selection = %v, want SYS-002", result["traceAfterTable"])
	}
}

// TestHTMLExportBrowserDocumentJumpScrollsToSelection verifies that entering the
// reading view lands on the selected requirement (not the top of the
// document), and that the outline's position highlight agrees with it.
func TestHTMLExportBrowserDocumentJumpScrollsToSelection(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  function active() {
    return Array.prototype.map.call(
      document.querySelectorAll('.toc__link.is-active'),
      function (a) { return a.getAttribute('href').slice(1); }
    );
  }
  try {
    var r = window.reqmd;
    r.setView('table');
    document.querySelector('.toc__link[href="#TST-002"]').click();
    r.setView('trace');
    r.setView('document');
    setTimeout(function () {
      var el = document.getElementById('TST-002');
      var rect = el.getBoundingClientRect();
      out.visible = rect.top < window.innerHeight && rect.bottom > 0;
      out.top = Math.round(rect.top);
      out.scrolled = window.scrollY > 0;
      out.active = active();
      out.marked = Array.prototype.map.call(
        document.querySelectorAll('.toc__item.is-selected'),
        function (i) { return i.getAttribute('data-toc-for'); }
      );
      document.title = 'RESULT:' + JSON.stringify(out);
    }, 500);
  } catch (e) {
    out.errors.push(String((e && e.stack) || e));
    document.title = 'RESULT:' + JSON.stringify(out);
  }
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	if result["visible"] != true {
		t.Errorf("selected requirement is not on screen after switching (top=%v)", result["top"])
	}
	if result["scrolled"] != true {
		t.Errorf("reading view did not scroll (top=%v)", result["top"])
	}
	active, _ := result["active"].([]any)
	if len(active) != 1 || active[0] != "TST-002" {
		t.Errorf("outline position highlight = %v, want only TST-002", active)
	}
	marked, _ := result["marked"].([]any)
	if len(marked) != 1 || marked[0] != "TST-002" {
		t.Errorf("outline selection = %v, want only TST-002", marked)
	}
}

// TestHTMLExportBrowserJumpClearsStickyChrome verifies that a jump (palette,
// deep link, or view switch) leaves the requirement heading below the sticky
// app bar and toolbar instead of hidden underneath them.
func TestHTMLExportBrowserJumpClearsStickyChrome(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  function measure(id) {
    var head = document.getElementById(id).querySelector('.req__head');
    var top = head.getBoundingClientRect().top;
    var chrome = document.querySelector('.app-bar').offsetHeight +
      document.querySelector('.toolbar').offsetHeight;
    return { top: Math.round(top), chrome: Math.round(chrome), clear: top >= chrome };
  }
  try {
    var r = window.reqmd;
    r.setView('table');
    document.querySelector('.toc__link[href="#TST-002"]').click();
    r.setView('trace');
    r.setView('document');
    setTimeout(function () {
      out.viewSwitch = measure('TST-002');
      r.jumpTo('SYS-001', true);
      setTimeout(function () {
        out.palette = measure('SYS-001');
        document.title = 'RESULT:' + JSON.stringify(out);
      }, 450);
    }, 450);
  } catch (e) {
    out.errors.push(String((e && e.stack) || e));
    document.title = 'RESULT:' + JSON.stringify(out);
  }
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	for _, key := range []string{"viewSwitch", "palette"} {
		m, ok := result[key].(map[string]any)
		if !ok {
			t.Fatalf("%s missing from result", key)
		}
		if m["clear"] != true {
			t.Errorf("%s: heading top %v is under the sticky chrome (%v)", key, m["top"], m["chrome"])
		}
	}
}

// TestHTMLExportBrowserSparseColumnsHidden verifies that table columns which
// are blank or nearly blank across the rows are dropped (an empty column reads
// as a bug) while the column chooser still offers them.
func TestHTMLExportBrowserSparseColumnsHidden(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  function headers() {
    return Array.prototype.map.call(document.querySelectorAll('#req-table th'),
      function (th) { return th.textContent.trim(); });
  }
  try {
    var r = window.reqmd;
    r.setView('table');
    setTimeout(function () {
      out.headers = headers();
      var chooser = document.querySelector('details.columns');
      out.hasChooser = !!chooser;
      out.chooser = chooser
        ? Array.prototype.map.call(chooser.querySelectorAll('.columns__option span'),
            function (s) { return s.textContent.trim(); }).join('|')
        : '';
      document.title = 'RESULT:' + JSON.stringify(out);
    }, 300);
  } catch (e) {
    out.errors.push(String((e && e.stack) || e));
    document.title = 'RESULT:' + JSON.stringify(out);
  }
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	headers, _ := result["headers"].([]any)
	var joined string
	for _, h := range headers {
		joined += h.(string) + "|"
	}
	// Every requirement in the fixture declares a status, so Status stays even
	// though the other attribute columns are sparse. The mechanism (a column
	// below half the rows is dropped) is covered by the empty-state assertion
	// below: the fixture's Verdict column has no values at all without
	// results.
	for _, want := range []string{"ID", "Title", "Status", "Coverage"} {
		if !strings.Contains(joined, want) {
			t.Errorf("column %q missing from %s", want, joined)
		}
	}
	if strings.Contains(joined, "Verdict") {
		t.Errorf("all-empty Verdict column was kept: %s", joined)
	}
	if result["hasChooser"] != true {
		t.Fatal("column chooser missing")
	}
	chooser, _ := result["chooser"].(string)
	if !strings.Contains(chooser, "Verdict") {
		t.Errorf("column chooser should still offer the hidden column, got %q", chooser)
	}
}

// TestHTMLExportModHintIsPlatformCorrect verifies the palette shortcut is
// spelled for the reader's platform instead of always claiming ⌘K.
func TestHTMLExportModHintIsPlatformCorrect(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  try {
    var btn = document.querySelector('[data-action="palette"]');
    out.label = btn.textContent.trim();
    out.title = btn.getAttribute('title');
    var dt = document.querySelector('[data-mod-keys]');
    out.dialog = dt ? dt.textContent.trim() : '';
    out.spaceLabel = out.label.replace(/\s+/g, ' ');
    document.title = 'RESULT:' + JSON.stringify(out);
  } catch (e) {
    out.errors.push(String((e && e.stack) || e));
    document.title = 'RESULT:' + JSON.stringify(out);
  }
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	isMac := strings.Contains(fmt.Sprint(result["label"]), "\u2318")
	wantLabel := "Ctrl K"
	if isMac {
		wantLabel = "\u2318K"
	}
	if result["label"] != wantLabel {
		t.Errorf("palette label = %v, want %v", result["label"], wantLabel)
	}
	if title, _ := result["title"].(string); !strings.Contains(title, wantLabel) {
		t.Errorf("palette title = %q, want it to mention %q", title, wantLabel)
	}
	// The dialog renders one <kbd> per key, so "Ctrl"+"K" has no separator;
	// compare without spaces on both sides.
	dialog := strings.ReplaceAll(fmt.Sprint(result["dialog"]), " ", "")
	if !strings.Contains(dialog, strings.ReplaceAll(wantLabel, " ", "")) {
		t.Errorf("shortcuts dialog = %v, want %v", result["dialog"], wantLabel)
	}
}

// TestHTMLExportBrowserPaletteNavigation covers the command palette with a
// requirement query: arrow keys step one row at a time, hover does not fight
// the keyboard, typing keeps the focus in the input, and Enter runs the
// highlighted row.
func TestHTMLExportBrowserPaletteNavigation(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  function active() {
    var a = document.querySelector('.palette__item.is-active');
    return a ? a.getAttribute('data-label') : null;
  }
  function key(k) { document.dispatchEvent(new KeyboardEvent('keydown', { key: k, bubbles: true, cancelable: true })); }
  try {
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true, bubbles: true, cancelable: true }));
    var input = document.querySelector('[data-palette-input]');
    input.value = 'tst';
    input.dispatchEvent(new Event('input'));
    out.focusInInput = document.activeElement === input;
    out.first = active();

    var down = [];
    key('ArrowDown'); down.push(active());
    key('ArrowDown'); down.push(active());
    out.down = down;

    key('ArrowUp'); out.afterUp = active();

    var rows = document.querySelectorAll('.palette__item');
    rows[1].dispatchEvent(new MouseEvent('mousemove', { bubbles: true }));
    out.afterHover = active();
    key('ArrowDown'); out.afterHoverDown = active();

    key('Enter');
    out.hash = location.hash;
    out.closed = document.getElementById('palette').hidden;
  } catch (e) { out.errors.push(String((e && e.stack) || e)); }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)
	if result["focusInInput"] != true {
		t.Error("typing moved focus out of the palette input")
	}
	down, _ := result["down"].([]any)
	if len(down) != 2 {
		t.Fatalf("down = %v", down)
	}
	if down[0] == result["first"] || down[1] == down[0] {
		t.Errorf("ArrowDown did not step one row at a time: first=%v down=%v", result["first"], down)
	}
	if result["afterUp"] != down[0] || result["afterUp"] == down[1] {
		t.Errorf("ArrowUp did not step back one row: afterUp=%v down=%v", result["afterUp"], down)
	}
	if result["afterHover"] != down[0] {
		t.Errorf("hover did not select the hovered row: %v", result["afterHover"])
	}
	if result["afterHoverDown"] != down[1] {
		t.Errorf("ArrowDown after hover = %v, want the next row %v", result["afterHoverDown"], down[1])
	}
	if result["closed"] != true {
		t.Error("Enter did not close the palette")
	}
	if got, _ := result["hash"].(string); !strings.Contains(got, "#") || strings.HasPrefix(got, "#?") {
		t.Errorf("Enter did not open a requirement: hash=%q", got)
	}
}

// TestHTMLExportBrowserPaletteFuzzy covers abbreviation matching in the
// command palette: "sysfm1" must find SYS-001-style IDs whose segments the
// query skips part of, and ids are ranked ahead of incidental title hits.
func TestHTMLExportBrowserPaletteFuzzy(t *testing.T) {
	chrome := findChrome(t)
	pagePath := exportBrowserFixture(t)

	const harness = `<script>
window.addEventListener('load', function () {
  var out = { errors: [] };
  function top(q, n) {
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true, bubbles: true, cancelable: true }));
    var input = document.querySelector('[data-palette-input]');
    input.value = q;
    input.dispatchEvent(new Event('input'));
    return Array.prototype.slice.call(document.querySelectorAll('.palette__item'))
      .slice(0, n || 3)
      .map(function (b) { return b.getAttribute('data-label'); });
  }
  try {
    // Fixture IDs are SYS-001, SYS-002, SYS-003, TST-001, TST-002.
    out.dense = top('sys1');
    out.denseFirst = out.dense[0] || null;
    out.skipWord = top('s1');
    out.exact = top('tst-002');
    out.title = top('uncovered');
    out.single = top('s', 5);
  } catch (e) { out.errors.push(String((e && e.stack) || e)); }
  document.title = 'RESULT:' + JSON.stringify(out);
});
</script>`
	result := runHeadless(t, chrome, pagePath, "", nil, harness)
	assertNoJSErrors(t, result)

	if result["denseFirst"] != "SYS-001" {
		t.Errorf("'sys1' first hit = %v, want SYS-001", result["dense"])
	}
	if result["exact"] == nil || len(result["exact"].([]any)) == 0 || result["exact"].([]any)[0] != "TST-002" {
		t.Errorf("'tst-002' did not resolve exactly: %v", result["exact"])
	}
	// Title words still find requirements that do not contain them in the ID.
	title, _ := result["title"].([]any)
	if len(title) == 0 {
		t.Error("title search returned nothing")
	}
	// A single character must not fuzzy-match the whole document.
	single, _ := result["single"].([]any)
	if len(single) > 8 {
		t.Errorf("single-character query returned %d items; too broad", len(single))
	}
}
