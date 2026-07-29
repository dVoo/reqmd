# Step 3 — Export

Export the spec tree from [Step 2](../02-trace-your-spec/) to CSV, HTML, and
graph formats.

## CSV export

```sh
reqmd export csv 02-trace-your-spec/ -o csv-out/
```

Produces one CSV per document directory (`stakeholder-requirements.csv`,
`system-requirements.csv`, etc.) with columns: ID, Title, all schema
attributes, Body, Rationale.

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

All export formats accept `--results` (see Steps 5 and 6):

```sh
reqmd export html 02-trace-your-spec/ \
  --results 05-verification-results-ctrf/ \
  --results 06-review-documentation/ \
  -o html-out/
```

HTML gets color-coded verdict badges on measure cards; CSV gets `Verdict` and
`Verdict Source` columns; Graph gets `RESULT:` nodes with `outcome` properties.

## What's next

You can export the standard spec tree. Now define your own schema — go to
[Step 4](../04-custom-templates/).