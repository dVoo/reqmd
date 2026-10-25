---
title: "Step 3 — Export"
description: "Export the boot-sequence spec tree from Step 2 to CSV, HTML, and graph formats."
weight: 3
---

Export the spec tree from [Step 2](/quickstart/02-trace-your-spec/) to CSV, HTML, and graph formats.

## CSV export

```sh
reqmd export csv 02-trace-your-spec/ -o csv-out/
```

Produces one CSV per document directory (`stakeholder-requirements.csv`, `system-requirements.csv`, etc.) with a leading `Type` column followed by ID, Title, all schema attributes, Body, Rationale. Rows are emitted in document order for every node — requirements **and** containers/info items (headings without `attr` blocks) — so the CSV mirrors the source document, not just its requirement subset.

## HTML export

```sh
reqmd export html 02-trace-your-spec/ -o html-out/
```

Produces a project home (`index.html`) plus one standalone HTML file per
document. Each page has:

- A server-rendered **Document** view with goldmark-rendered body text — works
  offline and without JavaScript
- **Table** (sortable, column chooser, expandable rows), **Trace** (selected
  requirement explorer plus a size-bounded matrix), and **Coverage** (actionable
  exception lists) views rendered on demand from one shared model
- Coverage, verification (with `no verdict` in the denominator), and link health
  that agree with `reqmd check`
- A ⌘K command palette to jump to any requirement or switch views
- Keyboard shortcuts (`/`, `g d`/`g t`/`g r`/`g c`, `t`, `f`, `?`) and a density toggle
- URL-synced state (`?q=&f=attr:value&view=&density=`) so a filtered view is shareable
- A compact sticky app bar with the document breadcrumb, outline toggle, and light/dark theme toggle
- No external scripts: the stylesheet and controller are embedded
- Collapsible container sections and plain info blocks for headings without `attr` blocks (folders, exactly as authored)
- Sidebar TOC with the same folder nesting
- Upstream/downstream trace links between files
- Document chain section (V-model navigation)
- A sticky search-and-filter toolbar: attribute menus are built from every attribute present on a requirement, including attributes declared in no `schema.yaml` (those also appear in the card grid marked `· undeclared`)

Open `html-out/index.html` in a browser for the project home, then any document page to see trace links.

<div class="callout">
<div class="callout-icon">
<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
</div>
<div class="callout-body">
<p><strong>See it live.</strong> Don't want to run the export yourself? Here's a real HTML export from the reqmd project's own spec, showing the card layout, coverage and verification state, the trace explorer, search, and theme toggle:</p>
<p><a href="/export-sample/02-system-requirements.html" target="_blank" rel="noopener"><strong>Open the system requirements export →</strong></a></p>
<p>Other documents in the same export: <a href="/export-sample/01-stakeholder-requirements.html" target="_blank" rel="noopener">Stakeholder needs</a>, <a href="/export-sample/03-software-requirements.html" target="_blank" rel="noopener">Software requirements</a>, <a href="/export-sample/04-tests-requirements.html" target="_blank" rel="noopener">Test specifications</a>. Try the trace links — they cross between documents.</p>
</div>
</div>

## Graph export (requires LadybugDB build)

```sh
go build -tags ladybug -o reqmd ./cmd/reqmd
reqmd export graph 02-trace-your-spec/ -o graph-out/
```

Creates a LadybugDB graph database with one node per requirement and `TracesTo` edges for every trace link. Query with:

```sh
lbug graph-out/reqmd-graph.lbug
```

## Export with verification results

All export formats accept `--results` (see Steps 11 and 12):

```sh
reqmd export html 02-trace-your-spec/ \
  --results 11-verification-results-ctrf/ \
  --results 12-review-documentation/ \
  -o html-out/
```

HTML gets color-coded verdict badges on measure cards; CSV gets `Verdict` and `Verdict Source` columns; Graph gets `RESULT:` nodes with `outcome` properties.

## Where to look things up

- Export flags and column/section details: [`reqmd export csv`](/cheat-sheet/#reqmd-export-csv--export-to-csv), [`reqmd export html`](/cheat-sheet/#reqmd-export-html--export-to-html), [`reqmd export graph`](/cheat-sheet/#reqmd-export-graph--export-to-ladybugdb-graph)

## What's next

You can export the standard spec tree. Now preview it live while you edit — go to [Step 4](/quickstart/04-live-preview/).
