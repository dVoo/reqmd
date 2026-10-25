# Step 3 — Export

Export the spec tree from [Step 2](../02-trace-your-spec/) to CSV, HTML, and
graph formats.

## CSV export

```sh
reqmd export csv 02-trace-your-spec/ -o csv-out/
```

Produces one CSV per document directory (`stakeholder-requirements.csv`,
`system-requirements.csv`, etc.) with a leading `Type` column followed by
ID, Title, all schema attributes, Body, Rationale. Rows are emitted in
document order for every node — requirements **and** containers/info
items (headings without `attr` blocks) — so the CSV mirrors the source
document, not just its requirement subset.

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
- A ⌘K command palette, keyboard shortcuts (`/`, `g d`/`g t`/`g r`/`g c`, `t`,
  `f`, `?`), and a density toggle
- URL-synced state (`?q=&f=attr:value&view=&density=`) so a filtered view is shareable
- A compact sticky app bar with the document breadcrumb, outline toggle, and light/dark theme toggle
- No external scripts: the stylesheet and controller are embedded
- Collapsible container sections and plain info blocks for headings
  without `attr` blocks (folders, exactly as authored)
- Sidebar TOC with the same folder nesting
- Upstream/downstream trace links between files
- Document chain section (V-model navigation)
- A sticky search-and-filter toolbar: attribute menus are built from every
  attribute present on a requirement, including attributes declared in no
  `schema.yaml` (those also appear in the card grid marked `· undeclared`)

Open `html-out/index.html` in a browser for the project home, then any document page to see trace links.

## Graph export (requires LadybugDB build)

```sh
go build -tags ladybug -o reqmd ./cmd/reqmd
reqmd export graph 02-trace-your-spec/ -o graph-out/
```

Creates a LadybugDB graph database with one node per requirement and
`TracesTo` edges for every trace link. Query with:

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

HTML gets color-coded verdict badges on measure cards; CSV gets `Verdict` and
`Verdict Source` columns; Graph gets `RESULT:` nodes with `outcome` properties.

## What's next

You can export the standard spec tree. Now preview it live while you edit —
go to [Step 4](../04-live-preview/).