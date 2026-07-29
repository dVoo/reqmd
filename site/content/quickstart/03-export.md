---
title: "Step 3 — Export"
description: "Export the spec tree to CSV, HTML, and graph formats."
weight: 7
---

Export the spec tree from [Step 2](/quickstart/02-trace-your-spec/) to CSV, HTML, and graph formats.

## CSV export

```sh
reqmd export csv 02-trace-your-spec/ -o csv-out/
```

Produces one CSV per document directory (`stakeholder-requirements.csv`, `system-requirements.csv`, etc.) with columns: ID, Title, all schema attributes, Body, Rationale.

## HTML export

```sh
reqmd export html 02-trace-your-spec/ -o html-out/
```

Produces standalone HTML files with:

- Card-based layout with goldmark-rendered body text
- Upstream/downstream trace links between files
- Document chain tab strip (V-model navigation)
- Search and filter (Alpine.js)
- Light/dark theme toggle
Open `html-out/tests-requirements.html` in a browser to see trace links.

<div class="callout">
<div class="callout-icon">
<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
</div>
<div class="callout-body">
<p><strong>See it live.</strong> Don't want to run the export yourself? Here's a real HTML export from the reqmd project's own spec, showing card layout, trace links, the document chain, search, and theme toggle:</p>
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

All export formats accept `--results` (see Steps 5 and 6):

```sh
reqmd export html 02-trace-your-spec/ \
  --results 05-verification-results-ctrf/ \
  --results 06-review-documentation/ \
  -o html-out/
```

HTML gets color-coded verdict badges on measure cards; CSV gets `Verdict` and `Verdict Source` columns; Graph gets `RESULT:` nodes with `outcome` properties.

## What's next

You can export the standard spec tree. Now define your own schema — go to [Step 4](/quickstart/04-custom-templates/).
