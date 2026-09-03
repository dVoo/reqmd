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

Produces standalone HTML files with:
- Requirement cards with goldmark-rendered body text
- Collapsible container sections and plain info blocks for headings
  without `attr` blocks (folders, exactly as authored)
- Sidebar TOC with the same folder nesting
- Upstream/downstream trace links between files
- Document chain tab strip (V-model navigation)
- Search and filter (Alpine.js, requirement-scoped)
- Light/dark theme toggle

Open `html-out/tests-requirements.html` in a browser to see trace links.

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